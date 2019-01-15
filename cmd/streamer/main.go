package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/gorilla/mux"
	"github.com/yolo3301/cnp/pkg/storage/file"
	"github.com/yolo3301/cnp/pkg/streamer"
	"go.opencensus.io/stats/view"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
}

var (
	port    = flag.Int("port", 8080, "The port.")
	root    = flag.String("storage_root", "", "The storage root.")
	cfgPath = flag.String("config_path", "", "The config file path.")
	project = flag.String("gcp_project", "", "Project to export OpenCensus data to StackDriver.")
)

func main() {
	flag.Parse()
	if *root == "" {
		log.Fatal("storage_root must be specified.")
	}
	if *cfgPath == "" {
		log.Fatal("config_path must be specified.")
	}
	if *project == "" {
		log.Fatal("project must be specified.")
	}

	sd, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: *project,
		// MetricPrefix helps uniquely identify your metrics.
		MetricPrefix: "cnp-streamer",
	})
	if err != nil {
		log.Fatalf("Failed to create the Stackdriver exporter: %v", err)
	}
	// It is imperative to invoke flush before your main function exits
	defer sd.Flush()

	// Register it as a metrics exporter
	view.RegisterExporter(sd)
	view.SetReportingPeriod(60 * time.Second)

	storage, err := file.NewStorage(*root)
	if err != nil {
		log.Fatalf("Failed to init storage: %v", err)
	}

	r := mux.NewRouter()
	if _, err = streamer.NewService(*cfgPath, r, storage); err != nil {
		log.Fatalf("Failed to start streamer service: %v", err)
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
