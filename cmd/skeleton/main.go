package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/yolo3301/cnp/pkg/service/skeleton"
	"github.com/yolo3301/cnp/pkg/streamer"

	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
)

var agentPort = flag.Int("agent_port", 8180, "The agent server port.")

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *agentPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	streamer.RegisterStreamerAgentServer(s, &skeleton.Agent{})
	log.Fatal(s.Serve(lis))
}
