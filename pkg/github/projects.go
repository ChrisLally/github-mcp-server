package github

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v69/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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
			client, err := getClient(ctx)
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

			query := fmt.Sprintf(`
				query {
					organization(login: "%s") {
						projectV2(number: %d) {
							id
							title
							shortDescription
							readme
							public
							items(first: 20) {
								nodes {
									id
									fieldValues(first: 8) {
										nodes {
											... on ProjectV2ItemFieldTextValue {
												text
												field {
													... on ProjectV2FieldCommon {
														name
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			`, owner, number)

			req, err := client.NewRequest("POST", "graphql", map[string]interface{}{
				"query": query,
			})
			if err != nil {
				return nil, err
			}

			var response map[string]interface{}
			_, err = client.Do(ctx, req, &response)
			if err != nil {
				return nil, err
			}

			r, err := json.Marshal(response)
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
			client, err := getClient(ctx)
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

			description, err := OptionalParam[string](request, "description")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			public, err := OptionalParam[bool](request, "public")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			mutation := fmt.Sprintf(`
				mutation {
					createProjectV2(
						input: {
							ownerId: "%s",
							title: "%s",
							description: "%s",
							public: %t
						}
					) {
						projectV2 {
							id
							title
							shortDescription
							public
						}
					}
				}
			`, owner, title, description, public)

			req, err := client.NewRequest("POST", "graphql", map[string]interface{}{
				"query": mutation,
			})
			if err != nil {
				return nil, err
			}

			var response map[string]interface{}
			_, err = client.Do(ctx, req, &response)
			if err != nil {
				return nil, err
			}

			r, err := json.Marshal(response)
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
			client, err := getClient(ctx)
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

			mutation := fmt.Sprintf(`
				mutation {
					addProjectV2ItemById(
						input: {
							projectId: "%s"
							contentId: "%s"
						}
					) {
						item {
							id
						}
					}
				}
			`, projectID, contentID)

			req, err := client.NewRequest("POST", "graphql", map[string]interface{}{
				"query": mutation,
			})
			if err != nil {
				return nil, err
			}

			var response map[string]interface{}
			_, err = client.Do(ctx, req, &response)
			if err != nil {
				return nil, err
			}

			r, err := json.Marshal(response)
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
			mcp.WithObject("value",
				mcp.Required(),
				mcp.Description("New field value"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, err := getClient(ctx)
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
			value, err := requiredParam[map[string]interface{}](request, "value")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			mutation := fmt.Sprintf(`
				mutation {
					updateProjectV2ItemFieldValue(
						input: {
							projectId: "%s"
							itemId: "%s"
							fieldId: "%s"
							value: %v
						}
					) {
						projectV2Item {
							id
						}
					}
				}
			`, projectID, itemID, fieldID, value)

			req, err := client.NewRequest("POST", "graphql", map[string]interface{}{
				"query": mutation,
			})
			if err != nil {
				return nil, err
			}

			var response map[string]interface{}
			_, err = client.Do(ctx, req, &response)
			if err != nil {
				return nil, err
			}

			r, err := json.Marshal(response)
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
			client, err := getClient(ctx)
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

			mutation := fmt.Sprintf(`
				mutation {
					deleteProjectV2Item(
						input: {
							projectId: "%s"
							itemId: "%s"
						}
					) {
						deletedItemId
					}
				}
			`, projectID, itemID)

			req, err := client.NewRequest("POST", "graphql", map[string]interface{}{
				"query": mutation,
			})
			if err != nil {
				return nil, err
			}

			var response map[string]interface{}
			_, err = client.Do(ctx, req, &response)
			if err != nil {
				return nil, err
			}

			r, err := json.Marshal(response)
			if err != nil {
				return nil, err
			}

			return mcp.NewToolResultText(string(r)), nil
		}
} 