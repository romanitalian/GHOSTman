package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/romanitalian/GHOSTman/v2/models"
)

const (
	preferenceCurrentForm     = "currentForm"
	preferenceCollectionPath  = "collectionPath"
	preferenceConfirmDeletion = "confirmDeletion"
	preferenceSaveOnDelete    = "saveOnDelete"
	minURLPathLength          = 2
	defaultWindowWidth        = 1024
	defaultWindowHeight       = 768

	logLevel      = zerolog.WarnLevel
	logFormatJSON = true

	appID    = "com.github.romanitalian.ghostman"
	appTitle = "GHOSTman"
)

//go:embed FyneApp.toml
var _ []byte

// AppState holds the application's state
type AppState struct {
	collection     *models.Collection
	collectionPath string
	isModified     bool
}

var (
	appState    AppState
	topWindow   fyne.Window
	httpMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "CONNECT", "TRACE"}

	unsavedLabel *widget.Label // Индикатор несохранённых изменений

	autoSaveInterval = 60 * time.Second
	autoSaveStopCh   chan struct{}
)

func (a *AppState) setModified(modified bool) {
	if a.isModified == modified {
		return
	}
	a.isModified = modified
	if modified {
		topWindow.SetTitle(appTitle + " *")
		if unsavedLabel != nil {
			unsavedLabel.SetText("● Unsaved changes")
			unsavedLabel.Show()
		}
	} else {
		topWindow.SetTitle(appTitle)
		if unsavedLabel != nil {
			unsavedLabel.SetText("")
			unsavedLabel.Hide()
		}
	}
}

func substituteVariables(s string, vars map[string]string) string {
	for k, v := range vars {
		placeholder := "{{" + k + "}}"
		s = strings.ReplaceAll(s, placeholder, v)
	}
	return s
}

func createForm(item *models.Item, vars map[string]string) fyne.CanvasObject {
	// Create form fields
	frm := &widget.Form{}

	// Add request info fields
	urlEntry := widget.NewEntry()
	urlEntry.SetText(substituteVariables(item.Request.URL.Raw, vars))
	urlEntry.OnChanged = func(s string) {
		item.Request.URL.Raw = s
		appState.setModified(true)
	}
	frm.Append(models.LabelURL, urlEntry)

	methodSelect := widget.NewSelect(httpMethods, func(value string) {
		item.Request.Method = value
		appState.setModified(true)
	})
	methodSelect.SetSelected(item.Request.Method)
	frm.Append(models.LabelMethod, methodSelect)

	var headersText strings.Builder
	for _, h := range item.Request.Header {
		headersText.WriteString(fmt.Sprintf("%s: %s\n", h.Key, substituteVariables(h.Value, vars)))
	}
	hdrsEntry := widget.NewMultiLineEntry()
	hdrsEntry.SetText(headersText.String())
	hdrsEntry.OnChanged = func(s string) {
		// This is a simplification. A more robust implementation would parse the headers.
		// For now, we'll just mark it as modified.
		// A proper implementation would need to update item.Request.Header
		appState.setModified(true)
	}
	frm.Append(models.LabelHeaders, hdrsEntry)

	// Create body field with fixed height
	bodyEntry := widget.NewMultiLineEntry()
	bodyEntry.SetText(substituteVariables(item.Request.Body.Raw, vars))
	bodyEntry.OnChanged = func(s string) {
		item.Request.Body.Raw = s
		appState.setModified(true)
	}

	// Calculate number of lines in JSON
	lines := strings.Count(item.Request.Body.Raw, "\n") + 1
	bodyEntry.SetMinRowsVisible(lines)

	frm.Append(models.LabelBody, bodyEntry)

	// Create response field
	textRS := widget.NewMultiLineEntry()
	textRS.Wrapping = fyne.TextWrapWord
	textRS.TextStyle = fyne.TextStyle{
		Bold:      true,
		Monospace: true,
	}
	textRS.Resize(fyne.NewSize(200, 200))
	textRS.SetMinRowsVisible(25)

	// Create progress bar
	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide() // Hide initially

	// Add submit button
	submitBtn := widget.NewButton(models.LabelSend, func() {
		// Clear response field and show progress
		textRS.SetText("")
		progressBar.Show()
		progressBar.Refresh()

		// Create request
		rq, err := http.NewRequest(methodSelect.Selected, urlEntry.Text, bytes.NewBufferString(bodyEntry.Text))
		if err != nil {
			progressBar.Hide()
			progressBar.Refresh()
			textRS.SetText(fmt.Sprintf(models.ErrCreatingRequest, err))
			return
		}

		// Add headers
		hdrs := strings.Split(hdrsEntry.Text, "\n")
		for _, h := range hdrs {
			if h == "" {
				continue
			}
			parts := strings.SplitN(h, ":", 2)
			if len(parts) == 2 {
				rq.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}

		// Send request in goroutine
		go func() {
			client := &http.Client{}
			resp, err := client.Do(rq)
			if err != nil {
				fyne.Do(func() {
					progressBar.Hide()
					progressBar.Refresh()
					textRS.SetText(fmt.Sprintf(models.ErrSendingRequest, err))
				})
				return
			}
			defer resp.Body.Close()

			// Read response
			bodyRS, err := io.ReadAll(resp.Body)
			if err != nil {
				fyne.Do(func() {
					progressBar.Hide()
					progressBar.Refresh()
					textRS.SetText(fmt.Sprintf(models.ErrReadingResponse, err))
				})
				return
			}

			// Format response
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, bodyRS, "", "    "); err != nil {
				fyne.Do(func() {
					textRS.SetText(string(bodyRS))
					progressBar.Hide()
					progressBar.Refresh()
				})
				return
			}

			// Update response and hide progress
			fyne.Do(func() {
				textRS.SetText(prettyJSON.String())
				progressBar.Hide()
				progressBar.Refresh()
			})
		}()
	})

	frm.Append("", submitBtn)

	frm.Append("", progressBar)

	// Create response container
	containerRS := container.NewVBox(
		progressBar,
		textRS,
	)

	// Add response field after submit button
	frm.Append(models.LabelResponse, containerRS)

	// Create vertical container with form
	return container.NewVBox(
		frm,
	)
}

func loadPostmanCollection(filePath string) ([]models.Form, error) {
	var forms []models.Form

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Error().Err(err).Msg(models.LogLoadingForms)
		return nil, fmt.Errorf(models.ErrReadingCollection, err)
	}

	var collection models.Collection
	if err := json.Unmarshal(data, &collection); err != nil {
		log.Error().Err(err).Msg(models.LogLoadingForms)
		return nil, fmt.Errorf(models.ErrParsingCollection, err)
	}

	appState.collection = &collection
	appState.collectionPath = filePath
	appState.setModified(false)

	// Store variables in map
	vars := make(map[string]string)
	for _, v := range collection.Variable {
		vars[v.Key] = v.Value
	}
	log.Info().Fields(vars).Msg(models.LogLoadedVariables)

	log.Info().Int("count", len(collection.Item)).Msg(models.LogTotalItems)

	for i, item := range collection.Item {
		log.Info().Int("idx", i+1).Str("name", item.Name).Msg(models.LogProcessingItem)
		log.Info().Interface("url_path", item.Request.URL.Path).Msg(models.LogURLPath)
		if len(item.Request.URL.Path) >= minURLPathLength {
			formID := item.Request.URL.Path[1]
			log.Info().Str("form_id", formID).Msg(models.LogFormID)

			// Create form with request info and variable substitution
			form := createForm(&collection.Item[i], vars)

			forms = append(forms, models.Form{
				ID:    formID,
				Title: item.Name,
				Intro: item.Request.Description,
				Form:  form,
			})
			log.Info().Str("form_id", formID).Str("name", item.Name).Msg(models.LogAddedForm)
		} else {
			log.Warn().Str("name", item.Name).Msg(models.LogSkippingItem)
		}
	}

	log.Info().Int("count", len(forms)).Msg(models.LogTotalForms)
	for _, form := range forms {
		log.Info().Str("form_id", form.ID).Str("title", form.Title).Msg(models.LogLoadedForm)
	}

	return forms, nil
}

func main() {
	// Set up zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.SetGlobalLevel(logLevel)

	// Configure log output format
	if logFormatJSON {
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05.000",
			NoColor:    false,
		})
	}

	log.Info().Msg(models.LogStartingApp)

	a := app.NewWithID(appID)
	w := a.NewWindow(appTitle)
	topWindow = w

	var forms []models.Form
	var filteredForms []models.Form
	var tree *widget.Tree
	var filterEntry *widget.Entry
	var selectedID string
	var confirmDeletion bool
	var saveOnDelete bool
	var undoBtn *widget.Button

	type deletedEntry struct {
		id                 string
		form               models.Form
		item               models.Item
		indexInForms       int
		indexInCollection  int
		previousSelectedID string
		filterBefore       string
	}
	var lastDeleted *deletedEntry

	content := container.NewStack()
	emptyLabel := widget.NewLabel("No forms. Add collection.")
	title := widget.NewLabel("Form Title")
	intro := widget.NewLabel("Form description goes here")
	intro.Wrapping = fyne.TextWrapWord

	setForm := func(form fyne.CanvasObject, formTitle string, formIntro string) {
		log.Info().Str("form_title", formTitle).Msg(models.LogSettingForm)
		title.SetText(formTitle)
		intro.SetText(formIntro)
		content.Objects = []fyne.CanvasObject{form}
		content.Refresh()
	}

	// Load initial collection from preferences
	collectionPath := a.Preferences().String(preferenceCollectionPath)
	if collectionPath != "" {
		var err error
		forms, err = loadPostmanCollection(collectionPath)
		if err != nil {
			log.Error().Err(err).Str("path", collectionPath).Msg("Failed to load collection from saved path")
			forms = []models.Form{} // Start with empty if load fails
		}
	}

	// Create filtered forms slice
	filteredForms = make([]models.Form, len(forms))
	copy(filteredForms, forms)

	// helper: find form index by ID in slice
	findFormIndexByID := func(list []models.Form, id string) int {
		for i := range list {
			if list[i].ID == id {
				return i
			}
		}
		return -1
	}

	// removed: replaced by preferIndexByTitle

	// helper: prefer collection index by matching title when multiple IDs match
	preferIndexByTitle := func(id string, title string) int {
		if appState.collection == nil {
			return -1
		}
		candidate := -1
		for i := range appState.collection.Item {
			it := appState.collection.Item[i]
			if len(it.Request.URL.Path) >= minURLPathLength && it.Request.URL.Path[1] == id {
				if candidate == -1 {
					candidate = i
				}
				if it.Name == title {
					return i
				}
			}
		}
		return candidate
	}

	// helper: remove form at index
	removeFormAt := func(list []models.Form, idx int) []models.Form {
		if idx < 0 || idx >= len(list) {
			return list
		}
		return append(list[:idx], list[idx+1:]...)
	}

	// helper: remove collection item at index
	removeCollectionItemAt := func(idx int) {
		if appState.collection == nil {
			return
		}
		if idx < 0 || idx >= len(appState.collection.Item) {
			return
		}
		appState.collection.Item = append(appState.collection.Item[:idx], appState.collection.Item[idx+1:]...)
	}

	// helper: clear selection and content
	clearSelection := func() {
		if len(filteredForms) == 0 {
			content.Objects = []fyne.CanvasObject{emptyLabel}
		} else {
			content.Objects = []fyne.CanvasObject{}
		}
		content.Refresh()
		title.SetText("")
		intro.SetText("")
		selectedID = ""
		a.Preferences().SetString(preferenceCurrentForm, "")
	}

	tree = &widget.Tree{
		ChildUIDs: func(uid string) []string {
			if uid == "" {
				keys := make([]string, len(filteredForms))
				for i, f := range filteredForms {
					keys[i] = f.ID
				}
				log.Info().Strs("keys", keys).Msg(models.LogTreeChildUIDs)
				return keys
			}
			return []string{}
		},
		IsBranch: func(uid string) bool {
			isRoot := uid == ""
			log.Debug().Str("uid", uid).Bool("is_root", isRoot).Msg(models.LogTreeIsBranch)
			return isRoot
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			log.Debug().Bool("branch", branch).Msg(models.LogTreeCreateNode)
			if branch {
				return widget.NewLabel(models.LabelForms)
			}
			nameLabel := widget.NewLabel(models.LabelForm)
			delBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {})
			delBtn.Importance = widget.LowImportance
			h := container.NewHBox(nameLabel, delBtn)
			return h
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			if uid == "" || branch {
				// root node
				if lbl, ok := obj.(*widget.Label); ok {
					log.Debug().Msg(models.LogTreeUpdateNodeRoot)
					lbl.SetText(models.LabelForms)
				}
				return
			}
			// leaf node container with label + delete button
			cont, ok := obj.(*fyne.Container)
			if !ok || len(cont.Objects) < 2 {
				return
			}
			nameLabel, ok := cont.Objects[0].(*widget.Label)
			if !ok {
				return
			}
			delBtn, ok := cont.Objects[1].(*widget.Button)
			if !ok {
				return
			}
			// update label text
			for _, f := range filteredForms {
				if f.ID == uid {
					log.Debug().Str("uid", uid).Str("title", f.Title).Msg(models.LogTreeUpdateNode)
					nameLabel.SetText(f.Title)
					break
				}
			}
			// bind delete handler (with optional confirm)
			delBtn.OnTapped = func() {
				doDelete := func() {
					log.Info().Str("uid", uid).Msg(models.LogDeleteClick)
					// compute candidate index by matching title to avoid id collisions
					ci := preferIndexByTitle(uid, nameLabel.Text)
					fiAll := findFormIndexByID(forms, uid)
					fiFiltered := findFormIndexByID(filteredForms, uid)
					log.Debug().Int("ci", ci).Int("fi_all", fiAll).Int("fi_filtered", fiFiltered).Msg(models.LogDeleteIndexes)
					// decide next selection before removal
					var nextID string
					if fiFiltered >= 0 && len(filteredForms) > 1 {
						if fiFiltered < len(filteredForms)-1 {
							nextID = filteredForms[fiFiltered+1].ID
						} else if fiFiltered-1 >= 0 {
							nextID = filteredForms[fiFiltered-1].ID
						}
					}

					// capture for undo
					if ci >= 0 && fiAll >= 0 {
						lastDeleted = &deletedEntry{
							id:                 uid,
							form:               forms[fiAll],
							item:               appState.collection.Item[ci],
							indexInForms:       fiAll,
							indexInCollection:  ci,
							previousSelectedID: selectedID,
							filterBefore:       filterEntry.Text,
						}
						if undoBtn != nil {
							undoBtn.Enable()
						}
					} else {
						lastDeleted = nil
						if undoBtn != nil {
							undoBtn.Disable()
						}
					}

					if ci >= 0 {
						removeCollectionItemAt(ci)
					}
					if fiAll >= 0 {
						forms = removeFormAt(forms, fiAll)
					}
					if fiFiltered >= 0 {
						filteredForms = removeFormAt(filteredForms, fiFiltered)
					}
					log.Debug().Int("forms_len", len(forms)).Int("filtered_len", len(filteredForms)).Msg(models.LogAfterDeleteCounts)

					// mark modified
					appState.setModified(true)

					// adjust selection if needed
					if selectedID == uid {
						if nextID != "" {
							selectedID = nextID
							a.Preferences().SetString(preferenceCurrentForm, nextID)
							tree.Select(nextID)
						} else {
							clearSelection()
						}
					}

					// refresh tree
					tree.Refresh()
					log.Info().Str("uid", uid).Msg(models.LogDeleteDone)
				}

				if confirmDeletion {
					// show method and URL if available + don't ask again
					method, urlRaw := "", ""
					idx := preferIndexByTitle(uid, nameLabel.Text)
					if idx >= 0 {
						method = appState.collection.Item[idx].Request.Method
						urlRaw = appState.collection.Item[idx].Request.URL.Raw
					}
					msg := fmt.Sprintf("Delete '%s' [%s %s]?", nameLabel.Text, method, urlRaw)
					dontAskChk := widget.NewCheck("Don't ask again", nil)
					content := container.NewVBox(
						widget.NewLabel(msg),
						dontAskChk,
					)
					dialog.ShowCustomConfirm("Delete", "Delete", "Cancel", content, func(ok bool) {
						if ok {
							if dontAskChk.Checked {
								confirmDeletion = false
								a.Preferences().SetBool(preferenceConfirmDeletion, false)
								// update UI checkbox if present
								// note: confirmChk is outside scope; user will see effect on next toggle
							}
							doDelete()
						}
					}, w)
					return
				}
				doDelete()
			}
		},
		OnSelected: func(uid string) {
			for _, f := range filteredForms {
				if f.ID == uid {
					log.Info().Str("uid", uid).Str("form", f.Title).Msg(models.LogTreeSelected)
					a.Preferences().SetString(preferenceCurrentForm, uid)
					selectedID = uid
					setForm(f.Form, f.Title, f.Intro)
					break
				}
			}
		},
	}

	filterEntry = widget.NewEntry()
	filterEntry.SetPlaceHolder(models.FilterPlaceholder)
	filterEntry.Resize(fyne.NewSize(200, 40)) // Set minimum size for filter
	filterEntry.OnChanged = func(input string) {
		filteredForms = make([]models.Form, 0)
		for _, f := range forms {
			if strings.Contains(strings.ToLower(f.Title), strings.ToLower(input)) {
				filteredForms = append(filteredForms, f)
			}
		}
		tree.Refresh()
	}

	// Theme switcher
	themeSelect := widget.NewSelect([]string{models.ThemeLight, models.ThemeDark}, func(value string) {
		if value == models.ThemeDark {
			a.Settings().SetTheme(fixedVariant(theme.VariantDark))
		} else {
			a.Settings().SetTheme(fixedVariant(theme.VariantLight))
		}
	})
	themeSelect.SetSelected(models.ThemeLight)

	// Кнопка для добавления коллекции
	addCollectionBtn := widget.NewButton("Add Collection", func() {
		fileDialog := dialog.NewFileOpen(
			func(reader fyne.URIReadCloser, err error) {
				if err != nil || reader == nil {
					return
				}
				defer reader.Close()

				filePath := reader.URI().Path()
				newForms, loadErr := loadPostmanCollection(filePath)
				if loadErr != nil {
					dialog.ShowError(fmt.Errorf("failed to load collection: %w", loadErr), w)
					return
				}

				forms = newForms
				filterEntry.SetText("")
				filteredForms = make([]models.Form, len(forms))
				copy(filteredForms, forms)
				tree.Refresh()
				if len(forms) > 0 {
					tree.Select(forms[0].ID)
				}

				a.Preferences().SetString(preferenceCollectionPath, filePath)
			},
			w,
		)
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
		fileDialog.Show()
	})

	// Select для выбора интервала автосохранения
	// intervalOptions := []string{"Выкл", "30 сек", "1 мин", "5 мин"}
	// intervalMap := map[string]time.Duration{
	// 	"Выкл":   0,
	// 	"30 сек": 30 * time.Second,
	// 	"1 мин":  60 * time.Second,
	// 	"5 мин":  5 * time.Minute,
	// }
	// autoSaveSelect := widget.NewSelect(intervalOptions, func(val string) {
	// 	if d, ok := intervalMap[val]; ok {
	// 		autoSaveInterval = d
	// 		if d > 0 {
	// 			setupAutoSave(d, w)
	// 		} else if autoSaveStopCh != nil {
	// 			close(autoSaveStopCh)
	// 			autoSaveStopCh = nil
	// 		}
	// 	}
	// })
	// autoSaveSelect.SetSelected("1 мин")

	// Unsaved changes indicator
	unsavedLabel = widget.NewLabel("")
	unsavedLabel.Hide()

	// Confirm deletion checkbox
	confirmDeletion = a.Preferences().Bool(preferenceConfirmDeletion)
	confirmChk := widget.NewCheck("Confirm Delete", func(b bool) {
		confirmDeletion = b
		a.Preferences().SetBool(preferenceConfirmDeletion, b)
	})
	confirmChk.SetChecked(confirmDeletion)

	// Save on delete checkbox
	saveOnDelete = a.Preferences().Bool(preferenceSaveOnDelete)
	saveOnDeleteChk := widget.NewCheck("Save on delete", func(b bool) {
		saveOnDelete = b
		a.Preferences().SetBool(preferenceSaveOnDelete, b)
	})
	saveOnDeleteChk.SetChecked(saveOnDelete)

	// Main application menu
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Save", func() {
				log.Info().Msg("Save menu item clicked!")
				err := saveCollection()
				if err != nil {
					dialog.ShowError(fmt.Errorf("save error: %w", err), w)
				} else {
					dialog.ShowInformation("Saved", "Collection saved successfully", w)
				}
			}),
			fyne.NewMenuItem("Save As...", func() {
				saveCollectionAs(w)
			}),
		),
	)
	mainMenu.Items[0].Items[0].Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierShortcutDefault}
	w.SetMainMenu(mainMenu)

	// Window close handling with unsaved changes warning
	w.SetCloseIntercept(func() {
		if appState.isModified {
			dialog.ShowCustomConfirm(
				"Unsaved Changes",
				"Save",
				"Exit Without Saving",
				widget.NewLabel("Save changes before exiting?"),
				func(save bool) {
					if save {
						err := saveCollection()
						if err != nil {
							dialog.ShowError(fmt.Errorf("save error: %w", err), w)
							return
						}
					}
					w.Close()
				},
				w,
			)
		} else {
			w.Close()
		}
	})

	// Undo delete button
	undoBtn = widget.NewButton("Undo delete", func() {
		if lastDeleted == nil || appState.collection == nil {
			return
		}
		// restore collection item
		if lastDeleted.indexInCollection < 0 || lastDeleted.indexInCollection > len(appState.collection.Item) {
			return
		}
		// insert back into collection
		idxC := lastDeleted.indexInCollection
		appState.collection.Item = append(appState.collection.Item[:idxC], append([]models.Item{lastDeleted.item}, appState.collection.Item[idxC:]...)...)

		// restore forms
		if lastDeleted.indexInForms < 0 || lastDeleted.indexInForms > len(forms) {
			return
		}
		idxF := lastDeleted.indexInForms
		forms = append(forms[:idxF], append([]models.Form{lastDeleted.form}, forms[idxF:]...)...)

		// rebuild filtered based on stored filter
		if filterEntry != nil {
			filterEntry.SetText(lastDeleted.filterBefore)
		}
		// selection
		if lastDeleted.previousSelectedID != "" {
			selectedID = lastDeleted.previousSelectedID
			a.Preferences().SetString(preferenceCurrentForm, selectedID)
			tree.Select(selectedID)
		}

		appState.setModified(true)
		tree.Refresh()
		lastDeleted = nil
		undoBtn.Disable()
	})
	undoBtn.Disable()

	top := container.NewVBox(
		container.NewHBox(
			themeSelect,
			addCollectionBtn,
			confirmChk,
			saveOnDeleteChk,
			undoBtn,
			// autoSaveSelect,
			unsavedLabel,
		),
		title,
		widget.NewSeparator(),
		intro,
	)

	if len(forms) > 0 {
		currentFormID := a.Preferences().String(preferenceCurrentForm)
		if currentFormID != "" {
			// check if form exists
			found := false
			for _, f := range forms {
				if f.ID == currentFormID {
					found = true
					break
				}
			}
			if found {
				tree.Select(currentFormID)
			} else {
				tree.Select(forms[0].ID)
			}
		} else {
			tree.Select(forms[0].ID)
		}
	}

	// Create scrollable container for tree
	treeScroll := container.NewVScroll(tree)
	treeScroll.SetMinSize(fyne.NewSize(200, 400)) // Set minimum size for tree container

	// Create left menu container with filter and tree
	leftMenu := container.NewVBox(
		filterEntry,
		treeScroll,
	)
	leftMenu.Resize(fyne.NewSize(200, defaultWindowHeight)) // Set minimum size for left menu

	split := container.NewHSplit(leftMenu, container.NewBorder(top, nil, nil, nil, content))
	split.Offset = 0.2 // Adjust split offset for better proportions
	w.SetContent(split)
	w.Resize(fyne.NewSize(defaultWindowWidth, defaultWindowHeight))
	log.Info().Msg(models.LogWindowReady)

	// Start autosave (default 1 min)
	setupAutoSave(autoSaveInterval)

	w.ShowAndRun()
}

// Saves collection to file with temporary file for atomicity
func saveCollection() error {
	if appState.collection == nil || appState.collectionPath == "" {
		return fmt.Errorf("no collection loaded")
	}

	// Collection serialization
	data, err := json.MarshalIndent(appState.collection, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal collection: %w", err)
	}

	// Save to temporary file
	tmpPath := appState.collectionPath + ".tmp"
	err = os.WriteFile(tmpPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Check integrity (can add additional checks if needed)
	// Rename temporary file to main
	err = os.Rename(tmpPath, appState.collectionPath)
	if err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	appState.setModified(false)
	return nil
}

// Saves collection to new file, updates path and resets modification flag
func saveCollectionAs(parent fyne.Window) error {
	if appState.collection == nil {
		return fmt.Errorf("no collection loaded")
	}

	dlg := dialog.NewFileSave(
		func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(fmt.Errorf("file selection error: %w", err), parent)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()

			data, err := json.MarshalIndent(appState.collection, "", "  ")
			if err != nil {
				dialog.ShowError(fmt.Errorf("serialization error: %w", err), parent)
				return
			}
			_, err = writer.Write(data)
			if err != nil {
				dialog.ShowError(fmt.Errorf("file write error: %w", err), parent)
				return
			}

			appState.collectionPath = writer.URI().Path()
			appState.setModified(false)
			dialog.ShowInformation("Saved", "Collection saved successfully", parent)
		},
		parent,
	)
	dlg.SetFileName("collection.json")
	dlg.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
	dlg.Show()
	return nil
}

// Creates backup copy of file
func createBackup(filePath string) error {
	if filePath == "" {
		return nil
	}
	bakPath := filePath + ".bak"
	in, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(bakPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// Starts autosave with given interval, can be stopped via channel
func setupAutoSave(interval time.Duration) {
	if autoSaveStopCh != nil {
		close(autoSaveStopCh)
	}
	autoSaveStopCh = make(chan struct{})
	go func(stopCh chan struct{}) {
		for {
			select {
			case <-time.After(interval):
				if appState.isModified && appState.collectionPath != "" {
					err := createBackup(appState.collectionPath)
					if err != nil {
						fyne.CurrentApp().SendNotification(&fyne.Notification{
							Title:   "Backup Error",
							Content: err.Error(),
						})
					}
					err = saveCollection()
					if err != nil {
						fyne.CurrentApp().SendNotification(&fyne.Notification{
							Title:   "Autosave Error",
							Content: err.Error(),
						})
					} else {
						fyne.CurrentApp().SendNotification(&fyne.Notification{
							Title:   "Autosave",
							Content: "Collection saved automatically",
						})
					}
				}
			case <-stopCh:
				return
			}
		}
	}(autoSaveStopCh)
}

// Fixed variant theme wrapper to avoid deprecated theme.DarkTheme/LightTheme
type fixedVariantTheme struct {
	base    fyne.Theme
	variant fyne.ThemeVariant
}

func (t fixedVariantTheme) Color(n fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return t.base.Color(n, t.variant)
}
func (t fixedVariantTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return t.base.Icon(n) }
func (t fixedVariantTheme) Font(s fyne.TextStyle) fyne.Resource     { return t.base.Font(s) }
func (t fixedVariantTheme) Size(n fyne.ThemeSizeName) float32       { return t.base.Size(n) }

func fixedVariant(v fyne.ThemeVariant) fyne.Theme {
	return fixedVariantTheme{base: theme.DefaultTheme(), variant: v}
}
