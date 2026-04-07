package notion

// PropertyBuilder helps construct property values for page creation/updates.
type PropertyBuilder struct {
	props map[string]any
}

func NewPropertyBuilder() *PropertyBuilder {
	return &PropertyBuilder{props: make(map[string]any)}
}

func (b *PropertyBuilder) Title(property, value string) *PropertyBuilder {
	b.props[property] = map[string]any{
		"title": []map[string]any{
			{"text": map[string]any{"content": value}},
		},
	}
	return b
}

func (b *PropertyBuilder) RichText(property, value string) *PropertyBuilder {
	b.props[property] = map[string]any{
		"rich_text": []map[string]any{
			{"text": map[string]any{"content": value}},
		},
	}
	return b
}

func (b *PropertyBuilder) Select(property, value string) *PropertyBuilder {
	b.props[property] = map[string]any{
		"select": map[string]any{"name": value},
	}
	return b
}

func (b *PropertyBuilder) Status(property, value string) *PropertyBuilder {
	b.props[property] = map[string]any{
		"status": map[string]any{"name": value},
	}
	return b
}

func (b *PropertyBuilder) MultiSelect(property string, values []string) *PropertyBuilder {
	options := make([]map[string]any, len(values))
	for i, v := range values {
		options[i] = map[string]any{"name": v}
	}
	b.props[property] = map[string]any{
		"multi_select": options,
	}
	return b
}

func (b *PropertyBuilder) Date(property, start string) *PropertyBuilder {
	b.props[property] = map[string]any{
		"date": map[string]any{"start": start},
	}
	return b
}

func (b *PropertyBuilder) DateRange(property, start, end string) *PropertyBuilder {
	b.props[property] = map[string]any{
		"date": map[string]any{"start": start, "end": end},
	}
	return b
}

func (b *PropertyBuilder) Checkbox(property string, value bool) *PropertyBuilder {
	b.props[property] = map[string]any{
		"checkbox": value,
	}
	return b
}

func (b *PropertyBuilder) URL(property, value string) *PropertyBuilder {
	b.props[property] = map[string]any{
		"url": value,
	}
	return b
}

func (b *PropertyBuilder) Number(property string, value float64) *PropertyBuilder {
	b.props[property] = map[string]any{
		"number": value,
	}
	return b
}

func (b *PropertyBuilder) Relation(property string, pageIDs []string) *PropertyBuilder {
	relations := make([]map[string]any, len(pageIDs))
	for i, id := range pageIDs {
		relations[i] = map[string]any{"id": id}
	}
	b.props[property] = map[string]any{
		"relation": relations,
	}
	return b
}

func (b *PropertyBuilder) Build() map[string]any {
	return b.props
}

// CreatePageBody builds the full body for POST /v1/pages.
func CreatePageBody(databaseID string, properties map[string]any) map[string]any {
	return map[string]any{
		"parent": map[string]any{
			"database_id": databaseID,
		},
		"properties": properties,
	}
}
