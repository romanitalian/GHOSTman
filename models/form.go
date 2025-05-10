package models

import "fyne.io/fyne/v2"

// Form represents a form in the application
type Form struct {
	ID    string
	Title string
	Intro string
	Form  fyne.CanvasObject
}
