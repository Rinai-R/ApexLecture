package initialize

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"

	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/cloudwego/kitex/pkg/klog"
)

func InitKey() (*rsa.PrivateKey, string) {
	privatePEM, err := os.ReadFile(consts.PrivateKey)
	if err != nil {
		klog.Fatal("initialize: open error ", err)
	}
	block, _ := pem.Decode(privatePEM)
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		klog.Fatalf("initialize: 解析私钥失败: %v", err)
	}

	publicPEM, _ := os.ReadFile(consts.PublicKey)
	if privateKey == nil {
		klog.Fatal("initialize: Private Key is nil")
	}
	return privateKey.(*rsa.PrivateKey), string(publicPEM)
}
