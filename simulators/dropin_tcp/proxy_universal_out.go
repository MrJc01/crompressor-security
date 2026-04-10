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
)

const (
	AlienSwarmTarget = "127.0.0.1:9999"
	BackendRealHost  = "127.0.0.1:8080" // Porta da "Api Legada" do JS ou PHP
)

type CrompressorSDKMock struct{}

func (c *CrompressorSDKMock) Unpack(ctx context.Context, input string, output string) error {
	time.Sleep(10 * time.Millisecond) 
	// Simula a Matrix reversa reconstruindo um pacote gigante e legível da poeira
	recon := "GET /api/data HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n"
	return os.WriteFile(output, []byte(recon), 0644)
}

func handleAlienToBackend(alienConn net.Conn) {
	defer alienConn.Close()
	log.Printf("[SDK-OMEGA-BRAIN] Detectado estilhaço espacial P2P vindos de %s", alienConn.RemoteAddr())

	ramDiskPath := os.TempDir()
	
	buf := make([]byte, 8192)
	for {
		n, err := alienConn.Read(buf)
		if err != nil {
			if err != io.EOF { log.Printf("Stream error: %v", err) }
			return
		}

		inPath := filepath.Join(ramDiskPath, "ingress_alien.crom")
		outPath := filepath.Join(ramDiskPath, "out_reconstruido.sock")
		os.WriteFile(inPath, buf[:n], 0644)

		start := time.Now()
		
		// Unpack Nativo
		engine := &CrompressorSDKMock{}
		engine.Unpack(context.Background(), inPath, outPath)

		// Bytes vivos e decodificados em formato de protocolo estandar
		trueBytes, _ := os.ReadFile(outPath)
		
		log.Printf("[SDK-METRICS] Alien Unpacked. Latency Matrix Hitting: %v", time.Since(start))

		// Empurra ladeira baixo pro Servidor HTTP real
		backendConn, err := net.Dial("tcp", BackendRealHost)
		if(err == nil) {
		    backendConn.Write(trueBytes)
		    
		    respBuf := make([]byte, 16000)
		    bn, _ := backendConn.Read(respBuf)
		    // Retorna a resposta esmagada pra nuvem (Omisso a logica esmagadora de saida pra resumir)
		    alienConn.Write(respBuf[:bn])
		    backendConn.Close()
		}
	}
}

func main() {
	fmt.Println("=================================================================")
	fmt.Println(" [ CROM ALIEN PROXY OUT-FLIGHT (EGRESS) ]")
	fmt.Println(" Resolvendo poeira p2p em pacotes TCP para Apis Backend (Node/PHP)")
	fmt.Println(" Escutando nuvem: localhost:9999 | Ocultando App em: 8080")
	fmt.Println("=================================================================")
	
	l, err := net.Listen("tcp", AlienSwarmTarget)
	if err != nil { log.Fatal(err) }
	for {
		conn, _ := l.Accept()
		go handleAlienToBackend(conn)
	}
}
