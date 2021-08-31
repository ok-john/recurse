package main

import (
	"bytes"
	"context"
	"encoding/pem"
	"log"
	"net/http"
	"net/url"
	"strings"

	"filippo.io/age"
	"github.com/ca-std/lib"
	"github.com/go-redis/redis/v8"
)

const (
	REDIS_ADDR = "trusted.recurse:6379"
	HTTP_ADDR  = "recurse:443"
	CERT       = "/etc/letsencrypt/live/1o.fyi/fullchain.pem"
	PK         = "/etc/letsencrypt/live/1o.fyi/privkey.pem"
)

var (
	instance = context.Background()
	client   = redis.NewClient(&redis.Options{
		Addr: REDIS_ADDR,
	})
)

func main() {
	defer instance.Done()
	mux := http.NewServeMux()
	mux.HandleFunc("/", home)
	mux.HandleFunc("/age", ega)
	mux.HandleFunc("/gen", gen)
	mux.HandleFunc("/get", get)
	mux.HandleFunc("/set", set)
	server := &http.Server{
		Addr:    HTTP_ADDR,
		Handler: mux,
	}
	log.Fatal(server.ListenAndServeTLS(CERT, PK))
}

func home(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000")
	w.Write([]byte("welcome, friend."))
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

func gen(rw http.ResponseWriter, r *http.Request) {
	rw.Write(lib.UniformDistributionRp(4, 256).EncodePEM())
}

func ega(w http.ResponseWriter, req *http.Request) {
	_u, err := age.GenerateX25519Identity()
	if err != nil {
		w.Write([]byte("bad."))
		return
	}
	w.Write(bytes.Join([][]byte{
		pem.EncodeToMemory(&pem.Block{
			Headers: make(map[string]string),
			Type:    "AGEv1 PRIVATE KEY",
			Bytes:   []byte(_u.String()),
		}),
		pem.EncodeToMemory(&pem.Block{
			Headers: make(map[string]string),
			Type:    "AGEv1 PUBLIC KEY",
			Bytes:   []byte(_u.Recipient().String()),
		}),
	}, []byte("\n")))
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
