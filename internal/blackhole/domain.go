package blackhole

const DatabaseID = "3097cc4a-ef13-80de-b79b-ec8552eb7d7e"

var ValidTypes = []string{"Idea", "Reference", "TBR"}

type Entry struct {
	ID        string
	URL       string
	Name      string
	Type      string
	Tags      []string
	Summary   string
	Processed bool
	EntryURL  string
}

type Filter struct {
	Type        string
	Tags        string // comma-separated, any match
	Name        string
	Processed   *bool
	HasURL      bool
	SortBy      string
	SortDir     string
	PageSize    int
	StartCursor string
}

type CreateParams struct {
	Name      string
	Type      string
	Summary   string
	Tags      string // comma-separated
	URL       string
	Processed bool
}
