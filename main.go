package main

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/romanitalian/GHOSTman/internal/collection"
	"github.com/romanitalian/GHOSTman/internal/logging"
	"github.com/romanitalian/GHOSTman/internal/ui"
)

const (
	preferenceCurrentForm = "currentForm"
	minURLPathLength      = 2
	defaultSplitOffset    = 0.2
	defaultWindowWidth    = 1024
	defaultWindowHeight   = 768

	appID    = "com.github.romanitalian.ghostman"
	appTitle = "GHOSTman"
)

func main() {
	logging.InitLogger()
	logging.Log.Info().Msg("Starting application...")

	a := app.NewWithID(appID)
	w := a.NewWindow(appTitle)

	content := container.NewStack()
	title := widget.NewLabel("Form Title")
	intro := widget.NewLabel("Form description goes here")
	intro.Wrapping = fyne.TextWrapWord
	top := container.NewVBox(title, widget.NewSeparator(), intro)

	setForm := func(form fyne.CanvasObject, formTitle string, formIntro string) {
		logging.Log.Info().Str("form_title", formTitle).Msg("Setting form")
		title.SetText(formTitle)
		intro.SetText(formIntro)
		content.Objects = []fyne.CanvasObject{form}
		content.Refresh()
	}

	// Загрузка коллекции
	coll, err := collection.LoadPostmanCollection("data/col.postman_collection.json")
	if err != nil {
		logging.Log.Error().Err(err).Msg("Error loading collection")
		return
	}

	// Переменные
	vars := make(map[string]string)
	for _, v := range coll.Variable {
		vars[v.Key] = v.Value
	}

	// Формы
	var forms []ui.Form
	for _, item := range coll.Item {
		if len(item.Request.URL.Path) >= minURLPathLength {
			formID := item.Request.URL.Path[1]
			form := ui.CreateForm(item, vars)
			forms = append(forms, ui.Form{
				ID:    formID,
				Title: item.Name,
				Intro: item.Request.Description,
				Form:  form,
			})
			logging.Log.Info().Str("form_id", formID).Str("name", item.Name).Msg("Added form")
		} else {
			logging.Log.Warn().Str("name", item.Name).Msg("Skipping item: invalid URL path length")
		}
	}

	filterEntry := widget.NewEntry()
	filterEntry.SetPlaceHolder("Filter by form name")
	filterEntry.Resize(fyne.NewSize(200, 40))

	filteredForms := make([]ui.Form, len(forms))
	copy(filteredForms, forms)

	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			if uid == "" {
				keys := make([]string, len(filteredForms))
				for i, f := range filteredForms {
					keys[i] = f.ID
				}
				return keys
			}
			return []string{}
		},
		IsBranch: func(uid string) bool {
			return uid == ""
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Form")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			if uid == "" {
				obj.(*widget.Label).SetText("Forms")
				return
			}
			for _, f := range filteredForms {
				if f.ID == uid {
					obj.(*widget.Label).SetText(f.Title)
					break
				}
			}
		},
		OnSelected: func(uid string) {
			for _, f := range filteredForms {
				if f.ID == uid {
					a.Preferences().SetString(preferenceCurrentForm, uid)
					setForm(f.Form, f.Title, f.Intro)
					break
				}
			}
		},
	}

	filterEntry.OnChanged = func(input string) {
		filteredForms = make([]ui.Form, 0)
		for _, f := range forms {
			if strings.Contains(strings.ToLower(f.Title), strings.ToLower(input)) {
				filteredForms = append(filteredForms, f)
			}
		}
		tree.Refresh()
	}

	if len(forms) > 0 {
		tree.Select(forms[0].ID)
	}

	treeScroll := container.NewVScroll(tree)
	treeScroll.SetMinSize(fyne.NewSize(200, 400))

	leftMenu := container.NewVBox(
		filterEntry,
		treeScroll,
	)
	leftMenu.Resize(fyne.NewSize(200, defaultWindowHeight))

	split := container.NewHSplit(leftMenu, container.NewBorder(top, nil, nil, nil, content))
	split.Offset = defaultSplitOffset
	w.SetContent(split)
	w.Resize(fyne.NewSize(defaultWindowWidth, defaultWindowHeight))
	logging.Log.Info().Msg("Application window created and ready")
	w.ShowAndRun()
}
