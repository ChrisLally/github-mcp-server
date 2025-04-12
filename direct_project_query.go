package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if token == "" {
		fmt.Println("GITHUB_PERSONAL_ACCESS_TOKEN environment variable not set")
		return
	}

	fmt.Printf("Using token: %s...%s (length: %d)\n", token[:5], token[len(token)-4:], len(token))

	// Define the GraphQL query for a user's project
	query := `
	{
		"query": "query($login: String!, $number: Int!) { user(login: $login) { projectV2(number: $number) { id title shortDescription public items(first: 10) { nodes { id content { ... on Issue { title } ... on PullRequest { title } ... on DraftIssue { title } } } } } } }",
		"variables": {
			"login": "manian0430",
			"number": 1
		}
	}`

	// Create a new request
	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer([]byte(query)))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	// Add headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	// Pretty print the JSON response
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
		fmt.Printf("Error formatting JSON: %v\n", err)
		fmt.Println(string(body))
		return
	}

	fmt.Println(prettyJSON.String())
} 