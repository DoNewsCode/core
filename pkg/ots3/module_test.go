package ots3

import (
	"net/http"
	"testing"

	"github.com/DoNewsCode/std/pkg/core"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestModule(t *testing.T) {
	c := core.New()
	c.AddCoreDependencies()
	c.AddDependencyFunc(ProvideManager)
	c.AddModuleFunc(New)
	router := mux.NewRouter()
	for _, provider := range c.GetHttpProviders() {
		provider(router)
	}
	request, _ := http.NewRequest("POST", "/upload", nil)
	assert.True(t, router.Match(request, &mux.RouteMatch{}))
}
