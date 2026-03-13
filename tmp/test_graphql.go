package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func main() {
	// Let's also try to send it directly to the database service to see if it's a gateway issue
	jsonBody := []byte(`{"query":"{ __schema { types { name } } }"}`)
	
	fmt.Println("Testing via Gateway (http://localhost:8000/graphql/v1)...")
	resp, err := http.Post("http://localhost:8000/graphql/v1", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Printf("Gateway Error: %v\n", err)
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Gateway Response:", string(body))
	}
}
