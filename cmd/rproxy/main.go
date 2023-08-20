package main

/*

to ChatGPT4 lol, and it worked without changes:

"write a go reverse proxy takes env vars on start
like "RPROXY0=/files@http://localhost:9321  \
RPROXY1=/files/init/ui=@ttp://localhost:3000 \
go run main.go" , to map prefixes to services. \
This lets me run a simple React server so that \
it can make reference to its services."

... well, until I ensured that they were sorted.
*/

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

type Proxy struct {
	prefix string
	target *url.URL
	proxy  *httputil.ReverseProxy
}

func main() {
	proxies := make([]*Proxy, 0)

	i := 0
	for {
		k := fmt.Sprintf("RPROXY%d", i)
		v := os.Getenv(k)
		if len(v) == 0 {
			break
		} else {
			routeParts := strings.Split(v, "@")

			target, err := url.Parse(routeParts[1])
			if err != nil {
				log.Panic("Invalid target URL: ", routeParts[1])
			}

			log.Println("Adding route: ", routeParts[0], " -> ", routeParts[1])
			proxies = append(proxies, &Proxy{
				prefix: routeParts[0],
				target: target,
				proxy:  httputil.NewSingleHostReverseProxy(target),
			})
		}
		i++
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		for _, p := range proxies {
			if strings.HasPrefix(r.URL.Path, p.prefix) {
				r.URL.Path = strings.TrimPrefix(r.URL.Path, p.prefix)
				p.proxy.ServeHTTP(w, r)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	})
	bindAddr := os.Getenv("BIND")
	if bindAddr == "" {
		bindAddr = ":8080"
	}
	if os.Getenv("X509_CERT") == "" {
		log.Fatal(http.ListenAndServe(bindAddr, nil))
	} else {
		certFile := os.Getenv("X509_CERT")
		keyFile := os.Getenv("X509_KEY")
		log.Fatal(http.ListenAndServeTLS(bindAddr, certFile, keyFile, nil))
	}
}
