package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	SwarmListenAddr = "127.0.0.1:9999"
	BackendRealHost = "127.0.0.1:8080"
	TenantSeed      = "CROM-SEC-TENANT-ALPHA-2026"
	CromMagic       = "CROM"
	JitterMagic     = "JITT"
)

// applyLLMSemanticExpansion reverte a lógica feita no Alpha.
func applyLLMSemanticExpansion(data []byte) []byte {
	str := string(data)
	str = strings.ReplaceAll(str, "⌬HTTP1", "HTTP/1.1")
	str = strings.ReplaceAll(str, "⌬CTJSON", "Accept: application/json")
	str = strings.ReplaceAll(str, "⌬CONKA", "Connection: keep-alive")
	str = strings.ReplaceAll(str, "⌬UA", "User-Agent")
	return []byte(str)
}

// cromDecryptPacket valida e descriptografa um pacote CROM via AES-256-GCM.
// Retorna (plaintext, isJitter). Se nil, pacote inválido (Hacker) → Silent Drop.
// O GCM Seal garante INTEGRIDADE + AUTENTICIDADE: qualquer bit adulterado = rejeição total.
func cromDecryptPacket(packet []byte) ([]byte, bool) {
	if len(packet) < 4 {
		return nil, false
	}

	magic := string(packet[:4])
	if magic != CromMagic && magic != JitterMagic {
		return nil, false
	}

	isJitter := magic == JitterMagic
	ciphertext := packet[4:]

	// Derivar chave AES-256 via HMAC-SHA256 da TenantSeed
	mac := hmac.New(sha256.New, []byte(TenantSeed))
	mac.Write([]byte("CROM_AES_GCM_KEY_V4"))
	key := mac.Sum(nil) // 32 bytes = AES-256

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, false
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, false
	}

	nonceSize := aesgcm.NonceSize() // 12 bytes padrão
	if len(ciphertext) < nonceSize {
		return nil, false
	}

	nonce := ciphertext[:nonceSize]
	sealed := ciphertext[nonceSize:]

	// Open valida o GCM Tag (16 bytes) — qualquer adulteração = erro = DROP
	decrypted, err := aesgcm.Open(nil, nonce, sealed, []byte(magic))
	if err != nil {
		// GCM Authentication FALHOU: pacote forjado, corrompido ou com Seed errada
		return nil, false
	}

	if isJitter {
		return decrypted, true
	}

	return applyLLMSemanticExpansion(decrypted), false
}

// cromEncrypt aplica AES-256-GCM autenticado para a resposta de volta ao Alpha.
// Compressão Semântica → AES-GCM Seal (Nonce aleatório + Tag de integridade 16B)
func cromEncrypt(data []byte) []byte {
	// Compressão semântica LLM
	str := string(data)
	str = strings.ReplaceAll(str, "HTTP/1.1", "⌬HTTP1")
	str = strings.ReplaceAll(str, "Accept: application/json", "⌬CTJSON")
	str = strings.ReplaceAll(str, "Connection: keep-alive", "⌬CONKA")
	str = strings.ReplaceAll(str, "User-Agent", "⌬UA")
	processedData := []byte(str)

	// Derivar chave AES-256 via HMAC-SHA256 da TenantSeed
	mac := hmac.New(sha256.New, []byte(TenantSeed))
	mac.Write([]byte("CROM_AES_GCM_KEY_V4"))
	key := mac.Sum(nil) // 32 bytes = AES-256

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("[OMEGA-CRYPTO] Falha ao criar cifra AES: %v", err)
		return nil
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("[OMEGA-CRYPTO] Falha ao criar GCM: %v", err)
		return nil
	}

	nonce := make([]byte, aesgcm.NonceSize()) // 12 bytes
	rand.Read(nonce)

	// Seal: cifra + autentica. AAD (Additional Authenticated Data) = magic header
	sealed := aesgcm.Seal(nil, nonce, processedData, []byte(CromMagic))

	// Pacote final: [MAGIC 4B][NONCE 12B][CIPHERTEXT+TAG]
	packet := make([]byte, 0, 4+len(nonce)+len(sealed))
	packet = append(packet, []byte(CromMagic)...)
	packet = append(packet, nonce...)
	packet = append(packet, sealed...)
	return packet
}

func handleAlienConnection(alienConn net.Conn) {
	defer alienConn.Close()

	// Ler o primeiro chunk para validar se é um pacote CROM válido
	// Prevenção inteligente contra Slowloris: Timeout de 3s na borda L4
	alienConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	initialBuf := make([]byte, 32768)
	n, err := alienConn.Read(initialBuf)
	
	// Limpar o deadline após o handshake para não afetar streams contínuos legítimos (como Chat/WebSocket)
	alienConn.SetReadDeadline(time.Time{})
	
	if err != nil {
		return
	}

	// ===== SILENT DROP =====
	// Se o pacote não tem a assinatura CROM, fechar a conexão sem responder NADA.
	// Isso impede banner grabbing e information leakage.
	plaintext, isJitt := cromDecryptPacket(initialBuf[:n])
	if plaintext == nil {
		log.Printf("[OMEGA-SILENT-DROP] Pacote inválido de %s (%d bytes). Dropped.", alienConn.RemoteAddr(), n)
		return
	}
	
	if isJitt {
		// Cover-Traffic cego P2P. Absorve a conexao sem mandar pro backend e encerra.
		log.Printf("[OMEGA] JITTER Cover-Traffic Inicial Recebido (%d bytes). Absorvido.", n)
		return
	}

	log.Printf("[OMEGA] Pacote CROM válido de %s (%d bytes)", alienConn.RemoteAddr(), len(plaintext))

	// Conectar ao backend real
	backendConn, err := net.Dial("tcp", BackendRealHost)
	if err != nil {
		log.Printf("[OMEGA] Backend indisponível: %v", err)
		return
	}
	defer backendConn.Close()

	backendConn.Write(plaintext)

	var wg sync.WaitGroup

	// Goroutine 1: Alien -> Decrypt -> Backend (upstream contínuo)
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 32768)
		for {
			rn, err := alienConn.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("[OMEGA] Alien read err: %v", err)
				}
				backendConn.Close()
				return
			}
			pt, isJt := cromDecryptPacket(buf[:rn])
			if pt == nil {
				log.Printf("[OMEGA-SILENT-DROP] Pacote corrompido mid-stream. Dropped.")
				continue
			}
			if isJt {
				// Cover traffic, absorvido na rede mas nunca incomoda a CPU local
				continue
			}
			_, werr := backendConn.Write(pt)
			if werr != nil {
				return
			}
		}
	}()

	// Goroutine 2: Backend -> Encrypt -> Alien (downstream contínuo)
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 32768)
		for {
			rn, err := backendConn.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("[OMEGA] Backend read err: %v", err)
				}
				alienConn.Close()
				return
			}
			encrypted := cromEncrypt(buf[:rn])
			_, werr := alienConn.Write(encrypted)
			if werr != nil {
				return
			}
		}
	}()

	wg.Wait()
}

func main() {
	fmt.Println("=================================================================")
	fmt.Println(" [ CROM ALIEN PROXY OUT-FLIGHT (v2 Full-Duplex + Silent Drop) ]")
	fmt.Println(" Resolvendo poeira P2P com validação criptográfica HMAC")
	fmt.Println(" Escutando nuvem: localhost:9999 | Backend: localhost:8080")
	fmt.Println("=================================================================")

	l, err := net.Listen("tcp", SwarmListenAddr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go handleAlienConnection(conn)
	}
}
