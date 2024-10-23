package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Redirect struct {
	Prefix string
	Target string
}

type Handler struct {
	Redirects []Redirect
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.URL.Path)
	client := &http.Client{}
	for _, redirect := range h.Redirects {
		if strings.HasPrefix(r.URL.Path, redirect.Prefix) {
			// only supporting http redirects for now, no https
			from := r.URL.Path
			if r.URL.Query().Encode() != "" {
				from += "?" + r.URL.Query().Encode()
			}
			to := redirect.Target + from[len(redirect.Prefix):]
			fmt.Printf("%s %s -> %s\n", r.Method, from, to)
			req, err := http.NewRequest(
				r.Method,
				to,
				r.Body,
			)
			if err != nil {
				fmt.Printf("Error creating request %s: %v\n", to, err)
				return
			}
			for k, v := range r.Header {
				req.Header[k] = v
			}
			res, err := client.Do(req)
			if err != nil {
				fmt.Printf("Error doing request %s: %v\n", to, err)
				return
			}
			for k, v := range res.Header {
				w.Header()[k] = v
			}
			w.WriteHeader(res.StatusCode)
			// POST/PUT requests have a body
			io.Copy(w, res.Body)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not found by rproxy"))
}

func parseRedirects() *Handler {
	handler := &Handler{
		Redirects: make([]Redirect, 0),
	}

	i := 0
	for {
		k := fmt.Sprintf("RPROXY%d", i)
		v := os.Getenv(k)
		if len(v) == 0 {
			break
		} else {
			routeParts := strings.Split(v, "@")
			_, err := url.Parse(routeParts[1])
			if err != nil {
				log.Panic("Invalid target URL: ", routeParts[1])
			}

			log.Println("Adding route: ", routeParts[0], " -> ", routeParts[1])
			handler.Redirects = append(handler.Redirects,
				Redirect{
					Prefix: routeParts[0],
					Target: routeParts[1],
				},
			)
		}
		i++
	}
	return handler
}

func main() {
	handler := parseRedirects()

	bindAddr := os.Getenv("BIND")
	if bindAddr == "" {
		bindAddr = ":8080"
	}
	if os.Getenv("X509_CERT") == "" {
		log.Fatal(http.ListenAndServe(bindAddr, handler))
	} else {
		certFile := os.Getenv("X509_CERT")
		keyFile := os.Getenv("X509_KEY")
		log.Fatal(http.ListenAndServeTLS(bindAddr, certFile, keyFile, handler))
	}
}
