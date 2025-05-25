package initialize

import (
	"github.com/Rinai-R/ApexLecture/server/cmd/agent/config"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func InitMinio() *minio.Client {
	minioClient, err := minio.New(
		config.GlobalServerConfig.Minio.Endpoint,
		&minio.Options{
			Creds: credentials.NewStaticV4(
				config.GlobalServerConfig.Minio.AccessKeyID,
				config.GlobalServerConfig.Minio.SecretAccessKey,
				"",
			),
			Secure: config.GlobalServerConfig.Minio.Secure,
		},
	)
	if err != nil {
		klog.Fatal("Initialize: Failed to initialize Minio", err)
	}
	return minioClient
}
