package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

type Redirect struct {
	Prefix string
	Target string
}

type Handler struct {
	Redirects []Redirect
	upgrader  websocket.Upgrader
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\n", r.Method, r.URL.Path)

	// Check if this is a websocket upgrade request
	if websocketRequest := r.Header.Get("Upgrade"); websocketRequest == "websocket" {
		h.handleWebsocket(w, r)
		return
	}

	h.handleNormalRequest(w, r)
}

func (h *Handler) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	for _, redirect := range h.Redirects {
		if strings.HasPrefix(r.URL.Path, redirect.Prefix) {
			targetURL := redirect.Target + r.URL.Path[len(redirect.Prefix):]
			if r.URL.RawQuery != "" {
				targetURL += "?" + r.URL.RawQuery
			}

			// Parse the target URL
			url, err := url.Parse(targetURL)
			if err != nil {
				log.Printf("Error parsing target URL %s: %v\n", targetURL, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Create the backend URL (convert http -> ws, https -> wss)
			backend := url.String()
			backend = strings.Replace(backend, "http://", "ws://", 1)
			backend = strings.Replace(backend, "https://", "wss://", 1)

			// Connect to the backend
			backConn, _, err := websocket.DefaultDialer.Dial(backend, nil)
			if err != nil {
				log.Printf("Error dialing websocket backend %s: %v\n", backend, err)
				http.Error(w, "Bad Gateway", http.StatusBadGateway)
				return
			}
			defer backConn.Close()

			// Upgrade the client connection
			conn, err := h.upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Printf("Error upgrading connection: %v\n", err)
				return
			}
			defer conn.Close()

			// Start copying data between connections
			errChan := make(chan error, 2)

			// Forward client messages to backend
			go func() {
				for {
					messageType, message, err := conn.ReadMessage()
					if err != nil {
						errChan <- err
						return
					}
					err = backConn.WriteMessage(messageType, message)
					if err != nil {
						errChan <- err
						return
					}
				}
			}()

			// Forward backend messages to client
			go func() {
				for {
					messageType, message, err := backConn.ReadMessage()
					if err != nil {
						errChan <- err
						return
					}
					err = conn.WriteMessage(messageType, message)
					if err != nil {
						errChan <- err
						return
					}
				}
			}()

			// Wait for either connection to close
			<-errChan
			return
		}
	}
	http.Error(w, "Not found by rproxy", http.StatusNotFound)
}

func (h *Handler) handleNormalRequest(w http.ResponseWriter, r *http.Request) {
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
