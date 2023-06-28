package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewMux()
	p := os.Getenv("PORT")
	http.ListenAndServe(fmt.Sprintf(":%s", p), r)
}
