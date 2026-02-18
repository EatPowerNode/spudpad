package main

import (
	"fmt"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	myWindow    fyne.Window
	textEditor  *widget.Entry
	currentFile fyne.URI
	status      *widget.Label
	wrapCheck   *widget.Check
)

func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(theme.DefaultTheme()) // or theme.DarkTheme() if preferred

	myWindow = myApp.NewWindow("SpudPad - Untitled")
	myWindow.Resize(fyne.NewSize(720, 540))

	textEditor = widget.NewMultiLineEntry()
	textEditor.Wrapping = fyne.TextWrapOff // classic Notepad style (no wrap by default)
	textEditor.OnChanged = updateStatus

	status = widget.NewLabel("Ln 1, Col 1   |   0 characters")
	status.Alignment = fyne.TextAlignTrailing

	wrapCheck = widget.NewCheck("Word Wrap", func(checked bool) {
		if checked {
			textEditor.Wrapping = fyne.TextWrapWord
		} else {
			textEditor.Wrapping = fyne.TextWrapOff
		}
		textEditor.Refresh()
		updateStatus(textEditor.Text)
	})

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentCreateIcon(), newDocument),
		widget.NewToolbarAction(theme.FolderOpenIcon(), openFile),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), saveFile),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ContentCutIcon(), func() { /*textEditor.Cut()*/ }),
		widget.NewToolbarAction(theme.ContentCopyIcon(), func() { /*textEditor.Copy()*/ }),
		widget.NewToolbarAction(theme.ContentPasteIcon(), func() { /*textEditor.Paste()*/ }),
		//widget.NewToolbarSpacer(),
		//container.NewCenter(wrapCheck),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			dialog.ShowInformation("About SpudPad", "Ultra-minimal text editor\nNo Markdown, no nonsense.", myWindow)
		}),
	)

	// Layout: toolbar top, editor fill, status bottom
	content := container.NewBorder(
		toolbar,
		status,
		nil, nil,
		textEditor,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func updateStatus(_ string) {
	row := textEditor.CursorRow + 1 // 0-based to 1-based
	col := textEditor.CursorColumn + 1
	charCount := len(textEditor.Text)
	status.SetText(fmt.Sprintf("Ln %d, Col %d   |   %d characters", row, col, charCount))
}

func newDocument() {
	if textEditor.Text != "" {
		d := dialog.NewConfirm("Unsaved changes", "Discard current text?", func(ok bool) {
			if ok {
				textEditor.SetText("")
				currentFile = nil
				myWindow.SetTitle("SpudPad - Untitled")
				updateStatus("")
			}
		}, myWindow)
		d.SetConfirmImportance(widget.DangerImportance)
		d.Show()
	} else {
		textEditor.SetText("")
		currentFile = nil
		myWindow.SetTitle("SpudPad - Untitled")
		updateStatus("")
	}
}

func openFile() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}

		textEditor.SetText(string(data))
		currentFile = reader.URI()
		myWindow.SetTitle("SpudPad - " + currentFile.Name())
		updateStatus(textEditor.Text)
	}, myWindow)
}

func saveFile() {
	if currentFile == nil {
		saveAs()
		return
	}

	// Quick save
	writer, err := storage.Writer(currentFile)
	if err != nil {
		dialog.ShowError(err, myWindow)
		return
	}
	defer writer.Close()

	_, err = io.WriteString(writer, textEditor.Text)
	if err != nil {
		dialog.ShowError(err, myWindow)
		return
	}

	// Optional: show brief "Saved" in status?
	// status.SetText(status.Text + "   Saved")
	// go func() {
	// 	<-fyne.CurrentApp().Driver().RunOnMainWhenIdle(func() {
	// 		updateStatus(textEditor.Text) // restore normal status
	// 	})
	// }()
}

func saveAs() {
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		_, err = io.WriteString(writer, textEditor.Text)
		if err != nil {
			dialog.ShowError(err, myWindow)
			return
		}

		currentFile = writer.URI()
		myWindow.SetTitle("SpudPad - " + currentFile.Name())
		updateStatus(textEditor.Text)
	}, myWindow)
}
