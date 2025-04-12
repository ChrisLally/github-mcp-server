package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type CallRequest struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  struct {
		Method string `json:"method"`
		Params struct {
			Owner       string `json:"owner"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Public      bool   `json:"public"`
		} `json:"params"`
	} `json:"params"`
}

func main() {
	// Create a request to call the create_project_v2 tool
	request := CallRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "mcp.callTool",
		Params: struct {
			Method string `json:"method"`
			Params struct {
				Owner       string `json:"owner"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Public      bool   `json:"public"`
			} `json:"params"`
		}{
			Method: "create_project_v2",
			Params: struct {
				Owner       string `json:"owner"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Public      bool   `json:"public"`
			}{
				Owner:       "manian0430",
				Title:       "Local MCP Project Test",
				Description: "Testing project creation via local MCP server",
				Public:      true,
			},
		},
	}

	// Convert the request to JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("Error marshaling request: %s\n", err)
		return
	}

	// Print the request
	fmt.Printf("Request: %s\n", string(requestJSON))

	// Create an HTTP POST request to the local MCP server
	req, err := http.NewRequest("POST", "http://localhost:3000/jsonrpc", bytes.NewBuffer(requestJSON))
	if err != nil {
		fmt.Printf("Error creating request: %s\n", err)
		return
	}

	// Set the content type
	req.Header.Set("Content-Type", "application/json")

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

	// Print the response
	fmt.Printf("Response: %s\n", string(body))
} 