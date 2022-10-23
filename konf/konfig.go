package konf

import (
	k8s "k8s.io/client-go/tools/clientcmd/api/v1"
)

type Config struct {
	Id         KonfID
	Kubeconfig k8s.Config
	StorePath  string
}
