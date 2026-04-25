//go:build ignore

package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

const (
	MiddleBrainListenAddr = "127.0.0.1:9998" // Ingress manda para cá
	AlienSwarmTargetNode  = "127.0.0.1:9999" // Daqui enviamos para o Egress/Omega
)

// O Nó Relay (Middle Brain) não possui o "Seed" do Cliente. 
// Para ele, a informação que trafega é literalmente alienígena. Ele apenas 
// injeta ruído temporal (Jitter) para quebrar a assinatura de tempos (Time-analysis)
func injectMixnetJitter(data []byte) {
	// Atrasos variando de 10ms a 150ms para que hackers 
	// percam o Tracking das requisições originais.
	delay := time.Duration(rand.Intn(140)+10) * time.Millisecond
	time.Sleep(delay)
}

func handleRelay(incomingConn net.Conn) {
	defer incomingConn.Close()
	log.Printf("[MIDDLE-BRAIN RELAY] Engolindo mixnet de: %s", incomingConn.RemoteAddr())

	outgoingConn, err := net.Dial("tcp", AlienSwarmTargetNode)
	if err != nil {
		log.Printf("[MIDDLE-BRAIN-ERRO] Falha ao encaminhar ao Cérebro Omega: %v", err)
		return
	}
	defer outgoingConn.Close()

	// Tráfego Ingress -> Egress
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := incomingConn.Read(buf)
			if err != nil {
				if err != io.EOF { log.Printf("Relay read err: %v", err) }
				return
			}
			
			// Fatiamento (Splits) e Jitter
			// Simula o mixnet reempacotando blocos e mudando latências.
			injectMixnetJitter(buf[:n])
			log.Printf("[RELAY->OMEGA] Repassado %d bytes Alien com ruído temporal", n)
			
			// TODO: Aqui a Merkle Tree assina a "viagem" do pacote
			_, err = outgoingConn.Write(buf[:n])
			if err != nil {
				break
			}
		}
	}()

	// Tráfego Egress -> Ingress (Respostas)
	bufResp := make([]byte, 4096)
	for {
		n, err := outgoingConn.Read(bufResp)
		if err != nil {
			if err != io.EOF { log.Printf("Relay resp read err: %v", err) }
			break
		}
		
		injectMixnetJitter(bufResp[:n])
		log.Printf("[OMEGA->RELAY] Resposta Alien de %d bytes com ruído temporal", n)
		
		_, err = incomingConn.Write(bufResp[:n])
		if err != nil {
			break
		}
	}
}

func main() {
	fmt.Println("==========================================================")
	fmt.Println("  [ Crompressor Relay Node (Middle-Brain / Mixnet) ]      ")
	fmt.Println("  Listening Mixnet Port:", MiddleBrainListenAddr)
	fmt.Println("  Forwarding Target    :", AlienSwarmTargetNode)
	fmt.Println("  Action: Temporal Jitter Injection and Blind Routing")
	fmt.Println("==========================================================")

	listener, err := net.Listen("tcp", MiddleBrainListenAddr)
	if err != nil {
		log.Fatal("Listen falhou:", err)
		os.Exit(1)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept erro:", err)
			continue
		}
		go handleRelay(conn)
	}
}
