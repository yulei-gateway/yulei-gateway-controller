package route

type HttpRoute struct {
	BindListenerName string
	Rules            []HttpRouteRule
	Name             string
}

type TcpRoute struct {
	BindListenerName string
	Rules            []TcpRouteRule
	Name             string
}

type Filter interface {
	GetName() string
}

type PathMatchType string

const (
	//Prefix 前缀匹配
	PathPrefix PathMatchType = "prefix"
	//Regex 正则匹配
	PathRegex PathMatchType = "regex"
	//Exact  完全匹配
	PathExact PathMatchType = "exact"
)

type HeaderMatchType string

const (
	//Regex 正则匹配
	HeaderRegex HeaderMatchType = "regex"
	//Exact  完全匹配
	HeaderExact HeaderMatchType = "exact"
)

type QueryMatchType string

const (
	//Regex 正则匹配
	QueryRegex QueryMatchType = "regex"
	//Exact  完全匹配
	QueryExact QueryMatchType = "exact"
)

type QueryParam struct {
	QueryMatchType QueryMatchType
	Name           string
	Value          string
}

type PathMatch struct {
	PathMatchType PathMatchType
	PathValue     string
}

type HeaderMatch struct {
	HeaderMatchType HeaderMatchType
	HeaderName      string
	HeaderValue     string
}

type HttpRouteRule struct {
	Filters          []Filter
	PathMatch        PathMatch
	HeaderMatchs     []HeaderMatch
	QueryParamMatchs []QueryParam
	ClusterConfigs   []ClusterConfig
}

type ClusterConfig struct {
	ClusterName string
	Weight      int
}

type TcpRouteRule struct {
	ClusterConfigs []ClusterConfig
}
