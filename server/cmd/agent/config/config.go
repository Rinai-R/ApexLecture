package config

type ServerConfig struct {
	Name         string         `json:"name"`
	Host         string         `json:"host"`
	Port         string         `json:"port"`
	Mysql        MysqlConfig    `json:"mysql"`
	Redis        RedisConfig    `json:"redis"`
	Kafka        KafkaConfig    `json:"kafka"`
	Minio        MinioConfig    `json:"minio"`
	RabbitMQ     RabbitMQConfig `json:"rabbitmq"`
	OtelEndpoint string         `json:"otel_endpoint"`
}

type EtcdConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Key  string `json:"key"`
}

type MysqlConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	Database int    `json:"database"`
}

type KafkaConfig struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Brokers  []string `json:"brokers"`
	Topic    string   `json:"topic"`
	Group    string   `json:"group"`
}

type MinioConfig struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	BucketName      string `json:"bucket_name"`
	Secure          bool   `json:"secure"`
}

type RabbitMQConfig struct {
	Host               string `json:"host"`
	Port               string `json:"port"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	Vhost              string `json:"vhost"`
	Exchange           string `json:"exchange"`
	DeadLetterExchange string `json:"dead_letter_exchange"`
}
