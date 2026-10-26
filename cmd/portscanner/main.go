package main

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/abhishekshinde2104/tcp-port-scanner/internal/scanner"
)

const version = "0.0.1"

type target struct {
	host string
	port int
}

func main() {

	targets := []target{
		{"127.0.0.1", 9999},
		{"scanme.nmap.org", 22},
		{"10.255.255.1", 80},
	}

	fmt.Println("tcp-port-scanner", "v"+version)

	for _, t := range targets {
		state := scanner.ScanPort(t.host, t.port, 2*time.Second)
		addr := net.JoinHostPort(t.host, strconv.Itoa(t.port))
		fmt.Printf("%-22s %s\n", addr, state)
	}

}
