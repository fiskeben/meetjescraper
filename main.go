package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"

	"github.com/fiskeben/scrapejestad"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var Version string

var sensorRegexp = regexp.MustCompile("[0-9]+")

func main() {
	_port := flag.String("port", "", "port to listen to")
	flag.Parse()

	printVersionAndExit()

	log.Printf("meetjescraper %s", Version)

	port := *_port
	if port == "" {
		port = "8080"
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	mux := http.NewServeMux()

	handler := http.HandlerFunc(handle)
	mux.Handle("/", handler)

	mux.Handle("/metrics", promhttp.Handler())

	server := http.Server{Addr: fmt.Sprintf(":%s", port), Handler: mux}

	go func() {
		log.Printf("listening to port %s", port)
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
			done <- true
		}
	}()

	select {
	case sig := <-sigs:
		log.Printf("received %v, shutting down", sig)
		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatal(err)
		}
	case <-done:
		log.Println("done")
		break
	}
	log.Print("exit")
}

func handle(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	w.Header().Add("Content-Type", "application/json")

	parameters := req.URL.Query()
	sensorID, err := getSensorID(parameters)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()))
		return
	}

	limit, err := getLimit(parameters)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()))
		return
	}
	if limit > 100 {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"error":"maximum allowed number of items is 100"}`)
		return
	}

	data, err := queryService(ctx, sensorID, limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()))
		return
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(&data)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()))
	}
}

func getSensorID(parameters url.Values) (string, error) {
	sensorID := parameters.Get("sensor")
	if sensorID == "" {
		return "", errors.New("missing sensor ID")
	}
	if !sensorRegexp.MatchString(sensorID) {
		return "", errors.New("sensor ID must be numeric")
	}
	return sensorID, nil
}

func getLimit(parameters url.Values) (int, error) {
	limit := parameters.Get("limit")
	if limit == "" {
		limit = "50"
	}
	val, err := strconv.Atoi(limit)
	if err != nil {
		return -1, fmt.Errorf("%s is not a number (%v)", limit, err)
	}
	return val, nil
}

func queryService(ctx context.Context, sensorID string, limit int) ([]scrapejestad.Reading, error) {
	raw := fmt.Sprintf("https://meetjestad.net/data/sensors_recent.php?sensor=%s&limit=%d", sensorID, limit)
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL '%s': %err", raw, err)
	}
	return scrapejestad.ReadWithContext(ctx, u)
}

func printVersionAndExit() {
	args := flag.Args()
	if len(args) == 0 || args[0] != "version" {
		return
	}
	fmt.Printf("meetjescraper version %s\n", Version)
	os.Exit(0)
}
