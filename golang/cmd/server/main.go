package main

import (
	"github.com/julianorinaldi/alexa-llm-go"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Try to load .env file if it exists
	_ = godotenv.Load("../../.env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	http.HandleFunc("/", alexallm.HandleAlexaRequest)

	// Força a criação do banco no boot
	alexallm.InitDB()

	fmt.Printf("📡 Servidor de Debug Local (GO) rodando na porta %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Erro no servidor: %v", err)
	}
}
