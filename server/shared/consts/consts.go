package consts

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

	OtelEndpoint = "localhost:4317"
)

// MysqlConfig
const (
	MysqlDNS      = "%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FShanghai"
	MysqlHost     = "localhost"
	MysqlUser     = "root"
	MysqlPassword = "123456"
	MysqlDatabase = "apex_db"
	MysqlPort     = "3306"
)

// MinioConfig
const (
	MinioEndpoint  = "localhost:9000"
	MinioAccessKey = "UhT0a4ETDSt5w6WMrdnL"
	MinioSecretKey = "fQTAGA6OhiU0PBgie6ReA9MZJGTht2ZV4frWhxvu"
	MinioBucket    = "lecture"
	MinioSecure    = false
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
	IvfFilePath  = "./tmp/record/%s/video.ivf"
	OggFilePath  = "./tmp/record/%s/audio.ogg"

	EtcdSnowFlakeNode         = 1
	UserIDSnowFlakeNode       = 2
	LectureIDSnowFlakeNode    = 3
	AttendanceIDSnowFlakeNode = 4
)
