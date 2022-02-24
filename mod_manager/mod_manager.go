package mod_manager

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func modLayout() fyne.CanvasObject {
	icon := &canvas.Image{}
	label := widget.NewLabel("Text Editor")
	icon.ScaleMode = canvas.ImageScaleFastest
	icon.SetMinSize(fyne.NewSize(70, 70))
	return container.NewHBox(icon, label)
}

// ModListWidget returns a list widget with all mods in a folder.
// TODO: improve the performance of the list by caching the thumbnail images, or building a mod index in a db
func ModListWidget() (modListWidget *widget.List, err error) {
	ck3Paths, err := Ck3Paths()
	if err != nil {
		return
	}
	modList, err := ModList(ck3Paths.userMods)
	if err != nil {
		return
	}
	modListWidget = widget.NewList(
		func() int {
			return len(modList)
		},
		func() fyne.CanvasObject {
			return modLayout()
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			// Load image if the URI exists
			if modList[i].image != nil {
				modImage := o.(*fyne.Container).Objects[0].(*canvas.Image)
				modImage.Resource = modList[i].image
				if err != nil {
					fmt.Printf("Error loading image: %v", err)
				}
				modImage.Refresh()
			}
			modName := o.(*fyne.Container).Objects[1].(*widget.Label)
			modName.SetText(modList[i].name)
		})
	return
}
