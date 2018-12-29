package streamer

import "io"

type Storage interface {
	Read(string, io.Writer) error
	Write(string, io.Reader) error
}

type Config struct {
	ServiceMap ServiceMap `yaml:"serviceMap,omitempty"`
}

type ServiceMap struct {
	RoutePrefix string `yaml:"routePrefix,omitempty"`
	AgentTarget string `yaml:"agentTarget,omitempty"`
}
