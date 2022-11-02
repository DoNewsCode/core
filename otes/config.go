package otes

// Config is the type of elasticsearch configurations.
type Config struct {
	URL         []string `json:"url" yaml:"url"`
	Index       string   `json:"index" yaml:"index"`
	Username    string   `json:"username" yaml:"username"`
	Password    string   `json:"password" yaml:"password"`
	Shards      int      `json:"shards" yaml:"shards"`
	Replicas    int      `json:"replicas" yaml:"replicas"`
	Sniff       *bool    `json:"sniff" yaml:"sniff"`
	Healthcheck *bool    `json:"healthcheck" yaml:"healthcheck"`
	// DebugLogLimitSize limit the explosive output of debug-level logs, default unlimited
	DebugLogLimitSize int `json:"debugLogLimitSize" yaml:"debugLogLimitSize"`
}
