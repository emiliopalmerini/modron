package page

// Property represents a single page property with its type and extracted value.
type Property struct {
	Name  string
	Type  string
	Value string
}

// Page represents a generic Notion page with its properties rendered for display.
type Page struct {
	ID             string
	URL            string
	CreatedTime    string
	LastEditedTime string
	Properties     []Property
}

// UpdateParams holds the update request. PropertyUpdates maps property names
// to their new values encoded as Notion API property objects.
type UpdateParams struct {
	PageID          string
	PropertyUpdates map[string]any
}
