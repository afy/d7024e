package main

import (
  "fmt"
  "net/http"
  "os"
  "io"
)

func exampleGet(w http.ResponseWriter, r *http.Request) {
  fmt.Println("Recieved request on /") 
  io.WriteString(w, "Hello\n")
}

func main() {
  port := os.Getenv("PORT")
  if port == "" {
    port = "8008"
  }
  fmt.Println(port)
  fmt.Printf("Starting server at port %s\n", port)
  http.HandleFunc("/", exampleGet)
  err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
  if err != nil {
    fmt.Println("Fatal error!")
  }
}
