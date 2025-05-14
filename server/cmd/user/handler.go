package main

import (
	"context"
	"crypto/rsa"
	"time"

	"github.com/Rinai-R/ApexLecture/server/cmd/user/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/user/model"
	"github.com/Rinai-R/ApexLecture/server/cmd/user/pkg/encrypt"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/base"
	user "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/user"
	"github.com/Rinai-R/ApexLecture/server/shared/rsp"
	"github.com/bwmarrin/snowflake"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/golang-jwt/jwt/v5"
)

// UserServiceImpl implements the last service interface defined in the IDL.
type UserServiceImpl struct {
	MysqlManager
	PrivateKey *rsa.PrivateKey
	PublicKey  string
}

type MysqlManager interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
}

var _ MysqlManager = (*dao.DM)(nil)

// Register implements the UserServiceImpl interface.
func (s *UserServiceImpl) Register(ctx context.Context, request *user.RegisterRequest) (*user.RegisterResponse, error) {
	sf, err := snowflake.NewNode(consts.UserIDSnowFlakeNode)
	if err != nil {
		klog.Errorf("user register: create snowflake node failed: %v", err)
		resp := &user.RegisterResponse{
			Base: &base.BaseResponse{
				Code:    rsp.SnowFalkeError,
				Message: "create snowflake node failed",
			},
		}
		return resp, nil
	}
	// 雪花算法生成用户ID
	userid := sf.Generate().Int64()
	// 加密密码
	password := encrypt.EncryptPassword(request.Password)
	err = s.CreateUser(ctx, &model.User{
		ID:       userid,
		Username: request.Username,
		Password: password,
	})
	if err != nil {
		klog.Errorf("user register: create user failed: %v", err)
		resp := &user.RegisterResponse{
			Base: &base.BaseResponse{
				Code:    rsp.UserCreateError,
				Message: "create user failed, maybe username already exists",
			},
		}
		return resp, nil
	}

	resp := &user.RegisterResponse{
		Base: &base.BaseResponse{
			Code:    rsp.Success,
			Message: "register success",
		},
		Id: userid,
	}
	return resp, nil
}

// Login implements the UserServiceImpl interface.
func (s *UserServiceImpl) Login(ctx context.Context, request *user.LoginRequest) (*user.LoginResponse, error) {
	userInfo, err := s.GetUserByUsername(ctx, request.Username)
	if err != nil {
		klog.Errorf("user login: get user by username failed: %v", err)
		resp := &user.LoginResponse{
			Base: &base.BaseResponse{
				Code:    rsp.UsernameNotExists,
				Message: "get user by username failed",
			},
		}
		return resp, nil
	}
	// 验证密码
	if encrypt.ComparePasswords(userInfo.Password, request.Password) {
		resp := &user.LoginResponse{
			Base: &base.BaseResponse{
				Code:    rsp.Success,
				Message: "login success",
			},
			Token: s.GenerateToken(userInfo.ID),
		}
		return resp, nil
	}
	resp := &user.LoginResponse{
		Base: &base.BaseResponse{
			Code:    rsp.PasswordError,
			Message: "password error",
		},
	}
	return resp, nil
}

func (a *UserServiceImpl) GenerateToken(uid int64) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": "user",
		"exp": time.Now().Add(time.Hour * 2).Unix(),
		"uid": uid,
	})

	tokenString, err := token.SignedString(a.PrivateKey)
	if err != nil {
		return ""
	}
	return tokenString
}

// GetPublicKey implements the UserServiceImpl interface.

func (s *UserServiceImpl) GetPublicKey(ctx context.Context, request *user.GetPublicKeyRequest) (*user.GetPublicKeyResponse, error) {
	resp := &user.GetPublicKeyResponse{
		PublicKey: s.PublicKey,
	}
	return resp, nil
}
