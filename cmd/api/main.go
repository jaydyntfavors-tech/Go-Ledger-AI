go
package main

import(
  "encoding/json"
  "log"
  "net/http"
  "os"
)

func main() {
  port := os.Getenv("PORT")
  if port == "" {
    port = "8080"
  }
  mux := http.NewServeMux()

  // Health check endpoint to verify the API is live
  mux.HandleFunc("GET /api/v1/heatlh", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
      "status":  "healthy",
      "service": "go-ledger-ai",
    })
  })

  log.Printf("Starting go-ledger-ai server on port %s...", port)
  if err := http.ListenAndServe(":"+port, mux); err != nil {
    log.Fatalf("Server failed to start: %v", err)
    }
  }
