package config

type MysqlConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type MinioConfig struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	BucketName      string `json:"bucket_name"`
	Secure          bool   `json:"secure"`
}

type EtcdConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Key  string `json:"key"`
}

type ServerConfig struct {
	Name               string       `json:"name"`
	Host               string       `json:"host"`
	Port               string       `json:"port"`
	Mysql              MysqlConfig  `json:"mysql"`
	Minio              MinioConfig  `json:"minio"`
	InteractionSrvInfo RPCSrvConfig `json:"interaction_srv_info"`
	OtelEndpoint       string       `json:"otel_endpoint"`
}

type RPCSrvConfig struct {
	Name string `json:"name"`
}
