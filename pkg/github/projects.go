package github

import (
	"context"
	"encoding/json"
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

			var query struct {
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

			variables := map[string]interface{}{
				"owner":  githubv4.String(owner),
				"number": githubv4.Int(number),
			}

			err = graphqlClient.Query(ctx, &query, variables)
			if err != nil {
				return nil, err
			}

			r, err := json.Marshal(query)
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
			_, graphqlClient, err := getClient(ctx)
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

			description, _, err := OptionalParamOK[string](request, "description")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			public, _, err := OptionalParamOK[bool](request, "public")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

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

			input := struct {
				OwnerID     githubv4.String
				Title       githubv4.String
				Description githubv4.String
				Public      githubv4.Boolean
			}{
				OwnerID:     githubv4.String(owner),
				Title:       githubv4.String(title),
				Description: githubv4.String(description),
				Public:      githubv4.Boolean(public),
			}

			variables := map[string]interface{}{
				"input": input,
			}

			err = graphqlClient.Mutate(ctx, &mutation, input, variables)
			if err != nil {
				return nil, err
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

			var mutation struct {
				AddProjectV2Item struct {
					Item struct {
						ID string
					}
				} `graphql:"addProjectV2Item(input: $input)"`
			}

			input := struct {
				ProjectID githubv4.ID
				ContentID githubv4.ID
			}{
				ProjectID: githubv4.ID(projectID),
				ContentID: githubv4.ID(contentID),
			}

			variables := map[string]interface{}{
				"input": input,
			}

			err = graphqlClient.Mutate(ctx, &mutation, input, variables)
			if err != nil {
				return nil, err
			}

			r, err := json.Marshal(mutation)
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// UpdateProjectV2Item creates a tool to update a project item
func UpdateProjectV2Item(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("update_project_v2_item",
			mcp.WithDescription(t("TOOL_UPDATE_PROJECT_V2_ITEM_DESCRIPTION", "Update a project item's field value")),
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
				mcp.Description("New field value"),
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

			var mutation struct {
				UpdateProjectV2ItemFieldValue struct {
					ProjectV2Item struct {
						ID string
					}
				} `graphql:"updateProjectV2ItemFieldValue(input: $input)"`
			}

			input := struct {
				ProjectID githubv4.ID
				ItemID    githubv4.ID
				FieldID   githubv4.ID
				Value     string
			}{
				ProjectID: githubv4.ID(projectID),
				ItemID:    githubv4.ID(itemID),
				FieldID:   githubv4.ID(fieldID),
				Value:     value,
			}

			variables := map[string]interface{}{
				"input": input,
			}

			err = graphqlClient.Mutate(ctx, &mutation, input, variables)
			if err != nil {
				return nil, err
			}

			r, err := json.Marshal(mutation)
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// DeleteProjectV2Item creates a tool to delete a project item
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

			var mutation struct {
				DeleteProjectV2Item struct {
					DeletedItemId string
				} `graphql:"deleteProjectV2Item(input: $input)"`
			}

			input := struct {
				ProjectID githubv4.ID
				ItemID    githubv4.ID
			}{
				ProjectID: githubv4.ID(projectID),
				ItemID:    githubv4.ID(itemID),
			}

			variables := map[string]interface{}{
				"input": input,
			}

			err = graphqlClient.Mutate(ctx, &mutation, input, variables)
			if err != nil {
				return nil, err
			}

			r, err := json.Marshal(mutation)
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(r)), nil
		}
} 