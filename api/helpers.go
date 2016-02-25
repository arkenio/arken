package api
import (
	"github.com/coreos/go-etcd/etcd"
	"github.com/arkenio/goarken"
	"github.com/spf13/viper"
)



func CreateEtcdClient() *etcd.Client {
	etcdAdress := viper.GetString("etcdAddress")
	return etcd.NewClient([]string{etcdAdress})
}


func CreateWatcherFromCli(client *etcd.Client) *goarken.Watcher {
	domainDir := viper.GetString("domainDir")
	serviceDir := viper.GetString("serviceDir")
	w := &goarken.Watcher{
		Client:        client,
		DomainPrefix:  domainDir,
		ServicePrefix: serviceDir,
		Domains:       make(map[string]*goarken.Domain),
		Services:      make(map[string]*goarken.ServiceCluster),
	}
	w.Init()
	return w
}

