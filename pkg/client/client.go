package client

import (
	"fmt"
	"github.com/ergoapi/util/exmap"
	"github.com/ergoapi/util/file"
	"github.com/ergoapi/util/ztime"
	"github.com/ergoapi/zlog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"next-terminal/models"
	"next-terminal/repository"
	"sync"
)

var (
	clusterManagerSets = &sync.Map{}
)

type ClusterManager struct {
	Cluster   *models.Cluster
	Clientset *kubernetes.Clientset
	Config    *rest.Config
}

func BuildApiserverClient() {
	newCLusters, err := repository.GClusterRepository.Gets("")
	if err != nil {
		zlog.Error("list k8s client err: %v", err)
		return
	}
	changed := clusterChanged(newCLusters)
	if changed {
		zlog.Info("cluster changed, try resync.")
		// 删除已离线集群
		shouldRemoveClusters(newCLusters)
		// build
		for i := 0; i < len(newCLusters); i++ {
			cluster := newCLusters[i]
			clientSet, config, err := buildClient(cluster.Kubeconfig, cluster.ID)
			if err != nil {
				zlog.Warn("build cluster [%s(%s)] client err :%v", cluster.Name, cluster.ID, err)
				continue
			}
			c := &ClusterManager{
				Cluster:   &cluster,
				Clientset: clientSet,
				Config:    config,
			}
			clusterManagerSets.Store(cluster.ID, c)
		}
		zlog.Info("sync cluster finished.")
	}
}

func shouldRemoveClusters(changedClusters []models.Cluster) {
	changedClusterMap := make(map[string]struct{})
	for _, cluster := range changedClusters {
		changedClusterMap[cluster.Name] = struct{}{}
	}

	clusterManagerSets.Range(func(key, value interface{}) bool {
		if _, ok := changedClusterMap[key.(string)]; !ok {
			clusterManagerSets.Delete(key)
		}
		return true
	})
}

func clusterChanged(clusters []models.Cluster) bool {
	if exmap.SyncMapLen(clusterManagerSets) != len(clusters) {
		zlog.Info("cluster length (%d) changed to (%d).", exmap.SyncMapLen(clusterManagerSets), len(clusters))
		return true
	}
	for _, cluster := range clusters {
		mc, ok := clusterManagerSets.Load(cluster.ID)
		if !ok {
			// maybe add new cluster
			return true
		}
		m := mc.(*ClusterManager)
		if m.Cluster.Kubeconfig != cluster.Kubeconfig {
			zlog.Info("cluster kubeconfig (%s) changed to (%s).", m.Cluster.Kubeconfig, cluster.Kubeconfig)
			return true
		}
		if m.Cluster.Status != cluster.Status {
			zlog.Info("cluster status (%d) changed to (%d).", m.Cluster.Status, cluster.Status)
			return true
		}
	}
	return false
}

func buildClient(kubecfg, id string) (*kubernetes.Clientset, *rest.Config, error) {
	idfile := fmt.Sprintf("/tmp/%v.%v", id, ztime.NowUnixString())
	defer func() {
		file.RemoveFiles(idfile)
	}()
	file.Writefile(idfile, kubecfg)
	config, err := clientcmd.BuildConfigFromFlags("", idfile)
	if err != nil {
		return nil, nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return clientset, config, nil
}
