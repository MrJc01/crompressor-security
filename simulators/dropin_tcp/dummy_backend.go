//go:build ignore

package main

import (
	"fmt"
	"log"
	"net/http"
)

// Esse arquivo simula qualquer app (ex: um blog Wordpress em PHP, ou um Server Adonis.js/Express)
// Ele serve dados crus via HTTP na porta local. Ele *NÃO TEM CONHECIMENTO* do Crompressor.
func mockNodeJSApp(w http.ResponseWriter, r *http.Request) {
	log.Printf("[NODE/PHP-APP] Recebeu request de: %s (Metodo: %s)", r.RemoteAddr, r.Method)
	
	// Payload hiper verbosa como costuma ser no mundo json legados
	payload := `{
		"status": "success",
		"tenant": "Legacy_App",
		"data": {
			"users": [
				{"id": 1, "name": "Admin", "permissions": "ALL"},
				{"id": 2, "name": "User", "permissions": "READ"}
			],
			"system_note": "Este payload não criptografado transitaria pela internet abrindo brechas para packet sniffers caso nao passasse pela rede Alienigena."
		}
	}`
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(payload))
}

func main() {
	port := ":8080"
	fmt.Println("=================================================================")
	fmt.Println("  [ Sistema Legado Simulado (PHP/NodeJS) rodando na porta 8080 ] ")
	fmt.Println("  O sistema está vulnerável caso a porta seja exposta publicamente")
	fmt.Println("=================================================================")
	
	http.HandleFunc("/api/data", mockNodeJSApp)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Erro ao subir a falsa API: %v", err)
	}
}
