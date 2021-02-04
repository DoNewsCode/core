package contract

import "fmt"

type contextKey string

const (
	IpKey         contextKey = "ip"
	TenantKey     contextKey = "tenant"
	TransportKey  contextKey = "transport"
	RequestUrlKey contextKey = "requestUrl"
)

type Tenant interface {
	KV() map[string]interface{}
	String() string
}

type MapTenant map[string]interface{}

func (d MapTenant) KV() map[string]interface{} {
	return d
}

func (d MapTenant) String() string {
	return fmt.Sprintf("%+v", map[string]interface{}(d))
}
