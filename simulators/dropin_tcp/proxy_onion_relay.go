package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

const (
	OnionListenAddr = "127.0.0.1:9955"
)

func handleRelayConnection(incoming net.Conn, nextHop string) {
	defer incoming.Close()

	// Onion Routing: Cego, não sabe descriptografar. Apenas repassa a casca CROM inteira ao próximo salto.
	outgoing, err := net.Dial("tcp", nextHop)
	if err != nil {
		log.Printf("[ONION-RELAY] Next hop %s indisponível: %v", nextHop, err)
		return
	}
	defer outgoing.Close()

	log.Printf("[ONION-RELAY] Mixnet Bridge: %s -> %s", incoming.RemoteAddr(), nextHop)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		io.Copy(outgoing, incoming)
		outgoing.(*net.TCPConn).CloseWrite()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		io.Copy(incoming, outgoing)
		incoming.(*net.TCPConn).CloseWrite()
	}()

	wg.Wait()
}

func main() {
	var nextHopTarget = os.Getenv("ONION_NEXT_HOP")
	if nextHopTarget == "" {
		nextHopTarget = "127.0.0.1:9999" // Default Omega
	}

	fmt.Println("=================================================================")
	fmt.Println(" [ CROM ALIEN PROXY ONION-RELAY (Middle-Brain/Mixnet) ]")
	fmt.Println(" Roteamento Cego de Camada. Zero Conhecimento Semântico.")
	fmt.Printf(" Escutando: %s | Repassando: %s\n", OnionListenAddr, nextHopTarget)
	fmt.Println("=================================================================")

	l, err := net.Listen("tcp", OnionListenAddr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go handleRelayConnection(conn, nextHopTarget)
	}
}
