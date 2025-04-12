package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/v57/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// DiagnoseTool creates a diagnostic tool to help understand issues with the MCP server
func DiagnoseTool(getClient GetClientFn) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("diagnose_github_mcp",
			mcp.WithDescription("Run diagnostics on the GitHub MCP server to debug issues"),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			// Start with basic diagnostics
			results := map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"request":   request,
			}

			// Check environment variables
			token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
			if token == "" {
				results["error"] = "GITHUB_PERSONAL_ACCESS_TOKEN not set"
				r, _ := json.MarshalIndent(results, "", "  ")
				return mcp.NewToolResultText(string(r)), nil
			}

			results["token_status"] = fmt.Sprintf("Token present (starts with %s..., length: %d)", token[:5], len(token))

			// Test token validity with GitHub API
			sts := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			)
			httpClient := oauth2.NewClient(context.Background(), sts)
			
			// Test REST API
			restClient := github.NewClient(httpClient)
			user, resp, err := restClient.Users.Get(ctx, "")
			if err != nil {
				results["rest_api_test"] = map[string]interface{}{
					"status":  "error",
					"error":   err.Error(),
					"headers": resp.Header,
				}
			} else {
				results["rest_api_test"] = map[string]interface{}{
					"status":    "success",
					"user_name": user.GetLogin(),
					"user_id":   user.GetID(),
				}
			}

			// Test GraphQL API
			gqlClient := githubv4.NewClient(httpClient)
			var query struct {
				Viewer struct {
					Login string
					ID    string
				}
			}
			err = gqlClient.Query(ctx, &query, nil)
			if err != nil {
				results["graphql_api_test"] = map[string]interface{}{
					"status": "error",
					"error":  err.Error(),
				}
			} else {
				results["graphql_api_test"] = map[string]interface{}{
					"status":    "success",
					"user_name": query.Viewer.Login,
					"user_id":   query.Viewer.ID,
				}
			}

			// Check token scopes with GitHub API
			req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
			req.Header.Add("Authorization", "token "+token)
			client := &http.Client{}
			resp, err = client.Do(req)
			if err != nil {
				results["token_scopes_test"] = map[string]interface{}{
					"status": "error",
					"error":  err.Error(),
				}
			} else {
				scopes := resp.Header.Get("X-OAuth-Scopes")
				results["token_scopes_test"] = map[string]interface{}{
					"status": "success",
					"scopes": scopes,
				}
			}

			// Format results as JSON
			r, _ := json.MarshalIndent(results, "", "  ")
			return mcp.NewToolResultText(string(r)), nil
		}
} 