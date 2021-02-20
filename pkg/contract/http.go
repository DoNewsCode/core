package contract

import "net/http"

// HttpDoer is the interface for a http client.
type HttpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}
