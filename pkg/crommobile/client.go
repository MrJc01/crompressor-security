package crommobile

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	CromMagic   = "CROM"
	JitterMagic = "JITT"
)

// [RT-06 FIX] Seed agora é unexported — nenhum package externo pode lê-la diretamente.
// [RT-01 FIX] O valor padrão é vazio. A seed DEVE ser configurada via SetTenantSeed()
//             ou via variável de ambiente CROM_TENANT_SEED antes de iniciar o túnel.
var globalTenantSeed string

// seedOnce garante que a configuração de fallback via env var aconteça apenas uma vez.
var seedOnce sync.Once

// GetTenantSeed retorna a seed configurada, com fallback para env var.
// Se nenhuma fonte estiver disponível, usa o valor hardcoded legado (com warning).
func GetTenantSeed() string {
	seedOnce.Do(func() {
		if globalTenantSeed == "" {
			envSeed := os.Getenv("CROM_TENANT_SEED")
			if envSeed != "" {
				globalTenantSeed = envSeed
				log.Println("[ALPHA-SECURITY] Seed carregada via CROM_TENANT_SEED env var.")
			} else {
				log.Fatal("[ALPHA-SECURITY] ⚠️  CROM_TENANT_SEED não definida. Abortando.")
			}
		}
	})
	return globalTenantSeed
}

// SetTenantSeed configura a seed do tenant. Deve ser chamada antes de StartTunnel().
func SetTenantSeed(seed string) {
	if seed == "" {
		log.Fatal("[ALPHA-SECURITY] Tentativa de configurar TenantSeed vazia. Abortando.")
	}
	globalTenantSeed = seed
}

// applyLLMSemanticCompression simula a tese do crompressor-sinapse:
// Substituindo texto de alta rotatividade por UUIDs neurais discretos reduzindo o payload.
func applyLLMSemanticCompression(data []byte) []byte {
	str := string(data)
	str = strings.ReplaceAll(str, "HTTP/1.1", "⌬HTTP1")
	str = strings.ReplaceAll(str, "Accept: application/json", "⌬CTJSON")
	str = strings.ReplaceAll(str, "Connection: keep-alive", "⌬CONKA")
	str = strings.ReplaceAll(str, "User-Agent", "⌬UA")
	return []byte(str)
}

// cromEncrypt aplica Tokenização Semântica → AES-256-GCM (Cifra Autenticada Bancária)
func cromEncrypt(data []byte, magic string) []byte {

	// Compressão Cognitiva (Se for CROM normal)
	var processedData []byte
	if magic == CromMagic {
		processedData = applyLLMSemanticCompression(data)
	} else {
		processedData = data
	}

	// Derivar chave AES-256 via HMAC-SHA256 da TenantSeed
	// [RT-01 FIX] Usa GetTenantSeed() que carrega de env var, não hardcode
	mac := hmac.New(sha256.New, []byte(GetTenantSeed()))
	mac.Write([]byte("CROM_AES_GCM_KEY_V4"))
	key := mac.Sum(nil) // 32 bytes = AES-256

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("[ALPHA-CRYPTO] Falha ao criar cifra AES: %v", err)
		return nil
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("[ALPHA-CRYPTO] Falha ao criar GCM: %v", err)
		return nil
	}

	nonce := make([]byte, aesgcm.NonceSize()) // 12 bytes
	// [RT-09 FIX] Checar retorno de rand.Read — nonce zero = catástrofe criptográfica
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Printf("[ALPHA-CRYPTO-FATAL] Falha ao gerar nonce aleatório: %v", err)
		return nil
	}

	// Seal: cifra + autentica. AAD (Additional Authenticated Data) = magic header
	sealed := aesgcm.Seal(nil, nonce, processedData, []byte(magic))

	// Pacote final: [MAGIC 4B][NONCE 12B][CIPHERTEXT+TAG]
	packet := make([]byte, 0, 4+len(nonce)+len(sealed))
	packet = append(packet, []byte(magic)...)
	packet = append(packet, nonce...)
	packet = append(packet, sealed...)
	return packet
}

func cromDecryptPacket(packet []byte) []byte {
	if len(packet) < 4 {
		return nil
	}
	if string(packet[:4]) != CromMagic {
		return nil
	}
	ciphertext := packet[4:]

	// Derivar chave AES-256 via HMAC-SHA256 da TenantSeed
	// [RT-01 FIX] Usa GetTenantSeed()
	mac := hmac.New(sha256.New, []byte(GetTenantSeed()))
	mac.Write([]byte("CROM_AES_GCM_KEY_V4"))
	key := mac.Sum(nil) // 32 bytes = AES-256

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil
	}

	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil
	}

	nonce := ciphertext[:nonceSize]
	sealed := ciphertext[nonceSize:]

	// Open valida o GCM Tag — qualquer adulteração = DROP
	decrypted, err := aesgcm.Open(nil, nonce, sealed, []byte(CromMagic))
	if err != nil {
		return nil
	}

	// Reverter compressão semântica
	str := string(decrypted)
	str = strings.ReplaceAll(str, "⌬HTTP1", "HTTP/1.1")
	str = strings.ReplaceAll(str, "⌬CTJSON", "Accept: application/json")
	str = strings.ReplaceAll(str, "⌬CONKA", "Connection: keep-alive")
	str = strings.ReplaceAll(str, "⌬UA", "User-Agent")

	return []byte(str)
}

// [RT-04 FIX] startJitterCoverTraffic agora aceita context.Context para cancelamento.
// Quando a conexão do cliente fechar, o context é cancelado e a goroutine termina.
func startJitterCoverTraffic(ctx context.Context, swarmAddr string) {
	for {
		select {
		case <-ctx.Done():
			// Conexão encerrada — parar a névoa criptográfica
			return
		case <-time.After(300 * time.Millisecond):
			conn, err := net.DialTimeout("tcp", swarmAddr, 500*time.Millisecond)
			if err == nil {
				fakeData := make([]byte, 64)
				// [RT-09 FIX] Checar erro do rand no jitter também
				if _, randErr := io.ReadFull(rand.Reader, fakeData); randErr != nil {
					conn.Close()
					continue
				}
				jittPacket := cromEncrypt(fakeData, JitterMagic)
				conn.Write(jittPacket)
				conn.Close()
			}
		}
	}
}

// StartTunnel exportado para SDK iOS/Android via GoMobile
func StartTunnel(listenAddr string, swarmAddr string) error {
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
			n, err := clientConn.Read(buf)
			if err != nil {
				swarmConn.Close()
				return
			}
			encrypted := cromEncrypt(buf[:n], CromMagic)
			_, werr := swarmConn.Write(encrypted)
			if werr != nil {
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 32768)
		for {
			n, err := swarmConn.Read(buf)
			if err != nil {
				clientConn.Close()
				return
			}
			plaintext := cromDecryptPacket(buf[:n])
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
