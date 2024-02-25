package conf

import "net/http"

type HttpConfig struct {
	Port     int
	CertFile string
	KeyFile  string
	Verbose  bool
	MaxConns int
	MaxBytes int64
	// milliseconds
	Timeout      int64
	CpuThreshold int64
	// TraceIgnorePaths is paths blacklist for trace middleware.
	TraceIgnorePaths []string

	Session        SessionConfig
	PeriodLimit    PeriodLimitConfig
	OAuth2         OAuth2Config
	StaticResource StaticResourceConfig
	Rpc            RpcConfig
}

type SessionConfig struct {
	Name      string
	Domain    string
	Path      string
	Secure    bool
	HttpOnly  bool
	MaxAge    int
	SameSite  http.SameSite
	StoreType string
}

type PeriodLimitConfig struct {
	Enable      bool
	Distributed bool
	Period      int
	Quota       int
}

type OAuth2Config struct {
	Enable   bool
	Expire   int
	TokenUri string
}

type RpcConfig struct {
	Token      string
	Type       string
	IpProxy    IpProxyConfig
	Kubernetes KubernetesConfig
}

type IpProxyConfig struct {
	Proxy map[string]string
}

type KubernetesConfig struct {
	Namespace string
	Proxy     map[string]string
	PortName  string
}

type StaticResourceConfig struct {
	Location string
	Fs       interface{}
}
