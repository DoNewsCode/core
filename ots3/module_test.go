package ots3

import (
	"net/http"
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestModule(t *testing.T) {
	c := core.New()
	c.ProvideEssentials()
	c.Provide(ProvideManager)
	c.AddModuleFunc(New)
	router := mux.NewRouter()
	c.ApplyRouter(router)
	request, _ := http.NewRequest("POST", "/upload", nil)
	assert.True(t, router.Match(request, &mux.RouteMatch{}))
}
