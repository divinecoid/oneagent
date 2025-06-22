package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/divinecoid/oneagent/pkg/config/keys"
)

func main() {
	// Parse command line flags
	env := flag.String("env", "", "Environment (dev|staging|prod)")
	version := flag.String("version", fmt.Sprintf("v%d", time.Now().Unix()), "Key version")
	flag.Parse()

	// Validate environment
	if *env == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *env != "dev" && *env != "staging" && *env != "prod" {
		log.Fatalf("Invalid environment: %s. Must be one of: dev, staging, prod", *env)
	}

	// Generate new key
	key, err := keys.GenerateKey(*env, *version)
	if err != nil {
		log.Fatalf("Failed to generate key: %v", err)
	}

	// Output key string
	fmt.Printf("Generated new encryption key for %s environment:\n", *env)
	fmt.Printf("ENCRYPTION_KEY_CURRENT=%s\n", key.String())
	fmt.Println("\nMake sure to update your .env file with this new key.")
	fmt.Println("If you're rotating keys, move your current key to ENCRYPTION_KEY_PREVIOUS")
	fmt.Printf("before updating ENCRYPTION_KEY_CURRENT.\n")
}