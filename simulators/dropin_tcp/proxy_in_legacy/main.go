//go:build ignore

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

// Configurações do ambiente simulado
const (
	ClientListenAddr = "127.0.0.1:5432"   // Porta local fingindo ser um DB Postgres ou App
	AlienSwarmTarget = "127.0.0.1:9999"   // Nó de entrada no Swarm/Alien Network (Pode ser o Relay ou Proxy Out)
	ClientTenantSeed = "Tenant_Alfa_Code_404" // Hash única geradora do Clone Mutante
)

// Mutação do Espaço Latente (Simulação do Codebook de-funcional)
// No modelo real aqui acionamos internal/codebook e internal/delta com XORs
func mutateToAlienTraffic(data []byte, seed string) []byte {
	// Cria uma derivação de ruído baseada na semente para embaralhar a Payload
	h := hmac.New(sha256.New, []byte(seed))
	alienStream := make([]byte, len(data))
	
	// Simulação de "Compressão + Alienação Estocástica"
	for i, b := range data {
		h.Write([]byte{b})
		hashChunk := h.Sum(nil)
		// O Byte sai como lixo criptográfico alienígena. Apenas quem tem a `seed` e o Histórico reverte
		alienStream[i] = b ^ hashChunk[0] 
	}
	return alienStream
}

func handleClientConnection(clientConn net.Conn) {
	defer clientConn.Close()
	log.Printf("[CÉREBRO-ALPHA] Interceptando conexão drop-in de: %s", clientConn.RemoteAddr())

	// Conectar na Red P2P/Swarm Hostil apontando para o próximo Cerebro/Relay
	alienConn, err := net.Dial("tcp", AlienSwarmTarget)
	if err != nil {
		log.Printf("[CÉREBRO-ALPHA-ERRO] Falha ao injetar no Swarm: %v", err)
		return
	}
	defer alienConn.Close()

	// Intercept and Mutate (Ingress to Swarm)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := clientConn.Read(buf)
			if err != nil {
				if err != io.EOF { log.Printf("Client read err: %v", err) }
				return
			}
			originalData := buf[:n]
			
			// Transformação Neural / Compressão Mutada
			start := time.Now()
			alienData := mutateToAlienTraffic(originalData, ClientTenantSeed)
			log.Printf("[CÉREBRO-ALPHA->ALIEN] Esmagado %d bytes para Ruído Semântico (Latency: %v)", n, time.Since(start))
			
			// Manda pra rede Hostil
			_, err = alienConn.Write(alienData)
			if err != nil {
				log.Println("Alien Network Write err:", err)
				return
			}
		}
	}()

	// Receive from Swarm and Decode back to Client
	buf := make([]byte, 4096)
	for {
		n, err := alienConn.Read(buf)
		if err != nil {
			if err != io.EOF { log.Printf("Alien net read err: %v", err) }
			break
		}
		alienData := buf[:n]
		
		// Reverte porque o Cérebro Ingress tem a lógica espelhada com o Egress 
		// (A matemátca do HMAC/XOR simula as chaves simétricas do Codebook/Swarm)
		originalResolved := mutateToAlienTraffic(alienData, ClientTenantSeed) 
		
		_, err = clientConn.Write(originalResolved)
		if err != nil {
			log.Println("Client write err:", err)
			break
		}
	}
}

func main() {
	fmt.Println("=========================================================")
	fmt.Println("  [ Crompressor Drop-In Ingress Node (Cérebro Alpha) ]   ")
	fmt.Println("  Listening (Drop-in):", ClientListenAddr)
	fmt.Println("  Target Swarm Node :", AlienSwarmTarget)
	fmt.Println("  Mutated Brain Seed:", hex.EncodeToString([]byte(ClientTenantSeed))[:12] + "...")
	fmt.Println("=========================================================")

	listener, err := net.Listen("tcp", ClientListenAddr)
	if err != nil {
		log.Fatal("Erro de DropIn Listen:", err)
		os.Exit(1)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept failed:", err)
			continue
		}
		go handleClientConnection(conn)
	}
}
