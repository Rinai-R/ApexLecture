package middleware

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/Rinai-R/ApexLecture/server/cmd/api/config"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/user"
	"github.com/Rinai-R/ApexLecture/server/shared/rsp"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/golang-jwt/jwt/v5"
)

func JwtAuth(ctx context.Context, c *app.RequestContext) {
	resp, _ := config.UserClient.GetPublicKey(ctx, &user.GetPublicKeyRequest{})
	block, _ := pem.Decode([]byte(resp.PublicKey))
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, rsp.ErrorInternalServer(err.Error()))
		hlog.Error("JwtAuth ParsePKIXPublicKey error: %s", err.Error())
		c.Abort()
		return
	}
	tokenString := c.Request.Header.Get("Authorization")
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return publicKey.(*rsa.PublicKey), nil
		},
		jwt.WithValidMethods([]string{"RS256"}),
	)
	if err != nil || !token.Valid {
		c.JSON(consts.StatusUnauthorized, rsp.ErrorUnAuthorized(err.Error()))
		c.Abort()
		return
	}
	userid, ok := claims["sub"]
	if !ok {
		c.JSON(consts.StatusUnauthorized, rsp.ErrorUnAuthorized("Invalid token"))
		c.Abort()
		return
	}
	c.Set("userid", userid)
	c.Next(ctx)
}
