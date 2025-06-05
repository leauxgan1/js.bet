package assets

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Assets struct {
	IconsSvgs map[string]string
}

func New() Assets {
	newAssets := Assets {
		IconsSvgs: make(map[string]string),
	}
	return newAssets
}

func (a *Assets) ReadIcons(iconsDirPath string) {
	icons, err := os.ReadDir(iconsDirPath)
	if err != nil {
		log.Panic(err)
	}

	for i := range(len(icons)) {
		icon := icons[i]
		name := icon.Name()
		cleanName := strings.TrimSuffix(name,".svg")
		path := fmt.Sprintf("../static/icons/%s",name)
		bytes,err := os.ReadFile(path)
		if err != nil {
			log.Panic(err)
		}
		a.IconsSvgs[cleanName] = string(bytes)
	}
}

func (a Assets) GetIcon(name string) string {
	return a.IconsSvgs[name]
}
