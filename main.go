package main

import (
	"fmt"
	"regexp"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func makeFormTab(w fyne.Window) fyne.CanvasObject {
	name := widget.NewEntry()
	name.SetPlaceHolder("John Smith")

	email := widget.NewEntry()
	email.SetPlaceHolder("test@example.com")
	email.Validator = func(s string) error {
		matched, err := regexp.MatchString(`\w{1,}@\w{1,}\.\w{1,4}`, s)
		if err != nil {
			return err
		}
		if !matched {
			return fmt.Errorf("not a valid email")
		}
		return nil
	}

	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Password")

	disabled := widget.NewRadioGroup([]string{"Option 1", "Option 2"}, func(string) {})
	disabled.Horizontal = true
	disabled.Disable()
	largeText := widget.NewMultiLineEntry()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Name", Widget: name, HintText: "Your full name"},
			{Text: "Email", Widget: email, HintText: "A valid email address"},
		},
		OnCancel: func() {
			fmt.Println("Cancelled")
		},
		OnSubmit: func() {
			fmt.Println("Form submitted")
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Form for: " + name.Text,
				Content: largeText.Text,
			})
		},
	}
	form.Append("Password", password)
	form.Append("Disabled", disabled)
	form.Append("Message", largeText)
	return form
}

func main() {
	myApp := app.New()
	window := myApp.NewWindow("Form Demo")

	content := makeFormTab(window)
	window.SetContent(container.NewVBox(content))
	window.Resize(fyne.NewSize(400, 500))
	window.ShowAndRun()
}
