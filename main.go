package main

import (
	"context"
	"encoding/pem"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"filippo.io/age"
	"github.com/ca-std/lib"
	"github.com/go-redis/redis/v8"
)

const (
	REDIS_ADDR    = "trusted.recurse:6379"
	HTTP_ADDR     = "recurse:443"
	CERT          = "/etc/letsencrypt/live/1o.fyi/fullchain.pem"
	PK            = "/etc/letsencrypt/live/1o.fyi/privkey.pem"
	AGE_SK_HEADER = "AGE-SECRET-KEY"
	AGE_PK_HEADER = "AGE-PUBLIC-KEY"
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
	mux.HandleFunc("/enc", enc)
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
		response, _ := client.Set(context.TODO(), key, value, 0).Result()
		w.Write([]byte(response))
	}
}

func get(w http.ResponseWriter, req *http.Request) {
	for key := range parseUri(req.URL) {
		response, _ := client.Get(context.TODO(), key).Result()
		w.Write([]byte(response))
	}
}

func gen(rw http.ResponseWriter, r *http.Request) {
	rw.Write(lib.UniformDistributionRp(4, 256).EncodePEM())
}

func enc(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		return
	}
	recip, rerr := age.ParseRecipients(req.Body)
	if rerr != nil {
		return
	}
	wrtr, err := age.Encrypt(w, recip...)
	if err != nil {
		return
	}
	for key := range parseUri(req.URL) {
		wrtr.Write([]byte(key))
	}
}

func ega(w http.ResponseWriter, req *http.Request) {
	_u, err := age.GenerateX25519Identity()
	if err != nil {
		w.Write([]byte("bad."))
		return
	}
	p0, p1 := pairEncoded(_u)
	w.Write(p0)
	w.Write(p1)
}

func pairEncoded(_u *age.X25519Identity) ([]byte, []byte) {
	return pem.EncodeToMemory(&pem.Block{
			Type:  AGE_PK_HEADER,
			Bytes: []byte(_u.String()),
		}),
		pem.EncodeToMemory(&pem.Block{
			Type:  AGE_SK_HEADER,
			Bytes: []byte(_u.Recipient().String()),
		})
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

func rdgaf(path string) *os.File {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	return f
}

func PEMDecodeIdentity(f *os.File) (*age.X25519Identity, error) {
	t := make([]byte, 255)
	_, err := f.Read(t)
	if err != nil {
		return nil, err
	}
	blocc, _ := pem.Decode(t)
	s := string(blocc.Bytes)
	_id, err := age.ParseX25519Identity(s)
	if err != nil {
		return nil, err
	}
	return _id, nil
}

func PEMDecodeRecipiant(f *os.File) (*age.X25519Recipient, error) {
	t := make([]byte, 255)
	_, err := f.Read(t)
	if err != nil {
		return nil, err
	}
	blocc, _ := pem.Decode(t)
	_id, err := age.ParseX25519Recipient(string(blocc.Bytes))
	if err != nil {
		return nil, err
	}
	return _id, nil
}
