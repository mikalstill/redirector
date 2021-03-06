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

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

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
	kv.Delete(ctx, "redirector/" + r.FormValue("short"))
	cancel()
	Info.Printf("Deleted %s", r.FormValue("short"))
	http.Redirect(w, r, "/", http.StatusFound)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

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
	kv.Put(ctx, "redirector/" + r.FormValue("short"), r.FormValue("url"))
	cancel()
	Info.Printf("Set value of /%s to %s", r.FormValue("short"),
		r.FormValue("url"))
	http.Redirect(w, r, "/", http.StatusFound)
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

	switch r.URL.Path {
	case "/":
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, "<h1>Write a new shortcut</h1>")
		fmt.Fprintf(w, "<form action='save' method='get'>")
		fmt.Fprintf(w, "<table>")
		fmt.Fprintf(w, "<tr><td>Short form:</td>")
		fmt.Fprintf(w, "<td><input type='text' name='short'></td></tr>")
		fmt.Fprintf(w, "<tr><td>URL:</td>")
		fmt.Fprintf(w, "<td><input type='text' name='url'></td></tr>")
		fmt.Fprintf(w, "<tr><td></td>")
		fmt.Fprintf(w, "<td><input type='submit' value='Save'></td></tr>")
		fmt.Fprintf(w, "</table>")
		fmt.Fprintf(w, "</form\n\n")

		fmt.Fprintf(w, "<h1>Currently saved URLs</h1><ul>")
		resp, err := kv.Get(ctx, "redirector", clientv3.WithPrefix())
		cancel()
		if err != nil {
			log.Fatal(err)
		}
		for _, ev := range resp.Kvs {
		        name := ev.Key[len("redirector/"):]
			fmt.Fprintf(w, "<li><a href=\"%s\">%s</a>: %s",
				name, name, ev.Value)
			fmt.Fprintf(w, "<form><input type='button' value='delete' onclick=\"window.location.href='/delete?short=%s'\" /></form>",
				string(name))
		}
		fmt.Fprintf(w, "</ul>")

	default:
		resp, err := kv.Get(ctx, "redirector" + r.URL.Path,
		                    clientv3.WithPrefix())
		cancel()
		if err != nil {
			log.Fatal(err)
		}

		Info.Printf("Search for %s found %d results",
			r.URL.Path, len(resp.Kvs))
		if len(resp.Kvs) == 0 {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Link not found")
		} else {
			Info.Printf("Redirecting %s to %s", r.URL.Path,
				string(resp.Kvs[0].Value))
			http.Redirect(w, r, string(resp.Kvs[0].Value),
				http.StatusFound)
		}
	}
}

func main() {
	Info = log.New(os.Stdout, "INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
