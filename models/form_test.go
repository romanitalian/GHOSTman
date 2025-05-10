package models

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func TestForm_New(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		title   string
		intro   string
		form    fyne.CanvasObject
		want    Form
		wantErr bool
	}{
		{
			name:  "valid form",
			id:    "test-form",
			title: "Test Form",
			intro: "Test Introduction",
			form:  widget.NewLabel("Test Form Content"),
			want: Form{
				ID:    "test-form",
				Title: "Test Form",
				Intro: "Test Introduction",
				Form:  widget.NewLabel("Test Form Content"),
			},
			wantErr: false,
		},
		{
			name:    "empty id",
			id:      "",
			title:   "Test Form",
			intro:   "Test Introduction",
			form:    widget.NewLabel("Test Form Content"),
			wantErr: true,
		},
		{
			name:    "empty title",
			id:      "test-form",
			title:   "",
			intro:   "Test Introduction",
			form:    widget.NewLabel("Test Form Content"),
			wantErr: true,
		},
		{
			name:    "nil form",
			id:      "test-form",
			title:   "Test Form",
			intro:   "Test Introduction",
			form:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := Form{
				ID:    tt.id,
				Title: tt.title,
				Intro: tt.intro,
				Form:  tt.form,
			}

			if tt.wantErr {
				if form.ID != "" && form.Title != "" && form.Form != nil {
					t.Errorf("Form.New() should have failed for invalid input")
				}
				return
			}

			if form.ID != tt.want.ID {
				t.Errorf("Form.ID = %v, want %v", form.ID, tt.want.ID)
			}
			if form.Title != tt.want.Title {
				t.Errorf("Form.Title = %v, want %v", form.Title, tt.want.Title)
			}
			if form.Intro != tt.want.Intro {
				t.Errorf("Form.Intro = %v, want %v", form.Intro, tt.want.Intro)
			}
			if form.Form == nil {
				t.Error("Form.Form is nil")
			}
		})
	}
}
