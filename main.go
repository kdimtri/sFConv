package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-yaml/yaml"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

var file = flag.String("in", "in.yaml", "path to input .yaml file ")
var addr = flag.String("addr", "localhost:8080", "http service adress")
var metrics *dto.MetricFamily

/*
First time readInput() runs to read and decode input.
Second run of readInput() would't do anything until input file
would be modifyed.
All errors woud be written in response - this way formating yaml could be speeds up
after first initial readng of input.
*/
func main() {
	flag.Parse()
	metrics = new(dto.MetricFamily)
	if err := readInput(*file, metrics); err != nil {
		log.Fatalf("Input data  erro: %v", err)
	}
	http.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := readInput(*file, metrics); err != nil {
			fmt.Fprintln(w, "Input data reading error:", err)
		} else {
			enc := expfmt.NewEncoder(w, expfmt.FmtOpenMetrics)
			if err := enc.Encode(metrics); err != nil {
				fmt.Fprintln(w, "Input data encoding error:", err)
			}
		}
	}))
	log.Println("Listening on", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

/*
readTime holds last time input were read
	and if readTime is zero (no reads) or input
	were modifyed after - input reads
modTime holds modification time from input file
	that were registred while last time input were read.
	This way happens overCompensation that shuld report
	if some thing were changed in input file or not
*/
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
		if err := dec.Decode(metrics); err != nil {
			return fmt.Errorf("Error decoding File: %q\n: %v\n", file, err)
		}
		log.Printf("File :%q has being read  at[ %v ]", file, readTime)
	} else {
		log.Printf("File :%q has't change since last time reading [ %v ]", file, readTime)
	}
	return nil
}
