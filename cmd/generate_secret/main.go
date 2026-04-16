package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
)

func main() {
	// Generate 32 bytes (256 bits) of random data
	secret := make([]byte, 32)
	_, err := rand.Read(secret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating secret: %v\n", err)
		os.Exit(1)
	}

	// Encode to base64 for easy storage
	encoded := base64.StdEncoding.EncodeToString(secret)

	fmt.Println("===========================================")
	fmt.Println("  Strong JWT Secret Generator")
	fmt.Println("===========================================")
	fmt.Println()
	fmt.Println("Generated JWT Secret (32 bytes, base64):")
	fmt.Println(encoded)
	fmt.Println()
	fmt.Println("Add this to your .env file:")
	fmt.Printf("JWT_SECRET=%s\n", encoded)
	fmt.Println()
	fmt.Println("⚠️  IMPORTANT:")
	fmt.Println("- Keep this secret safe and never commit to git")
	fmt.Println("- Use different secrets for dev/staging/production")
	fmt.Println("- Changing this secret will invalidate all existing tokens")
	fmt.Println("===========================================")
}
