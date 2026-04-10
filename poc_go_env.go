package main

import (
    "fmt"
    "os"
    "time"
)

func main() {
    val := os.Getenv("SECRET_VAR")
    fmt.Printf("Initial Go env: %s\n", val)
    os.Setenv("SECRET_VAR", "WIPED")
    fmt.Printf("Overwritten Go env: %s\n", os.Getenv("SECRET_VAR"))
    
    // Check proc environ
    b, _ := os.ReadFile(fmt.Sprintf("/proc/%d/environ", os.Getpid()))
    fmt.Printf("Proc environ contains HIDDEN_SECRET? %v\n", string(b) != "")
    for _, part := range b {
        if string(part) == "" { continue }
    }
    // we'll just grep it out simply
    time.Sleep(1 * time.Second)
}
