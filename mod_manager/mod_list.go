package mod_manager

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"github.com/ck3-mods/yet-another-launcher/ck3_parser"
)

type Mod struct {
	description string
	image       fyne.Resource
	modType     string
	name        string
	path        fyne.URI
}

// getModData retrieves a mod folder's data, such as the name of the mod, it's picture if any, the mod type, etc.
func getModData(modUri fyne.URI) (mod Mod, err error) {
	// Get list of files/dir in the mod directory
	modListable, err := storage.ListerForURI(modUri)
	// fmt.Printf("getModData() \n URI = %v\nScheme = %v\nMimeType = %v\nExtension = %v\n", modUri, modUri.Scheme(), modUri.MimeType(), modUri.Extension())
	if err != nil {
		return
	}
	modList, err := modListable.List()
	if err != nil {
		return
	}
	// Get mod data
	mod.name = modListable.Name()
	mod.path = modUri
	// mod.imageUri = path/to/placeholder
	// Get the first png file found, if any
	for _, modFile := range modList {
		if modFile.Extension() == ".png" {
			mod.image, err = storage.LoadResourceFromURI(modFile)
			if err != nil {
				// mod.image = path/to/placeholder
			}
			break
		}
	}
	return
}

func getModFolders(modsFolderUri fyne.URI) (modFolders []fyne.URI, err error) {
	modsFolderListable, err := storage.ListerForURI(modsFolderUri)
	if err != nil {
		fmt.Printf("ModList Error: URI = %v\n", modsFolderUri)
		return
	}
	// modsFolderFiles, err := modsFolderListable.List()
	modFolders, err = modsFolderListable.List()
	if err != nil {
		return
	}
	// for _, file := range modsFolderFiles {
	// 	fileInfo, _ := os.Stat(file.Path())
	// 	if fileInfo.IsDir() {
	// 		modFolders = append(modFolders, file)
	// 	}
	// }
	return
}

func ModList(modFolderUri fyne.URI) (modList []Mod, err error) {
	modFolders, err := getModFolders(modFolderUri)
	for _, modFolder := range modFolders {
		if modFolder.Extension() == ".mod" {
			file, _ := os.Open(modFolder.Path())
			tokens := ck3_parser.Lex(file)
			fmt.Printf("ModList token : %v\n", <-tokens)
		}
		modData, modDataErr := getModData(modFolder)
		if modDataErr != nil {
			// We just ignore the data on error
			continue
		}
		modList = append(modList, modData)
	}
	return
}
