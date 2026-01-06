package main

import (
    "time"
    "github.com/your-username/obsidian-vault/internal/logger"
)

func main() {
    logger := logger.NewGoogleLogger("obsidian-bot-test")
    
    // Test basic logging
    logger.LogUserAction("test_action", "test-user-123", map[string]interface{}{
        "test": "basic_logging_test",
    })
    
    // Test structured logging
    logger.LogRequest("GET", "/test", "", "", "", "test-agent", "127.0.0.1", 200, time.Millisecond*10, "test-user-123")
    
    print_status "OK" "Google Logger test completed"
}
