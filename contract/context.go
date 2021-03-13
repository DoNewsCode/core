package contract

import "fmt"

type contextKey string

const (
	IpKey         contextKey = "ip"         // IP address
	TenantKey     contextKey = "tenant"     // Tenant
	TransportKey  contextKey = "transport"  // Transport, such as HTTP
	RequestUrlKey contextKey = "requestUrl" // Request url
)

// Tenant is interface representing a user or a consumer.
type Tenant interface {
	// KV contains key values about this tenant.
	KV() map[string]interface{}
	// String should uniquely represent this user. It should be human friendly.
	String() string
}

// MapTenant is an demo Tenant implementation. Useful for testing.
type MapTenant map[string]interface{}

// KV contains key values about this tenant.
func (d MapTenant) KV() map[string]interface{} {
	return d
}

// String should uniquely represent this user. It should be human friendly.
func (d MapTenant) String() string {
	return fmt.Sprintf("%+v", map[string]interface{}(d))
}
