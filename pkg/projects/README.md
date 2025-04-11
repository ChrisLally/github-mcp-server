# GitHub Projects v2 GraphQL Client

This package provides a Go client for interacting with GitHub Projects v2 using GraphQL. It supports all major operations for managing projects, items, and fields.

## Installation

```bash
go get github.com/your-org/your-repo/pkg/projects
```

## Usage

### Creating a Client

```go
import "github.com/your-org/your-repo/pkg/projects"

// Create a new client with your GitHub token
client := projects.NewClient("your-github-token")
```

### Finding a Project

```go
// Find a project by organization and number
project, err := client.FindProjectByNumber(ctx, "your-org", 1)
if err != nil {
    log.Fatal(err)
}
```

### Managing Project Items

```go
// Add an item to a project
item, err := client.AddItemToProject(ctx, projectID, contentID)
if err != nil {
    log.Fatal(err)
}

// Update a field value
err = client.UpdateItemField(ctx, projectID, itemID, fieldID, "new value")
if err != nil {
    log.Fatal(err)
}

// Delete an item
err = client.DeleteItemFromProject(ctx, projectID, itemID)
if err != nil {
    log.Fatal(err)
}
```

### Managing Project Fields

```go
// Get all fields in a project
fields, err := client.GetProjectFields(ctx, projectID)
if err != nil {
    log.Fatal(err)
}

// Get all items in a project
items, err := client.GetProjectItems(ctx, projectID)
if err != nil {
    log.Fatal(err)
}
```

### Project Settings

```go
// Update project settings
err = client.UpdateProjectSettings(ctx, projectID, "New Title", true, "# Project README", "Short description")
if err != nil {
    log.Fatal(err)
}

// Create a new project
newProject, err := client.CreateProject(ctx, ownerID, "Project Title")
if err != nil {
    log.Fatal(err)
}
```

### Draft Issues

```go
// Add a draft issue
draft, err := client.AddDraftIssue(ctx, projectID, "Draft Title", "Draft Body")
if err != nil {
    log.Fatal(err)
}
```

## Authentication

The client requires a GitHub token with appropriate scopes:
- `read:project` for read-only operations
- `project` for read and write operations

## Error Handling

All methods return errors that can be handled appropriately. Common errors include:
- Authentication failures
- Invalid project IDs
- Permission issues
- Network errors

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 