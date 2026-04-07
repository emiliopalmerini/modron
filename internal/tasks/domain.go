package tasks

const DatabaseID = "30a7cc4a-ef13-8151-8d39-d44ff1118cf7"

var ValidStatuses = []string{"Not Started", "In progress", "Done", "Archived"}

type Task struct {
	ID          string
	URL         string
	Name        string
	Status      string
	Due         string
	ProjectIDs  []string
	ParentTask  []string
	SubTasks    []string
	IsBlocking  []string
	BlockedBy   []string
}

type Filter struct {
	Status      string
	Name        string
	DueBefore   string
	DueAfter    string
	ProjectID   string
	SortBy      string
	SortDir     string
	PageSize    int
	StartCursor string
}

type CreateParams struct {
	Name         string
	Due          string
	Status       string
	ProjectID    string
	ParentTaskID string
}
