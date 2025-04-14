package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/github/github-mcp-server/pkg/github"
	"github.com/github/github-mcp-server/pkg/translations"
	gogithub "github.com/google/go-github/v69/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

func TestGetProjectV2(t *testing.T) {
	// Set up environment variables
	token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if token == "" {
		t.Fatal("GITHUB_PERSONAL_ACCESS_TOKEN not set")
	}

	// Create HTTP client with auth
	httpClient := &http.Client{
		Transport: &oauth2.Transport{
			Base:   http.DefaultTransport,
			Source: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}),
		},
	}

	// Create GitHub REST client
	ghClient := gogithub.NewClient(httpClient)
	ghClient.UserAgent = "github-mcp-server/test"

	// Create GitHub GraphQL client
	graphqlClient := githubv4.NewClient(httpClient)

	getClient := func(_ context.Context) (*gogithub.Client, *githubv4.Client, error) {
		return ghClient, graphqlClient, nil
	}

	trans, _ := translations.TranslationHelper()

	// Create the GetProjectV2 tool handler
	tool, handler := github.GetProjectV2(getClient, trans)

	// Create a test context
	ctx := context.Background()

	// Create a test request
	request := struct {
		Params struct {
			Arguments map[string]interface{} `json:"arguments"`
		} `json:"params"`
	}{
		Params: struct {
			Arguments map[string]interface{} `json:"arguments"`
		}{
			Arguments: map[string]interface{}{
				"owner":  "manian0430",
				"number": "1",
			},
		},
	}

	// Call the handler
	result, err := handler(ctx, request)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	// Check the result
	if result == nil {
		t.Fatal("handler returned nil result")
	}

	// Print the result
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Result: %s\n", string(resultJSON))

	// Check if the result is correct
	expectedResult := `{"content":[{"type":"text","text":"{\"message\":\"GetProjectV2 handler executed, skipping actual API call.\"}"}],"isError":false}`
	actualResult := string(resultJSON)

	if actualResult != expectedResult {
		t.Errorf("result is incorrect, got: %v, want: %v", actualResult, expectedResult)
	}

	fmt.Printf("Tool Name: %s\n", tool.Name)
	fmt.Printf("Tool Description: %s\n", tool.Description)
}

func main() {
	testing.Main(nil, nil, nil)
}
