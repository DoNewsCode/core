package otetcd

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/DoNewsCode/core/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Option is a type that holds all of available etcd configurations.
type Option struct {
	// Endpoints is a list of URLs.
	Endpoints []string `json:"endpoints" yaml:"endpoints"`

	// AutoSyncInterval is the interval to update endpoints with its latest members.
	// 0 disables auto-sync. By default auto-sync is disabled.
	AutoSyncInterval config.Duration `json:"autoSyncInterval" yaml:"autoSyncInterval"`

	// DialTimeout is the timeout for failing to establish a connection.
	DialTimeout config.Duration `json:"dialTimeout" yaml:"dialTimeout"`

	// DialKeepAliveTime is the time after which client pings the server to see if
	// transport is alive.
	DialKeepAliveTime config.Duration `json:"dialKeepAliveTime" yaml:"dialKeepAliveTime"`

	// DialKeepAliveTimeout is the time that the client waits for a response for the
	// keep-alive probe. If the response is not received in this time, the connection is closed.
	DialKeepAliveTimeout config.Duration `json:"dialKeepAliveTimeout" yaml:"dialKeepAliveTimeout"`

	// MaxCallSendMsgSize is the client-side request send limit in bytes.
	// If 0, it defaults to 2.0 MiB (2 * 1024 * 1024).
	// Make sure that "MaxCallSendMsgSize" < server-side default send/recv limit.
	// ("--max-request-bytes" flag to etcd or "embed.Config.MaxRequestBytes").
	MaxCallSendMsgSize int `json:"maxCallSendMsgSize" yaml:"maxCallSendMsgSize"`

	// MaxCallRecvMsgSize is the client-side response receive limit.
	// If 0, it defaults to "math.MaxInt32", because range response can
	// easily exceed request send limits.
	// Make sure that "MaxCallRecvMsgSize" >= server-side default send/recv limit.
	// ("--max-request-bytes" flag to etcd or "embed.Config.MaxRequestBytes").
	MaxCallRecvMsgSize int `json:"maxCallRecvMsgSize" yaml:"MaxCallRecvMsgSize"`

	// TLS holds the client secure credentials, if any.
	TLS *tls.Config `json:"-" yaml:"-"`

	// Username is a user name for authentication.
	Username string `json:"username" yaml:"username"`

	// Password is a password for authentication.
	Password string `json:"password" yaml:"password"`

	// RejectOldCluster when set will refuse to create a client against an outdated cluster.
	RejectOldCluster bool `json:"rejectOldCluster" yaml:"rejectOldCluster"`

	// DialOptions is a list of dial options for the grpc client (e.g., for interceptors).
	// For example, pass "grpc.WithBlock()" to block until the underlying connection is up.
	// Without this, Dial returns immediately and connecting the server happens in background.
	DialOptions []grpc.DialOption `json:"-" yaml:"-"`

	// Context is the default client context; it can be used to cancel grpc dial out and
	// other operations that do not have an explicit context.
	Context context.Context `json:"-" yaml:"-"`

	// LogConfig configures client-side logger.
	// If nil, use the default logger.
	// TODO: configure gRPC logger
	LogConfig *zap.Config `json:"-" yaml:"-"`

	// PermitWithoutStream when set will allow client to send keepalive pings to server without any active streams(RPCs).
	PermitWithoutStream bool `json:"permitWithoutStream" yaml:"permitWithoutStream"`
}

// GetDialTimeout provide default timeout duration, avoid program blocking.
func (o *Option) dialTimeout() config.Duration {
	if o.DialTimeout.IsZero() {
		return config.Duration{Duration: 10 * time.Second}
	}
	return o.DialTimeout
}
