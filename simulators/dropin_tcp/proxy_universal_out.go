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
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"runtime"
	"runtime/debug"
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

	// [GEN-6 RT-08 FIX / GEN-7] Aumentado massivamente para suportar Jitter Burst.
	MaxConnsPerIP = 500

	// [GEN-8 RT-202 FIX] Limite máximo de entradas no nonce cache para prevenir OOM.
	MaxNonceCacheEntries = 100000

	// [GEN-8 RT-204 FIX] Idle timeout para mid-stream reads (segundos).
	// Previne goroutine exhaustion via slow-feed após handshake.
	MidStreamIdleTimeoutSecs = 120

	// [GEN-8 RT-205 FIX] Tamanho máximo de pacote no framing.
	MaxFramedPacketSize = 35000
)

// [GEN-7 RT-02 FIX] Seed não é mais mantida globalmente em string imutável.
// Apenas a AEAD global será mantida em memória local, dificultando ptrace extraction.

// [GEN-7 RT-08 FIX] Bloqueio robusto anti-Race Condition (Mutex L4).
var ipMutex sync.Mutex
var perIPConns = make(map[string]int)

// [GEN-8 RT-202 FIX] Contador atômico de entradas no nonce cache.
var nonceCacheCount int64

// [GEN-8 RT-208 FIX] KDF label ofuscada em byte-level para prevenir extração via strings.
// Decodificada em runtime via XOR com chave de rotação.
var kdfLabelObfuscated = []byte{
	0x1a, 0x0b, 0x16, 0x14, 0x06, 0x18, 0x1c, 0x0a,
	0x06, 0x1e, 0x1a, 0x14, 0x06, 0x12, 0x1c, 0x00,
	0x06, 0x0f, 0x6d,
}
var kdfLabelXORKey = byte(0x59)

func decodeKDFLabel() []byte {
	out := make([]byte, len(kdfLabelObfuscated))
	for i, b := range kdfLabelObfuscated {
		out[i] = b ^ kdfLabelXORKey
	}
	return out
}

// [GEN-8 RT-203 FIX] Watchdog anti-debug contínuo.
// Verifica TracerPid a cada 500ms em vez de apenas no startup.
func startAntiDebugWatchdog() {
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			status, err := os.ReadFile("/proc/self/status")
			if err != nil {
				continue
			}
			statusStr := string(status)
			if strings.Contains(statusStr, "TracerPid:\t") &&
				!strings.Contains(statusStr, "TracerPid:\t0") {
				log.Fatal("[OMEGA-FATAL-DRM] ptrace() detectado em runtime. Abort imediato!")
			}
		}
	}()
}

// secureReadSeed lê a chave de forma segura no startup e joga tudo para bytes apagáveis
func secureReadSeedAndInitAEAD() cipher.AEAD {
	// [GEN-8 RT-206 FIX] Bloquear /proc/PID/mem reads e core dumps.
	// PR_SET_DUMPABLE = 4, valor 0 = não-dumpable.
	_, _, errno := syscall.RawSyscall(syscall.SYS_PRCTL, 4, 0, 0)
	if errno != 0 {
		log.Printf("[OMEGA-SECURITY] Aviso: prctl(PR_SET_DUMPABLE,0) falhou: %v", errno)
	}

	// [GEN-7 DRM] Anti-Trace (TracerPid Check) — verificação inicial no startup
	status, err := os.ReadFile("/proc/self/status")
	if err == nil && strings.Contains(string(status), "TracerPid:\t") {
		if !strings.Contains(string(status), "TracerPid:\t0") {
			log.Fatal("[OMEGA-FATAL-DRM] ptrace() interceptado. Execução terminada para evitar Memory Dumps!")
		}
	}

	// [GEN-8 RT-203 FIX] Iniciar watchdog contínuo após check inicial
	startAntiDebugWatchdog()

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		log.Fatal("[OMEGA-SECURITY] MODO INSEGURO DETECTADO. Forneça a Seed de Tenant APENAS via STDIN pipe (echo SEED | proxy_out).")
	}

	rawBytes := make([]byte, 1024)
	n, err := os.Stdin.Read(rawBytes)
	if err != nil && err != io.EOF {
		log.Fatal("[OMEGA-FATAL] Falha lendo pipe STDIN.")
	}

	// Localizar trim
	seedBytes := rawBytes[:n]
	var validCount int
	for _, b := range seedBytes {
		if b >= 32 && b <= 126 {
			validCount++
		}
	}
	trimmed := make([]byte, 0, validCount)
	for _, b := range seedBytes {
		if b >= 32 && b <= 126 {
			trimmed = append(trimmed, b)
		}
	}

	if len(trimmed) == 0 {
		log.Fatal("[OMEGA-FATAL] Seed STDIN vazia.")
	}

	// [GEN-8 RT-208 FIX] Derivar AES key com label ofuscada.
	kdfLabel := decodeKDFLabel()
	mac := hmac.New(sha256.New, trimmed)
	mac.Write(kdfLabel)
	key := mac.Sum(nil)

	// [GEN-8] Zeroize label decodificada imediatamente
	for i := range kdfLabel { kdfLabel[i] = 0 }

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

	// [GEN-7] Força coleta de lixo da cópia do key state na lib crypto original
	runtime.GC()
	debug.FreeOSMemory()

	log.Println("[OMEGA-SECURITY] KMS Armado. Runtime Zeroize concluído.")
	return aead
}

var globalAEAD cipher.AEAD
var onceAEAD sync.Once

// [RT-17 FIX] Nonce Anti-Replay Map
var globalNonceCache sync.Map

func getAEAD() cipher.AEAD {
	onceAEAD.Do(func() {
		globalAEAD = secureReadSeedAndInitAEAD()

		// [GEN-8 RT-202 FIX] Janitor com TTL granular + limite de entradas.
		go func() {
			for {
				time.Sleep(10 * time.Second)
				now := time.Now().Unix()
				var cleaned int64
				globalNonceCache.Range(func(key, value interface{}) bool {
					if ts, ok := value.(int64); ok {
						if now-ts > 60 {
							globalNonceCache.Delete(key)
							cleaned++
						}
					}
					return true
				})
				if cleaned > 0 {
					atomic.AddInt64(&nonceCacheCount, -cleaned)
				}
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
	if packetLen > 35000 {
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
// [GEN-8 RT-205 FIX] Guarda contra uint16 overflow silencioso.
func writeFramedPacket(conn net.Conn, packet []byte) error {
	if len(packet) > MaxFramedPacketSize {
		return fmt.Errorf("packet too large for framing: %d > %d", len(packet), MaxFramedPacketSize)
	}
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
// [GEN-7] Formato: [MAGIC 4B][DIR 1B][TIMESTAMP 8B][NONCE 12B][CIPHERTEXT+GCM_TAG]
func cromDecryptPacket(packet []byte) ([]byte, bool) {
	// Mínimo: 4 (magic) + 1 (dir) + 8 (timestamp) + 12 (nonce) + 16 (gcm tag) = 41 bytes
	if len(packet) < 41 {
		return nil, false
	}

	magic := string(packet[:4])
	if magic != CromMagic && magic != JitterMagic {
		return nil, false
	}

	dirFlag := packet[4]
	if dirFlag != 'C' && dirFlag != 'J' {
		// [GEN-7 VULN-1 FIX] Anti-Reflection Attack. Se não for C ou J, reject.
		return nil, false
	}

	isJitter := magic == JitterMagic

	// AAD Gen-7: [MAGIC 4B][DIR 1B][TIMESTAMP 8B] — autenticados pelo GCM Tag
	aad := packet[:13]
	ciphertext := packet[13:]

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
	packetTime := int64(binary.BigEndian.Uint64(packet[5:13]))
	now := time.Now().Unix()
	drift := now - packetTime
	if drift < 0 {
		drift = -drift
	}
	if drift > MaxTimestampDriftSecs {
		log.Printf("[OMEGA-SECURITY] Pacote autenticado expirado (drift=%ds). Replay Window bloqueado.", drift)
		return nil, false
	}

	// [GEN-8 RT-202 FIX] Verificar limite de entradas ANTES de inserir.
	if atomic.LoadInt64(&nonceCacheCount) >= MaxNonceCacheEntries {
		log.Println("[OMEGA-SECURITY] Nonce cache saturado. Rejeitando pacote para prevenir OOM.")
		return nil, false
	}

	// [ENGULF-FIX VULN-3] SOMENTE após autenticação criptográfica, verificar replay no cache.
	nonceStr := string(nonce)
	if _, used := globalNonceCache.LoadOrStore(nonceStr, time.Now().Unix()); used {
		log.Println("[OMEGA-SECURITY-FATAL] Ataque de REPLAY L7 interceptado! Bloqueado.")
		return nil, false
	}
	atomic.AddInt64(&nonceCacheCount, 1)

	if isJitter {
		return decrypted, true
	}

	// [ENGULF-FIX VULN-2] Payload transparente — sem expansão semântica cega em L4
	return decrypted, false
}

// cromEncrypt aplica AES-256-GCM autenticado para a resposta de volta ao Alpha.
// [GEN-7] Formato: [MAGIC 4B][DIR 1B][TIMESTAMP 8B][NONCE 12B][CIPHERTEXT+TAG]
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

	// Timestamp Gen-7 para AAD autenticado contra Replay
	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(time.Now().Unix()))

	// AAD Gen-7: [MAGIC 4B][DIR 1B][TIMESTAMP 8B]
	aad := make([]byte, 0, 13)
	aad = append(aad, []byte(CromMagic)...)
	aad = append(aad, 'S') // Direcional: S = Server -> Client
	aad = append(aad, tsBytes...)

	// Seal: cifra + autentica com AAD expandido
	sealed := aesgcm.Seal(nil, nonce, processedData, aad)

	// Pacote final Gen-7: [MAGIC 4B][DIR 1B][TIMESTAMP 8B][NONCE 12B][CIPHERTEXT+TAG]
	packet := make([]byte, 0, 4+1+8+len(nonce)+len(sealed))
	packet = append(packet, []byte(CromMagic)...)
	packet = append(packet, 'S')
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
			// [GEN-8 RT-204 FIX] Idle timeout no mid-stream para prevenir goroutine exhaustion.
			// Diferente do handshake (3s), este timeout é longo (120s) para suportar WS idle.
			alienConn.SetReadDeadline(time.Now().Add(time.Duration(MidStreamIdleTimeoutSecs) * time.Second))
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
			// [GEN-9 RT-300 FIX] Deadline contra Backend Hang
			backendConn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			_, werr := backendConn.Write(pt)
			if werr != nil {
				// [GEN-9 RT-300 FIX] Cross-close imediato para acordar Theard Oposta Lida
				alienConn.Close()
				backendConn.Close()
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
			// [GEN-9 RT-300 FIX] Zero-Window Anti-DoS Pipeline limit (Deadline)
			alienConn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			werr := writeFramedPacket(alienConn, encrypted)
			if werr != nil {
				// [GEN-9 RT-300 FIX] Cross-close imediato para limpar recursos e slots MaxConn
				alienConn.Close()
				backendConn.Close()
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
	// [GEN-8 RT-213 FIX] Go 1.22+ usa auto-seeding. mrand.Seed() é deprecated.
	_ = mrand.Intn(1) // Referência para manter import sem dead-code

	// [RT-01 FIX] Cache do cipher AES e limpeza automática
	_ = getAEAD()

	// [GEN-8 RT-210 FIX] Banner sanitizado Gen-8.
	fmt.Println("=================================================================")
	fmt.Println(" [ CROM PROXY OMEGA (Gen-8 Hardened) ]")
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
