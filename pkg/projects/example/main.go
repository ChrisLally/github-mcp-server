package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/your-org/your-repo/pkg/projects"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	// Get GitHub token from environment variable
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required")
	}

	// Get organization name from environment variable
	org := os.Getenv("GITHUB_ORG")
	if org == "" {
		log.Fatal("GITHUB_ORG environment variable is required")
	}

	// Get project number from environment variable
	projectNumberStr := os.Getenv("PROJECT_NUMBER")
	if projectNumberStr == "" {
		log.Fatal("PROJECT_NUMBER environment variable is required")
	}
	projectNumber, err := strconv.Atoi(projectNumberStr)
	if err != nil {
		log.Fatalf("Invalid PROJECT_NUMBER: %v", err)
	}

	// Create a new client
	client := projects.NewClient(token)

	// Create a context
	ctx := context.Background()

	// Example 1: Find a project
	project, err := client.FindProjectByNumber(ctx, org, projectNumber)
	if err != nil {
		log.Fatalf("Failed to find project: %v", err)
	}
	fmt.Printf("Found project: %s (ID: %s)\n", project.Title, project.ID)

	// Example 2: Get project fields
	fields, err := client.GetProjectFields(ctx, project.ID)
	if err != nil {
		log.Fatalf("Failed to get project fields: %v", err)
	}
	fmt.Println("\nProject fields:")
	for _, field := range fields {
		fmt.Printf("- %s (Type: %s)\n", field.Name, field.Type)
	}

	// Example 3: Get project items
	items, err := client.GetProjectItems(ctx, project.ID)
	if err != nil {
		log.Fatalf("Failed to get project items: %v", err)
	}
	fmt.Println("\nProject items:")
	for _, item := range items {
		fmt.Printf("- Item ID: %s\n", item.ID)
		for _, fieldValue := range item.FieldValues {
			fmt.Printf("  - %s: %v\n", fieldValue.Name, fieldValue.Value)
		}
	}

	// Example 4: Create a draft issue
	draft, err := client.AddDraftIssue(ctx, project.ID, "New Task", "This is a new task description")
	if err != nil {
		log.Fatalf("Failed to create draft issue: %v", err)
	}
	fmt.Printf("\nCreated draft issue with ID: %s\n", draft.ID)

	// Example 5: Update a field value (assuming we have a field ID)
	if len(fields) > 0 {
		fieldID := fields[0].ID
		err = client.UpdateItemField(ctx, project.ID, draft.ID, fieldID, "Updated value")
		if err != nil {
			log.Fatalf("Failed to update field value: %v", err)
		}
		fmt.Printf("Updated field %s for item %s\n", fieldID, draft.ID)
	}

	// Example 6: Update project settings
	err = client.UpdateProjectSettings(ctx, project.ID, project.Title, true, 
		"# Project README\n\nUpdated project description", 
		"Short description of the project")
	if err != nil {
		log.Fatalf("Failed to update project settings: %v", err)
	}
	fmt.Println("\nUpdated project settings successfully")
} 