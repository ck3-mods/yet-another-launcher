package mod_manager

import (
	"log"
	"os"

	"fyne.io/fyne/v2/storage"
)

func Ck3Paths() (paths ck3Paths, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
		return
	}
	steamDir := homeDir + "/Library/Application Support/Steam"

	paths = ck3Paths{
		steamRoot: storage.NewFileURI(steamDir),
		game:      storage.NewFileURI(steamDir + "/steamapps/common/Crusader Kings III"),
		steamMods: storage.NewFileURI(steamDir + "/steamapps/workshop/content/1158310/"),
		user:      storage.NewFileURI(homeDir + "/Documents/Paradox Interactive/Crusader Kings III"),
		userMods:  storage.NewFileURI(homeDir + "/Documents/Paradox Interactive/Crusader Kings III/mod"),
	}
	return
}
