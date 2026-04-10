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



	// [GEN-6 RT-11 FIX] Banner sanitizado — sem revelar mecanismos internos.
	fmt.Println("=================================================================")
	fmt.Println(" [ CROM PROXY ALPHA (Gen-6 Hardened) ]")
	fmt.Printf(" Ouvindo porta %s | Target: %s\n", ListeningPort, cloudTarget)
	fmt.Println("=================================================================")

	// Utiliza o pacote SDK criado para poder ser portado ao iOS/Android futuramente.
	err := crommobile.StartTunnel(ListeningPort, cloudTarget)
	if err != nil {
		log.Fatalf("Falha crítica no motor Alpha: %v", err)
	}
}
