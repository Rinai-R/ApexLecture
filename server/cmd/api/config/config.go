package config

type ServerConfig struct {
	Name           string       `json:"name"`
	Host           string       `json:"host"`
	Port           string       `json:"port"`
	UserSrvInfo    RPCSrvConfig `json:"user_srv"`
	LectureSrvInfo RPCSrvConfig `json:"lecture_srv"`
	OtelEndpoint   string       `json:"otel_endpoint"`
}

type RPCSrvConfig struct {
	Name string `json:"name"`
}

type EtcdConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Key  string `json:"key"`
}
