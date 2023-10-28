package store

import (
	"fmt"
)

// KubeConfigOverload describes a state in which a kubeconfig has multiple Contexts or Clusters
// This can be undesirable for konf when such a kubeconfig is in its store
type KubeConfigOverload struct {
	path string
}

func (k *KubeConfigOverload) Error() string {
	return fmt.Sprintf("Impure Store: The kubeconfig %q contains multiple contexts and/or clusters. Please only use 'konf import' for populating the store\n", k.path)
}

// EmptyStore describes a state in which no kubeconfig is inside the store
// It makes sense to have this in a separate case as it does not matter for some operations (e.g. importing) but detrimental for others (e.g. running the selection prompt)
type EmptyStore struct {
	storepath string
}

func (e *EmptyStore) Error() string {
	return fmt.Sprintf("The konf store at %q is empty. Please run 'konf import' to populate it", e.storepath)
}

// NoMatch describes a state in which no konf was found matching the supplied glob
// It makes sense to have this in a separate case as it does not matter for some operations (e.g. importing) but detrimental for others (e.g. running the selection prompt)
type NoMatch struct {
	Pattern string
}

func (k *NoMatch) Error() string {
	return fmt.Sprintf("No konf file matched your search pattern %q", k.Pattern)
}
