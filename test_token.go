package main

import (
	"encoding/json"
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

	// Create a request to the GitHub API
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		fmt.Printf("Error creating request: %s\n", err)
		return
	}

	// Add the token as a bearer token
	req.Header.Add("Authorization", "token "+token)
	req.Header.Add("Accept", "application/vnd.github+json")

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
		return
	}

	// Parse the response
	var user map[string]interface{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		fmt.Printf("Error parsing response: %s\n", err)
		return
	}

	// Print the user information
	fmt.Println("Successfully authenticated!")
	fmt.Printf("Username: %s\n", user["login"])
	fmt.Printf("Name: %s\n", user["name"])

	// Check the scopes from the response headers
	fmt.Println("\nToken scopes:")
	scopes := resp.Header.Get("X-OAuth-Scopes")
	if scopes == "" {
		fmt.Println("No scopes found in response headers")
	} else {
		fmt.Println(scopes)
	}
	
	// Try to list projects using REST API
	fmt.Println("\nTrying to list projects using REST API...")
	reqProjects, err := http.NewRequest("GET", "https://api.github.com/user/projects", nil)
	if err != nil {
		fmt.Printf("Error creating request: %s\n", err)
		return
	}
	
	reqProjects.Header.Add("Authorization", "token "+token)
	reqProjects.Header.Add("Accept", "application/vnd.github+json")
	
	respProjects, err := client.Do(reqProjects)
	if err != nil {
		fmt.Printf("Error sending request: %s\n", err)
		return
	}
	defer respProjects.Body.Close()
	
	bodyProjects, err := io.ReadAll(respProjects.Body)
	if err != nil {
		fmt.Printf("Error reading response: %s\n", err)
		return
	}
	
	if respProjects.StatusCode != http.StatusOK {
		fmt.Printf("Error listing projects: %s\n", respProjects.Status)
		fmt.Printf("Response body: %s\n", string(bodyProjects))
		return
	}
	
	fmt.Printf("Projects response: %s\n", string(bodyProjects))
} 