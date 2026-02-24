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

		// Cut
		widget.NewToolbarAction(theme.ContentCutIcon(), func() {
			selected := textEditor.SelectedText()
			if selected == "" {
				return
			}

			allRunes := []rune(textEditor.Text)
			selRuneLen := len([]rune(selected))

			selectionEnd := textEditor.CursorColumn
			selectionStart := selectionEnd - selRuneLen

			if selectionStart < 0 || selectionStart > len(allRunes) {
				fmt.Printf("Cut: invalid selection bounds (start=%d, end=%d, textLen=%d)\n",
					selectionStart, selectionEnd, len(allRunes))
				selectionStart = 0
				selectionEnd = len(allRunes)
			}

			fmt.Printf("Cut debug: start=%d end=%d selected=%q\n", selectionStart, selectionEnd, selected)

			newRunes := append(allRunes[:selectionStart], allRunes[selectionEnd:]...)

			myApp.Clipboard().SetContent(selected)
			textEditor.SetText(string(newRunes))

			textEditor.CursorColumn = selectionStart
			textEditor.Refresh()
			updateStatus(textEditor.Text)
		}),

		// Copy
		widget.NewToolbarAction(theme.ContentCopyIcon(), func() {
			selected := textEditor.SelectedText()
			if selected != "" {
				myApp.Clipboard().SetContent(selected)
			}
		}),

		// Paste
		widget.NewToolbarAction(theme.ContentPasteIcon(), func() {
			clip := myApp.Clipboard().Content()
			if clip == "" {
				return
			}

			runes := []rune(textEditor.Text)

			selected := textEditor.SelectedText()
			selRuneLen := len([]rune(selected))

			start := textEditor.CursorColumn - selRuneLen

			if start < 0 {
				start = 0
			}

			if start > len(runes) {
				start = len(runes)
			}

			clipRunes := []rune(clip)

			newRunes := make([]rune, 0, len(runes)+len(clipRunes))
			newRunes = append(newRunes, runes[:start]...)
			newRunes = append(newRunes, clipRunes...)
			newRunes = append(newRunes, runes[textEditor.CursorColumn:]...)

			newText := string(newRunes)

			textEditor.SetText(newText)

			textEditor.CursorColumn = start + len(clipRunes)
			textEditor.Refresh()
			updateStatus(textEditor.Text)
		}),

		widget.NewToolbarSpacer(),

		widget.NewToolbarAction(theme.HelpIcon(), func() {
			dialog.ShowInformation("About SpudPad", "Ultra-minimal text editor\nNo Markdown, no nonsense.", myWindow)
		}),
	)

	// Scroll container around the editor
	editorScroll := container.NewVScroll(textEditor)
	editorScroll.SetMinSize(fyne.NewSize(0, 0))

	// Bottom row: word wrap checkbox on left, status on right
	bottomRow := container.NewBorder(
		nil, nil,
		wrapCheck, // left
		status,    // right
	)

	// Main layout
	content := container.NewBorder(
		toolbar,   // top
		bottomRow, // bottom
		nil, nil,
		editorScroll, // center
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

func updateStatus(_ string) {
	row := textEditor.CursorRow + 1
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
