package notion

// FilterBuilder constructs Notion API filter objects.
type FilterBuilder struct {
	conditions []map[string]any
}

func NewFilter() *FilterBuilder {
	return &FilterBuilder{}
}

// StatusEquals adds a status equals filter.
func (f *FilterBuilder) StatusEquals(property, value string) *FilterBuilder {
	f.conditions = append(f.conditions, map[string]any{
		"property": property,
		"status":   map[string]any{"equals": value},
	})
	return f
}

// StatusNotEquals adds a status not-equals filter.
func (f *FilterBuilder) StatusNotEquals(property, value string) *FilterBuilder {
	f.conditions = append(f.conditions, map[string]any{
		"property": property,
		"status":   map[string]any{"does_not_equal": value},
	})
	return f
}

// SelectEquals adds a select equals filter.
func (f *FilterBuilder) SelectEquals(property, value string) *FilterBuilder {
	f.conditions = append(f.conditions, map[string]any{
		"property": property,
		"select":   map[string]any{"equals": value},
	})
	return f
}

// TextContains adds a rich_text contains filter.
func (f *FilterBuilder) TextContains(property, value string) *FilterBuilder {
	f.conditions = append(f.conditions, map[string]any{
		"property":  property,
		"rich_text": map[string]any{"contains": value},
	})
	return f
}

// TitleContains adds a title contains filter.
func (f *FilterBuilder) TitleContains(property, value string) *FilterBuilder {
	f.conditions = append(f.conditions, map[string]any{
		"property": property,
		"title":    map[string]any{"contains": value},
	})
	return f
}

// DateOnOrBefore adds a date on_or_before filter.
func (f *FilterBuilder) DateOnOrBefore(property, value string) *FilterBuilder {
	f.conditions = append(f.conditions, map[string]any{
		"property": property,
		"date":     map[string]any{"on_or_before": value},
	})
	return f
}

// DateOnOrAfter adds a date on_or_after filter.
func (f *FilterBuilder) DateOnOrAfter(property, value string) *FilterBuilder {
	f.conditions = append(f.conditions, map[string]any{
		"property": property,
		"date":     map[string]any{"on_or_after": value},
	})
	return f
}

// DateBefore adds a date before filter.
func (f *FilterBuilder) DateBefore(property, value string) *FilterBuilder {
	f.conditions = append(f.conditions, map[string]any{
		"property": property,
		"date":     map[string]any{"before": value},
	})
	return f
}

// DateAfter adds a date after filter.
func (f *FilterBuilder) DateAfter(property, value string) *FilterBuilder {
	f.conditions = append(f.conditions, map[string]any{
		"property": property,
		"date":     map[string]any{"after": value},
	})
	return f
}

// DateIsNotEmpty adds a date is_not_empty filter.
func (f *FilterBuilder) DateIsNotEmpty(property string) *FilterBuilder {
	f.conditions = append(f.conditions, map[string]any{
		"property": property,
		"date":     map[string]any{"is_not_empty": true},
	})
	return f
}

// CheckboxEquals adds a checkbox equals filter.
func (f *FilterBuilder) CheckboxEquals(property string, value bool) *FilterBuilder {
	f.conditions = append(f.conditions, map[string]any{
		"property": property,
		"checkbox": map[string]any{"equals": value},
	})
	return f
}

// MultiSelectContains adds a multi_select contains filter.
func (f *FilterBuilder) MultiSelectContains(property, value string) *FilterBuilder {
	f.conditions = append(f.conditions, map[string]any{
		"property":     property,
		"multi_select": map[string]any{"contains": value},
	})
	return f
}

// RelationContains adds a relation contains filter.
func (f *FilterBuilder) RelationContains(property, pageID string) *FilterBuilder {
	f.conditions = append(f.conditions, map[string]any{
		"property": property,
		"relation": map[string]any{"contains": pageID},
	})
	return f
}

// URLIsNotEmpty adds a url is_not_empty filter.
func (f *FilterBuilder) URLIsNotEmpty(property string) *FilterBuilder {
	f.conditions = append(f.conditions, map[string]any{
		"property": property,
		"url":      map[string]any{"is_not_empty": true},
	})
	return f
}

// Build returns the filter object for the Notion API.
// Returns nil if no conditions were added.
func (f *FilterBuilder) Build() map[string]any {
	if len(f.conditions) == 0 {
		return nil
	}
	if len(f.conditions) == 1 {
		return f.conditions[0]
	}
	return map[string]any{
		"and": f.conditions,
	}
}

// Sort represents a sort specification.
type Sort struct {
	Property  string `json:"property,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Direction string `json:"direction"`
}

// BuildQueryBody constructs the full query request body.
func BuildQueryBody(filter map[string]any, sorts []Sort, pageSize int, startCursor string) map[string]any {
	body := map[string]any{}
	if filter != nil {
		body["filter"] = filter
	}
	if len(sorts) > 0 {
		body["sorts"] = sorts
	}
	if pageSize > 0 {
		body["page_size"] = pageSize
	}
	if startCursor != "" {
		body["start_cursor"] = startCursor
	}
	return body
}
