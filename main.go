package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/ca-std/lib"
	"github.com/go-redis/redis/v8"
)

const (
	REDIS_ADDR = "trusted.recurse:6379"
	HTTP_ADDR  = "recurse:443"
)

var (
	instance = context.Background()
	client   = redis.NewClient(&redis.Options{
		Addr: REDIS_ADDR,
	})
)

func main() {
	defer instance.Done()
	log.Fatal(server(HTTP_ADDR).ListenAndServeTLS("/etc/letsencrypt/live/1o.fyi/fullchain.pem", "/etc/letsencrypt/live/1o.fyi/privkey.pem"))
}

func set(w http.ResponseWriter, req *http.Request) {
	for key, value := range parseUri(req.URL) {
		log.Println(key, value)
		response, err := client.Set(context.TODO(), key, value, 0).Result()
		if err != nil {
			w.Write([]byte(err.Error()))
		}
		w.Write([]byte(response))
	}
}

func get(w http.ResponseWriter, req *http.Request) {
	for key := range parseUri(req.URL) {
		log.Println(key)
		response, err := client.Get(context.TODO(), key).Result()
		if err != nil {
			w.Write([]byte(err.Error()))
		}
		w.Write([]byte(response))
	}
}

func parseUri(u *url.URL) map[string]string {
	uri, res := u.RawQuery, make(map[string]string)
	pairs := strings.Split(uri, "?")
	for _, pair := range pairs {
		split := strings.Split(pair, "=")
		if len(split) == 1 {
			split = append(split, "")
		}
		res[split[0]] = split[1]
	}
	return res
}

func server(addr string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Strict-Transport-Security", "max-age=63072000")
		w.Write([]byte("welcome, friend."))
	})
	mux.HandleFunc("/gen", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write(lib.UniformDistributionRp(4, 256).EncodePEM())
	})
	mux.HandleFunc("/get", get)
	mux.HandleFunc("/set", set)
	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
