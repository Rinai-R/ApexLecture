package main

import (
	"context"
	"crypto/rsa"
	"time"

	"github.com/Rinai-R/ApexLecture/server/cmd/user/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/user/model"
	"github.com/Rinai-R/ApexLecture/server/cmd/user/pkg/encrypt"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
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
	// 创建雪花算法节点
	sf, err := snowflake.NewNode(consts.UserIDSnowFlakeNode)
	if err != nil {
		klog.Errorf("user register: create snowflake node failed: %v", err)
		resp := &user.RegisterResponse{
			Base: rsp.ErrorSnowFalke(err.Error()),
		}
		return resp, nil
	}
	// 雪花算法生成用户ID
	userid := sf.Generate().Int64()
	// 加密密码
	password := encrypt.EncryptPassword(request.Password)
	// 数据库操作
	err = s.CreateUser(ctx, &model.User{
		ID:       userid,
		Username: request.Username,
		Password: password,
	})
	if err != nil {
		klog.Errorf("user register: create user failed: %v", err)
		resp := &user.RegisterResponse{
			Base: rsp.ErrorUserCreate(err.Error()),
		}
		return resp, nil
	}
	// 返回注册成功响应
	resp := &user.RegisterResponse{
		Base: rsp.OK(),
		Id:   userid,
	}
	return resp, nil
}

// Login implements the UserServiceImpl interface.
func (s *UserServiceImpl) Login(ctx context.Context, request *user.LoginRequest) (*user.LoginResponse, error) {
	// 数据库查询
	userInfo, err := s.GetUserByUsername(ctx, request.Username)
	if err != nil {
		klog.Errorf("user login: get user by username failed: %v", err)
		resp := &user.LoginResponse{
			Base: rsp.ErrorUsernameNotExists(),
		}
		return resp, nil
	}
	// 验证密码
	if encrypt.ComparePasswords(userInfo.Password, request.Password) {
		resp := &user.LoginResponse{
			Base:  rsp.OK(),
			Token: s.GenerateToken(userInfo.ID),
		}
		return resp, nil
	}
	resp := &user.LoginResponse{
		Base: rsp.ErrorPasswordError(),
	}
	return resp, nil
}

func (a *UserServiceImpl) GenerateToken(uid int64) string {
	// 产生 Token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": uid,
		"exp": time.Now().Add(time.Hour * 2).Unix(),
	})
	// 私钥加密
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
