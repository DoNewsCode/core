package main

import (
	"context"
	"github.com/DoNewsCode/std/pkg/core"
	"github.com/DoNewsCode/std/pkg/otgorm"
	"github.com/DoNewsCode/std/pkg/srvhttp"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"net/http"
)

type User struct {
	Id   string
	Name string
}

type Repository struct {
	DB *gorm.DB
}

func (r Repository) Find(id string) (*User, error) {
	var user User
	if err := r.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

type Handler struct {
	R Repository
}

func (h Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	encoder := srvhttp.NewResponseEncoder(writer)
	encoder.Encode(h.R.Find(request.URL.Query().Get("id")))
}

type Module struct {
	H Handler
}

func New(db *gorm.DB) Module {
	return Module{Handler{Repository{db}}}
}

func (m Module) ProvideHttp(router *mux.Router) {
	router.Handle("/", m.H)
}

func main() {
	// Phase One: creating a core from a configuration file
	c := core.New(core.WithYamlFile("config.yaml"))

	// Phase two: bootstrapping dependencies
	c.Provide(otgorm.Provide)

	// Phase three: define service
	c.AddModuleFunc(New)

	// Phase four: run!
	c.Serve(context.Background())
}
