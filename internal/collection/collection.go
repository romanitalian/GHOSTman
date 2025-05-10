package collection

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Cllns represents the structure of the collection JSON (Postman collection format)
type Cllns struct {
	Info struct {
		Name string `json:"name"`
	} `json:"info"`
	Item []struct {
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
	} `json:"item"`
	Variable []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
		Type  string `json:"type"`
	} `json:"variable"`
}

// substituteVariables replaces {{var}} in a string with values from vars
func SubstituteVariables(s string, vars map[string]string) string {
	for k, v := range vars {
		placeholder := "{{" + k + "}}"
		s = strings.ReplaceAll(s, placeholder, v)
	}
	return s
}

// LoadPostmanCollection loads and parses the Postman collection from file
type FormMeta struct {
	ID    string
	Title string
	Intro string
}

func LoadPostmanCollection(path string) (*Cllns, error) {
	data, err := os.ReadFile(filepath.Join(path))
	if err != nil {
		return nil, fmt.Errorf("error reading Postman collection: %v", err)
	}
	var collection Cllns
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("error parsing Postman collection: %v", err)
	}
	return &collection, nil
}
