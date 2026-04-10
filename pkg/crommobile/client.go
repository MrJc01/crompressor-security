package crommobile

import (
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

// cromEncrypt aplica Tokenização Semântica -> XOR -> HMAC
func cromEncrypt(data []byte, magic string) []byte {
	
	// Compressão Cognitiva (Se for CROM normal)
	var processedData []byte
	if magic == CromMagic {
		processedData = applyLLMSemanticCompression(data)
	} else {
		processedData = data
	}

	mac := hmac.New(sha256.New, []byte(GlobalTenantSeed))
	mac.Write([]byte("CROM_SESSION_KEY"))
	key := mac.Sum(nil)

	nonce := make([]byte, 8)
	rand.Read(nonce)

	encrypted := make([]byte, len(processedData))
	for i, b := range processedData {
		encrypted[i] = b ^ key[i%len(key)] ^ nonce[i%len(nonce)]
	}

	packet := make([]byte, 0, 4+8+len(encrypted))
	packet = append(packet, []byte(magic)...)
	packet = append(packet, nonce...)
	packet = append(packet, encrypted...)
	return packet
}

func cromDecryptPacket(packet []byte) []byte {
	if len(packet) < 13 {
		return nil
	}
	if string(packet[:4]) != CromMagic { // No client não recebe Jitter
		return nil
	}
	nonce := packet[4:12]
	encrypted := packet[12:]

	mac := hmac.New(sha256.New, []byte(GlobalTenantSeed))
	mac.Write([]byte("CROM_SESSION_KEY"))
	key := mac.Sum(nil)

	decrypted := make([]byte, len(encrypted))
	for i, b := range encrypted {
		decrypted[i] = b ^ key[i%len(key)] ^ nonce[i%len(nonce)]
	}

	// Reverter compressão (O Omega não inverte nada pro Alpha que seja texto "cru", maximo que pode chegar é o Omega mandando DB plain e chegando encriptado no Alpha)
	// Na verdade a reversao ocorreria no Omega e aqui tbm (Simetrico)
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
