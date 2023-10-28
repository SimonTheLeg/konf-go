package konf

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/simontheleg/konf-go/config"
)

// ID unifies ID management that konf uses
// Currently an ID is defined by the context and clustername of the config, separated by an underscore
// I have chosen this combination as it is fairly unique among multiple configs. I decided against using just context.name as a lot of times the context is just called "default", which results in lots of naming collisions
// Some special characters that are reserved by the filesystem, will be replaced by a "-" character
type KonfID string

// IDFromClusterAndContext creates an id based on the cluster and context
// It escapes any illegal file characters and is filesafe
func IDFromClusterAndContext(cluster, context string) KonfID {
	id := context + "_" + cluster

	illegalChars := []string{"/", ":"}
	for _, c := range illegalChars {
		id = strings.ReplaceAll(id, c, "-")
	}

	return KonfID(id)
}

// IDFromProcessID creates a KonfID based on the supplied processID
func IDFromProcessID(pid int) KonfID {
	// since the pid is an int, no illegal character replacement is needed
	return KonfID(fmt.Sprint(pid))
}

// IDFromFileInfo creates an ID from the name of a file
func IDFromFileInfo(fi fs.FileInfo) KonfID {
	return KonfID(strings.TrimSuffix(fi.Name(), filepath.Ext(fi.Name())))
}

// StorePathForID creates a valid filepath inside the configured storeDir
func (id KonfID) StorePath() string {
	return genIDPath(config.StoreDir(), string(id))
}

// ActivePath creates a valid filepath inside the configured activeDir
func (id KonfID) ActivePath() string {
	return genIDPath(config.ActiveDir(), string(id))
}

func genIDPath(path, id string) string {
	return path + "/" + id + ".yaml"
}
