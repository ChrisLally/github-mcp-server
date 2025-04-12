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

	// First get the viewer ID
	viewerQuery := `{
		"query": "query { viewer { login id }}"
	}`

	// Create a request to the GitHub GraphQL API
	viewerReq, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer([]byte(viewerQuery)))
	if err != nil {
		fmt.Printf("Error creating request: %s\n", err)
		return
	}

	// Add the token as a bearer token
	viewerReq.Header.Add("Authorization", "Bearer "+token)
	viewerReq.Header.Add("Content-Type", "application/json")
	viewerReq.Header.Add("Accept", "application/json")

	// Send the request
	client := &http.Client{}
	viewerResp, err := client.Do(viewerReq)
	if err != nil {
		fmt.Printf("Error sending request: %s\n", err)
		return
	}
	defer viewerResp.Body.Close()

	// Read the response
	viewerBody, err := io.ReadAll(viewerResp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %s\n", err)
		return
	}

	// Check if the request was successful
	if viewerResp.StatusCode != http.StatusOK {
		fmt.Printf("Error: %s\n", viewerResp.Status)
		fmt.Printf("Response body: %s\n", string(viewerBody))
		return
	}

	fmt.Printf("Viewer Query Response: %s\n", string(viewerBody))

	// Now create a project
	createProjectQuery := `{
		"query": "mutation($input: CreateProjectV2Input!) { createProjectV2(input: $input) { projectV2 { id title } } }",
		"variables": {
			"input": {
				"ownerId": "MDQ6VXNlcjczOTI4OTIy",
				"title": "Test Project via Direct GraphQL"
			}
		}
	}`

	// Create a request to the GitHub GraphQL API
	projectReq, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer([]byte(createProjectQuery)))
	if err != nil {
		fmt.Printf("Error creating project request: %s\n", err)
		return
	}

	// Add the token as a bearer token
	projectReq.Header.Add("Authorization", "Bearer "+token)
	projectReq.Header.Add("Content-Type", "application/json")
	projectReq.Header.Add("Accept", "application/json")

	// Send the request
	projectResp, err := client.Do(projectReq)
	if err != nil {
		fmt.Printf("Error sending project request: %s\n", err)
		return
	}
	defer projectResp.Body.Close()

	// Read the response
	projectBody, err := io.ReadAll(projectResp.Body)
	if err != nil {
		fmt.Printf("Error reading project response: %s\n", err)
		return
	}

	// Check if the request was successful
	if projectResp.StatusCode != http.StatusOK {
		fmt.Printf("Error creating project: %s\n", projectResp.Status)
		fmt.Printf("Response body: %s\n", string(projectBody))
		return
	}

	// Print the response
	fmt.Printf("Project Creation Response: %s\n", string(projectBody))
	
	// Check the scopes from the response headers
	fmt.Println("\nToken scopes:")
	scopes := projectResp.Header.Get("X-OAuth-Scopes")
	if scopes == "" {
		fmt.Println("No scopes found in response headers")
	} else {
		fmt.Println(scopes)
	}
} 