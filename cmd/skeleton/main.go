package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/yolo3301/cnp/pkg/service/skeleton"
	"github.com/yolo3301/cnp/pkg/streamer"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
)

var (
	httpPort     = flag.Int("http_port", 8280, "The http port.")
	agentPort    = flag.Int("agent_port", 8180, "The agent server port.")
	serviceName  = flag.String("service_name", "", "The service name")
	streamerHost = flag.String("streamer_host", "", "The streamer host.")
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
}

func main() {
	flag.Parse()
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
