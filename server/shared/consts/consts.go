package consts

// ApiConfig
const (
	SrvLen       = 2
	UserSrvno    = 0
	LectureSrvno = 1
)

// ServerConfig
const (
	Host        = "0.0.0.0"
	ApiPort     = "10000"
	UserPort    = "10001"
	LecturePort = "10002"

	ApiConfig     = "./server/cmd/api/config.yaml"
	UserConfig    = "./server/cmd/user/config.yaml"
	LectureConfig = "./server/cmd/lecture/config.yaml"

	ApiSrvPrefix     = "api"
	UserSrvPrefix    = "user"
	LectureSrvPrefix = "lecture"
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

// key
const (
	PrivateKey = "./server/cmd/user/keys/private_key.pem"
	PublicKey  = "./server/cmd/user/keys/public_key.pem"
)

// Other
const (
	HlogFilePath = "./tmp/hlog/"
	KlogFilePath = "./tmp/klog/"

	UserSrvSnowFlakeNode = 1
	UserIDSnowFlakeNode  = 2
)
