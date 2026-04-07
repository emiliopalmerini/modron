package events

const DatabaseID = "30c7cc4a-ef13-80d8-81ec-ffe5d241d8fd"

type Event struct {
	ID        string
	URL       string
	Name      string
	DateStart string
	DateEnd   string
}

type Filter struct {
	Name        string
	DateAfter   string
	DateBefore  string
	SortBy      string
	SortDir     string
	PageSize    int
	StartCursor string
}

type CreateParams struct {
	Name       string
	DateStart  string
	DateEnd    string
	IsDatetime bool
}
