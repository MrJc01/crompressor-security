package crommobile

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	CromMagic    = "CROM"
	JitterMagic  = "JITT"
)

var GlobalTenantSeed = "CROM-SEC-TENANT-ALPHA-2026"

func SetTenantSeed(seed string) {
	GlobalTenantSeed = seed
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
	mac := hmac.New(sha256.New, []byte(GlobalTenantSeed))
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
	rand.Read(nonce)

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
	mac := hmac.New(sha256.New, []byte(GlobalTenantSeed))
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

// startJitterCoverTraffic injeta lixo criptografado na rede via conexoes pararelas cegas (Smoke screen)
func startJitterCoverTraffic(swarmAddr string) {
	for {
		time.Sleep(300 * time.Millisecond) // Rajadas paraleas de fumaça
		conn, err := net.Dial("tcp", swarmAddr)
		if err == nil {
			fakeData := make([]byte, 64)
			rand.Read(fakeData)
			jittPacket := cromEncrypt(fakeData, JitterMagic)
			conn.Write(jittPacket)
			conn.Close()
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

	// Inicia a névoa!
	go startJitterCoverTraffic(swarmAddr)

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
