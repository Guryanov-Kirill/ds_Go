package handler

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.json
var openapiSpec []byte

//go:embed swagger.html
var swaggerPage []byte

func serveSwagger(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(swaggerPage)
}

func serveOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(openapiSpec)
}
