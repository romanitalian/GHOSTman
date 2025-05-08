package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	preferenceCurrentForm = "currentForm"
	minURLPathLength      = 2
	defaultSplitOffset    = 0.2
	responseHeightRatio   = 0.3
	defaultWindowWidth    = 1024
	defaultWindowHeight   = 768

	logLevel      = zerolog.WarnLevel
	logFormatJSON = true

	appID    = "com.github.romanitalian.ghostman"
	appTitle = "GHOSTman"
)

var (
	topWindow   fyne.Window
	httpMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "CONNECT", "TRACE"}
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

func substituteVariables(s string, vars map[string]string) string {
	for k, v := range vars {
		placeholder := "{{" + k + "}}"
		s = strings.ReplaceAll(s, placeholder, v)
	}
	return s
}

func createForm(item struct {
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
}, vars map[string]string) fyne.CanvasObject {
	// Create form fields
	frm := &widget.Form{}

	// Add request info fields
	urlEntry := widget.NewEntry()
	urlEntry.SetText(substituteVariables(item.Request.URL.Raw, vars))
	frm.Append("URL", urlEntry)

	methodSelect := widget.NewSelect(httpMethods, func(value string) {})
	methodSelect.SetSelected(item.Request.Method)
	frm.Append("Method", methodSelect)

	var headersText strings.Builder
	for _, h := range item.Request.Header {
		headersText.WriteString(fmt.Sprintf("%s: %s\n", h.Key, substituteVariables(h.Value, vars)))
	}
	hdrsEntry := widget.NewMultiLineEntry()
	hdrsEntry.SetText(headersText.String())
	frm.Append("Headers", hdrsEntry)

	// Create body field with fixed height
	bodyEntry := widget.NewMultiLineEntry()
	bodyEntry.SetText(substituteVariables(item.Request.Body.Raw, vars))

	// Calculate number of lines in JSON
	lines := strings.Count(item.Request.Body.Raw, "\n") + 1
	bodyEntry.SetMinRowsVisible(lines)

	frm.Append("Body", bodyEntry)

	// Create response field
	textRS := widget.NewMultiLineEntry()
	textRS.Wrapping = fyne.TextWrapWord
	textRS.Disable() // Make it read-only
	textRS.TextStyle = fyne.TextStyle{
		Bold:      true,
		Monospace: true,
	}
	textRS.Resize(fyne.NewSize(200, 200))
	textRS.SetMinRowsVisible(10)

	// Add submit button
	submitBtn := widget.NewButton("Send", func() {
		// Clear response field
		textRS.SetText("")

		// Create request
		rq, err := http.NewRequest(methodSelect.Selected, urlEntry.Text, bytes.NewBufferString(bodyEntry.Text))
		if err != nil {
			textRS.SetText(fmt.Sprintf("Error creating request: %v", err))
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

		// Send request
		client := &http.Client{}
		resp, err := client.Do(rq)
		if err != nil {
			textRS.SetText(fmt.Sprintf("Error sending request: %v", err))
			return
		}
		defer resp.Body.Close()

		// Read response
		bodyRS, err := io.ReadAll(resp.Body)
		if err != nil {
			textRS.SetText(fmt.Sprintf("Error reading response: %v", err))
			return
		}

		// Format response
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, bodyRS, "", "    "); err != nil {
			textRS.SetText(string(bodyRS))
		} else {
			textRS.SetText(prettyJSON.String())
		}
	})
	frm.Append("", submitBtn)

	// Create response container
	containerRS := container.NewVBox(
		widget.NewLabel("Response"),
		textRS,
	)

	// Add response field after submit button
	frm.Append("", containerRS)

	// Create vertical container with form
	return container.NewVBox(
		frm,
	)
}

func loadPostmanCollection() (map[string]struct {
	title string
	intro string
	form  fyne.CanvasObject
}, error) {
	forms := make(map[string]struct {
		title string
		intro string
		form  fyne.CanvasObject
	})

	data, err := os.ReadFile(filepath.Join("data", "col.postman_collection.json"))
	if err != nil {
		log.Error().Err(err).Msg("error reading Postman collection")
		return nil, fmt.Errorf("error reading Postman collection: %v", err)
	}

	var collection Cllns
	if err := json.Unmarshal(data, &collection); err != nil {
		log.Error().Err(err).Msg("error parsing Postman collection")
		return nil, fmt.Errorf("error parsing Postman collection: %v", err)
	}

	// Store variables in map
	vars := make(map[string]string)
	for _, v := range collection.Variable {
		vars[v.Key] = v.Value
	}
	log.Info().Fields(vars).Msg("Loaded Postman variables")

	log.Info().Int("count", len(collection.Item)).Msg("Total items in collection")

	for i, item := range collection.Item {
		log.Info().Int("idx", i+1).Str("name", item.Name).Msg("Processing item")
		log.Info().Interface("url_path", item.Request.URL.Path).Msg("URL Path")
		if len(item.Request.URL.Path) >= minURLPathLength {
			formID := item.Request.URL.Path[1]
			log.Info().Str("form_id", formID).Msg("Form ID")

			// Create form with request info and variable substitution
			form := createForm(item, vars)

			forms[formID] = struct {
				title string
				intro string
				form  fyne.CanvasObject
			}{
				title: item.Name,
				intro: item.Request.Description,
				form:  form,
			}
			log.Info().Str("form_id", formID).Str("name", item.Name).Msg("Added form")
		} else {
			log.Warn().Str("name", item.Name).Msg("Skipping item: invalid URL path length")
		}
	}

	log.Info().Int("count", len(forms)).Msg("Total forms loaded")
	for id, form := range forms {
		log.Info().Str("form_id", id).Str("title", form.title).Msg("Loaded form")
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

	log.Info().Msg("Starting application...")

	a := app.NewWithID(appID)
	w := a.NewWindow(appTitle)
	topWindow = w

	content := container.NewStack()
	title := widget.NewLabel("Form Title")
	intro := widget.NewLabel("Form description goes here")
	intro.Wrapping = fyne.TextWrapWord

	top := container.NewVBox(title, widget.NewSeparator(), intro)

	setForm := func(form fyne.CanvasObject, formTitle string, formIntro string) {
		log.Info().Str("form_title", formTitle).Msg("Setting form")
		title.SetText(formTitle)
		intro.SetText(formIntro)
		content.Objects = []fyne.CanvasObject{form}
		content.Refresh()
	}

	forms, err := loadPostmanCollection()
	if err != nil {
		log.Error().Err(err).Msg("Error loading forms")
		return
	}

	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			if uid == "" {
				keys := make([]string, 0, len(forms))
				for k := range forms {
					keys = append(keys, k)
				}
				log.Info().Strs("keys", keys).Msg("Tree ChildUIDs called for root")
				return keys
			}
			return []string{}
		},
		IsBranch: func(uid string) bool {
			isRoot := uid == ""
			log.Debug().Str("uid", uid).Bool("is_root", isRoot).Msg("Tree IsBranch called")
			return isRoot
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			log.Debug().Bool("branch", branch).Msg("Tree CreateNode called")
			return widget.NewLabel("Form")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			if uid == "" {
				log.Debug().Msg("Tree UpdateNode called for root")
				obj.(*widget.Label).SetText("Forms")
				return
			}
			if f, ok := forms[uid]; ok {
				log.Debug().Str("uid", uid).Str("title", f.title).Msg("Tree UpdateNode called")
				obj.(*widget.Label).SetText(f.title)
			}
		},
		OnSelected: func(uid string) {
			if f, ok := forms[uid]; ok {
				log.Info().Str("uid", uid).Str("form", f.title).Msg("Tree OnSelected called")
				a.Preferences().SetString(preferenceCurrentForm, uid)
				setForm(f.form, f.title, f.intro)
			}
		},
	}

	if len(forms) > 0 {
		for k := range forms {
			log.Info().Str("form_id", k).Msg("Selecting initial form")
			tree.Select(k)
			break
		}
	}

	split := container.NewHSplit(tree, container.NewBorder(top, nil, nil, nil, content))
	split.Offset = defaultSplitOffset
	w.SetContent(split)
	w.Resize(fyne.NewSize(defaultWindowWidth, defaultWindowHeight))
	log.Info().Msg("Application window created and ready")
	w.ShowAndRun()
}
