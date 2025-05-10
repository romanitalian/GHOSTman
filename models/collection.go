package models

// Collection represents the structure of the collection JSON (Postman collection format)
type Collection struct {
	Info struct {
		Name string `json:"name"`
	} `json:"info"`
	Item     []Item `json:"item"`
	Variable []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
		Type  string `json:"type"`
	} `json:"variable"`
}

// Item represents a single item in the Postman collection
type Item struct {
	Name    string `json:"name"`
	Request struct {
		Method      string `json:"method"`
		Description string `json:"description"`
		Header      []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"header"`
		Body struct {
			Mode string `json:"mode"`
			Raw  string `json:"raw"`
		} `json:"body"`
		URL struct {
			Raw  string   `json:"raw"`
			Host []string `json:"host"`
			Path []string `json:"path"`
		} `json:"url"`
	} `json:"request"`
}
