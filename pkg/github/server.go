package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v69/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/shurcooL/githubv4"
)

type GetClientFn func(context.Context) (*github.Client, *githubv4.Client, error)

// RateLimitError represents a GitHub API rate limit error
type RateLimitError struct {
	Reset time.Time
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("GitHub API rate limit exceeded. Reset at %v", e.Reset)
}

// handleRateLimit checks the rate limit from the response and handles it appropriately
func handleRateLimit(resp *github.Response) error {
	if resp == nil {
		return nil
	}

	// Check if we've hit the rate limit
	if resp.Rate.Remaining == 0 {
		return &RateLimitError{
			Reset: resp.Rate.Reset.Time,
		}
	}

	// If we're getting close to the rate limit (less than 10% remaining), log a warning
	if float64(resp.Rate.Remaining)/float64(resp.Rate.Limit) < 0.1 {
		// You might want to log this warning or handle it in some way
		fmt.Printf("Warning: GitHub API rate limit is low. %d/%d requests remaining. Reset at %v\n",
			resp.Rate.Remaining, resp.Rate.Limit, resp.Rate.Reset.Time)
	}

	return nil
}

// withRateLimitRetry wraps a GitHub API call with rate limit handling and retry logic
func withRateLimitRetry(ctx context.Context, maxRetries int, fn func() (*github.Response, error)) error {
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		resp, err := fn()
		if err != nil {
			var rateLimitErr *github.RateLimitError
			if errors.As(err, &rateLimitErr) {
				if i == maxRetries {
					return fmt.Errorf("max retries exceeded waiting for rate limit: %w", err)
				}
				
				// Calculate sleep duration (with exponential backoff)
				sleepDuration := time.Until(rateLimitErr.Rate.Reset.Time)
				if sleepDuration < 0 {
					sleepDuration = time.Second * time.Duration(1<<uint(i))
				}
				
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(sleepDuration):
					continue
				}
			}
			lastErr = err
			break
		}
		
		if err := handleRateLimit(resp); err != nil {
			var rateLimitErr *RateLimitError
			if errors.As(err, &rateLimitErr) {
				if i == maxRetries {
					return fmt.Errorf("max retries exceeded waiting for rate limit reset: %w", err)
				}
				
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Until(rateLimitErr.Reset)):
					continue
				}
			}
			lastErr = err
			break
		}
		
		return nil
	}
	
	return lastErr
}

// NewServer creates a new GitHub MCP server with the specified GH client and logger.
func NewServer(getClient GetClientFn, version string, readOnly bool, t translations.TranslationHelperFunc) *server.MCPServer {
	// Create a new MCP server
	s := server.NewMCPServer(
		"github-mcp-server",
		version,
		server.WithResourceCapabilities(true, true),
		server.WithLogging())

	// Add GitHub Resources
	s.AddResourceTemplate(GetRepositoryResourceContent(getClient, t))
	s.AddResourceTemplate(GetRepositoryResourceBranchContent(getClient, t))
	s.AddResourceTemplate(GetRepositoryResourceCommitContent(getClient, t))
	s.AddResourceTemplate(GetRepositoryResourceTagContent(getClient, t))
	s.AddResourceTemplate(GetRepositoryResourcePrContent(getClient, t))

	// Add GitHub tools - Issues
	s.AddTool(GetIssue(getClient, t))
	s.AddTool(SearchIssues(getClient, t))
	s.AddTool(ListIssues(getClient, t))
	s.AddTool(GetIssueComments(getClient, t))
	if !readOnly {
		s.AddTool(CreateIssue(getClient, t))
		s.AddTool(AddIssueComment(getClient, t))
		s.AddTool(UpdateIssue(getClient, t))
	}

	// Add GitHub tools - Pull Requests
	s.AddTool(GetPullRequest(getClient, t))
	s.AddTool(ListPullRequests(getClient, t))
	s.AddTool(GetPullRequestFiles(getClient, t))
	s.AddTool(GetPullRequestStatus(getClient, t))
	s.AddTool(GetPullRequestComments(getClient, t))
	s.AddTool(GetPullRequestReviews(getClient, t))
	if !readOnly {
		s.AddTool(MergePullRequest(getClient, t))
		s.AddTool(UpdatePullRequestBranch(getClient, t))
		s.AddTool(CreatePullRequestReview(getClient, t))
		s.AddTool(CreatePullRequest(getClient, t))
		s.AddTool(UpdatePullRequest(getClient, t))
	}

	// Add GitHub tools - Projects
	s.AddTool(GetProjectV2(getClient, t))
	if !readOnly {
		s.AddTool(CreateProjectV2(getClient, t))
		s.AddTool(AddProjectV2Item(getClient, t))
		s.AddTool(UpdateProjectV2Item(getClient, t))
		s.AddTool(DeleteProjectV2Item(getClient, t))
	}

	// Add GitHub tools - Repositories
	s.AddTool(SearchRepositories(getClient, t))
	s.AddTool(GetFileContents(getClient, t))
	s.AddTool(ListCommits(getClient, t))
	if !readOnly {
		s.AddTool(CreateOrUpdateFile(getClient, t))
		s.AddTool(CreateRepository(getClient, t))
		s.AddTool(ForkRepository(getClient, t))
		s.AddTool(CreateBranch(getClient, t))
		s.AddTool(PushFiles(getClient, t))
	}

	// Add GitHub tools - Search
	s.AddTool(SearchCode(getClient, t))
	s.AddTool(SearchUsers(getClient, t))

	// Add GitHub tools - Users
	s.AddTool(GetMe(getClient, t))

	// Add GitHub tools - Code Scanning
	s.AddTool(GetCodeScanningAlert(getClient, t))
	s.AddTool(ListCodeScanningAlerts(getClient, t))
	return s
}

// GetMe creates a tool to get details of the authenticated user.
func GetMe(getClient GetClientFn, t translations.TranslationHelperFunc) (tool mcp.Tool, handler server.ToolHandlerFunc) {
	return mcp.NewTool("get_me",
			mcp.WithDescription(t("TOOL_GET_ME_DESCRIPTION", "Get details of the authenticated GitHub user. Use this when a request include \"me\", \"my\"...")),
			mcp.WithString("reason",
				mcp.Description("Optional: reason the session was created"),
			),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			client, _, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			user, resp, err := client.Users.Get(ctx, "")
			if err != nil {
				return nil, fmt.Errorf("failed to get user: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return mcp.NewToolResultError(fmt.Sprintf("failed to get user: %s", string(body))), nil
			}

			r, err := json.Marshal(user)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal user: %w", err)
			}

			return mcp.NewToolResultText(string(r)), nil
		}
}

// OptionalParamOK is a helper function that can be used to fetch a requested parameter from the request.
// It returns the value, a boolean indicating if the parameter was present, and an error if the type is wrong.
func OptionalParamOK[T any](r mcp.CallToolRequest, p string) (value T, ok bool, err error) {
	// Check if the parameter is present in the request
	val, exists := r.Params.Arguments[p]
	if !exists {
		// Not present, return zero value, false, no error
		return
	}

	// Check if the parameter is of the expected type
	value, ok = val.(T)
	if !ok {
		// Present but wrong type
		err = fmt.Errorf("parameter %s is not of type %T, is %T", p, value, val)
		ok = true // Set ok to true because the parameter *was* present, even if wrong type
		return
	}

	// Present and correct type
	ok = true
	return
}

// isAcceptedError checks if the error is an accepted error.
func isAcceptedError(err error) bool {
	var acceptedError *github.AcceptedError
	return errors.As(err, &acceptedError)
}

// requiredParam is a helper function that can be used to fetch a requested parameter from the request.
// It does the following checks:
// 1. Checks if the parameter is present in the request.
// 2. Checks if the parameter is of the expected type.
// 3. Checks if the parameter is not empty, i.e: non-zero value
func requiredParam[T comparable](r mcp.CallToolRequest, p string) (T, error) {
	var zero T

	// Check if the parameter is present in the request
	if _, ok := r.Params.Arguments[p]; !ok {
		return zero, fmt.Errorf("missing required parameter: %s", p)
	}

	// Check if the parameter is of the expected type
	if _, ok := r.Params.Arguments[p].(T); !ok {
		return zero, fmt.Errorf("parameter %s is not of type %T", p, zero)
	}

	if r.Params.Arguments[p].(T) == zero {
		return zero, fmt.Errorf("missing required parameter: %s", p)

	}

	return r.Params.Arguments[p].(T), nil
}

// requiredInt is a helper function that can be used to fetch a requested parameter from the request.
// It does the following checks:
// 1. Checks if the parameter is present in the request.
// 2. Checks if the parameter is of the expected type.
// 3. Checks if the parameter is not empty, i.e: non-zero value
func requiredInt(r mcp.CallToolRequest, p string) (int, error) {
	// Check if the parameter is present in the request
	if _, ok := r.Params.Arguments[p]; !ok {
		return 0, fmt.Errorf("missing required parameter: %s", p)
	}

	// Convert parameter to int based on its type
	switch val := r.Params.Arguments[p].(type) {
	case float64:
		return int(val), nil
	case int:
		return val, nil
	case int64:
		return int(val), nil
	case json.Number:
		i, err := val.Int64()
		if err != nil {
			return 0, fmt.Errorf("failed to parse %s as int: %v", p, err)
		}
		return int(i), nil
	default:
		return 0, fmt.Errorf("invalid type for %s, expected number, got %T", p, r.Params.Arguments[p])
	}
}

// OptionalParam is a helper function that can be used to fetch a requested parameter from the request.
// It does the following checks:
// 1. Checks if the parameter is present in the request, if not, it returns its zero-value
// 2. If it is present, it checks if the parameter is of the expected type and returns it
func OptionalParam[T any](r mcp.CallToolRequest, p string) (T, error) {
	var zero T

	// Check if the parameter is present in the request
	if _, ok := r.Params.Arguments[p]; !ok {
		return zero, nil
	}

	// Check if the parameter is of the expected type
	if _, ok := r.Params.Arguments[p].(T); !ok {
		return zero, fmt.Errorf("parameter %s is not of type %T, is %T", p, zero, r.Params.Arguments[p])
	}

	return r.Params.Arguments[p].(T), nil
}

// OptionalIntParam is a helper function that can be used to fetch a requested parameter from the request.
// It does the following checks:
// 1. Checks if the parameter is present in the request, if not, it returns its zero-value
// 2. If it is present, it checks if the parameter is of the expected type and returns it
func OptionalIntParam(r mcp.CallToolRequest, p string) (int, error) {
	v, err := OptionalParam[float64](r, p)
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

// OptionalIntParamWithDefault is a helper function that can be used to fetch a requested parameter from the request
// similar to optionalIntParam, but it also takes a default value.
func OptionalIntParamWithDefault(r mcp.CallToolRequest, p string, d int) (int, error) {
	v, err := OptionalIntParam(r, p)
	if err != nil {
		return 0, err
	}
	if v == 0 {
		return d, nil
	}
	return v, nil
}

// OptionalStringArrayParam is a helper function that can be used to fetch a requested parameter from the request.
// It does the following checks:
// 1. Checks if the parameter is present in the request, if not, it returns its zero-value
// 2. If it is present, iterates the elements and checks each is a string
func OptionalStringArrayParam(r mcp.CallToolRequest, p string) ([]string, error) {
	// Check if the parameter is present in the request
	if _, ok := r.Params.Arguments[p]; !ok {
		return []string{}, nil
	}

	switch v := r.Params.Arguments[p].(type) {
	case nil:
		return []string{}, nil
	case []string:
		return v, nil
	case []any:
		strSlice := make([]string, len(v))
		for i, v := range v {
			s, ok := v.(string)
			if !ok {
				return []string{}, fmt.Errorf("parameter %s is not of type string, is %T", p, v)
			}
			strSlice[i] = s
		}
		return strSlice, nil
	default:
		return []string{}, fmt.Errorf("parameter %s could not be coerced to []string, is %T", p, r.Params.Arguments[p])
	}
}

// WithPagination returns a ToolOption that adds "page" and "perPage" parameters to the tool.
// The "page" parameter is optional, min 1. The "perPage" parameter is optional, min 1, max 100.
func WithPagination() mcp.ToolOption {
	return func(tool *mcp.Tool) {
		mcp.WithNumber("page",
			mcp.Description("Page number for pagination (min 1)"),
			mcp.Min(1),
		)(tool)

		mcp.WithNumber("perPage",
			mcp.Description("Results per page for pagination (min 1, max 100)"),
			mcp.Min(1),
			mcp.Max(100),
		)(tool)
	}
}

type PaginationParams struct {
	page    int
	perPage int
}

// OptionalPaginationParams returns the "page" and "perPage" parameters from the request,
// or their default values if not present, "page" default is 1, "perPage" default is 30.
// In future, we may want to make the default values configurable, or even have this
// function returned from `withPagination`, where the defaults are provided alongside
// the min/max values.
func OptionalPaginationParams(r mcp.CallToolRequest) (PaginationParams, error) {
	page, err := OptionalIntParamWithDefault(r, "page", 1)
	if err != nil {
		return PaginationParams{}, err
	}
	perPage, err := OptionalIntParamWithDefault(r, "perPage", 30)
	if err != nil {
		return PaginationParams{}, err
	}
	return PaginationParams{
		page:    page,
		perPage: perPage,
	}, nil
}
