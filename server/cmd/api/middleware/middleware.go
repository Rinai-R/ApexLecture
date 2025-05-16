package middleware

import (
	"context"

	"github.com/Rinai-R/ApexLecture/server/cmd/api/config"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/user"
	"github.com/Rinai-R/ApexLecture/server/shared/rsp"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/golang-jwt/jwt/v5"
)

func JwtAuth(ctx context.Context, c *app.RequestContext) {
	resp, _ := config.UserClient.GetPublicKey(ctx, &user.GetPublicKeyRequest{})

	publickey := resp.PublicKey
	tokenString := c.Request.Header.Get("Authorization")
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return publickey, nil
		},
		jwt.WithValidMethods([]string{"RS256"}),
	)
	if err != nil || !token.Valid {
		c.JSON(consts.StatusUnauthorized, rsp.ErrorUnAuthorized(err.Error()))
		c.Abort()
		return
	}
	c.Set("userid", claims["sub"])
	c.Next(ctx)
}
