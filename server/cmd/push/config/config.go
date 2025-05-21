package config

type ServerConfig struct {
	Name         string      `json:"name"`
	Host         string      `json:"host"`
	Port         string      `json:"port"`
	Redis        RedisConfig `json:"redis"`
	OtelEndpoint string      `json:"otel_endpoint"`
}

type EtcdConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Key  string `json:"key"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	Database int    `json:"database"`
}
