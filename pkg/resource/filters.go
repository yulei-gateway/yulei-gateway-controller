package resource

import (
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	cors "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/cors/v3"
	fault "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/fault/v3"
	grpcstats "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/grpc_stats/v3"
	grpcweb "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/grpc_web/v3"
	lua "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/lua/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	httpwasm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/wasm/v3"
	httpinspector "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/http_inspector/v3"
	originaldst "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/original_dst/v3"
	originalsrc "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/original_src/v3"
	tlsinspector "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	wasm "github.com/envoyproxy/go-control-plane/envoy/extensions/wasm/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	OriginalSrcFilterName = "envoy.filters.listener.original_src"
	// Alpn HTTP filter name which will override the ALPN for upstream TLS connection.
	AlpnFilterName = "istio.alpn"

	TLSTransportProtocol       = "tls"
	RawBufferTransportProtocol = "raw_buffer"

	MxFilterName          = "istio.metadata_exchange"
	StatsFilterName       = "istio.stats"
	StackdriverFilterName = "istio.stackdriver"
)

type RouterFilterContext struct {
	StartChildSpan bool
}

// Define static filters to be reused across the codebase. This avoids duplicate marshaling/unmarshaling
// This should not be used for filters that will be mutated
var (
	Router = &hcm.HttpFilter{
		Name: wellknown.Router,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: MessageToAny(&router.Router{}),
		},
	}
	TLSInspector = &listener.ListenerFilter{
		Name: wellknown.TlsInspector,
		ConfigType: &listener.ListenerFilter_TypedConfig{
			TypedConfig: MessageToAny(&tlsinspector.TlsInspector{}),
		},
	}
	HTTPInspector = &listener.ListenerFilter{
		Name: wellknown.HttpInspector,
		ConfigType: &listener.ListenerFilter_TypedConfig{
			TypedConfig: MessageToAny(&httpinspector.HttpInspector{}),
		},
	}
	OriginalDestination = &listener.ListenerFilter{
		Name: wellknown.OriginalDestination,
		ConfigType: &listener.ListenerFilter_TypedConfig{
			TypedConfig: MessageToAny(&originaldst.OriginalDst{}),
		},
	}
	OriginalSrc = &listener.ListenerFilter{
		Name: OriginalSrcFilterName,
		ConfigType: &listener.ListenerFilter_TypedConfig{
			TypedConfig: MessageToAny(&originalsrc.OriginalSrc{
				Mark: 1337,
			}),
		},
	}
	//TODO ALPN Support
	//Alpn = &hcm.HttpFilter{
	//	Name: AlpnFilterName,
	//	ConfigType: &hcm.HttpFilter_TypedConfig{
	//		TypedConfig: MessageToAny(&alpn.FilterConfig{
	//			AlpnOverride: []*alpn.FilterConfig_AlpnOverride{
	//				{
	//					UpstreamProtocol: alpn.FilterConfig_HTTP10,
	//					AlpnOverride:     mtlsHTTP10ALPN,
	//				},
	//				{
	//					UpstreamProtocol: alpn.FilterConfig_HTTP11,
	//					AlpnOverride:     mtlsHTTP11ALPN,
	//				},
	//				{
	//					UpstreamProtocol: alpn.FilterConfig_HTTP2,
	//					AlpnOverride:     mtlsHTTP2ALPN,
	//				},
	//			},
	//		}),
	//	},
	//}

	//tcpMx = MessageToAny(&metadata_exchange.MetadataExchange{Protocol: "istio-peer-exchange"})
	//
	//TCPListenerMx = &listener.Filter{
	//	Name:       MxFilterName,
	//	ConfigType: &listener.Filter_TypedConfig{TypedConfig: tcpMx},
	//}
	//
	//TCPClusterMx = &cluster.Filter{
	//	Name:        MxFilterName,
	//	TypedConfig: tcpMx,
	//}

	HTTPMx = buildHTTPMxFilter()
)

func BuildRouterFilter(ctx *RouterFilterContext) *hcm.HttpFilter {
	if ctx == nil {
		return Router
	}

	return &hcm.HttpFilter{
		Name: wellknown.Router,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: MessageToAny(&router.Router{
				StartChildSpan: ctx.StartChildSpan,
			}),
		},
	}
}

// this not support istio extensions filter
func buildHTTPMxFilter() *hcm.HttpFilter {
	httpMxConfigProto := &httpwasm.Wasm{
		Config: &wasm.PluginConfig{
			Vm: ConstructVMConfig("/etc/istio/extensions/metadata-exchange-filter.compiled.wasm", "envoy.wasm.metadata_exchange"),
			//	Configuration: MessageToAny(&metadata_exchange.MetadataExchange{}),
		},
	}
	return &hcm.HttpFilter{
		Name:       MxFilterName,
		ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: MessageToAny(httpMxConfigProto)},
	}
}

//https://github.com/api7/envoy-apisix
//https://github.com/envoyproxy/envoy/tree/4b5b3386c6b0d2d284bb1f71639c8e0972659867/examples/lua
//https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/lua_filter
// ConstructVMConfig constructs a VM config. If WASM is enabled, the wasm plugin at filename will be used.
// If not, the builtin (null vm) extension, name, will be used.
func ConstructVMConfig(filename, name string) *wasm.PluginConfig_VmConfig {
	var vmConfig *wasm.PluginConfig_VmConfig
	if filename != "" {
		vmConfig = &wasm.PluginConfig_VmConfig{
			VmConfig: &wasm.VmConfig{
				Runtime:          "envoy.wasm.runtime.v8",
				AllowPrecompiled: true,
				Code: &core.AsyncDataSource{Specifier: &core.AsyncDataSource_Local{
					Local: &core.DataSource{
						Specifier: &core.DataSource_Filename{
							Filename: filename,
						},
					},
				}},
			},
		}
	} else {
		vmConfig = &wasm.PluginConfig_VmConfig{
			VmConfig: &wasm.VmConfig{
				Runtime: "envoy.wasm.runtime.null",
				Code: &core.AsyncDataSource{Specifier: &core.AsyncDataSource_Local{
					Local: &core.DataSource{
						Specifier: &core.DataSource_InlineString{
							InlineString: name,
						},
					},
				}},
			},
		}
	}
	return vmConfig
}

// MessageToAnyWithError converts from proto message to proto Any
func MessageToAnyWithError(msg proto.Message) (*anypb.Any, error) {
	b, err := proto.MarshalOptions{Deterministic: true}.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return &anypb.Any{
		// nolint: staticcheck
		TypeUrl: "type.googleapis.com/" + string(msg.ProtoReflect().Descriptor().FullName()),
		Value:   b,
	}, nil
}

// MessageToAny converts from proto message to proto Any
func MessageToAny(msg proto.Message) *anypb.Any {
	out, err := MessageToAnyWithError(msg)
	if err != nil {
		//log.Error(fmt.Sprintf("error marshaling Any %s: %v", prototext.Format(msg), err))
		return nil
	}
	return out
}

type ListenerFilter interface {
	GetName() string
	Build() *listener.ListenerFilter
}

type HttpFilter interface {
	GetName() string
	Build() *hcm.HttpFilter
}

type WASMFilter struct {
	Name     string
	FilePath string
	RootId   string
	FailOpen bool
}

func (f *WASMFilter) GetName() string {
	return f.Name
}

func (f *WASMFilter) Build() *hcm.HttpFilter {
	httpMxConfigProto := &httpwasm.Wasm{
		Config: &wasm.PluginConfig{
			Vm: ConstructVMConfig(f.FilePath, "envoy.wasm.metadata_exchange"),
			//TODO wasm plugin configuration support
			//Configuration: MessageToAny(&metadata_exchange.MetadataExchange{}),
			Name:     f.Name,
			RootId:   f.RootId,
			FailOpen: f.FailOpen,
		},
	}
	return &hcm.HttpFilter{
		Name:       f.Name,
		ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: MessageToAny(httpMxConfigProto)},
	}
}

type LuaFilter struct {
	InlineCode    string
	LuaDataSource []LuaDataSource
}

type LuaDataSource struct {
	Name           string
	DataSourceType string
	Value          string
}

func (f *LuaFilter) GetName() string {
	return wellknown.Lua
}

func (f *LuaFilter) Build() *hcm.HttpFilter {
	var sourceCodes = map[string]*core.DataSource{}
	for _, item := range f.LuaDataSource {
		switch item.DataSourceType {
		case "Filename":
			sourceCodes[item.Name] = &core.DataSource{
				Specifier: &core.DataSource_Filename{
					Filename: item.Value,
				},
			}
		case "InlineBytes":
			sourceCodes[item.Name] = &core.DataSource{
				Specifier: &core.DataSource_InlineBytes{
					InlineBytes: []byte(item.Value),
				},
			}
		case "InlineString":
			sourceCodes[item.Name] = &core.DataSource{
				Specifier: &core.DataSource_InlineString{
					InlineString: item.Value,
				},
			}
		}
	}
	return &hcm.HttpFilter{
		Name: wellknown.Lua,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: MessageToAny(&lua.Lua{
				InlineCode:  f.InlineCode,
				SourceCodes: sourceCodes,
			}),
		},
	}
}

type GrpcStatsFilter struct {
}

func (f *GrpcStatsFilter) GetName() string {
	return wellknown.HTTPGRPCStats
}

func (f *GrpcStatsFilter) Build() *hcm.HttpFilter {
	return &hcm.HttpFilter{
		Name: wellknown.HTTPGRPCStats,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: MessageToAny(&grpcstats.FilterConfig{
				EmitFilterState: true,
				PerMethodStatSpecifier: &grpcstats.FilterConfig_StatsForAllMethods{
					StatsForAllMethods: &wrapperspb.BoolValue{Value: false},
				},
			}),
		},
	}
}

type GrpcWebFilter struct {
}

func (f *GrpcWebFilter) GetName() string {
	return wellknown.GRPCWeb
}

func (f *GrpcWebFilter) Build() *hcm.HttpFilter {
	return &hcm.HttpFilter{
		Name: wellknown.GRPCWeb,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: MessageToAny(&grpcweb.GrpcWeb{}),
		},
	}
}

//RouterFilter this is all default filter
type RouterFilter struct {
	// Specifies a list of HTTP headers to strictly validate. Envoy will reject a
	// request and respond with HTTP status 400 if the request contains an invalid
	// value for any of the headers listed in this field. Strict header checking
	// is only supported for the following headers:
	//
	// Value must be a ','-delimited list (i.e. no spaces) of supported retry
	// policy values:
	//
	// * :ref:`config_http_filters_router_x-envoy-retry-grpc-on`
	// * :ref:`config_http_filters_router_x-envoy-retry-on`
	//
	// Value must be an integer:
	//
	// * :ref:`config_http_filters_router_x-envoy-max-retries`
	// * :ref:`config_http_filters_router_x-envoy-upstream-rq-timeout-ms`
	// * :ref:`config_http_filters_router_x-envoy-upstream-rq-per-try-timeout-ms`
	StrictCheckHeaders []string
	// Defaults to true
	DynamicStats             bool
	StartChildSpan           bool
	UpstreamLog              *AccessLog
	SuppressEnvoyHeaders     bool
	RespectExpectedRqTimeout bool
}

type AccessLog struct {
	Name string
	// TODO this build have some problem
	// StatusCodeFilter
	FilterType string
	// EQ 0
	// GE 1
	// LE 2
	Op int
	// AccessLog_TypedConfig or AccessLog_HiddenEnvoyDeprecatedConfig
	ConfigType string
}

func (f *RouterFilter) GetName() string {
	return wellknown.Router
}

func (f *RouterFilter) Build() *hcm.HttpFilter {
	var dynamicStats = true
	if !f.DynamicStats {
		dynamicStats = f.DynamicStats
	}
	//TODO AccessLog AND StrictCheckHeaders logic
	return &hcm.HttpFilter{
		Name: wellknown.Router,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: MessageToAny(&router.Router{
				DynamicStats:             &wrappers.BoolValue{Value: dynamicStats},
				StrictCheckHeaders:       f.StrictCheckHeaders,
				StartChildSpan:           f.StartChildSpan,
				SuppressEnvoyHeaders:     f.SuppressEnvoyHeaders,
				RespectExpectedRqTimeout: f.RespectExpectedRqTimeout,
			}),
		},
	}
}

type FaultFilter struct {
	Delay FaultDelay
	Abort FaultAbort

	UpstreamCluster                 string
	Headers                         HeaderRoute
	DownstreamNodes                 []string
	MaxActiveFaults                 *uint32
	ResponseRateLimit               *FaultRateLimit
	DelayPercentRuntime             string
	AbortPercentRuntime             string
	DelayDurationRuntime            string
	AbortHttpStatusRuntime          string
	MaxActiveFaultsRuntime          string
	ResponseRateLimitPercentRuntime string
	AbortGrpcStatusRuntime          string
	DisableDownstreamClusterStats   bool
}
type FaultRateLimit struct {
	// FixedLimit or HeaderLimit
	LimitType string

	// if type is FixedLimit use this
	LimitKbps uint64

	Percentage uint32
	// 1/100 = 1%. 0
	// 1/10000 = 0.01%. 1
	// 1/1000000 = 0.0001%. 2
	DenominatorType int
}

type FaultAbort struct {
	// HttpStatus or GrpcStatus or HeaderAbort
	ErrorType string
	// if type is httpStatus is set value to httpStatus if GrpcStatus setValue to GrpcStatus if HeaderAbort ignore
	Status uint32

	Percentage uint32
	// 1/100 = 1%. 0
	// 1/10000 = 0.01%. 1
	// 1/1000000 = 0.0001%. 2
	DenominatorType int
}
type FaultDelay struct {
	//FixedDelay or HeaderDelay
	Type       string
	FixedDelay time.Duration
}

func (f *FaultFilter) GetName() string {
	return wellknown.Fault
}

func (f *FaultFilter) Build() *hcm.HttpFilter {
	//TODO build fault.HTTPFault instance from FaultFilter
	return &hcm.HttpFilter{
		Name: wellknown.Fault,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: MessageToAny(&fault.HTTPFault{}),
		},
	}
}

type CORSFilter struct {
}

func (f *CORSFilter) GetName() string {
	return wellknown.CORS
}

func (f *CORSFilter) Build() *hcm.HttpFilter {
	return &hcm.HttpFilter{
		Name: wellknown.CORS,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: MessageToAny(&cors.Cors{}),
		},
	}
}
