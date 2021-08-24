package main

import (
	"goto/web"
	"net/http"
)

func main() {

	http.HandleFunc("/", web.Redirect)
	http.HandleFunc("/add", web.Add)
	http.ListenAndServe(":8080", nil)
}
