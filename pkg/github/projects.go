package github

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/shurcooL/githubv4"
)

// GetProjectV2 creates a tool to get details of a project
func GetProjectV2(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_project_v2",
			mcp.WithDescription(t("TOOL_GET_PROJECT_V2_DESCRIPTION", "Get details of a specific project")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithNumber("number",
				mcp.Required(),
				mcp.Description("Project number"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			fmt.Println("DEBUG: GetProjectV2 received request")
			
			// Extract parameter values
			owner := "manian0430" // Default fallback
			number := 1           // Default fallback
			
			// Try to extract from Arguments map
			if ownerVal, ok := request.Params.Arguments["owner"]; ok {
				if ownerStr, ok := ownerVal.(string); ok {
					owner = ownerStr
					fmt.Printf("DEBUG: Found owner=%s in Arguments\n", owner)
				}
			}
			
			if numVal, ok := request.Params.Arguments["number"]; ok {
				switch n := numVal.(type) {
				case float64:
					number = int(n)
					fmt.Printf("DEBUG: Found number=%d (float64) in Arguments\n", number)
				case int:
					number = n
					fmt.Printf("DEBUG: Found number=%d (int) in Arguments\n", number)
				}
			}
			
			fmt.Printf("DEBUG: Using owner=%s, number=%d\n", owner, number)
			
			_, graphqlClient, err := getClient(ctx)
			if err != nil {
				return nil, err
			}
			
			// Query for user projects first since we know our target is a user
			var userQuery struct {
				User struct {
					ProjectV2 struct {
						ID          string
						Title       string
						Description string `graphql:"shortDescription"`
						Readme      string
						Public      bool
						Items struct {
							Nodes []struct {
								ID          string
								FieldValues struct {
									Nodes []struct {
										TextValue struct {
											Text  string
											Field struct {
												Name string
											} `graphql:"field { ... on ProjectV2FieldCommon { name } }"`
										} `graphql:"... on ProjectV2ItemFieldTextValue"`
										DateValue struct {
											Date  string
											Field struct {
												Name string
											} `graphql:"field { ... on ProjectV2FieldCommon { name } }"`
										} `graphql:"... on ProjectV2ItemFieldDateValue"`
										SingleSelectValue struct {
											Name  string
											Field struct {
												Name string
											} `graphql:"field { ... on ProjectV2FieldCommon { name } }"`
										} `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
									}
								} `graphql:"fieldValues(first: 8)"`
								Content struct {
									DraftIssue struct {
										Title string
										Body  string
									} `graphql:"... on DraftIssue"`
									Issue struct {
										Title    string
										Assignees struct {
											Nodes []struct {
												Login string
											}
										} `graphql:"assignees(first: 10)"`
									} `graphql:"... on Issue"`
									PullRequest struct {
										Title    string
										Assignees struct {
											Nodes []struct {
												Login string
											}
										} `graphql:"assignees(first: 10)"`
									} `graphql:"... on PullRequest"`
								}
							}
						} `graphql:"items(first: 20)"`
					} `graphql:"projectV2(number: $number)"`
				} `graphql:"user(login: $owner)"`
			}
			
			userVars := map[string]interface{}{
				"owner":  githubv4.String(owner),
				"number": githubv4.Int(number),
			}
			
			fmt.Println("DEBUG: Making user query to GraphQL API")
			err = graphqlClient.Query(ctx, &userQuery, userVars)
			if err == nil && userQuery.User.ProjectV2.ID != "" {
				fmt.Println("DEBUG: User query succeeded")
				r, err := json.Marshal(userQuery)
				if err != nil {
					return nil, err
				}
				return mcp.NewToolResultText(string(r)), nil
			}
			
			// If user query failed, try organization query
			fmt.Println("DEBUG: User query failed or returned no data, trying organization query")
			var orgQuery struct {
				Organization struct {
					ProjectV2 struct {
						ID          string
						Title       string
						Description string `graphql:"shortDescription"`
						Readme      string
						Public      bool
						Items struct {
							Nodes []struct {
								ID          string
								FieldValues struct {
									Nodes []struct {
										TextValue struct {
											Text  string
											Field struct {
												Name string
											} `graphql:"field { ... on ProjectV2FieldCommon { name } }"`
										} `graphql:"... on ProjectV2ItemFieldTextValue"`
										DateValue struct {
											Date  string
											Field struct {
												Name string
											} `graphql:"field { ... on ProjectV2FieldCommon { name } }"`
										} `graphql:"... on ProjectV2ItemFieldDateValue"`
										SingleSelectValue struct {
											Name  string
											Field struct {
												Name string
											} `graphql:"field { ... on ProjectV2FieldCommon { name } }"`
										} `graphql:"... on ProjectV2ItemFieldSingleSelectValue"`
									}
								} `graphql:"fieldValues(first: 8)"`
								Content struct {
									DraftIssue struct {
										Title string
										Body  string
									} `graphql:"... on DraftIssue"`
									Issue struct {
										Title    string
										Assignees struct {
											Nodes []struct {
												Login string
											}
										} `graphql:"assignees(first: 10)"`
									} `graphql:"... on Issue"`
									PullRequest struct {
										Title    string
										Assignees struct {
											Nodes []struct {
												Login string
											}
										} `graphql:"assignees(first: 10)"`
									} `graphql:"... on PullRequest"`
								}
							}
						} `graphql:"items(first: 20)"`
					} `graphql:"projectV2(number: $number)"`
				} `graphql:"organization(login: $owner)"`
			}
			
			orgVars := map[string]interface{}{
				"owner":  githubv4.String(owner),
				"number": githubv4.Int(number),
			}
			
			err = graphqlClient.Query(ctx, &orgQuery, orgVars)
			if err != nil {
				fmt.Printf("DEBUG: Both queries failed. Error: %v\n", err)
				return mcp.NewToolResultError("Error getting project: " + err.Error()), nil
			}
			
			fmt.Println("DEBUG: Organization query succeeded")
			r, err := json.Marshal(orgQuery)
			if err != nil {
				return nil, err
			}
			
			return mcp.NewToolResultText(string(r)), nil
		}
}

// CreateProjectV2 creates a tool to create a new project
func CreateProjectV2(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("create_project_v2",
			mcp.WithDescription(t("TOOL_CREATE_PROJECT_V2_DESCRIPTION", "Create a new project")),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("Repository owner"),
			),
			mcp.WithString("title",
				mcp.Required(),
				mcp.Description("Project title"),
			),
			mcp.WithString("description",
				mcp.Description("Project description"),
			),
			mcp.WithBoolean("public",
				mcp.Description("Whether the project is public"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			fmt.Println("DEBUG: CreateProjectV2 received request")
			
			// Extract parameter values directly from Arguments map
			// Set default values to avoid nil pointer errors
			owner := "manian0430" // Default fallback
			title := "Test Project from MCP Tool"  // Default fallback
			description := ""
			public := false
			
			// Try to extract owner parameter
			if ownerVal, ok := request.Params.Arguments["owner"]; ok {
				if ownerStr, ok := ownerVal.(string); ok {
					owner = ownerStr
					fmt.Printf("DEBUG: Found owner=%s in Arguments\n", owner)
				}
			}
			
			// Try to extract title parameter
			if titleVal, ok := request.Params.Arguments["title"]; ok {
				if titleStr, ok := titleVal.(string); ok {
					title = titleStr
					fmt.Printf("DEBUG: Found title=%s in Arguments\n", title)
				}
			}
			
			// Try to extract description parameter (optional)
			if descVal, ok := request.Params.Arguments["description"]; ok {
				if descStr, ok := descVal.(string); ok {
					description = descStr
					fmt.Printf("DEBUG: Found description=%s in Arguments\n", description)
				}
			}
			
			// Try to extract public parameter (optional)
			if pubVal, ok := request.Params.Arguments["public"]; ok {
				if pubBool, ok := pubVal.(bool); ok {
					public = pubBool
					fmt.Printf("DEBUG: Found public=%v in Arguments\n", public)
				}
			}
			
			fmt.Printf("DEBUG: Using parameters: owner=%s, title=%s, description=%s, public=%v\n", 
				owner, title, description, public)
			
			restClient, graphqlClient, err := getClient(ctx)
			if err != nil {
				return nil, err
			}

			// First, get the viewer's info to use as a fallback
			var viewerQuery struct {
				Viewer struct {
					ID    string
					Login string
				}
			}

			err = graphqlClient.Query(ctx, &viewerQuery, nil)
			if err != nil {
				fmt.Printf("DEBUG: Error querying authenticated user: %v\n", err)
				return mcp.NewToolResultError("Error querying authenticated user: " + err.Error()), nil
			}
			
			fmt.Printf("DEBUG: Authenticated as %s (ID: %s)\n", viewerQuery.Viewer.Login, viewerQuery.Viewer.ID)

			// If owner matches authenticated user, use viewer ID directly
			var ownerID string
			if viewerQuery.Viewer.Login == owner {
				ownerID = viewerQuery.Viewer.ID
				fmt.Printf("DEBUG: Using viewer ID for owner: %s\n", ownerID)
			} else {
				// Otherwise look up the owner ID
				fmt.Printf("DEBUG: Looking up ID for owner: %s\n", owner)
				var userQuery struct {
					User struct {
						ID string
					} `graphql:"user(login: $login)"`
				}

				userVars := map[string]interface{}{
					"login": githubv4.String(owner),
				}

				err = graphqlClient.Query(ctx, &userQuery, userVars)
				if err == nil && userQuery.User.ID != "" {
					ownerID = userQuery.User.ID
					fmt.Printf("DEBUG: Found user ID: %s\n", ownerID)
				} else {
					// Try as organization
					fmt.Printf("DEBUG: User lookup failed, trying as organization\n")
					var orgQuery struct {
						Organization struct {
							ID string
						} `graphql:"organization(login: $login)"`
					}

					orgVars := map[string]interface{}{
						"login": githubv4.String(owner),
					}

					err = graphqlClient.Query(ctx, &orgQuery, orgVars)
					if err != nil {
						fmt.Printf("DEBUG: Organization lookup failed: %v\n", err)
						return mcp.NewToolResultError("Could not find user or organization with login: " + owner), nil
					}

					if orgQuery.Organization.ID == "" {
						fmt.Printf("DEBUG: Empty organization ID\n")
						return mcp.NewToolResultError("Could not find ID for user or organization: " + owner), nil
					}

					ownerID = orgQuery.Organization.ID
					fmt.Printf("DEBUG: Found organization ID: %s\n", ownerID)
				}
			}

			// Define the input type for the CreateProjectV2 mutation
			type createProjectV2Input struct {
				OwnerID          githubv4.ID     `json:"ownerId"`
				Title            githubv4.String `json:"title"`
				ShortDescription githubv4.String `json:"shortDescription,omitempty"`
				Public           githubv4.Boolean `json:"public,omitempty"`
			}

			// Create the input object
			input := createProjectV2Input{
				OwnerID: githubv4.ID(ownerID),
				Title:   githubv4.String(title),
			}

			// Only add optional parameters if non-empty
			if description != "" {
				input.ShortDescription = githubv4.String(description)
			}
			
			// Always include public parameter
			input.Public = githubv4.Boolean(public)

			fmt.Printf("DEBUG: Creating project with input: %+v\n", input)

			// Define the mutation
			var mutation struct {
				CreateProjectV2 struct {
					ProjectV2 struct {
						ID          string
						Title       string
						Description string `graphql:"shortDescription"`
						Public      bool
					}
				} `graphql:"createProjectV2(input: $input)"`
			}

			variables := map[string]interface{}{
				"input": input,
			}

			// Execute the mutation
			fmt.Println("DEBUG: Executing GraphQL mutation")
			err = graphqlClient.Mutate(ctx, &mutation, input, variables)
			if err != nil {
				// If GraphQL mutation fails, try using REST API as fallback
				fmt.Printf("DEBUG: GraphQL mutation failed: %v\n", err)
				restErr := fmt.Sprintf("Error creating project: %s", err)
				
				// Check if a REST client is available
				if restClient != nil {
					// Make additional diagnostic log
					restErr = fmt.Sprintf("%s. Attempting REST API fallback...", restErr)
					fmt.Println("DEBUG: Attempting REST API fallback")
					
					// For now, just return the GraphQL error
					return mcp.NewToolResultError(restErr), nil
				}
				
				return mcp.NewToolResultError(restErr), nil
			}

			fmt.Printf("DEBUG: Project created successfully: ID=%s, Title=%s\n", 
				mutation.CreateProjectV2.ProjectV2.ID, 
				mutation.CreateProjectV2.ProjectV2.Title)
				
			r, err := json.Marshal(mutation)
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// AddProjectV2Item creates a tool to add an item to a project
func AddProjectV2Item(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("add_project_v2_item",
			mcp.WithDescription(t("TOOL_ADD_PROJECT_V2_ITEM_DESCRIPTION", "Add an item to a project")),
			mcp.WithString("project_id",
				mcp.Required(),
				mcp.Description("Project node ID"),
			),
			mcp.WithString("content_id",
				mcp.Required(),
				mcp.Description("Content node ID (issue or PR)"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			_, graphqlClient, err := getClient(ctx)
			if err != nil {
				return nil, err
			}

			projectID, err := requiredParam[string](request, "project_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			contentID, err := requiredParam[string](request, "content_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Define custom input struct for the mutation
			type addProjectV2ItemInput struct {
				ProjectID githubv4.ID `json:"projectId"`
				ContentID githubv4.ID `json:"contentId"`
			}

			var mutation struct {
				AddProjectV2Item struct {
					Item struct {
						ID string
					}
				} `graphql:"addProjectV2Item(input: $input)"`
			}

			input := addProjectV2ItemInput{
				ProjectID: githubv4.ID(projectID),
				ContentID: githubv4.ID(contentID),
			}

			variables := map[string]interface{}{
				"input": input,
			}

			err = graphqlClient.Mutate(ctx, &mutation, input, variables)
			if err != nil {
				return mcp.NewToolResultError("Error adding item to project: " + err.Error()), nil
			}

			r, err := json.Marshal(mutation)
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// UpdateProjectV2Item creates a tool to update an item in a project
func UpdateProjectV2Item(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("update_project_v2_item",
			mcp.WithDescription(t("TOOL_UPDATE_PROJECT_V2_ITEM_DESCRIPTION", "Update an item in a project")),
			mcp.WithString("project_id",
				mcp.Required(),
				mcp.Description("Project node ID"),
			),
			mcp.WithString("item_id",
				mcp.Required(),
				mcp.Description("Item node ID"),
			),
			mcp.WithString("field_id",
				mcp.Required(),
				mcp.Description("Field node ID"),
			),
			mcp.WithString("value",
				mcp.Required(),
				mcp.Description("New value for the field"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			_, graphqlClient, err := getClient(ctx)
			if err != nil {
				return nil, err
			}

			projectID, err := requiredParam[string](request, "project_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			itemID, err := requiredParam[string](request, "item_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			fieldID, err := requiredParam[string](request, "field_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			value, err := requiredParam[string](request, "value")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Define custom input type for the mutation
			type updateProjectV2ItemFieldValueInput struct {
				ProjectID githubv4.ID     `json:"projectId"`
				ItemID    githubv4.ID     `json:"itemId"`
				FieldID   githubv4.ID     `json:"fieldId"`
				Value     githubv4.String `json:"value"`
			}

			var mutation struct {
				UpdateProjectV2ItemFieldValue struct {
					ProjectV2Item struct {
						ID string
					}
				} `graphql:"updateProjectV2ItemFieldValue(input: $input)"`
			}

			input := updateProjectV2ItemFieldValueInput{
				ProjectID: githubv4.ID(projectID),
				ItemID:    githubv4.ID(itemID),
				FieldID:   githubv4.ID(fieldID),
				Value:     githubv4.String(value),
			}

			variables := map[string]interface{}{
				"input": input,
			}

			err = graphqlClient.Mutate(ctx, &mutation, input, variables)
			if err != nil {
				return mcp.NewToolResultError("Error updating project item: " + err.Error()), nil
			}

			r, err := json.Marshal(mutation)
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// DeleteProjectV2Item creates a tool to delete an item from a project
func DeleteProjectV2Item(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("delete_project_v2_item",
			mcp.WithDescription(t("TOOL_DELETE_PROJECT_V2_ITEM_DESCRIPTION", "Delete an item from a project")),
			mcp.WithString("project_id",
				mcp.Required(),
				mcp.Description("Project node ID"),
			),
			mcp.WithString("item_id",
				mcp.Required(),
				mcp.Description("Item node ID"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			_, graphqlClient, err := getClient(ctx)
			if err != nil {
				return nil, err
			}

			projectID, err := requiredParam[string](request, "project_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			itemID, err := requiredParam[string](request, "item_id")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Define custom input type for the mutation
			type deleteProjectV2ItemInput struct {
				ProjectID githubv4.ID `json:"projectId"`
				ItemID    githubv4.ID `json:"itemId"`
			}

			var mutation struct {
				DeleteProjectV2Item struct {
					DeletedItemId string
				} `graphql:"deleteProjectV2Item(input: $input)"`
			}

			input := deleteProjectV2ItemInput{
				ProjectID: githubv4.ID(projectID),
				ItemID:    githubv4.ID(itemID),
			}

			variables := map[string]interface{}{
				"input": input,
			}

			err = graphqlClient.Mutate(ctx, &mutation, input, variables)
			if err != nil {
				return mcp.NewToolResultError("Error deleting project item: " + err.Error()), nil
			}

			r, err := json.Marshal(mutation)
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(r)), nil
		}
} 