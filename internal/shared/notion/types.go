package notion

import "encoding/json"

// QueryResponse is the response from POST /v1/databases/{id}/query.
type QueryResponse struct {
	Results    []Page  `json:"results"`
	HasMore    bool    `json:"has_more"`
	NextCursor *string `json:"next_cursor"`
}

// Page represents a Notion page.
type Page struct {
	ID             string                        `json:"id"`
	URL            string                        `json:"url"`
	CreatedTime    string                        `json:"created_time"`
	LastEditedTime string                        `json:"last_edited_time"`
	Parent         json.RawMessage               `json:"parent"`
	Properties     map[string]json.RawMessage    `json:"properties"`
}

// PropertyValue extracts a typed property from raw JSON.
// Callers use the specific Extract* functions below.

type PropertyType struct {
	Type string `json:"type"`
}

// Title extracts the plain text from a title property.
func ExtractTitle(raw json.RawMessage) string {
	var prop struct {
		Title []struct {
			PlainText string `json:"plain_text"`
		} `json:"title"`
	}
	if err := json.Unmarshal(raw, &prop); err != nil {
		return ""
	}
	var result string
	for _, t := range prop.Title {
		result += t.PlainText
	}
	return result
}

// ExtractRichText extracts plain text from a rich_text property.
func ExtractRichText(raw json.RawMessage) string {
	var prop struct {
		RichText []struct {
			PlainText string `json:"plain_text"`
		} `json:"rich_text"`
	}
	if err := json.Unmarshal(raw, &prop); err != nil {
		return ""
	}
	var result string
	for _, t := range prop.RichText {
		result += t.PlainText
	}
	return result
}

// ExtractSelect extracts the name from a select property.
func ExtractSelect(raw json.RawMessage) string {
	var prop struct {
		Select *struct {
			Name string `json:"name"`
		} `json:"select"`
	}
	if err := json.Unmarshal(raw, &prop); err != nil || prop.Select == nil {
		return ""
	}
	return prop.Select.Name
}

// ExtractStatus extracts the name from a status property.
func ExtractStatus(raw json.RawMessage) string {
	var prop struct {
		Status *struct {
			Name string `json:"name"`
		} `json:"status"`
	}
	if err := json.Unmarshal(raw, &prop); err != nil || prop.Status == nil {
		return ""
	}
	return prop.Status.Name
}

// ExtractMultiSelect extracts the names from a multi_select property.
func ExtractMultiSelect(raw json.RawMessage) []string {
	var prop struct {
		MultiSelect []struct {
			Name string `json:"name"`
		} `json:"multi_select"`
	}
	if err := json.Unmarshal(raw, &prop); err != nil {
		return nil
	}
	names := make([]string, len(prop.MultiSelect))
	for i, s := range prop.MultiSelect {
		names[i] = s.Name
	}
	return names
}

// DateValue represents a Notion date property value.
type DateValue struct {
	Start string `json:"start"`
	End   string `json:"end,omitempty"`
}

// ExtractDate extracts the date value from a date property.
func ExtractDate(raw json.RawMessage) *DateValue {
	var prop struct {
		Date *DateValue `json:"date"`
	}
	if err := json.Unmarshal(raw, &prop); err != nil {
		return nil
	}
	return prop.Date
}

// ExtractCheckbox extracts the boolean from a checkbox property.
func ExtractCheckbox(raw json.RawMessage) bool {
	var prop struct {
		Checkbox bool `json:"checkbox"`
	}
	if err := json.Unmarshal(raw, &prop); err != nil {
		return false
	}
	return prop.Checkbox
}

// ExtractURL extracts the URL string from a url property.
func ExtractURL(raw json.RawMessage) string {
	var prop struct {
		URL *string `json:"url"`
	}
	if err := json.Unmarshal(raw, &prop); err != nil || prop.URL == nil {
		return ""
	}
	return *prop.URL
}

// ExtractNumber extracts a number from a number property.
func ExtractNumber(raw json.RawMessage) *float64 {
	var prop struct {
		Number *float64 `json:"number"`
	}
	if err := json.Unmarshal(raw, &prop); err != nil {
		return nil
	}
	return prop.Number
}

// ExtractRelation extracts page IDs from a relation property.
func ExtractRelation(raw json.RawMessage) []string {
	var prop struct {
		Relation []struct {
			ID string `json:"id"`
		} `json:"relation"`
	}
	if err := json.Unmarshal(raw, &prop); err != nil {
		return nil
	}
	ids := make([]string, len(prop.Relation))
	for i, r := range prop.Relation {
		ids[i] = r.ID
	}
	return ids
}

// GetPropertyType returns the type string of a property.
func GetPropertyType(raw json.RawMessage) string {
	var pt PropertyType
	if err := json.Unmarshal(raw, &pt); err != nil {
		return ""
	}
	return pt.Type
}
