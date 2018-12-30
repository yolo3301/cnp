package streamer

import (
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

type Service struct {
	storage Storage
	config  *Config
}

func NewService(cfgPath string, router *mux.Router, storage Storage) (*Service, error) {
	cfg, err := loadConfig(cfgPath)
	if err != nil {
		return nil, err
	}
	s := &Service{
		storage: storage,
		config:  cfg,
	}
	s.registerRoutes(router)
	http.Handle("/", router)
	return s, nil
}

func loadConfig(cfgPath string) (*Config, error) {
	b, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (s *Service) registerRoutes(router *mux.Router) {
	r := router.PathPrefix("/streamer").Subrouter()
	r.PathPrefix("/uploads").HandlerFunc(s.handleUpload).Methods(http.MethodPost, http.MethodPut, http.MethodPatch)
	r.PathPrefix("/downloads").HandlerFunc(s.handleDownload).Methods(http.MethodGet)
}

func (s *Service) handleUpload(w http.ResponseWriter, r *http.Request) {
	log.Infof("Received upload request %s: %s", r.Method, r.RequestURI)
}

func (s *Service) handleDownload(w http.ResponseWriter, r *http.Request) {
	log.Infof("Received download request %s: %s", r.Method, r.RequestURI)
}
