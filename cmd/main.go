package main

import (
	"net/http"

	"github.com/wosai/ultron/transport/rest"
)

func main() {
	handler := rest.BuildHTTPRouter()
	http.ListenAndServe(":2017", handler)
}
