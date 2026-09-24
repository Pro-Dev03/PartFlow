package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
)

// This command prints a one-time key pair. Keep the private key in the cloud
// secret manager and distribute only the public key with trusted desktop builds.
func main() {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("PARTFLOW_OFFLINE_GRANT_PRIVATE_KEY=%s\n", base64.StdEncoding.EncodeToString(privateKey))
	fmt.Printf("PARTFLOW_OFFLINE_GRANT_PUBLIC_KEY=%s\n", base64.StdEncoding.EncodeToString(publicKey))
}
