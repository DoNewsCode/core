package contract

import "fmt"

var (
	IpKey = struct{}{}
	TenantKey = struct{}{}
	TransportKey = struct {}{}
	RequestUrlKey = struct{}{}
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

