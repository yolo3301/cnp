package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/yolo3301/cnp/pkg/storage/file"
	"github.com/yolo3301/cnp/pkg/streamer"

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
)

func main() {
	if *root == "" {
		log.Fatal("storage_root must be specified.")
	}
	if *cfgPath == "" {
		log.Fatal("config_path must be specified.")
	}

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
