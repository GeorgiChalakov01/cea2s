package main

import (
	"fmt"
	"net/http"
	
	"github.com/a-h/templ"
	"github.com/GeorgiChalakov01/cea2s/pages/home"
	"github.com/GeorgiChalakov01/cea2s/pages/part1"
)

func main() {
	http.Handle("/", http.RedirectHandler("/home", http.StatusSeeOther))
	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(home.Home()).ServeHTTP(w, r)
	})
	http.HandleFunc("/part1", func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(part1.Part1()).ServeHTTP(w, r)
	})
	
	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
