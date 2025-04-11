package projects

import (
	"context"
	"fmt"

	"github.com/shurcooL/graphql"
)

// GetProjectFields retrieves all fields for a project
func (c *Client) GetProjectFields(ctx context.Context, projectID string) ([]Field, error) {
	var query struct {
		Node struct {
			ProjectV2 struct {
				Fields struct {
					Nodes []struct {
						ID   string
						Name string
						Type string
					}
				} `graphql:"fields(first: 100)"`
			} `graphql:"... on ProjectV2"`
		} `graphql:"node(id: $projectId)"`
	}

	variables := map[string]interface{}{
		"projectId": graphql.ID(projectID),
	}

	err := c.client.Query(ctx, &query, variables)
	if err != nil {
		return nil, err
	}

	fields := make([]Field, len(query.Node.ProjectV2.Fields.Nodes))
	for i, node := range query.Node.ProjectV2.Fields.Nodes {
		fields[i] = Field{
			ID:   node.ID,
			Name: node.Name,
			Type: node.Type,
		}
	}

	return fields, nil
}

// GetProjectItems retrieves items from a project
func (c *Client) GetProjectItems(ctx context.Context, projectID string) ([]ProjectItem, error) {
	var query struct {
		Node struct {
			ProjectV2 struct {
				Items struct {
					Nodes []struct {
						ID         string
						ContentID  string `graphql:"contentId"`
						FieldValues struct {
							Nodes []struct {
								ID    string
								Name  string
								Value interface{}
							}
						} `graphql:"fieldValues(first: 100)"`
					}
				} `graphql:"items(first: 100)"`
			} `graphql:"... on ProjectV2"`
		} `graphql:"node(id: $projectId)"`
	}

	variables := map[string]interface{}{
		"projectId": graphql.ID(projectID),
	}

	err := c.client.Query(ctx, &query, variables)
	if err != nil {
		return nil, err
	}

	items := make([]ProjectItem, len(query.Node.ProjectV2.Items.Nodes))
	for i, node := range query.Node.ProjectV2.Items.Nodes {
		fieldValues := make([]FieldValue, len(node.FieldValues.Nodes))
		for j, fv := range node.FieldValues.Nodes {
			fieldValues[j] = FieldValue{
				ID:    fv.ID,
				Name:  fv.Name,
				Value: fv.Value,
			}
		}

		items[i] = ProjectItem{
			ID:          node.ID,
			ContentID:   node.ContentID,
			FieldValues: fieldValues,
		}
	}

	return items, nil
}

// UpdateProjectSettings updates project settings
func (c *Client) UpdateProjectSettings(ctx context.Context, projectID string, title string, public bool, readme string, shortDescription string) error {
	var mutation struct {
		UpdateProjectV2 struct {
			ProjectV2 struct {
				ID string
			}
		} `graphql:"updateProjectV2(input: {projectId: $projectId, title: $title, public: $public, readme: $readme, shortDescription: $shortDescription})"`
	}

	variables := map[string]interface{}{
		"projectId":        graphql.ID(projectID),
		"title":           graphql.String(title),
		"public":          graphql.Boolean(public),
		"readme":          graphql.String(readme),
		"shortDescription": graphql.String(shortDescription),
	}

	return c.client.Mutate(ctx, &mutation, variables)
}

// CreateProject creates a new project
func (c *Client) CreateProject(ctx context.Context, ownerID string, title string) (*Project, error) {
	var mutation struct {
		CreateProjectV2 struct {
			ProjectV2 struct {
				ID    string
				Title string
			}
		} `graphql:"createProjectV2(input: {ownerId: $ownerId, title: $title})"`
	}

	variables := map[string]interface{}{
		"ownerId": graphql.ID(ownerID),
		"title":   graphql.String(title),
	}

	err := c.client.Mutate(ctx, &mutation, variables)
	if err != nil {
		return nil, err
	}

	return &Project{
		ID:    mutation.CreateProjectV2.ProjectV2.ID,
		Title: mutation.CreateProjectV2.ProjectV2.Title,
	}, nil
}

// AddDraftIssue adds a draft issue to a project
func (c *Client) AddDraftIssue(ctx context.Context, projectID string, title string, body string) (*ProjectItem, error) {
	var mutation struct {
		AddProjectV2DraftIssue struct {
			ProjectItem struct {
				ID string
			}
		} `graphql:"addProjectV2DraftIssue(input: {projectId: $projectId, title: $title, body: $body})"`
	}

	variables := map[string]interface{}{
		"projectId": graphql.ID(projectID),
		"title":    graphql.String(title),
		"body":     graphql.String(body),
	}

	err := c.client.Mutate(ctx, &mutation, variables)
	if err != nil {
		return nil, err
	}

	return &ProjectItem{
		ID: mutation.AddProjectV2DraftIssue.ProjectItem.ID,
	}, nil
} 