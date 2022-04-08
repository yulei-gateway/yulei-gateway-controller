package listener

type ProtocolType string

const (
	HTTP  ProtocolType = "HTTP"
	TCP   ProtocolType = "TCP"
	UDP   ProtocolType = "UDP"
	MTLS  ProtocolType = "MTLS"
	HTTPS ProtocolType = "HTTPS"
)

type Listener struct {
	NodeID      string
	Metadata    map[string]string
	Port        uint32
	BindAddress string
	Protocol    ProtocolType
	Name        string
}
