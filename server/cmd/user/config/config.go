package config

type MysqlConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type EtcdConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Key  string `json:"key"`
}

type ServerConfig struct {
	Name  string      `json:"name"`
	Host  string      `json:"host"`
	Port  string      `json:"port"`
	Mysql MysqlConfig `json:"mysql"`
}
