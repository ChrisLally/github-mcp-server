package projects

import (
	"context"
	"fmt"
	"net/http"

	"github.com/shurcooL/graphql"
)

// Client represents a GitHub Projects v2 GraphQL client
type Client struct {
	client *graphql.Client
}

// NewClient creates a new GitHub Projects v2 GraphQL client
func NewClient(token string) *Client {
	httpClient := &http.Client{
		Transport: &transport{
			token: token,
		},
	}
	return &Client{
		client: graphql.NewClient("https://api.github.com/graphql", httpClient),
	}
}

// transport implements http.RoundTripper
type transport struct {
	token string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.token))
	return http.DefaultTransport.RoundTrip(req)
}

// FindProjectByNumber finds a project by organization and number
func (c *Client) FindProjectByNumber(ctx context.Context, org string, number int) (*Project, error) {
	var query struct {
		Organization struct {
			ProjectV2 struct {
				ID    string
				Title string
				Number int
			} `graphql:"projectV2(number: $number)"`
		} `graphql:"organization(login: $org)"`
	}

	variables := map[string]interface{}{
		"org":    graphql.String(org),
		"number": graphql.Int(number),
	}

	err := c.client.Query(ctx, &query, variables)
	if err != nil {
		return nil, err
	}

	return &Project{
		ID:     query.Organization.ProjectV2.ID,
		Title:  query.Organization.ProjectV2.Title,
		Number: query.Organization.ProjectV2.Number,
	}, nil
}

// AddItemToProject adds an item to a project
func (c *Client) AddItemToProject(ctx context.Context, projectID, contentID string) (*ProjectItem, error) {
	var mutation struct {
		AddProjectV2ItemById struct {
			Item struct {
				ID string
			}
		} `graphql:"addProjectV2ItemById(input: {projectId: $projectId, contentId: $contentId})"`
	}

	variables := map[string]interface{}{
		"projectId": graphql.ID(projectID),
		"contentId": graphql.ID(contentID),
	}

	err := c.client.Mutate(ctx, &mutation, variables)
	if err != nil {
		return nil, err
	}

	return &ProjectItem{
		ID: mutation.AddProjectV2ItemById.Item.ID,
	}, nil
}

// UpdateItemField updates a field value for a project item
func (c *Client) UpdateItemField(ctx context.Context, projectID, itemID, fieldID string, value interface{}) error {
	var mutation struct {
		UpdateProjectV2ItemFieldValue struct {
			ProjectV2Item struct {
				ID string
			}
		} `graphql:"updateProjectV2ItemFieldValue(input: {projectId: $projectId, itemId: $itemId, fieldId: $fieldId, value: $value})"`
	}

	variables := map[string]interface{}{
		"projectId": graphql.ID(projectID),
		"itemId":    graphql.ID(itemID),
		"fieldId":   graphql.ID(fieldID),
		"value":     value,
	}

	return c.client.Mutate(ctx, &mutation, variables)
}

// DeleteItemFromProject deletes an item from a project
func (c *Client) DeleteItemFromProject(ctx context.Context, projectID, itemID string) error {
	var mutation struct {
		DeleteProjectV2Item struct {
			DeletedItemId string
		} `graphql:"deleteProjectV2Item(input: {projectId: $projectId, itemId: $itemId})"`
	}

	variables := map[string]interface{}{
		"projectId": graphql.ID(projectID),
		"itemId":    graphql.ID(itemID),
	}

	return c.client.Mutate(ctx, &mutation, variables)
} 