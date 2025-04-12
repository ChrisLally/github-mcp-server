package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// Custom type for the input
type CreateProjectV2Input struct {
	OwnerID     githubv4.ID     `json:"ownerId"`
	Title       githubv4.String `json:"title"`
	Description githubv4.String `json:"description,omitempty"`
	Public      githubv4.Boolean `json:"public,omitempty"`
}

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

	// Setup GitHub client with explicit endpoint
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	
	// Create a transport that adds the Authorization header
	transport := &oauth2.Transport{
		Base: http.DefaultTransport,
		Source: oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		),
	}
	
	// Create a client with the transport
	httpClient = &http.Client{Transport: transport}
	
	// Print the endpoint we're using
	fmt.Println("Using GraphQL endpoint: https://api.github.com/graphql")
	
	// Create the GraphQL client
	client := githubv4.NewClient(httpClient)

	// First, try to query the viewer to check authentication
	var viewerQuery struct {
		Viewer struct {
			Login string
			ID    string
		}
	}

	fmt.Println("Testing viewer query...")
	err := client.Query(context.Background(), &viewerQuery, nil)
	if err != nil {
		fmt.Printf("Error querying viewer: %s\n", err)
		if strings.Contains(strings.ToLower(err.Error()), "unauthorized") {
			fmt.Println("Authentication failed. Check that your token is valid and has the 'project' scope.")
		}
		// Try to get more details from the error
		if respErr, ok := err.(*githubv4.ResponseError); ok {
			fmt.Printf("GraphQL response errors:\n")
			for i, gqlErr := range respErr.Errors {
				fmt.Printf("  %d: %s\n", i+1, gqlErr.Message)
			}
		}
		return
	}

	fmt.Printf("Successfully authenticated as: %s (ID: %s)\n", viewerQuery.Viewer.Login, viewerQuery.Viewer.ID)

	// Owner who will own the project
	owner := "manian0430"
	fmt.Printf("Querying for user: %s\n", owner)

	// First try to find the owner ID from the user query
	var userQuery struct {
		User struct {
			ID    string
			Login string
		} `graphql:"user(login: $login)"`
	}

	userVars := map[string]interface{}{
		"login": githubv4.String(owner),
	}

	fmt.Println("Querying for user ID...")
	err = client.Query(context.Background(), &userQuery, userVars)
	if err != nil {
		fmt.Printf("Error querying user: %s\n", err)
		// Try to get more details from the error
		if respErr, ok := err.(*githubv4.ResponseError); ok {
			fmt.Printf("GraphQL response errors:\n")
			for i, gqlErr := range respErr.Errors {
				fmt.Printf("  %d: %s\n", i+1, gqlErr.Message)
			}
		}
		return
	}

	fmt.Printf("User query result: ID=%s, Login=%s\n", userQuery.User.ID, userQuery.User.Login)

	if userQuery.User.ID == "" {
		fmt.Println("User ID not found, trying as organization")
		var orgQuery struct {
			Organization struct {
				ID    string
				Login string
			} `graphql:"organization(login: $login)"`
		}

		orgVars := map[string]interface{}{
			"login": githubv4.String(owner),
		}

		err := client.Query(context.Background(), &orgQuery, orgVars)
		if err != nil {
			fmt.Printf("Error querying organization: %s\n", err)
			return
		}

		if orgQuery.Organization.ID == "" {
			fmt.Println("Could not find ID for user or organization")
			return
		}

		fmt.Printf("Organization found: ID=%s, Login=%s\n", orgQuery.Organization.ID, orgQuery.Organization.Login)
	}

	// Use the owner ID to create the project
	ownerID := userQuery.User.ID
	fmt.Printf("Using owner ID: %s\n", ownerID)

	title := "Test Project via API"
	description := "This is a test project created via the GraphQL API"
	isPublic := true

	// Create the input for the mutation
	input := CreateProjectV2Input{
		OwnerID:     githubv4.ID(ownerID),
		Title:       githubv4.String(title),
		Description: githubv4.String(description),
		Public:      githubv4.Boolean(isPublic),
	}

	// Convert to JSON for debugging
	inputJSON, _ := json.MarshalIndent(input, "", "  ")
	fmt.Printf("Mutation input:\n%s\n", string(inputJSON))

	// Define the mutation
	var m struct {
		CreateProjectV2 struct {
			ProjectV2 struct {
				ID          string
				Title       string
				Description string `graphql:"shortDescription"`
				Public      bool
			}
		} `graphql:"createProjectV2(input: $input)"`
	}

	// Define variables
	variables := map[string]interface{}{
		"input": input,
	}

	// Execute the mutation
	fmt.Println("Executing create project mutation...")
	err = client.Mutate(context.Background(), &m, input, variables)
	if err != nil {
		fmt.Printf("Error creating project: %s\n", err)
		// Try to get more details from the error
		if respErr, ok := err.(*githubv4.ResponseError); ok {
			fmt.Printf("GraphQL response errors:\n")
			for i, gqlErr := range respErr.Errors {
				fmt.Printf("  %d: %s\n", i+1, gqlErr.Message)
			}
		}
		return
	}

	// Print the result
	result, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling result: %s\n", err)
		return
	}

	fmt.Printf("Project created successfully:\n%s\n", string(result))
} 