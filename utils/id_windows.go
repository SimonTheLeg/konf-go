// +build windows

package utils

// As windows has the great limitation that a filepath can only be 260 characters long,
// it makes a bit of sense to do a little trimming on the ID

func IDFromClusterAndContext(cluster, context string) string {
	id := context + "_" + cluster
	if len(id) > 120 { // tbf at this point 120 is an arbitrary number
		id = id[:120] // since we are dealing with filenames which are UTF-8 this should be fine
	}
	return id
}
