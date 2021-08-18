package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-yaml/yaml"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

const (
	file = "in.yaml"        // input file
	addr = "localhost:8080" // server address
)

var metrics *dto.MetricFamily

func main() {
	metrics = new(dto.MetricFamily)
	if err := readInput(file, metrics); err != nil {
		log.Fatalf("Input data  erro: %v", err)
	}
	http.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := readInput(file, metrics); err != nil {
			fmt.Fprintln(w, "Input data reading error:", err)
		} else {
			enc := expfmt.NewEncoder(w, expfmt.FmtOpenMetrics)
			if err := enc.Encode(metrics); err != nil {
				fmt.Fprintln(w, "Input data encoding error:", err)
			}
		}
	}))
	log.Println("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

var readTime, modTime time.Time

func readInput(file string, metrics *dto.MetricFamily) error {
	stat, err := os.Lstat(file)
	if err != nil {
		return fmt.Errorf("Reading file info: %q error:\n%v", file, err)
	}
	if stat.ModTime().Equal(modTime) == false || readTime.IsZero() || stat.ModTime().After(readTime) {
		r, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("Opening file: %q error:\n%v", file, err)
		}
		defer r.Close()
		modTime = stat.ModTime()
		readTime = time.Now()
		dec := yaml.NewDecoder(r)
		if err := dec.Decode(metrics); err != nil /* && err != io.EOF */ {
			return fmt.Errorf("Error decoding File: %q\n: %v\n", file, err)
		}
		log.Printf("File :%q has being read  at[ %v ]", file, readTime)
	} else {
		log.Printf("File :%q has't change since last time reading [ %v ]", file, readTime)
	}
	return nil
}
