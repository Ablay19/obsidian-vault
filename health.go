package main

import (
"fmt"
"log"
"net/http"
"time"
)

var lastActivity time.Time

func startHealthServer() {
lastActivity = time.Now()
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
uptime := time.Since(lastActivity)
fmt.Fprintf(w, `{"status": "healthy", "uptime": "%s"}`, uptime)
})
log.Println("Health check at :8080/health")
go http.ListenAndServe(":8080", nil)
}

func updateActivity() {
lastActivity = time.Now()
}
