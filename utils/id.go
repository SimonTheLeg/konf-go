package utils

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// ID unifies ID and File management that konf uses
// Currently an ID is defined by the context and clustername of the config, separated by an underscore
// I have chosen this combination as it is fairly unique among multiple configs. I decided against using just context.name as a lot of times the context is just called "default", which results in lots of naming collisions

func IDFromFileInfo(fi fs.FileInfo) string {
	return strings.TrimSuffix(fi.Name(), filepath.Ext(fi.Name()))
}

func StorePathForID(id string) string {
	return genIDPath(viper.GetString("storeDir"), id)
}

func ActivePathForID(id string) string {
	return genIDPath(viper.GetString("activeDir"), id)
}

func genIDPath(path, id string) string {
	return path + "/" + id + ".yaml"
}
