package main

import (
	"fmt"
	"log"
	"os"

	"github.com/neonpc/go-amule-webui/internal/api"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	host := getEnv("AMULE_HOST", "127.0.0.1")
	portStr := getEnv("AMULE_PORT", "4712")
	listen := getEnv("LISTEN", ":8080")
	password := getEnv("AMULE_PASSWORD", "")

	port := 4712
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		log.Fatalf("invalid AMULE_PORT: %s", portStr)
	}

	if password == "" {
		log.Fatal("AMULE_PASSWORD is required")
	}

	server := api.NewServer(host, port, password, listen)

	if err := server.Run(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
