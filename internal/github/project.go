package github

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/shurcooL/githubv4"
)

func (c *Client) CreateProjectV2(ctx context.Context, owner string, title string, description string, public bool) (*ProjectV2, error) {
	log.Printf("Attempting to create project. Owner: %s, Title: %s", owner, title)
	
	if owner == "" {
		return nil, fmt.Errorf("owner cannot be empty")
	}
	if title == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}

	var query struct {
		CreateProjectV2 struct {
			ProjectV2 ProjectV2
		} `graphql:"createProjectV2(input: $input)"`
	}

	input := CreateProjectV2Input{
		OwnerID:    owner,
		Title:      title,
		RepositoryID: nil,
	}
	if description != "" {
		input.Description = &description
	}
	input.Public = &public

	err := c.client.Mutate(ctx, &query, input, nil)
	if err != nil {
		log.Printf("Error creating project: %v", err)
		// Check for specific error types
		if strings.Contains(err.Error(), "NOT_FOUND") {
			return nil, fmt.Errorf("owner %s not found", owner)
		}
		if strings.Contains(err.Error(), "FORBIDDEN") {
			return nil, fmt.Errorf("insufficient permissions to create project for %s", owner)
		}
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	log.Printf("Successfully created project. ID: %s", query.CreateProjectV2.ProjectV2.ID)
	return &query.CreateProjectV2.ProjectV2, nil
}

func (c *Client) GetProjectV2(ctx context.Context, owner string, number int) (*ProjectV2, error) {
	log.Printf("Attempting to get project. Owner: %s, Number: %d", owner, number)
	
	if owner == "" {
		return nil, fmt.Errorf("owner cannot be empty")
	}
	if number <= 0 {
		return nil, fmt.Errorf("project number must be positive")
	}

	var query struct {
		User struct {
			ProjectV2 *ProjectV2 `graphql:"projectV2(number: $number)"`
		} `graphql:"user(login: $login)"`
	}

	variables := map[string]interface{}{
		"login":  githubv4.String(owner),
		"number": githubv4.Int(number),
	}

	err := c.client.Query(ctx, &query, variables)
	if err != nil {
		log.Printf("Error getting project: %v", err)
		if strings.Contains(err.Error(), "NOT_FOUND") {
			return nil, fmt.Errorf("project %d not found for owner %s", number, owner)
		}
		if strings.Contains(err.Error(), "FORBIDDEN") {
			return nil, fmt.Errorf("insufficient permissions to view project %d for %s", number, owner)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if query.User.ProjectV2 == nil {
		return nil, fmt.Errorf("project %d not found for user %s", number, owner)
	}

	log.Printf("Found project. ID: %s", query.User.ProjectV2.ID)
	return query.User.ProjectV2, nil
}

func (c *Client) DeleteProjectV2Item(ctx context.Context, projectID string, itemID string) error {
	log.Printf("Attempting to delete project item. Project ID: %s, Item ID: %s", projectID, itemID)
	
	if projectID == "" {
		return fmt.Errorf("project ID cannot be empty")
	}
	if itemID == "" {
		return fmt.Errorf("item ID cannot be empty")
	}

	var mutation struct {
		DeleteProjectV2Item struct {
			DeletedItemId string
		} `graphql:"deleteProjectV2Item(input: $input)"`
	}

	input := DeleteProjectV2ItemInput{
		ProjectID: projectID,
		ItemID:    itemID,
	}

	err := c.client.Mutate(ctx, &mutation, input, nil)
	if err != nil {
		log.Printf("Error deleting project item: %v", err)
		if strings.Contains(err.Error(), "NOT_FOUND") {
			return fmt.Errorf("project item not found. Project ID: %s, Item ID: %s", projectID, itemID)
		}
		if strings.Contains(err.Error(), "FORBIDDEN") {
			return fmt.Errorf("insufficient permissions to delete project item")
		}
		return fmt.Errorf("failed to delete project item: %w", err)
	}

	log.Printf("Successfully deleted project item. Deleted ID: %s", mutation.DeleteProjectV2Item.DeletedItemId)
	return nil
}

func (c *Client) AddProjectV2Item(ctx context.Context, projectID string, contentID string) error {
	log.Printf("Attempting to add item to project. Project ID: %s, Content ID: %s", projectID, contentID)
	
	if projectID == "" {
		return fmt.Errorf("project ID cannot be empty")
	}
	if contentID == "" {
		return fmt.Errorf("content ID cannot be empty")
	}

	var mutation struct {
		AddProjectV2Item struct {
			Item struct {
				ID string
			}
		} `graphql:"addProjectV2Item(input: $input)"`
	}

	input := AddProjectV2ItemInput{
		ProjectID: projectID,
		ContentID: contentID,
	}

	err := c.client.Mutate(ctx, &mutation, input, nil)
	if err != nil {
		log.Printf("Error adding project item: %v", err)
		if strings.Contains(err.Error(), "NOT_FOUND") {
			return fmt.Errorf("project or content not found. Project ID: %s, Content ID: %s", projectID, contentID)
		}
		if strings.Contains(err.Error(), "FORBIDDEN") {
			return fmt.Errorf("insufficient permissions to add item to project")
		}
		return fmt.Errorf("failed to add project item: %w", err)
	}

	log.Printf("Successfully added item to project. Item ID: %s", mutation.AddProjectV2Item.Item.ID)
	return nil
}

func (c *Client) UpdateProjectV2Item(ctx context.Context, projectID string, itemID string, fieldID string, value string) error {
	log.Printf("Attempting to update project item. Project ID: %s, Item ID: %s, Field ID: %s", projectID, itemID, fieldID)
	
	if projectID == "" {
		return fmt.Errorf("project ID cannot be empty")
	}
	if itemID == "" {
		return fmt.Errorf("item ID cannot be empty")
	}
	if fieldID == "" {
		return fmt.Errorf("field ID cannot be empty")
	}

	var mutation struct {
		UpdateProjectV2ItemFieldValue struct {
			ProjectV2Item struct {
				ID string
			}
		} `graphql:"updateProjectV2ItemFieldValue(input: $input)"`
	}

	input := UpdateProjectV2ItemFieldValueInput{
		ProjectID: projectID,
		ItemID:    itemID,
		FieldID:   fieldID,
		Value:     value,
	}

	err := c.client.Mutate(ctx, &mutation, input, nil)
	if err != nil {
		log.Printf("Error updating project item: %v", err)
		if strings.Contains(err.Error(), "NOT_FOUND") {
			return fmt.Errorf("project item or field not found. Project ID: %s, Item ID: %s, Field ID: %s", projectID, itemID, fieldID)
		}
		if strings.Contains(err.Error(), "FORBIDDEN") {
			return fmt.Errorf("insufficient permissions to update project item")
		}
		if strings.Contains(err.Error(), "INVALID") {
			return fmt.Errorf("invalid value for field. Value: %s", value)
		}
		return fmt.Errorf("failed to update project item: %w", err)
	}

	log.Printf("Successfully updated project item. Item ID: %s", mutation.UpdateProjectV2ItemFieldValue.ProjectV2Item.ID)
	return nil
} 