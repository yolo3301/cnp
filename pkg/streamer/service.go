package streamer

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

type Service struct {
	storage     Storage
	config      *Config
	cachedConns *sync.Map
}

func NewService(cfgPath string, router *mux.Router, storage Storage) (*Service, error) {
	cfg, err := loadConfig(cfgPath)
	if err != nil {
		return nil, err
	}
	s := &Service{
		storage:     storage,
		config:      cfg,
		cachedConns: &sync.Map{},
	}
	s.registerRoutes(router)
	http.Handle("/", router)
	return s, nil
}

// TODO: add config validation.
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

func (s *Service) getConn(service string) (*grpc.ClientConn, error) {
	if conn, ok := s.cachedConns.Load(service); ok {
		return conn.(*grpc.ClientConn), nil
	}
	if cfg, ok := s.config.ServiceMaps[service]; ok {
		// TODO: make it secure?
		conn, err := grpc.Dial(cfg.AgentTarget, grpc.WithInsecure())
		if err != nil {
			return nil, err
		}
		s.cachedConns.Store(service, conn)
		return conn, nil
	}
	return nil, fmt.Errorf("didn't find configuration for service %q", service)
}

func (s *Service) matchService(relativePath string) (string, bool) {
	for key, val := range s.config.ServiceMaps {
		if strings.HasPrefix(relativePath, val.RoutePrefix) {
			return key, true
		}
	}
	return "", false
}

func (s *Service) handleUpload(w http.ResponseWriter, r *http.Request) {
	log.Infof("Received upload request %s: %s", r.Method, r.RequestURI)

	relativePath := strings.TrimPrefix(r.RequestURI, "/streamer/uploads/")
	svc, ok := s.matchService(relativePath)
	if !ok {
		log.Infof("didn't find matching service for URL %s", r.RequestURI)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	conn, err := s.getConn(svc)
	if err != nil {
		log.Errorf("failed to create RPC connection to backend agent for service %q: %v", svc, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	client := NewStreamerAgentClient(conn)
	reqID := uuid.New().String()

	firstReq := initNotificationReq(reqID, r)
	firstReq.First = true
	firstRes, err := client.OnNotification(r.Context(), firstReq)
	if err != nil {
		log.Errorf("failed first RPC request to backend agent for request %s: %v", r.RequestURI, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if firstRes.GetResponse() != nil {
		convertToHTTPResponse(firstRes.GetResponse(), w)
		return
	}

	if firstRes.GetDropTarget() == nil {
		log.Errorf("first RPC request didn't return a drop target for request %s", r.RequestURI)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dstFile := firstRes.GetDropTarget().GetFileTarget().GetPath()
	err = s.storage.Write(dstFile, r.Body)
	defer r.Body.Close()

	lastReq := initNotificationReq(reqID, r)
	lastReq.AgentNote = firstRes.AgentNote

	if err != nil {
		lastReq.FinalStatus = int32(StreamNotificationRequest_INTERNAL_ERROR)
	} else {
		lastReq.FinalStatus = int32(StreamNotificationRequest_OK)
		// TODO: add final size and hash
	}

	lastRes, err := client.OnNotification(r.Context(), lastReq)
	if err != nil {
		log.Errorf("failed last RPC request to backend agent for request %s: %v", r.RequestURI, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if lastRes.GetResponse() == nil {
		log.Errorf("last RPC request didn't return HTTP response for request %s", r.RequestURI)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	convertToHTTPResponse(lastRes.GetResponse(), w)
}

func (s *Service) handleDownload(w http.ResponseWriter, r *http.Request) {
	log.Infof("Received download request %s: %s", r.Method, r.RequestURI)

	relativePath := strings.TrimPrefix(r.RequestURI, "/streamer/downloads/")
	svc, ok := s.matchService(relativePath)
	if !ok {
		log.Infof("didn't find matching service for URL %s", r.RequestURI)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	conn, err := s.getConn(svc)
	if err != nil {
		log.Errorf("failed to create RPC connection to backend agent for service %q: %v", svc, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	client := NewStreamerAgentClient(conn)
	reqID := uuid.New().String()

	firstReq := initNotificationReq(reqID, r)
	firstReq.First = true
	firstRes, err := client.OnNotification(r.Context(), firstReq)
	if err != nil {
		log.Errorf("failed first RPC request to backend agent for request %s: %v", r.RequestURI, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if firstRes.GetResponse() == nil {
		log.Errorf("first RPC request didn't return HTTP response for request %s", r.RequestURI)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if firstRes.GetResponse().GetPayload() != nil {
		src := firstRes.GetResponse().GetPayload().GetFileObject().GetPath()
		s.storage.Read(src, w)
	}

	convertToHTTPResponse(firstRes.GetResponse(), w)
}

func initNotificationReq(id string, r *http.Request) *StreamNotificationRequest {
	return &StreamNotificationRequest{
		RequestId: id,
		Request: &HttpRequestInfo{
			Method: r.Method,
			ReqUri: r.RequestURI,
			Header: convertToHeader(r.Header),
		},
	}
}

func convertToHeader(header http.Header) []*KeyValue {
	res := make([]*KeyValue, 0)
	for k, v := range header {
		for _, val := range v {
			res = append(res, &KeyValue{Key: k, Value: val})
		}
	}
	return res
}

func convertToHTTPResponse(res *HttpResponseInfo, w http.ResponseWriter) {
	for _, kv := range res.GetHeader() {
		w.Header().Add(kv.GetKey(), kv.GetValue())
	}
	w.WriteHeader(int(res.GetHttpStatus()))
}
