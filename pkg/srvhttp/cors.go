package srvhttp

import (
	"net/http"

	"github.com/rs/cors"
)

func MakeUnsafeCorsMiddleware() func(handler http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "OPTIONS", "HEAD", "DELETE", "PATCH"},
	}).Handler
}
