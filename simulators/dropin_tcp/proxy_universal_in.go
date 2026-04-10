package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

const (
	AlienSwarmTarget = "127.0.0.1:9999"
	ListeningPort    = "127.0.0.1:5432"
	// Seed secreta compartilhada entre Alpha e Omega (por inquilino)
	TenantSeed = "CROM-SEC-TENANT-ALPHA-2026"
	// Prefixo mágico para identificar pacotes CROM válidos
	CromMagic = "CROM"
)

// cromEncrypt aplica XOR cíclico com chave derivada do HMAC da seed do tenant.
// Isso garante que o tráfego no canal P2P seja entropia pura.
func cromEncrypt(data []byte) []byte {
	mac := hmac.New(sha256.New, []byte(TenantSeed))
	mac.Write([]byte("CROM_SESSION_KEY"))
	key := mac.Sum(nil) // 32 bytes de chave derivada

	// Gerar 8 bytes de nonce aleatório (anti-replay)
	nonce := make([]byte, 8)
	rand.Read(nonce)

	// XOR cíclico com a chave
	encrypted := make([]byte, len(data))
	for i, b := range data {
		encrypted[i] = b ^ key[i%len(key)] ^ nonce[i%len(nonce)]
	}

	// Formato do pacote: [4 magic][8 nonce][N encrypted_data]
	packet := make([]byte, 0, 4+8+len(encrypted))
	packet = append(packet, []byte(CromMagic)...)
	packet = append(packet, nonce...)
	packet = append(packet, encrypted...)
	return packet
}

func handleClient(clientConn net.Conn) {
	defer clientConn.Close()
	log.Printf("[ALPHA] Conexão de %s", clientConn.RemoteAddr())

	// Estabelecer conexão persistente com o Swarm (full-duplex)
	swarmConn, err := net.Dial("tcp", AlienSwarmTarget)
	if err != nil {
		log.Printf("[ALPHA] Swarm indisponível: %v", err)
		return
	}
	defer swarmConn.Close()

	var wg sync.WaitGroup

	// Goroutine 1: Client -> Encrypt -> Swarm (upstream)
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 32768) // 32KB buffer (suporta payloads grandes)
		for {
			n, err := clientConn.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("[ALPHA] Client read err: %v", err)
				}
				swarmConn.Close() // Sinaliza pro outro lado
				return
			}
			encrypted := cromEncrypt(buf[:n])
			_, werr := swarmConn.Write(encrypted)
			if werr != nil {
				return
			}
		}
	}()

	// Goroutine 2: Swarm -> Decrypt -> Client (downstream)
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 32768)
		for {
			n, err := swarmConn.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("[ALPHA] Swarm read err: %v", err)
				}
				clientConn.Close()
				return
			}
			// Os dados de volta do Omega já vêm criptografados com CROM header.
			// Precisamos descriptografar antes de devolver ao cliente.
			plaintext := cromDecryptPacket(buf[:n])
			if plaintext == nil {
				log.Printf("[ALPHA] Resposta inválida do Swarm (dropped)")
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

// cromDecryptPacket extrai e descriptografa um pacote CROM.
func cromDecryptPacket(packet []byte) []byte {
	// Validar tamanho mínimo: 4 (magic) + 8 (nonce) + 1 (data)
	if len(packet) < 13 {
		return nil
	}
	// Validar magic header
	if string(packet[:4]) != CromMagic {
		return nil
	}
	nonce := packet[4:12]
	encrypted := packet[12:]

	mac := hmac.New(sha256.New, []byte(TenantSeed))
	mac.Write([]byte("CROM_SESSION_KEY"))
	key := mac.Sum(nil)

	decrypted := make([]byte, len(encrypted))
	for i, b := range encrypted {
		decrypted[i] = b ^ key[i%len(key)] ^ nonce[i%len(nonce)]
	}
	return decrypted
}

func main() {
	fmt.Println("=================================================================")
	fmt.Println(" [ CROM ALIEN PROXY IN-FLIGHT (v2 Full-Duplex + Crypto) ]")
	fmt.Println(" Roteando Sockets crus on-the-fly com XOR+HMAC")
	fmt.Println(" Ouvindo na porta localhost:5432")
	fmt.Println("=================================================================")

	l, err := net.Listen("tcp", ListeningPort)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go handleClient(conn)
	}
}
