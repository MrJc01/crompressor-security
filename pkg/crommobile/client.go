package crommobile

import (
	"context"
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
	CromMagic   = "CROM"
	JitterMagic = "JITT"

	// [GEN-6 RT-06 FIX] Janela de drift máximo do timestamp autenticado (segundos)
	MaxTimestampDriftSecs = 5

	// [GEN-8 RT-202 FIX] Limite máximo de entradas no nonce cache para prevenir OOM.
	MaxNonceCacheEntries = 100000

	// [GEN-8 RT-204 FIX] Idle timeout para mid-stream reads (segundos).
	MidStreamIdleTimeoutSecs = 120

	// [GEN-8 RT-205 FIX] Tamanho máximo de pacote no framing.
	MaxFramedPacketSize = 35000
)

// [GEN-7 RT-06 FIX] Seed agora é recebida apenas em bytes ou pipe para prevenir extração de strings (Memory Dump).
var globalTenantSeedBytes []byte
var seedMutex sync.Mutex

// [GEN-8 RT-202 FIX] Contador atômico de entradas no nonce cache.
var nonceCacheCount int64

// [GEN-8 RT-208 FIX] KDF label ofuscada em byte-level para prevenir extração via strings.
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
				log.Fatal("[ALPHA-FATAL-DRM] ptrace() detectado em runtime. Abort imediato!")
			}
		}
	}()
}

// SetTenantSeedBytes insere a seed de forma segura. O desenvolvedor deve zerar `seed` no seu array original após isso.
func SetTenantSeedBytes(seed []byte) {
	seedMutex.Lock()
	defer seedMutex.Unlock()
	globalTenantSeedBytes = make([]byte, len(seed))
	copy(globalTenantSeedBytes, seed)
}

// Deprecated: SetTenantSeed is vulnerable to memory dumps because strings are immutable.
func SetTenantSeed(seed string) {
	SetTenantSeedBytes([]byte(seed))
}

func secureReadSeedAndInitAEAD() cipher.AEAD {
	// [GEN-8 RT-206 FIX] Bloquear /proc/PID/mem reads e core dumps.
	_, _, errno := syscall.RawSyscall(syscall.SYS_PRCTL, 4, 0, 0)
	if errno != 0 {
		log.Printf("[ALPHA-SECURITY] Aviso: prctl(PR_SET_DUMPABLE,0) falhou: %v", errno)
	}

	// [GEN-7 DRM] Anti-Trace (TracerPid Check)
	status, err := os.ReadFile("/proc/self/status")
	if err == nil && strings.Contains(string(status), "TracerPid:\t") {
		if !strings.Contains(string(status), "TracerPid:\t0") {
			log.Fatal("[ALPHA-FATAL-DRM] ptrace() interceptado. Execução terminada para evitar Memory Dumps!")
		}
	}

	// [GEN-8 RT-203 FIX] Iniciar watchdog contínuo
	startAntiDebugWatchdog()

	seedMutex.Lock()
	defer seedMutex.Unlock()

	var activeSeed []byte

	if len(globalTenantSeedBytes) > 0 {
		activeSeed = globalTenantSeedBytes
	} else {
		// Prioridade 2: Ler seed via STDIN pipe
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			log.Fatal("[ALPHA-SECURITY] MODO INSEGURO DETECTADO. Sem globalTenantSeedBytes e sem STDIN pipe. Abortando.")
		}

		rawBytes := make([]byte, 1024)
		n, err := os.Stdin.Read(rawBytes)
		if err != nil && err != io.EOF {
			log.Fatal("[ALPHA-FATAL] Falha lendo pipe STDIN.")
		}

		// Alloc estático para Memory Leak
		var validCount int
		for _, b := range rawBytes[:n] {
			if b >= 32 && b <= 126 {
				validCount++
			}
		}
		activeSeed = make([]byte, 0, validCount)
		for _, b := range rawBytes[:n] {
			if b >= 32 && b <= 126 {
				activeSeed = append(activeSeed, b)
			}
		}
		// Zeroize STDIN
		for i := range rawBytes { rawBytes[i] = 0 }
	}

	if len(activeSeed) == 0 {
		log.Fatal("[ALPHA-FATAL] Nenhuma Seed válida fornecida.")
	}

	// [GEN-8 RT-208 FIX] Derivar AES key com label ofuscada.
	kdfLabel := decodeKDFLabel()
	mac := hmac.New(sha256.New, activeSeed)
	mac.Write(kdfLabel)
	key := mac.Sum(nil)

	// [GEN-8] Zeroize label decodificada imediatamente
	for i := range kdfLabel { kdfLabel[i] = 0 }

	// [GEN-7 RT-01] Zeroize
	for i := range activeSeed { activeSeed[i] = 0 }
	globalTenantSeedBytes = nil

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalf("[ALPHA-FATAL] Falha no AES: %v", err)
	}

	for i := range key { key[i] = 0 }

	aead, err := cipher.NewGCM(block)
	if err != nil {
		log.Fatalf("[ALPHA-FATAL] Falha no GCM: %v", err)
	}

	// [GEN-7] Força coleta de lixo
	runtime.GC()
	debug.FreeOSMemory()

	log.Println("[ALPHA-SECURITY] KMS Armado. Runtime Zeroize concluído.")
	return aead
}

var globalAEAD cipher.AEAD
var onceAEAD sync.Once
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

// [GEN-6 RT-05 FIX] Atomic write: combina length + payload em um único buffer.
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

// Removed duplicated SetTenantSeed

// applyLLMSemanticCompression simula a tese do crompressor-sinapse:
// [ENGULF-FIX VULN-2] DESATIVADA: A substituição cega de strings em payload TCP
// alterava o tamanho do body sem atualizar Content-Length, permitindo HTTP Request Smuggling.
// Proxies L4 NUNCA devem modificar payload L7. Retorna dados intactos.
func applyLLMSemanticCompression(data []byte) []byte {
	return data
}

// cromEncrypt aplica AES-256-GCM autenticado.
// [GEN-7] Formato: [MAGIC 4B][DIR 1B][TIMESTAMP 8B][NONCE 12B][CIPHERTEXT+TAG]
func cromEncrypt(data []byte, magic string) []byte {
	// [ENGULF-FIX VULN-2] Payload transparente — sem compressão semântica cega em L4.
	processedData := data

	aesgcm := getAEAD()
	nonce := make([]byte, aesgcm.NonceSize()) // 12 bytes
	// [RT-09 FIX] Checar retorno de rand.Read — nonce zero = catástrofe criptográfica
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Printf("[ALPHA-CRYPTO-FATAL] Falha ao gerar nonce aleatório: %v", err)
		return nil
	}

	// Timestamp Gen-7 para AAD autenticado contra Replay
	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(time.Now().Unix()))

	dirFlag := byte('C')
	if magic == JitterMagic {
		dirFlag = 'J'
	}

	// AAD Gen-7: [MAGIC 4B][DIR 1B][TIMESTAMP 8B]
	aad := make([]byte, 0, 13)
	aad = append(aad, []byte(magic)...)
	aad = append(aad, dirFlag)
	aad = append(aad, tsBytes...)

	// Seal: cifra + autentica com AAD expandido (magic + dir + timestamp)
	sealed := aesgcm.Seal(nil, nonce, processedData, aad)

	// Pacote final Gen-7: [MAGIC 4B][DIR 1B][TIMESTAMP 8B][NONCE 12B][CIPHERTEXT+TAG]
	packet := make([]byte, 0, 4+1+8+len(nonce)+len(sealed))
	packet = append(packet, []byte(magic)...)
	packet = append(packet, dirFlag)
	packet = append(packet, tsBytes...)
	packet = append(packet, nonce...)
	packet = append(packet, sealed...)
	return packet
}

// cromDecryptPacket valida e descriptografa um pacote CROM via AES-256-GCM.
// [GEN-7] Formato: [MAGIC 4B][DIR 1B][TIMESTAMP 8B][NONCE 12B][CIPHERTEXT+GCM_TAG]
func cromDecryptPacket(packet []byte) []byte {
	// Mínimo: 4 (magic) + 1 (dir) + 8 (ts) + 12 (nonce) + 16 (tag) = 41
	if len(packet) < 41 {
		return nil
	}
	if string(packet[:4]) != CromMagic {
		return nil
	}

	dirFlag := packet[4]
	if dirFlag != 'S' {
		return nil
	}

	// AAD Gen-7: [MAGIC 4B][DIR 1B][TIMESTAMP 8B]
	aad := packet[:13]
	ciphertext := packet[13:]

	aesgcm := getAEAD()
	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil
	}

	nonce := ciphertext[:nonceSize]
	sealed := ciphertext[nonceSize:]

	// [ENGULF-FIX VULN-3] PRIMEIRO validar AES-GCM. Nunca guardar estado de pacotes não-autenticados.
	decrypted, err := aesgcm.Open(nil, nonce, sealed, aad)
	if err != nil {
		return nil
	}

	// [GEN-7 RT-03] Avaliar drift APÓS AEAD para não permitir Logs Forjados
	packetTime := int64(binary.BigEndian.Uint64(packet[5:13]))
	now := time.Now().Unix()
	drift := now - packetTime
	if drift < 0 {
		drift = -drift
	}
	if drift > MaxTimestampDriftSecs {
		log.Printf("[ALPHA-SECURITY] Pacote autenticado expirado (drift=%ds). Replay Window bloqueado.", drift)
		return nil
	}

	// [GEN-8 RT-202 FIX] Verificar limite de entradas ANTES de inserir.
	if atomic.LoadInt64(&nonceCacheCount) >= MaxNonceCacheEntries {
		log.Println("[ALPHA-SECURITY] Nonce cache saturado. Rejeitando pacote para prevenir OOM.")
		return nil
	}

	// [ENGULF-FIX VULN-3] SOMENTE após autenticação criptográfica, verificar replay no cache.
	nonceStr := string(nonce)
	if _, used := globalNonceCache.LoadOrStore(nonceStr, time.Now().Unix()); used {
		log.Println("[ALPHA-SECURITY-FATAL] Ataque de REPLAY L7 interceptado! Bloqueado.")
		return nil
	}
	atomic.AddInt64(&nonceCacheCount, 1)

	// [ENGULF-FIX VULN-2] Payload transparente — sem expansão semântica cega
	return decrypted
}

// [GEN-8 RT-207 FIX] Jitter agora é multiplexado na conexão Swarm existente.
// Em vez de criar 1 conexão TCP por pacote (self-DoS), usa a conexão persistente.
func startJitterCoverTraffic(ctx context.Context, swarmConn net.Conn, mu *sync.Mutex) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(100+mrand.Intn(400)) * time.Millisecond):
			// [GEN-6 RT-10 FIX] Tamanho aleatório entre 32 e 2048 bytes
			fakeData := make([]byte, 32+mrand.Intn(2016))
			if _, randErr := io.ReadFull(rand.Reader, fakeData); randErr != nil {
				continue
			}
			jittPacket := cromEncrypt(fakeData, JitterMagic)
			if jittPacket == nil {
				continue
			}
			// [GEN-8] Mutex para serializar writes na conexão compartilhada
			mu.Lock()
			err := writeFramedPacket(swarmConn, jittPacket)
			mu.Unlock()
			if err != nil {
				return // Conexão morreu
			}
		}
	}
}

// StartTunnel exportado para SDK iOS/Android via GoMobile
func StartTunnel(listenAddr string, swarmAddr string) error {
	// [RT-01 FIX] Force cache init to prevent memory leak
	_ = getAEAD()

	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	fmt.Printf("[GoMobile CROM] Escutando %s -> Borda P2P %s\n", listenAddr, swarmAddr)

	for {
		clientConn, err := l.Accept()
		if err != nil {
			continue
		}
		go handleClient(clientConn, swarmAddr)
	}
}

func handleClient(clientConn net.Conn, swarmAddr string) {
	defer clientConn.Close()

	swarmConn, err := net.Dial("tcp", swarmAddr)
	if err != nil {
		log.Printf("[CROM] Swarm Edge inacessível: %v", err)
		return
	}
	defer swarmConn.Close()

	// [RT-04 FIX] Criar context cancelável vinculado ao ciclo de vida da conexão.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// [GEN-8 RT-207 FIX] Mutex para serializar writes na conexão Swarm compartilhada.
	var swarmWriteMu sync.Mutex

	// Inicia a névoa multiplexada na conexão existente
	go startJitterCoverTraffic(ctx, swarmConn, &swarmWriteMu)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 32768)
		for {
			n, err := clientConn.Read(buf)
			if err != nil {
				swarmConn.Close()
				return
			}
			encrypted := cromEncrypt(buf[:n], CromMagic)
			// [GEN-8] Lock para serializar com Jitter
			swarmWriteMu.Lock()
			werr := writeFramedPacket(swarmConn, encrypted)
			swarmWriteMu.Unlock()
			if werr != nil {
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		invalidCount := 0
		for {
			// [GEN-8 RT-204 FIX] Idle timeout mid-stream.
			swarmConn.SetReadDeadline(time.Now().Add(time.Duration(MidStreamIdleTimeoutSecs) * time.Second))
			packet, err := readFramedPacket(swarmConn)
			if err != nil {
				clientConn.Close()
				return
			}
			plaintext := cromDecryptPacket(packet)
			if plaintext == nil {
				invalidCount++
				if invalidCount >= 5 {
					log.Printf("[ALPHA-SECURITY] Muitos pacotes L7 corrompidos do Swarm (%d). Fechando TCP.", invalidCount)
					clientConn.Close()
					return
				}
				continue
			}
			invalidCount = 0
			_, werr := clientConn.Write(plaintext)
			if werr != nil {
				return
			}
		}
	}()

	wg.Wait()
}
