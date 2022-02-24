package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"github.com/ck3-mods/yet-another-launcher/mod_manager"
)

func main() {
	a := app.New()
	w := a.NewWindow("Yet Another Launcher")

	// path := mod_manager.Ck3Paths()

	var content fyne.Widget
	modListWidget, err := mod_manager.ModListWidget()
	if err != nil {
		content = widget.NewLabel(err.Error())
	} else {
		content = modListWidget
	}
	w.SetContent(content)
	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}
