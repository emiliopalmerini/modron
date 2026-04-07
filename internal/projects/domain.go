package projects

const DatabaseID = "30a7cc4a-ef13-81fd-8b36-fe632b889b70"

var (
	ValidStatuses = []string{"Planning", "In Progress", "Paused", "Backlog", "Done", "Canceled"}
	ValidTags     = []string{"Content", "Dev", "Marketing", "Community", "Business", "Work"}
)

type Project struct {
	ID         string
	URL        string
	Name       string
	Status     string
	Tag        string
	Summary    string
	DatesStart string
	DatesEnd   string
	LaunchDate string
	TaskIDs    []string
	IsBlocking []string
	BlockedBy  []string
}

type Filter struct {
	Status          string
	Tag             string
	Name            string
	DateStartAfter  string
	DateStartBefore string
	HasLaunchDate   bool
	SortBy          string
	SortDir         string
	PageSize        int
	StartCursor     string
}

type CreateParams struct {
	Name       string
	Status     string
	Tag        string
	Summary    string
	DatesStart string
	DatesEnd   string
	LaunchDate string
}
