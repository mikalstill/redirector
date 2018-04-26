package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	dialTimeout    = 2 * time.Second
	requestTimeout = time.Second
)

var (
	Info *log.Logger
)

func saveHandler(w http.ResponseWriter, r *http.Request) {
	//title := r.URL.Path[len("/save/"):]
	//body := r.FormValue("body")
	//http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Connect to etcd
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"192.168.50.1:2379"},
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	// We want to deal in key values
	kv := clientv3.NewKV(cli)

	// Write a value
	//kv.Put(ctx, "/testing", "testing123")

	switch r.URL.Path {
	case "/":
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, "<h1>Write a new shortcut</h1>")
		fmt.Fprintf(w, "<form action='save' method='post'>")
		fmd.Fprintf(w, "</form\n\n")

		fmt.Fprintf(w, "<h1>Currently saved URLs</h1><ul>")
		resp, err := kv.Get(ctx, "", clientv3.WithPrefix())
		cancel()
		if err != nil {
			log.Fatal(err)
		}
		for _, ev := range resp.Kvs {
			fmt.Fprintf(w, "<li><a href=\"%s\">%s</a>: %s\n",
				ev.Key, ev.Key, ev.Value)
		}
		fmt.Fprintf(w, "</ul>")

	default:
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		resp, err := kv.Get(ctx, r.URL.Path, clientv3.WithPrefix())
		cancel()
		if err != nil {
			log.Fatal(err)
		}

		Info.Printf("Search for %s found %d results",
			r.URL.Path, len(resp.Kvs))
		if len(resp.Kvs) == 0 {
			fmt.Fprintf(w, "Link not found")
		} else {
			http.Redirect(w, r, string(resp.Kvs[0].Value),
				http.StatusFound)
		}
	}
}

func main() {
	Info = log.New(os.Stdout, "INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
