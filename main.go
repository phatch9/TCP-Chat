package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	s := newServer()
	go s.run()

	listener, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Fatalf("unable to start server: %s", err.Error())
	}

	defer listener.Close()
	log.Printf("server started on :8888")
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║        TCP-Chat Server Started         ║")
	fmt.Println("║      Listening on localhost:8888       ║")
	fmt.Println("╚════════════════════════════════════════╝")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\nShutting down server...")
		listener.Close()
		os.Exit(0)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %s", err.Error())
			continue
		}

		c := s.newClient(conn)
		go c.readInput()
	}
}
