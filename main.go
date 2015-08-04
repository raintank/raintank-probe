package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/raintank/raintank-probe/checks"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", handler)
	var port int
	flag.IntVar(&port, "p", 8080, "TCP port to listen on")
	flag.Parse()
	fmt.Println("raintank-probe server starting up.")
	err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	checkType := r.URL.Path[1:]
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendError(w, 500, fmt.Errorf("failed to read request body."))
		return
	}

	var result interface{}

	switch checkType {
	case "ping":
		result, err = checks.Ping(body)
	//case "dns":
	//#	result, err = checks.Dns(body)
	default:
		sendError(w, 500, fmt.Errorf("unknown check type. "+checkType))
		return
	}

	json, err := json.Marshal(result)
	if err != nil {
		sendError(w, 500, fmt.Errorf("failed to convert results to json: "+err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
	return
}

func sendError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	w.Write([]byte(err.Error()))
	return
}
