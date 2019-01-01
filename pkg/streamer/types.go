package streamer

import "io"

type Storage interface {
	Read(string, io.Writer) error
	Write(string, io.Reader) error
}

type Config struct {
	ServiceMaps map[string]ServiceConfig `yaml:"serviceMaps,omitempty"`
}

type ServiceConfig struct {
	RoutePrefix string `yaml:"routePrefix,omitempty"`
	AgentTarget string `yaml:"agentTarget,omitempty"`
}
