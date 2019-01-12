package skeleton

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

type Service struct {
	name         string
	streamerHost string
}

func NewService(r *mux.Router, name, streamerHost string) *Service {
	s := &Service{
		streamerHost: streamerHost,
		name:         name,
	}
	sr := r.PathPrefix("/" + name).Subrouter()
	sr.PathPrefix("/u").HandlerFunc(s.handleUpload).Methods(http.MethodPost, http.MethodPut, http.MethodPatch)
	sr.PathPrefix("/d").HandlerFunc(s.handleDownload).Methods(http.MethodGet)
	http.Handle("/", r)
	return s
}

func (s *Service) handleUpload(w http.ResponseWriter, r *http.Request) {
	log.Infof("Received upload request %s: %s", r.Method, r.RequestURI)
	relativePath := strings.TrimPrefix(r.RequestURI, fmt.Sprintf("/%s/u/", s.name))
	redirectURL := fmt.Sprintf("http://%s/streamer/uploads/%s/%s", s.streamerHost, s.name, relativePath)
	log.Infof("Redirecting upload request to %s", redirectURL)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (s *Service) handleDownload(w http.ResponseWriter, r *http.Request) {
	log.Infof("Received download request %s: %s", r.Method, r.RequestURI)
	relativePath := strings.TrimPrefix(r.RequestURI, fmt.Sprintf("/%s/d/", s.name))
	redirectURL := fmt.Sprintf("http://%s/streamer/downloads/%s/%s", s.streamerHost, s.name, relativePath)
	log.Infof("Redirecting download request to %s", redirectURL)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
