package main

import (
	"goto/web"
	"log"
	"net/http"
)

func main() {

	log.SetFlags(log.Llongfile)

	http.HandleFunc("/", web.Redirect)
	http.HandleFunc("/add", web.Add)
	http.ListenAndServe(":8080", nil)
}
