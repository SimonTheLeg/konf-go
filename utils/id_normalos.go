// +build !windows

package utils

func IDFromClusterAndContext(cluster, context string) string {
	return context + "_" + cluster
}
