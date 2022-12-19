package web

import (
	"github.com/gorilla/mux"
	"net/http"
)

type View struct {
	TemplateDir string
	Route       string
}

func (v *View) RegisterRoutes(router *mux.Router) {
	dir := http.Dir(v.TemplateDir)
	fileServer := http.FileServer(dir)
	router.Handle(v.Route, fileServer)
}
