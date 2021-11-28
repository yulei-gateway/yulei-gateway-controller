package dynamicdiscovery

// DiscoveryClient this is a dynamic discovery cluster  endpoint support interface, impl this and do registry
type DiscoveryClient interface {

	// GetEndpoint get the cluster endpoints
	GetEndpoint(clusterName string) ([]string, error)

	// IsSupportNotice get the discovery server is support watch event
	IsSupportNotice() bool
	// GetChangeNotice get the watch chan
	GetChangeNotice() chan string

	Registry()
}

var dynamicDiscoveries = map[string]DiscoveryClient{}

func RegistryDynamicDiscoveryClient(name string, impl DiscoveryClient) {
	dynamicDiscoveries[name] = impl
}

func GetAllDynamicDiscoveryClient() map[string]DiscoveryClient {
	return dynamicDiscoveries
}

func GetDynamicDiscoveryClient(name string) DiscoveryClient {
	value, ok := dynamicDiscoveries[name]
	if ok {
		return value
	}
	return nil
}
