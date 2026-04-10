package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	// Importando da engine original (Excluido do go.mod local por ser PoC)
	// "github.com/MrJc01/crompressor/pkg/sdk"
)

const (
	AlienSwarmTarget = "127.0.0.1:9999"
	ListeningPort    = "127.0.0.1:5432" // Porta drop-in 
)

// CrompressorSDKMock simula a interface do SDK real baseada no arquivo pkg/sdk/api.go auditoado
type CrompressorSDKMock struct{}

func (c *CrompressorSDKMock) Pack(ctx context.Context, input string, output string) error {
	time.Sleep(15 * time.Millisecond) // Simula engine neural processando chunk (O(1))
	// Simula a compressão jogando lixo alienígena no arquivo de output e diminuindo o tamanho
	data, _ := os.ReadFile(input)
	compressed := fmt.Sprintf("XOR_ALIEN_CROM:[LEN=>%d]", len(data))
	return os.WriteFile(output, []byte(compressed), 0644)
}

func handleTCPStreamToSDK(clientConn net.Conn) {
	defer clientConn.Close()
	log.Printf("[SDK-ALPHA-BRAIN] Sequestrando conexão cliente de %s", clientConn.RemoteAddr())

	// Truque SRE: Se o Crompressor SDK espera HD, usamos TempFS Mount (Na memória RAM /dev/shm)
	// Evita latência de 14ms num SSD e vira 0.01ms na memoria DIMM
	ramDiskPath := os.TempDir()
	
	buf := make([]byte, 8192) // Lendo o máximo do stream
	for {
		n, err := clientConn.Read(buf)
		if err != nil {
			if err != io.EOF { log.Printf("Stream error: %v", err) }
			return
		}

		// 1. Grava no disco de RAM temporário
		inPath := filepath.Join(ramDiskPath, "ingress_chunk.crom")
		outPath := filepath.Join(ramDiskPath, "out_chunk.alien")
		os.WriteFile(inPath, buf[:n], 0644)

		start := time.Now()
		
		// 2. Invoca o Monstro Native do Crompressor SDK
		engine := &CrompressorSDKMock{}
		engine.Pack(context.Background(), inPath, outPath)

		// 3. Lê o lixo esmagado e injeta na MixNet TCP (Swarm Network)
		alienBytes, _ := os.ReadFile(outPath)
		
		log.Printf("[SDK-METRICS] Original: %d bytes | Comprimido/Mutado: %d bytes | Latency: %v", 
			n, len(alienBytes), time.Since(start))

		// Dispara na nuvem p2p (Omitindo handling do dial pra focar no SDK flow)
		swarmConn, _ := net.Dial("tcp", AlienSwarmTarget)
		swarmConn.Write(alienBytes)
		
		// Ler resposta alien:
		respBuf := make([]byte, 8192)
		_, _ = swarmConn.Read(respBuf)
		
		// Unpack Reverso (Mock do Proxy In para receber as respostas da DB)
		clientConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n Resposta alienigena simulada devolta"))
		swarmConn.Close()
	}
}

func main() {
	fmt.Println("=================================================================")
	fmt.Println(" [ CROM ALIEN PROXY IN-FLIGHT ]")
	fmt.Println(" Roteando Sockets crus on-the-fly usando TempFS + SDK nativo")
	fmt.Println(" Ouvindo na porta localhost:5432")
	fmt.Println("=================================================================")
	
	l, err := net.Listen("tcp", ListeningPort)
	if err != nil { log.Fatal(err) }
	for {
		conn, _ := l.Accept()
		go handleTCPStreamToSDK(conn)
	}
}
