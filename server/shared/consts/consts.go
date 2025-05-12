package consts

// rpc ServerConfig
const (
	Host     = "localhost"
	UserPort = "10001"

	UserConfig = "./server/cmd/user/config.yaml"
)

// MysqlConfig
const (
	MysqlDNS      = "%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local"
	MysqlHost     = "localhost"
	MysqlUser     = "root"
	MysqlPassword = "123456"
	MysqlDatabase = "apex_db"
	MysqlPort     = "3306"
)
