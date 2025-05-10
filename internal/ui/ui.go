package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/romanitalian/GHOSTman/internal/collection"
	"github.com/romanitalian/GHOSTman/internal/httpclient"
)

type Form struct {
	ID    string
	Title string
	Intro string
	Form  fyne.CanvasObject
}

func CreateForm(item struct {
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
	frm := &widget.Form{}

	urlEntry := widget.NewEntry()
	urlEntry.SetText(collection.SubstituteVariables(item.Request.URL.Raw, vars))
	frm.Append("URL", urlEntry)

	methodSelect := widget.NewSelect([]string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "CONNECT", "TRACE"}, func(value string) {})
	methodSelect.SetSelected(item.Request.Method)
	frm.Append("Method", methodSelect)

	var headersText strings.Builder
	for _, h := range item.Request.Header {
		headersText.WriteString(fmt.Sprintf("%s: %s\n", h.Key, collection.SubstituteVariables(h.Value, vars)))
	}
	hdrsEntry := widget.NewMultiLineEntry()
	hdrsEntry.SetText(headersText.String())
	frm.Append("Headers", hdrsEntry)

	bodyEntry := widget.NewMultiLineEntry()
	bodyEntry.SetText(collection.SubstituteVariables(item.Request.Body.Raw, vars))
	lines := strings.Count(item.Request.Body.Raw, "\n") + 1
	bodyEntry.SetMinRowsVisible(lines)
	frm.Append("Body", bodyEntry)

	textRS := widget.NewMultiLineEntry()
	textRS.Wrapping = fyne.TextWrapWord
	textRS.TextStyle = fyne.TextStyle{
		Bold:      true,
		Monospace: true,
	}
	textRS.Resize(fyne.NewSize(200, 200))
	textRS.SetMinRowsVisible(25)

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide()

	submitBtn := widget.NewButton("Send", func() {
		textRS.SetText("")
		progressBar.Show()
		progressBar.Refresh()

		rq, err := httpclient.NewRequest(methodSelect.Selected, urlEntry.Text, bodyEntry.Text, hdrsEntry.Text)
		if err != nil {
			progressBar.Hide()
			progressBar.Refresh()
			textRS.SetText(fmt.Sprintf("Error creating request: %v", err))
			return
		}

		go func() {
			statusLine, prettyBody, _, err := httpclient.SendRequest(rq)
			fyne.Do(func() {
				progressBar.Hide()
				progressBar.Refresh()
				if err != nil {
					textRS.SetText(fmt.Sprintf("%s", err))
					return
				}
				textRS.SetText(statusLine + prettyBody)
			})
		}()
	})

	frm.Append("", submitBtn)
	frm.Append("", progressBar)

	containerRS := container.NewVBox(
		progressBar,
		textRS,
	)
	frm.Append("Response", containerRS)

	return container.NewVBox(frm)
}
