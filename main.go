package main

import (
	"io"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

type config struct {
	Editor      *widget.Entry
	Previwer    *widget.RichText
	CurrentFile fyne.URI
	SaveMenu    *fyne.MenuItem
}

var cfg config
var filter = storage.NewExtensionFileFilter([]string{".md", ".MD"})

func main() {
	a := app.New()

	window := a.NewWindow("Markedly")

	edit, preview := cfg.makeUI()

	cfg.createMenuItems(window)

	window.SetContent(container.NewHSplit(edit, preview))
	window.Resize(fyne.Size{Width: 800, Height: 600})
	window.CenterOnScreen()

	window.ShowAndRun()
}

func (app *config) makeUI() (*widget.Entry, *widget.RichText) {
	edit := widget.NewMultiLineEntry()
	preview := widget.NewRichTextFromMarkdown("")
	app.Editor = edit
	app.Previwer = preview

	edit.OnChanged = preview.ParseMarkdown

	return edit, preview
}

func (app *config) createMenuItems(window fyne.Window) {
	openMenu := fyne.NewMenuItem("Open...", app.openFunc(window))
	saveMenu := fyne.NewMenuItem("Save", app.saveFunc(window))
	app.SaveMenu = saveMenu
	app.SaveMenu.Disabled = true
	saveAsMenu := fyne.NewMenuItem("SaveAs..", app.saveAsFunc(window))

	fileMenu := fyne.NewMenu("File", openMenu, saveMenu, saveAsMenu)

	menu := fyne.NewMainMenu(fileMenu)

	window.SetMainMenu(menu)
}

func (app *config) saveAsFunc(win fyne.Window) func() {
	return func() {
		saveDialog := dialog.NewFileSave(func(write fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
			}

			if write == nil {
				return
			}

			if !strings.HasSuffix(strings.ToLower(write.URI().Name()), ".md") {
				dialog.ShowInformation("Error", "File extension must be .MD", win)
				return
			}

			write.Write([]byte(app.Editor.Text))
			app.CurrentFile = write.URI()

			defer write.Close()

			win.SetTitle(win.Title() + " - " + write.URI().Name())

			app.SaveMenu.Disabled = false
		}, win)

		saveDialog.SetFileName("untitled.md")
		saveDialog.SetFilter(filter)
		saveDialog.Show()
	}
}

func (app *config) openFunc(win fyne.Window) func() {
	return func() {
		openDialog := dialog.NewFileOpen(func(read fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}

			if read == nil {
				return
			}

			defer read.Close()

			data, err := io.ReadAll(read)

			if err != nil {
				dialog.ShowError(err, win)
				return
			}

			app.Editor.SetText(string(data))

			win.SetTitle(win.Title() + " - " + read.URI().Name())

			app.SaveMenu.Disabled = false
		}, win)

		openDialog.SetFilter(filter)
		openDialog.Show()
	}
}

func (app *config) saveFunc(win fyne.Window) func() {
	return func() {
		if app.CurrentFile == nil {
			app.saveAsFunc(win)()
			return
		}

		write, err := storage.Writer(app.CurrentFile)

		if err != nil {
			dialog.ShowError(err, win)
			return
		}

		write.Write([]byte(app.Editor.Text))
		defer write.Close()
	}
}
