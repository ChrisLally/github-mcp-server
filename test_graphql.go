package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	// Get token from environment variable
	token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if token == "" {
		fmt.Println("GITHUB_PERSONAL_ACCESS_TOKEN environment variable is not set")
		return
	}

	// Print token info (masking most of it for security)
	if len(token) > 10 {
		fmt.Printf("Using token: %s...%s (length: %d)\n", token[:4], token[len(token)-4:], len(token))
	}

	// Create a GraphQL query directly using the HTTP API
	graphqlQuery := `{
		"query": "query { viewer { login id }}"
	}`

	// Create a request to the GitHub GraphQL API
	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer([]byte(graphqlQuery)))
	if err != nil {
		fmt.Printf("Error creating request: %s\n", err)
		return
	}

	// Add the token as a bearer token
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %s\n", err)
		return
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %s\n", err)
		return
	}

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", resp.Status)
		fmt.Printf("Response body: %s\n", string(body))
		fmt.Println("\nHeaders:")
		for k, v := range resp.Header {
			fmt.Printf("%s: %s\n", k, v)
		}
		return
	}

	// Print the response
	fmt.Printf("GraphQL Response: %s\n", string(body))
	
	// Check the scopes from the response headers
	fmt.Println("\nToken scopes:")
	scopes := resp.Header.Get("X-OAuth-Scopes")
	if scopes == "" {
		fmt.Println("No scopes found in response headers")
	} else {
		fmt.Println(scopes)
	}
} 