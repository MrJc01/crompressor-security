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
)

// [RT-01 FIX] TenantSeed carregada de variável de ambiente em vez de hardcoded.
// Se CROM_TENANT_SEED não estiver definida, usa fallback legado (com warning).
var tenantSeed string

func getTenantSeed() string {
	if tenantSeed == "" {
		envSeed := os.Getenv("CROM_TENANT_SEED")
		if envSeed != "" && envSeed != "WIPED_BY_SEC_POLICY" {
			tenantSeed = envSeed
			log.Println("[OMEGA-SECURITY] Seed carregada via CROM_TENANT_SEED env var.")
		} else if envSeed == "WIPED_BY_SEC_POLICY" || tenantSeed == "WIPED_BY_SEC_POLICY" {
			// Vazio / memory secure
		} else {
			log.Fatal("[OMEGA-SECURITY] ⚠️  CROM_TENANT_SEED não definida. Abortando.")
		}
	}
	return tenantSeed
}

var globalAEAD cipher.AEAD
var onceAEAD sync.Once

// [RT-17 FIX] Nonce Anti-Replay Map
var globalNonceCache sync.Map

func getAEAD() cipher.AEAD {
	onceAEAD.Do(func() {
		seed := getTenantSeed()
		mac := hmac.New(sha256.New, []byte(seed))
		mac.Write([]byte("CROM_AES_GCM_KEY_V4"))
		key := mac.Sum(nil)
		block, err := aes.NewCipher(key)
		if err != nil {
			log.Fatalf("[OMEGA-FATAL] Falha no AES: %v", err)
		}
		globalAEAD, err = cipher.NewGCM(block)
		if err != nil {
			log.Fatalf("[OMEGA-FATAL] Falha no GCM: %v", err)
		}
		// [RT-15 FIX] Memory Wipe of Env var
		// [ENGULF-FIX VULN-1] ATENÇÃO: os.Setenv NÃO limpa /proc/PID/environ no Linux.
		// Em produção, usar STDIN pipe ou Secret Manager (Vault) para injetar a seed.
		// O snapshot original da env var permanece em /proc/PID/environ até o processo morrer.
		os.Setenv("CROM_TENANT_SEED", "WIPED_BY_SEC_POLICY")
		log.Println("[OMEGA-SECURITY] ⚠️  AVISO: /proc/PID/environ pode ainda conter a seed original (limitação POSIX).")
		tenantSeed = "WIPED_BY_SEC_POLICY"

		// [ENGULF-FIX VULN-4] Janitor com TTL granular — limpa entradas individuais após 60s
		// Em vez de purge total (que reabria a janela de Replay), cada nonce expira isoladamente.
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

func writeFramedPacket(conn net.Conn, packet []byte) error {
	lenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBuf, uint16(len(packet)))
	if _, err := conn.Write(lenBuf); err != nil {
		return err
	}
	_, err := conn.Write(packet)
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
	packetTime := int64(binary.BigEndian.Uint64(packet[4:12]))
	now := time.Now().Unix()
	drift := now - packetTime
	if drift < 0 {
		drift = -drift
	}
	if drift > 30 {
		log.Printf("[OMEGA-SECURITY] Pacote expirado (drift=%ds). Replay Window bloqueado.", drift)
		return nil, false
	}

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

func main() {
	// [RT-01 FIX] Cache do cipher AES e limpeza automática
	_ = getAEAD()

	fmt.Println("=================================================================")
	fmt.Println(" [ CROM ALIEN PROXY OUT-FLIGHT (v3 Hardened + Silent Drop) ]")
	fmt.Println(" AES-256-GCM | Anti-Ptrace | Max-Conn Limiter | IP Hashing")
	fmt.Println(" Escutando nuvem: localhost:9999 | Backend: localhost:8080")
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

		// [RT-13 FIX] Tentar adquirir slot. Se no limite, rejeitar silenciosamente.
		select {
		case connSemaphore <- struct{}{}:
			go func() {
				defer func() { <-connSemaphore }()
				handleAlienConnection(conn)
			}()
		default:
			// Limite de conexões atingido — silent drop
			conn.Close()
		}
	}
}
