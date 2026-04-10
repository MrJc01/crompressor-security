package main

import (
	"fmt"
	"log"
	"os"

	"crompressor-security/pkg/crommobile"
)

const (
	ListeningPort = "127.0.0.1:5432"
)

func main() {
	var cloudTarget = os.Getenv("SWARM_CLOUD_TARGET")
	if cloudTarget == "" {
		cloudTarget = "127.0.0.1:9999"
	}

	fmt.Println("=================================================================")
	fmt.Println(" [ CROM ALIEN PROXY IN-FLIGHT (Gen-3 Cloud & Mobile) ]")
	fmt.Println(" Roteando Sockets crus via GoMobile SDK (LLM + Jitter Cover)")
	fmt.Printf(" Ouvindo porta %s | Target: %s\n", ListeningPort, cloudTarget)
	fmt.Println("=================================================================")

	// Utiliza o pacote SDK criado para poder ser portado ao iOS/Android futuramente.
	err := crommobile.StartTunnel(ListeningPort, cloudTarget)
	if err != nil {
		log.Fatalf("Falha crítica no motor Alpha: %v", err)
	}
}
