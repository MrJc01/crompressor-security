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
	AlienSwarmListenAddr = "127.0.0.1:9999"    // Escuta tráfego hostil P2P da Rede/Mixnet
	TrueBackendAddr      = "127.0.0.1:8080"    // O banco de dados ou Apache/Nginx real rodando na máquina destino
	ServerTenantSeed     = "Tenant_Alfa_Code_404" // A seed de mutação tem que ser exatamente igual para inflar e decodificar coerentemente
)

// Mutação do Espaço Latente Reversa/Direta (Simulação Gêmea)
// O Crompressor trabalha com deltas, o XOR de um tráfego XOR com a mesma Seed simula a restauração em O(1) do CROMFS
func resolveFromAlienTraffic(data []byte, seed string) []byte {
	h := hmac.New(sha256.New, []byte(seed))
	resolvedStream := make([]byte, len(data))
	
	for i, b := range data {
		h.Write([]byte{b})
		hashChunk := h.Sum(nil)
		// Revertendo o "Lixo" devolta a JSON/Strings
		resolvedStream[i] = b ^ hashChunk[0] 
	}
	return resolvedStream
}

func handleAlienConnection(alienConn net.Conn) {
	defer alienConn.Close()
	log.Printf("[CÉREBRO-OMEGA] Ingestão de tráfego espacial P2P remoto: %s", alienConn.RemoteAddr())

	// Conectar na Aplicação Verdadeira de forma transparente (Drop-in Egress)
	backendConn, err := net.Dial("tcp", TrueBackendAddr)
	if err != nil {
		log.Printf("[CÉREBRO-OMEGA-DB-ERRO] Falha ao bater no banco local protegido: %v", err)
		return
	}
	defer backendConn.Close()

	// Tráfego Alien -> Decodifica -> Backend
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := alienConn.Read(buf)
			if err != nil {
				if err != io.EOF { log.Printf("Alien read err: %v", err) }
				return
			}
			alienData := buf[:n]
			
			// Reversão Simétrica
			start := time.Now()
			trueData := resolveFromAlienTraffic(alienData, ServerTenantSeed)
			log.Printf("[CÉREBRO-OMEGA->DB] Descifrado/Inflado %d bytes para formato nativo HTTP/TCP (Latency: %v)", n, time.Since(start))
			
			_, err = backendConn.Write(trueData)
			if err != nil {
				log.Println("Backend Write err:", err)
				return
			}
		}
	}()

	// Backend Responde -> Mutate/Comprime -> Retorna Cérebro/Alien Swarm
	buf := make([]byte, 4096)
	for {
		n, err := backendConn.Read(buf)
		if err != nil {
			if err != io.EOF { log.Printf("Backend read err: %v", err) }
			break
		}
		rawDbOutput := buf[:n]
		
		alienReply := resolveFromAlienTraffic(rawDbOutput, ServerTenantSeed) 
		
		_, err = alienConn.Write(alienReply)
		if err != nil {
			log.Println("Alien Network Repy err:", err)
			break
		}
	}
}

func main() {
	fmt.Println("==========================================================")
	fmt.Println("  [ Crompressor Drop-In Egress Node (Cérebro Omega)  ]    ")
	fmt.Println("  Alien Port Listening :", AlienSwarmListenAddr)
	fmt.Println("  Shielded DB/App      :", TrueBackendAddr)
	fmt.Println("  Mutated Brain Verify :", hex.EncodeToString([]byte(ServerTenantSeed))[:12] + "...")
	fmt.Println("==========================================================")

	listener, err := net.Listen("tcp", AlienSwarmListenAddr)
	if err != nil {
		log.Fatal("Erro no Listen da Swarm hostil:", err)
		os.Exit(1)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept failed:", err)
			continue
		}
		go handleAlienConnection(conn)
	}
}
