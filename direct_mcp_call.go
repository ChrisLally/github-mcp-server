package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

type GetProjectParams struct {
	Owner  string `json:"owner"`
	Number int    `json:"number"`
}

func main() {
	// Create the JSON-RPC request
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "callTool",
		Params: map[string]interface{}{
			"name": "get_project_v2",
			"input": GetProjectParams{
				Owner:  "manian0430",
				Number: 1,
			},
		},
		ID: 1,
	}

	// Marshal the request to JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		return
	}

	fmt.Printf("Request: %s\n", string(requestJSON))

	// Send the request to the MCP server
	resp, err := http.Post("http://localhost:8080/jsonrpc", "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
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

	fmt.Println("Response:")
	fmt.Println(prettyJSON.String())
} 