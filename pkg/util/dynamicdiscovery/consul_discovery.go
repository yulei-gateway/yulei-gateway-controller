package dynamicdiscovery

import "fmt"

// ConsulDiscovery consul discovery service client
//TODO ADD SCHEMA VERIFY
type ConsulDiscovery struct {
	IPs           []string
	Port          uint32
	AclToken      string
	PublicKeyPath string
	CertPath      string
	noticeChan    chan string
	name          string
}

// NewConsulDiscovery create a new consul client
func NewConsulDiscovery(name string, ips []string, port uint32, token string) (*ConsulDiscovery, error) {
	if name == "" {
		return nil, fmt.Errorf("name not set for create instancd")
	}
	var instance = &ConsulDiscovery{}
	instance.IPs = ips
	instance.Port = port
	instance.noticeChan = make(chan string)
	instance.name = name
	instance.AclToken = token
	//TODO Create consul client and try connect ,if error do panic
	return instance, nil
}

//NewConsulDiscoveryWithTLS use tls connect the client
func NewConsulDiscoveryWithTLS(name string, ips []string, port uint32, token string) (*ConsulDiscovery, error) {
	if name == "" {
		return nil, fmt.Errorf("name not set for create instancd")
	}
	var instance = &ConsulDiscovery{}
	instance.IPs = ips
	instance.Port = port
	instance.noticeChan = make(chan string)
	instance.name = name
	//TODO Create consul client and try connect ,if error do panic
	return instance, nil
}

// GetEndpoint get the cluster endpoints
func (c *ConsulDiscovery) GetEndpoint(clusterName string) ([]string, error) {
	return nil, nil
}

// IsSupportNotice get the discovery server is support watch event
// TODO: the code impl in https://juejin.cn/post/6984378158347157512
func (c *ConsulDiscovery) IsSupportNotice() bool {
	return true
}

// GetChangeNotice get the watch chan
func (c *ConsulDiscovery) GetChangeNotice() chan string {
	return c.noticeChan
}

func (c *ConsulDiscovery) Registry() {
	RegistryDynamicDiscoveryClient(c.name, c)
}
