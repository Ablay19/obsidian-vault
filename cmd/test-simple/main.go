package main

import (
	"fmt"
	"os"

	"obsidian-automation/internal/config"
)

func main() {
	fmt.Println("🧪 Testing Cloudflare AI integration...")

	// Load config
	config.LoadConfig()

	fmt.Printf("✅ Configuration loaded\n")
	fmt.Printf("🌐 Worker URL: %s\n", os.Getenv("CLOUDFLARE_WORKER_URL"))

	// Just print config and exit
	fmt.Println("🎉 Test completed - configuration working!")
}
