package models

import (
	"encoding/json"
	"testing"
)

func TestCollection_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    Collection
		wantErr bool
	}{
		{
			name: "valid collection",
			json: `{
				"info": {
					"name": "Test Collection"
				},
				"item": [
					{
						"name": "Test Request",
						"request": {
							"method": "GET",
							"description": "Test Description",
							"header": [
								{
									"key": "Content-Type",
									"value": "application/json"
								}
							],
							"body": {
								"mode": "raw",
								"raw": "{\"test\": \"value\"}"
							},
							"url": {
								"raw": "http://example.com",
								"host": ["example.com"],
								"path": ["test"]
							}
						}
					}
				],
				"variable": [
					{
						"key": "baseUrl",
						"value": "http://example.com",
						"type": "string"
					}
				]
			}`,
			want: Collection{
				Info: struct {
					Name string `json:"name"`
				}{
					Name: "Test Collection",
				},
				Item: []Item{
					{
						Name: "Test Request",
						Request: struct {
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
						}{
							Method:      "GET",
							Description: "Test Description",
							Header: []struct {
								Key   string `json:"key"`
								Value string `json:"value"`
							}{
								{
									Key:   "Content-Type",
									Value: "application/json",
								},
							},
							Body: struct {
								Mode string `json:"mode"`
								Raw  string `json:"raw"`
							}{
								Mode: "raw",
								Raw:  "{\"test\": \"value\"}",
							},
							URL: struct {
								Raw  string   `json:"raw"`
								Host []string `json:"host"`
								Path []string `json:"path"`
							}{
								Raw:  "http://example.com",
								Host: []string{"example.com"},
								Path: []string{"test"},
							},
						},
					},
				},
				Variable: []struct {
					Key   string `json:"key"`
					Value string `json:"value"`
					Type  string `json:"type"`
				}{
					{
						Key:   "baseUrl",
						Value: "http://example.com",
						Type:  "string",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid json",
			json:    `{invalid json}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Collection
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Collection.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Info.Name != tt.want.Info.Name {
					t.Errorf("Collection.Info.Name = %v, want %v", got.Info.Name, tt.want.Info.Name)
				}
				if len(got.Item) != len(tt.want.Item) {
					t.Errorf("Collection.Item length = %v, want %v", len(got.Item), len(tt.want.Item))
				}
				if len(got.Variable) != len(tt.want.Variable) {
					t.Errorf("Collection.Variable length = %v, want %v", len(got.Variable), len(tt.want.Variable))
				}
			}
		})
	}
}
