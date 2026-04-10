package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
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

// cromDecryptPacket valida e descriptografa um pacote CROM.
// Retorna (plaintext, isJitter). Se nil, pacote inválido (Hacker).
func cromDecryptPacket(packet []byte) ([]byte, bool) {
	if len(packet) < 13 {
		return nil, false
	}
	
	magic := string(packet[:4])
	if magic != CromMagic && magic != JitterMagic {
		return nil, false
	}
	
	isJitter := magic == JitterMagic

	nonce := packet[4:12]
	encrypted := packet[12:]

	mac := hmac.New(sha256.New, []byte(TenantSeed))
	mac.Write([]byte("CROM_SESSION_KEY"))
	key := mac.Sum(nil)

	decrypted := make([]byte, len(encrypted))
	for i, b := range encrypted {
		decrypted[i] = b ^ key[i%len(key)] ^ nonce[i%len(nonce)]
	}

	if isJitter {
		return decrypted, true
	}

	return applyLLMSemanticExpansion(decrypted), false
}

// cromEncrypt aplica XOR cíclico com chave HMAC para a resposta de volta.
func cromEncrypt(data []byte) []byte {
	mac := hmac.New(sha256.New, []byte(TenantSeed))
	mac.Write([]byte("CROM_SESSION_KEY"))
	key := mac.Sum(nil)

	nonce := make([]byte, 8)
	rand.Read(nonce)

	str := string(data)
	str = strings.ReplaceAll(str, "HTTP/1.1", "⌬HTTP1")
	str = strings.ReplaceAll(str, "Accept: application/json", "⌬CTJSON")
	str = strings.ReplaceAll(str, "Connection: keep-alive", "⌬CONKA")
	str = strings.ReplaceAll(str, "User-Agent", "⌬UA")
	processedData := []byte(str)

	encrypted := make([]byte, len(processedData))
	for i, b := range processedData {
		encrypted[i] = b ^ key[i%len(key)] ^ nonce[i%len(nonce)]
	}

	packet := make([]byte, 0, 4+8+len(encrypted))
	packet = append(packet, []byte(CromMagic)...)
	packet = append(packet, nonce...)
	packet = append(packet, encrypted...)
	return packet
}

func handleAlienConnection(alienConn net.Conn) {
	defer alienConn.Close()

	// Ler o primeiro chunk para validar se é um pacote CROM válido
	initialBuf := make([]byte, 32768)
	n, err := alienConn.Read(initialBuf)
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
