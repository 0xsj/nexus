package main

import (
	"fmt"
	"os"

	"github.com/0xsj/nexus/platform/internal/integration/infrastructure/crypto"
)

// Generate encryption key utility
// Usage: go run keygen/main.go

func main() {
	key, err := crypto.GenerateEncryptionKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate key: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=================================================================")
	fmt.Println("Generated AES-256 Encryption Key (Base64)")
	fmt.Println("=================================================================")
	fmt.Println()
	fmt.Println(key)
	fmt.Println()
	fmt.Println("Add this to your environment variables:")
	fmt.Printf("export INTEGRATION_TOKEN_ENCRYPTION_KEY=%s\n", key)
	fmt.Println()
	fmt.Println("⚠️  KEEP THIS KEY SECURE!")
	fmt.Println("   - Never commit to version control")
	fmt.Println("   - Store in a secrets manager (AWS Secrets Manager, Vault, etc.)")
	fmt.Println("   - Rotate regularly")
	fmt.Println("=================================================================")
}
