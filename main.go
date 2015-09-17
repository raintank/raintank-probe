package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/raintank/raintank-probe/checks"
)

type RaintankProbeCheck interface {
	Run() error
	Results() interface{}
}

func main() {
	http.HandleFunc("/", handler)
	var port int
	flag.IntVar(&port, "p", 8080, "TCP port to listen on")
	flag.Parse()
	fmt.Println("raintank-probe server starting up on port " + strconv.Itoa(port))
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

	check, err := GetCheck(checkType, body)
	if err != nil {
		sendError(w, 500, err)
		return
	}
	result, err := RunCheck(check)
	if err != nil {
		sendError(w, 500, err)
		return
	}

	//encode the result as a JSON string.
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

func GetCheck(checkType string, body []byte) (RaintankProbeCheck, error) {
	switch checkType {
	case "ping":
		return checks.NewRaintankPingProbe(body)
	case "dns":
		return checks.NewRaintankDnsProbe(body)
	case "http":
		return checks.NewRaintankHTTPProbe(body)
	case "https":
		return checks.NewRaintankHTTPSProbe(body)
	default:
		return nil, fmt.Errorf("unknown check type. " + checkType)
	}

}

func RunCheck(check RaintankProbeCheck) (interface{}, error) {
	resultChan := make(chan error)
	//run the check in a goroutine.
	go func() {
		//push the return into the resultChan
		resultChan <- check.Run()
	}()

	err := <-resultChan
	return check.Results(), err
}
