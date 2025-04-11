package projects

// Project represents a GitHub Project v2
type Project struct {
	ID             string
	Title          string
	Number         int
	Public         bool
	Readme         string
	ShortDescription string
}

// ProjectItem represents an item in a GitHub Project
type ProjectItem struct {
	ID         string
	ContentID  string
	FieldValues []FieldValue
}

// FieldValue represents a field value in a project item
type FieldValue struct {
	ID    string
	Name  string
	Value interface{}
}

// Field represents a field in a project
type Field struct {
	ID    string
	Name  string
	Type  string
}

// SingleSelectOption represents an option in a single select field
type SingleSelectOption struct {
	ID   string
	Name string
}

// Iteration represents an iteration in an iteration field
type Iteration struct {
	ID        string
	StartDate string
} 