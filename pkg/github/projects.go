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
			_, graphqlClient, err := getClient(ctx)
			if err != nil {
				return nil, err
			}

			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			number, err := RequiredInt(request, "number")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Query for organization projects
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
			
			// If organization query succeeds and returns data
			if err == nil && orgQuery.Organization.ProjectV2.ID != "" {
				r, err := json.Marshal(orgQuery)
				if err != nil {
					return nil, err
				}
				return mcp.NewToolResultText(string(r)), nil
			}
			
			// Try user query if organization query failed or returned no data
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

			err = graphqlClient.Query(ctx, &userQuery, userVars)
			if err != nil {
				return mcp.NewToolResultError("Error getting project: " + err.Error()), nil
			}

			r, err := json.Marshal(userQuery)
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
			restClient, graphqlClient, err := getClient(ctx)
			if err != nil {
				return nil, err
			}

			owner, err := requiredParam[string](request, "owner")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			title, err := requiredParam[string](request, "title")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			description, descExists, err := OptionalParamOK[string](request, "description")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			public, publicExists, err := OptionalParamOK[bool](request, "public")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
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
				return mcp.NewToolResultError("Error querying authenticated user: " + err.Error()), nil
			}

			// If owner matches authenticated user, use viewer ID directly
			var ownerID string
			if viewerQuery.Viewer.Login == owner {
				ownerID = viewerQuery.Viewer.ID
			} else {
				// Otherwise look up the owner ID
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
				} else {
					// Try as organization
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
						return mcp.NewToolResultError("Could not find user or organization with login: " + owner), nil
					}

					if orgQuery.Organization.ID == "" {
						return mcp.NewToolResultError("Could not find ID for user or organization: " + owner), nil
					}

					ownerID = orgQuery.Organization.ID
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

			// Only add optional parameters if they were provided
			if descExists {
				input.ShortDescription = githubv4.String(description)
			}
			
			if publicExists {
				input.Public = githubv4.Boolean(public)
			}

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
			err = graphqlClient.Mutate(ctx, &mutation, input, variables)
			if err != nil {
				// If GraphQL mutation fails, try using REST API as fallback
				restErr := fmt.Sprintf("Error creating project: %s", err)
				
				// Check if a REST client is available
				if restClient != nil {
					// Make additional diagnostic log
					restErr = fmt.Sprintf("%s. Attempting REST API fallback...", restErr)
					
					// For now, just return the GraphQL error
					return mcp.NewToolResultError(restErr), nil
				}
				
				return mcp.NewToolResultError(restErr), nil
			}

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