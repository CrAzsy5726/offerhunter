package main

import (
	"log"

	"offerhunter/internal/config"
	"offerhunter/internal/server"
)

func main() {
	r := server.NewRouter()
	addr := config.ServerHost + ":" + config.ServerPort
	log.Printf("Offer-Hunter LangChain Edition starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}
