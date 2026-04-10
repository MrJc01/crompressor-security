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
	"sync"
	"time"
)

const (
	CromMagic   = "CROM"
	JitterMagic = "JITT"

	// [GEN-6 RT-06 FIX] Janela de drift máximo do timestamp autenticado (segundos)
	MaxTimestampDriftSecs = 5
)

// [GEN-7 RT-06 FIX] Seed agora é recebida apenas em bytes ou pipe para prevenir extração de strings (Memory Dump).
var globalTenantSeedBytes []byte
var seedMutex sync.Mutex

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
	seedMutex.Lock()
	defer seedMutex.Unlock()

	var activeSeed []byte

	if len(globalTenantSeedBytes) > 0 {
		activeSeed = globalTenantSeedBytes
	} else {
		// Prioridade 2: Ler seed via STDIN pipe
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			log.Fatal("[ALPHA-SECURITY] ⚠️ MODO INSEGURO DETECTADO. Sem globalTenantSeedBytes e sem STDIN pipe. Abortando.")
		}

		rawBytes := make([]byte, 1024)
		n, err := os.Stdin.Read(rawBytes)
		if err != nil && err != io.EOF {
			log.Fatal("[ALPHA-FATAL] Falha lendo pipe STDIN.")
		}

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

	mac := hmac.New(sha256.New, activeSeed)
	mac.Write([]byte("CROM_AES_GCM_KEY_V4"))
	key := mac.Sum(nil)

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
	log.Println("[ALPHA-SECURITY] KMS Inicializado e memória higienizada (Zeroize).")
	return aead
}

var globalAEAD cipher.AEAD
var onceAEAD sync.Once
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

// Removed duplicated SetTenantSeed

// applyLLMSemanticCompression simula a tese do crompressor-sinapse:
// [ENGULF-FIX VULN-2] DESATIVADA: A substituição cega de strings em payload TCP
// alterava o tamanho do body sem atualizar Content-Length, permitindo HTTP Request Smuggling.
// Proxies L4 NUNCA devem modificar payload L7. Retorna dados intactos.
func applyLLMSemanticCompression(data []byte) []byte {
	return data
}

// cromEncrypt aplica AES-256-GCM autenticado.
// [ENGULF Gen-5] Formato: [MAGIC 4B][TIMESTAMP 8B][NONCE 12B][CIPHERTEXT+TAG]
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

	// [ENGULF-FIX VULN-4] Timestamp Gen-5 para AAD autenticado contra Replay
	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(time.Now().Unix()))

	// AAD Gen-5: [MAGIC 4B][TIMESTAMP 8B]
	aad := make([]byte, 0, 12)
	aad = append(aad, []byte(magic)...)
	aad = append(aad, tsBytes...)

	// Seal: cifra + autentica com AAD expandido (magic + timestamp)
	sealed := aesgcm.Seal(nil, nonce, processedData, aad)

	// Pacote final Gen-5: [MAGIC 4B][TIMESTAMP 8B][NONCE 12B][CIPHERTEXT+TAG]
	packet := make([]byte, 0, 4+8+len(nonce)+len(sealed))
	packet = append(packet, []byte(magic)...)
	packet = append(packet, tsBytes...)
	packet = append(packet, nonce...)
	packet = append(packet, sealed...)
	return packet
}

// cromDecryptPacket valida e descriptografa um pacote CROM via AES-256-GCM.
// [ENGULF Gen-5] Formato: [MAGIC 4B][TIMESTAMP 8B][NONCE 12B][CIPHERTEXT+GCM_TAG]
func cromDecryptPacket(packet []byte) []byte {
	// Mínimo: 4 (magic) + 8 (timestamp) + 12 (nonce) + 16 (gcm tag) = 40 bytes
	if len(packet) < 40 {
		return nil
	}
	if string(packet[:4]) != CromMagic {
		return nil
	}

	// AAD Gen-5: [MAGIC 4B][TIMESTAMP 8B]
	aad := packet[:12]
	ciphertext := packet[12:]

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
	packetTime := int64(binary.BigEndian.Uint64(packet[4:12]))
	now := time.Now().Unix()
	drift := now - packetTime
	if drift < 0 {
		drift = -drift
	}
	if drift > MaxTimestampDriftSecs {
		log.Printf("[ALPHA-SECURITY] Pacote autenticado expirado (drift=%ds). Replay Window bloqueado.", drift)
		return nil
	}

	// [ENGULF-FIX VULN-3] SOMENTE após autenticação criptográfica, verificar replay no cache.
	nonceStr := string(nonce)
	if _, used := globalNonceCache.LoadOrStore(nonceStr, time.Now().Unix()); used {
		log.Println("[ALPHA-SECURITY-FATAL] Ataque de REPLAY L7 interceptado! Bloqueado.")
		return nil
	}

	// [ENGULF-FIX VULN-2] Payload transparente — sem expansão semântica cega
	return decrypted
}

// [RT-04 FIX] startJitterCoverTraffic agora aceita context.Context para cancelamento.
// Quando a conexão do cliente fechar, o context é cancelado e a goroutine termina.
func startJitterCoverTraffic(ctx context.Context, swarmAddr string) {
	for {
		select {
		case <-ctx.Done():
			// Conexão encerrada — parar a névoa criptográfica
			return
		// [GEN-6 RT-10 FIX] Jitter anti-fingerprinting: tempo aleatório entre 100ms e 500ms
		case <-time.After(time.Duration(100+mrand.Intn(400)) * time.Millisecond):
			conn, err := net.DialTimeout("tcp", swarmAddr, 500*time.Millisecond)
			if err == nil {
				// [GEN-6 RT-10 FIX] Tamanho aleatório entre 32 e 2048 bytes
				fakeData := make([]byte, 32+mrand.Intn(2016))
				// [RT-09 FIX] Checar erro do rand no jitter também
				if _, randErr := io.ReadFull(rand.Reader, fakeData); randErr != nil {
					conn.Close()
					continue
				}
				jittPacket := cromEncrypt(fakeData, JitterMagic)
				writeFramedPacket(conn, jittPacket)
				conn.Close()
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
	// Quando handleClient() terminar (defer cancel()), a goroutine de jitter para.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Inicia a névoa! (agora com context para cancelamento limpo)
	go startJitterCoverTraffic(ctx, swarmAddr)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 32768)
		for {
			// [RT-18 FIX] Retorno do timeout 
			clientConn.SetReadDeadline(time.Now().Add(10 * time.Second))
			n, err := clientConn.Read(buf)
			if err != nil {
				swarmConn.Close()
				return
			}
			encrypted := cromEncrypt(buf[:n], CromMagic)
			if err := writeFramedPacket(swarmConn, encrypted); err != nil {
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// [RT-18 FIX] Retorno do timeout
			swarmConn.SetReadDeadline(time.Now().Add(10 * time.Second))
			packet, err := readFramedPacket(swarmConn)
			if err != nil {
				clientConn.Close()
				return
			}
			plaintext := cromDecryptPacket(packet)
			if plaintext == nil {
				continue
			}
			_, werr := clientConn.Write(plaintext)
			if werr != nil {
				return
			}
		}
	}()

	wg.Wait()
}
