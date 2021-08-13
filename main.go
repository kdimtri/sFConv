package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/common/expfmt"
)

const filepath = "./in.yaml"
const format = "text/yaml"

func main() {
	http.Handle("/metrics", http.HandlerFunc(openMetricsResponse))
	log.Println("localhost:8080")
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
func openMetricsResponse(w http.ResponseWriter, r *http.Request) {

	log.Println(r.Header, r.RequestURI, r.URL)
	b, err := os.Open(filepath)
	if err != nil {
		fmt.Fprintf(w, "Open file: %v, %v", filepath, err)
		log.Println("file at ", filepath, ": ", err)
	}
	defer b.Close()
	log.Println("file at ", filepath, ": ok")
	encoder := expfmt.NewEncoder(w, expfmt.FmtOpenMetrics)
	parser := expfmt.TextParser{}
	m, err := parser.TextToMetricFamilies(b)
	if err != nil {
		fmt.Fprintf(w, "parser.TextToMetricFamilies: %v", err)
		log.Printf("parser.TextToMetricFamilies: %v", err)
	}
	for _, v := range m {
		encoder.Encode(v)
	}
}
