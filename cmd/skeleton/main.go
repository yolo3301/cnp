package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/gorilla/mux"
	"github.com/yolo3301/cnp/pkg/service/skeleton"
	"github.com/yolo3301/cnp/pkg/streamer"
	"go.opencensus.io/stats/view"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
)

var (
	httpPort     = flag.Int("http_port", 8280, "The http port.")
	agentPort    = flag.Int("agent_port", 8180, "The agent server port.")
	serviceName  = flag.String("service_name", "", "The service name")
	streamerHost = flag.String("streamer_host", "", "The streamer host.")
	project      = flag.String("gcp_project", "", "Project to export OpenCensus data to StackDriver.")
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
}

func main() {
	flag.Parse()
	if *project == "" {
		log.Fatal("project must be specified.")
	}

	sd, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: *project,
		// MetricPrefix helps uniquely identify your metrics.
		MetricPrefix: *serviceName,
	})
	if err != nil {
		log.Fatalf("Failed to create the Stackdriver exporter: %v", err)
	}
	// It is imperative to invoke flush before your main function exits
	defer sd.Flush()

	// Register it as a metrics exporter
	view.RegisterExporter(sd)
	view.SetReportingPeriod(60 * time.Second)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *agentPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	streamer.RegisterStreamerAgentServer(s, &skeleton.Agent{})

	if *serviceName == "" {
		log.Fatal("service_name must be specified.")
	}
	if *streamerHost == "" {
		log.Fatal("streamer_host must be specified.")
	}
	r := mux.NewRouter()
	skeleton.NewService(r, *serviceName, *streamerHost)

	var g errgroup.Group
	g.Go(func() error {
		return s.Serve(lis)
	})
	g.Go(func() error {
		return http.ListenAndServe(fmt.Sprintf(":%d", *httpPort), nil)
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
