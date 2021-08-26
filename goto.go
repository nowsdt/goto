package main

import (
	"flag"
	"fmt"
	"goto/client"
	st "goto/store"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"time"
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
var shutdown = make(chan int, 1)

func main() {
	log.Println(*port, *host, *rpcEnabled, *masterAddr, *fileName)

	if len(*masterAddr) != 0 {
		store = client.NewProxyStore(*masterAddr)
	} else {
		store = st.NewURLStore(*fileName)
	}

	if *rpcEnabled {
		rpc.RegisterName("Store", store)
		rpc.HandleHTTP()
	}

	http.HandleFunc("/", Redirect)
	http.HandleFunc("/add", Add)
	http.HandleFunc("/shutdown", Shutdown)

	err := http.ListenAndServe(fmt.Sprintf(":%s", *port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func waitForShutdown() {
	go func() {
		<-shutdown
		log.Println("receive shutdown")
		for i := 5; i > 0; i-- {
			time.Sleep(1 * time.Second)
			log.Printf("shutdown count %d", i)
		}
		log.Println("bye")
		os.Exit(0)
	}()
}

func init() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Llongfile)
	waitForShutdown()
}

func Add(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	if len(url) == 0 {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, AddForm)
		return
	}

	var k string
	err := store.Put(&url, &k)
	if err != nil {
		fmt.Fprintf(w, "add failed，err:%s", err)
		return
	}
	fmt.Fprintf(w, "http://localhost:8080/%s", k)
}

func Shutdown(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "server will shutdown after %s s", "5")
	// shutdown signal
	shutdown <- 1
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
