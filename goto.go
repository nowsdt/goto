package main

import (
	"flag"
	"fmt"
	"goto/client"
	st "goto/store"
	"log"
	"net/http"
	"net/rpc"
)

var AddForm = `
	<form method="POST" action="/add">
	URL: <input type="text" name="url">
	<input type="submit" value="Add">
	</form>`

var (
	port       = flag.String("p", "8080", "http listen port")
	host       = flag.String("h", "localhost", "http host name")
	rpcEnabled = flag.Bool("rpc", false, "enable RPC server")
	masterAddr = flag.String("m", "", "RPC master address")
	fileName   = flag.String("f", "store.json", "file name")
)

var store st.Store

func main() {

	http.HandleFunc("/", Redirect)
	http.HandleFunc("/add", Add)
	http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), nil)

	if len(*masterAddr) != 0 {
		store = client.NewProxyStore(*masterAddr)
	} else {
		store = st.NewURLStore(*fileName)
	}

	if *rpcEnabled {
		rpc.RegisterName("Store", store)
		rpc.HandleHTTP()
	}
}

func init() {
	flag.Parse()
	log.SetFlags(log.Llongfile)
}

func Add(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	if len(url) == 0 {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, AddForm)
		return
	}

	var k string
	key := store.Put(&url, &k)
	fmt.Fprintf(w, "http://localhost:8080/%s", key)
}

func Redirect(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]
	var url string
	if err := store.Get(&key, &url); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)

}
