package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"net"
	"os"
	"sync"
	"time"
)

const (
	SwarmListenAddr = "127.0.0.1:9999"
	BackendRealHost = "127.0.0.1:8080"
	CromMagic       = "CROM"
	JitterMagic     = "JITT"

	// [RT-13 FIX] Limite máximo de conexões simultâneas para evitar FD exhaustion
	MaxConcurrentConns = 2048

	// [RT-11 FIX] Máximo de pacotes inválidos mid-stream antes de encerrar conexão
	MaxInvalidMidStream = 5

	// [GEN-6 RT-06 FIX] Janela de drift máximo do timestamp autenticado (segundos).
	// Reduzido de 30s para 5s para minimizar janela de replay.
	MaxTimestampDriftSecs = 5

	// [GEN-6 RT-08 FIX] Máximo de conexões simultâneas por IP de origem.
	MaxConnsPerIP = 10
)

// [GEN-7 RT-02 FIX] Seed não é mais mantida globalmente em string imutável.
// Apenas a AEAD global será mantida em memória local, dificultando ptrace extraction.

// [GEN-7 RT-08 FIX] Bloqueio robusto anti-Race Condition (Mutex L4).
var ipMutex sync.Mutex
var perIPConns = make(map[string]int)

// secureReadSeed lê a chave de forma segura no startup e joga tudo para bytes apagáveis
func secureReadSeedAndInitAEAD() cipher.AEAD {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		log.Fatal("[OMEGA-SECURITY] ⚠️ MODO INSEGURO DETECTADO. Forneça a Seed de Tenant APENAS via STDIN pipe (echo SEED | proxy_out).")
	}

	rawBytes := make([]byte, 1024)
	n, err := os.Stdin.Read(rawBytes)
	if err != nil && err != io.EOF {
		log.Fatal("[OMEGA-FATAL] Falha lendo pipe STDIN.")
	}

	// Localizar trim
	seedBytes := rawBytes[:n]
	var trimmed []byte
	for _, b := range seedBytes {
		if b >= 32 && b <= 126 {
			trimmed = append(trimmed, b)
		}
	}

	if len(trimmed) == 0 {
		log.Fatal("[OMEGA-FATAL] Seed STDIN vazia.")
	}

	// [GEN-7] Derivar AES key IMEDIATAMENTE.
	mac := hmac.New(sha256.New, trimmed)
	mac.Write([]byte("CROM_AES_GCM_KEY_V4"))
	key := mac.Sum(nil)

	// [GEN-7] MANUALLY ZERO THE SEED BUFFER AND RAW BYTES
	for i := range trimmed { trimmed[i] = 0 }
	for i := range rawBytes { rawBytes[i] = 0 }

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalf("[OMEGA-FATAL] Falha no AES: %v", err)
	}

	// Zero key stack
	for i := range key { key[i] = 0 }

	aead, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatalf("[OMEGA-FATAL] Falha no GCM: %v", err)
	}
	log.Println("[OMEGA-SECURITY] Módulo KMS Criptográfico Armado. Limpeza de memória executada com sucesso (Runtime Zeroize).")
	return aead
}

var globalAEAD cipher.AEAD
var onceAEAD sync.Once

// [RT-17 FIX] Nonce Anti-Replay Map
var globalNonceCache sync.Map

func getAEAD() cipher.AEAD {
	onceAEAD.Do(func() {
		globalAEAD = secureReadSeedAndInitAEAD()

		// [ENGULF-FIX VULN-4] Janitor com TTL granular — limpa entradas individuais após 60s
		go func() {
			for {
				time.Sleep(10 * time.Second)
				now := time.Now().Unix()
				globalNonceCache.Range(func(key, value interface{}) bool {
					if ts, ok := value.(int64); ok {
						if now-ts > 60 {
							globalNonceCache.Delete(key)
						}
					}
					return true
				})
			}
		}()
	})
	return globalAEAD
}

// [RT-14 FIX] TCP Length-Prefix Framing + [RT-16 FIX] Block Limit Bounds
func readFramedPacket(conn net.Conn) ([]byte, error) {
	lenBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, lenBuf); err != nil {
		return nil, err
	}
	packetLen := binary.BigEndian.Uint16(lenBuf)
	if packetLen > 32768 {
		return nil, fmt.Errorf("length buffer oversize attack")
	}
	packetBuf := make([]byte, packetLen)
	if _, err := io.ReadFull(conn, packetBuf); err != nil {
		return nil, err
	}
	return packetBuf, nil
}

// [GEN-6 RT-05 FIX] Atomic write: combina length + payload em um único buffer
// para eliminar risco de interleaving em cenários de concurrent write.
func writeFramedPacket(conn net.Conn, packet []byte) error {
	frame := make([]byte, 2+len(packet))
	binary.BigEndian.PutUint16(frame[:2], uint16(len(packet)))
	copy(frame[2:], packet)
	_, err := conn.Write(frame)
	return err
}

// [RT-10 FIX] hashAddr gera um hash truncado do endereço para logging seguro.
// Impede que IPs de atacantes sejam logados em plaintext (OPSEC).
func hashAddr(addr net.Addr) string {
	h := sha256.Sum256([]byte(addr.String()))
	return fmt.Sprintf("%x", h[:6])
}

// applyLLMSemanticExpansion reverte a lógica feita no Alpha.
// [ENGULF-FIX VULN-2] DESATIVADA: A substituição cega de strings em payload TCP
// alterava o tamanho do body sem atualizar Content-Length, permitindo HTTP Request Smuggling.
// Proxies L4 NUNCA devem modificar payload L7. Retorna dados intactos.
func applyLLMSemanticExpansion(data []byte) []byte {
	return data
}

// cromDecryptPacket valida e descriptografa um pacote CROM via AES-256-GCM.
// [ENGULF Gen-5] Formato: [MAGIC 4B][TIMESTAMP 8B][NONCE 12B][CIPHERTEXT+GCM_TAG]
// Retorna (plaintext, isJitter). Se nil, pacote inválido (Hacker) → Silent Drop.
// O GCM Seal garante INTEGRIDADE + AUTENTICIDADE: qualquer bit adulterado = rejeição total.
func cromDecryptPacket(packet []byte) ([]byte, bool) {
	// Mínimo: 4 (magic) + 8 (timestamp) + 12 (nonce) + 16 (gcm tag) = 40 bytes
	if len(packet) < 40 {
		return nil, false
	}

	magic := string(packet[:4])
	if magic != CromMagic && magic != JitterMagic {
		return nil, false
	}

	isJitter := magic == JitterMagic

	// [ENGULF-FIX VULN-4] Extrair e validar Timestamp autenticado contra Replay Window
	// AAD Gen-5: [MAGIC 4B][TIMESTAMP 8B] — ambos autenticados pelo GCM Tag
	aad := packet[:12]
	ciphertext := packet[12:]

	aesgcm := getAEAD()
	nonceSize := aesgcm.NonceSize() // 12 bytes padrão
	if len(ciphertext) < nonceSize {
		return nil, false
	}

	nonce := ciphertext[:nonceSize]
	sealed := ciphertext[nonceSize:]

	// [ENGULF-FIX VULN-3] PRIMEIRO validar AES-GCM. Nunca guardar estado de pacotes não-autenticados.
	// Isso impede OOM DoS: atacantes não-autenticados não conseguem poluir o sync.Map.
	decrypted, err := aesgcm.Open(nil, nonce, sealed, aad)
	if err != nil {
		// GCM Authentication FALHOU: pacote forjado, corrompido ou com Seed errada
		return nil, false
	}

	// [GEN-7 RT-03] Apenas desempacotar e avaliar lógicas L7 complexas (e Logs) SE O PACOTE FOR AUTÊNTICO.
	packetTime := int64(binary.BigEndian.Uint64(packet[4:12]))
	now := time.Now().Unix()
	drift := now - packetTime
	if drift < 0 {
		drift = -drift
	}
	if drift > MaxTimestampDriftSecs {
		log.Printf("[OMEGA-SECURITY] Pacote autenticado expirado (drift=%ds). Replay Window bloqueado.", drift)
		return nil, false
	}

	// [ENGULF-FIX VULN-3] SOMENTE após autenticação criptográfica, verificar replay no cache.
	nonceStr := string(nonce)
	if _, used := globalNonceCache.LoadOrStore(nonceStr, time.Now().Unix()); used {
		log.Println("[OMEGA-SECURITY-FATAL] Ataque de REPLAY L7 interceptado! Bloqueado.")
		return nil, false
	}

	if isJitter {
		return decrypted, true
	}

	// [ENGULF-FIX VULN-2] Payload transparente — sem expansão semântica cega em L4
	return decrypted, false
}

// cromEncrypt aplica AES-256-GCM autenticado para a resposta de volta ao Alpha.
// [ENGULF Gen-5] Formato: [MAGIC 4B][TIMESTAMP 8B][NONCE 12B][CIPHERTEXT+TAG]
func cromEncrypt(data []byte) []byte {
	// [ENGULF-FIX VULN-2] Payload transparente — sem compressão semântica cega em L4.
	// A mutação de tamanho de payload destruía Content-Length HTTP, permitindo Request Smuggling.
	processedData := data

	aesgcm := getAEAD()

	nonce := make([]byte, aesgcm.NonceSize()) // 12 bytes
	// [RT-09 FIX] Checar retorno de rand.Read — nonce zero = catástrofe GCM
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Printf("[OMEGA-CRYPTO-FATAL] Falha ao gerar nonce aleatório: %v", err)
		return nil
	}

	// [ENGULF-FIX VULN-4] Timestamp Gen-5 para AAD autenticado contra Replay
	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(time.Now().Unix()))

	// AAD Gen-5: [MAGIC 4B][TIMESTAMP 8B]
	aad := make([]byte, 0, 12)
	aad = append(aad, []byte(CromMagic)...)
	aad = append(aad, tsBytes...)

	// Seal: cifra + autentica com AAD expandido (magic + timestamp)
	sealed := aesgcm.Seal(nil, nonce, processedData, aad)

	// Pacote final Gen-5: [MAGIC 4B][TIMESTAMP 8B][NONCE 12B][CIPHERTEXT+TAG]
	packet := make([]byte, 0, 4+8+len(nonce)+len(sealed))
	packet = append(packet, []byte(CromMagic)...)
	packet = append(packet, tsBytes...)
	packet = append(packet, nonce...)
	packet = append(packet, sealed...)
	return packet
}

func handleAlienConnection(alienConn net.Conn) {
	defer alienConn.Close()

	// Ler o prefixo TCP Framer para lidar com desincronização em stream TCP (Nagle)
	// Prevenção inteligente contra Slowloris: Timeout de 3s na borda L4
	alienConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	initialPacket, err := readFramedPacket(alienConn)

	// Limpar o deadline após o handshake para não afetar streams contínuos legítimos (como Chat/WebSocket)
	alienConn.SetReadDeadline(time.Time{})

	if err != nil {
		return
	}

	// ===== SILENT DROP =====
	// Se o pacote não tem a assinatura CROM, fechar a conexão sem responder NADA.
	// Isso impede banner grabbing e information leakage.
	plaintext, isJitt := cromDecryptPacket(initialPacket)
	n := len(initialPacket)
	if plaintext == nil {
		// [RT-10 FIX] Hash do endereço nos logs para OPSEC
		log.Printf("[OMEGA-SILENT-DROP] Pacote inválido de %s (%d bytes). Dropped.", hashAddr(alienConn.RemoteAddr()), n)
		return
	}

	if isJitt {
		// Cover-Traffic cego P2P. Absorve a conexao sem mandar pro backend e encerra.
		log.Printf("[OMEGA] JITTER Cover-Traffic Inicial Recebido (%d bytes). Absorvido.", n)
		return
	}

	// [RT-10 FIX] Hash do endereço
	log.Printf("[OMEGA] Pacote CROM válido de %s (%d bytes)", hashAddr(alienConn.RemoteAddr()), len(plaintext))

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
		// [RT-11 FIX] Contador de pacotes inválidos mid-stream
		invalidCount := 0
		for {
			// [RT-18 FIX] Retorno do timeout para prevenir Slowloris post-handshake e goroutine explosion
			alienConn.SetReadDeadline(time.Now().Add(10 * time.Second))
			packet, err := readFramedPacket(alienConn)
			if err != nil {
				if err != io.EOF {
					log.Printf("[OMEGA] Alien read err: %v", err)
				}
				backendConn.Close()
				return
			}
			pt, isJt := cromDecryptPacket(packet)
			if pt == nil {
				// [RT-11 FIX] Fechar conexão após MaxInvalidMidStream pacotes inválidos consecutivos
				invalidCount++
				if invalidCount >= MaxInvalidMidStream {
					log.Printf("[OMEGA-SECURITY] %d pacotes inválidos consecutivos mid-stream. Encerrando conexão.", invalidCount)
					backendConn.Close()
					return
				}
				log.Printf("[OMEGA-SILENT-DROP] Pacote corrompido mid-stream (%d/%d). Dropped.", invalidCount, MaxInvalidMidStream)
				continue
			}
			invalidCount = 0 // Reset no sucesso
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
			werr := writeFramedPacket(alienConn, encrypted)
			if werr != nil {
				return
			}
		}
	}()

	wg.Wait()
}

// [GEN-6 RT-08 FIX] Helpers para tracking per-IP.
func incrementIPConn(addr string) bool {
	ipMutex.Lock()
	defer ipMutex.Unlock()
	
	if perIPConns[addr] >= MaxConnsPerIP {
		return false
	}
	perIPConns[addr]++
	return true
}

func decrementIPConn(addr string) {
	ipMutex.Lock()
	defer ipMutex.Unlock()
	
	perIPConns[addr]--
	if perIPConns[addr] <= 0 {
		delete(perIPConns, addr)
	}
}

func extractIP(addr net.Addr) string {
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return addr.String()
	}
	return host
}

func main() {
	// Seed do PRNG para jitter anti-fingerprint
	mrand.Seed(time.Now().UnixNano())

	// [RT-01 FIX] Cache do cipher AES e limpeza automática
	_ = getAEAD()

	// [GEN-6 RT-11 FIX] Banner sanitizado — sem revelar mecanismos de segurança.
	fmt.Println("=================================================================")
	fmt.Println(" [ CROM PROXY OMEGA (Gen-6 Hardened) ]")
	fmt.Printf("  Escutando: %s | Backend: %s\n", SwarmListenAddr, BackendRealHost)
	fmt.Println("=================================================================")

	l, err := net.Listen("tcp", SwarmListenAddr)
	if err != nil {
		log.Fatal(err)
	}

	// [RT-13 FIX] Semáforo de conexões máximas para prevenir FD exhaustion
	connSemaphore := make(chan struct{}, MaxConcurrentConns)

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}

		// [GEN-6 RT-08 FIX] Per-IP connection limiting
		clientIP := extractIP(conn.RemoteAddr())
		if !incrementIPConn(clientIP) {
			conn.Close()
			continue
		}

		// [RT-13 FIX] Tentar adquirir slot. Se no limite, rejeitar silenciosamente.
		select {
		case connSemaphore <- struct{}{}:
			go func() {
				defer func() {
					<-connSemaphore
					decrementIPConn(clientIP)
				}()
				handleAlienConnection(conn)
			}()
		default:
			// Limite de conexões atingido — silent drop
			decrementIPConn(clientIP)
			conn.Close()
		}
	}
}
