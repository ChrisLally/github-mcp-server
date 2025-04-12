package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	// The exact format for the JSON-RPC request
	jsonStr := []byte(`{
		"jsonrpc": "2.0",
		"method": "callTool",
		"params": {
			"name": "get_project_v2",
			"input": {
				"owner": "manian0430",
				"number": 1
			}
		},
		"id": 1
	}`)

	// Create a new request
	req, err := http.NewRequest("POST", "http://localhost:8080/jsonrpc", bytes.NewBuffer(jsonStr))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	// Add headers
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