package resource

import (
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	cors "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/cors/v3"
	fault "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/fault/v3"
	grpcstats "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/grpc_stats/v3"
	grpcweb "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/grpc_web/v3"
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
	Cors = &hcm.HttpFilter{
		Name: wellknown.CORS,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: MessageToAny(&cors.Cors{}),
		},
	}
	Fault = &hcm.HttpFilter{
		Name: wellknown.Fault,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: MessageToAny(&fault.HTTPFault{}),
		},
	}
	Router = &hcm.HttpFilter{
		Name: wellknown.Router,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: MessageToAny(&router.Router{}),
		},
	}
	GrpcWeb = &hcm.HttpFilter{
		Name: wellknown.GRPCWeb,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: MessageToAny(&grpcweb.GrpcWeb{}),
		},
	}
	GrpcStats = &hcm.HttpFilter{
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
