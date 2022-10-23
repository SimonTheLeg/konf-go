package konf

import (
	"io"

	k8s "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/yaml"
)

// KonfsFromKubeconfig takes in the content of a kubeconfig and splits it into
// one or multiple konfs.
//
// No error is being returned if the kubeconfig contains no contexts, instead
// konfs is simply an empty slice
func KonfsFromKubeconfig(kubeconfig io.Reader) (konfs []*Konfig, err error) {
	konfs = []*Konfig{}

	b, err := io.ReadAll(kubeconfig)
	if err != nil {
		return nil, err
	}

	var origConf k8s.Config
	err = yaml.Unmarshal(b, &origConf)
	if err != nil {
		return nil, err
	}

	// basically should be as simple as
	// 1. Loop through all the contexts
	// 2. Find the corresponding cluster for each context
	// 3. Find the corresponding user for each context
	// 4. Create a new konfigFile for each context mapped to its cluster

	for _, curCon := range origConf.Contexts {

		cluster := k8s.NamedCluster{}
		for _, curCl := range origConf.Clusters {
			if curCl.Name == curCon.Context.Cluster {
				cluster = curCl
				break
			}
		}
		user := k8s.NamedAuthInfo{}
		for _, curU := range origConf.AuthInfos {
			if curU.Name == curCon.Context.AuthInfo {
				user = curU
				break
			}
		}

		var k Konfig
		id := IDFromClusterAndContext(cluster.Name, curCon.Name)
		// TODO need to remove this. StorePath should only be setable by store pkg later on
		k.StorePath = id.StorePath()
		k.Id = id
		k.Kubeconfig.AuthInfos = append(k.Kubeconfig.AuthInfos, user)
		k.Kubeconfig.Clusters = append(k.Kubeconfig.Clusters, cluster)
		k.Kubeconfig.Contexts = append(k.Kubeconfig.Contexts, curCon)

		k.Kubeconfig.APIVersion = origConf.APIVersion
		k.Kubeconfig.Kind = origConf.Kind
		k.Kubeconfig.CurrentContext = curCon.Name

		konfs = append(konfs, &k)
	}

	return konfs, nil
}
