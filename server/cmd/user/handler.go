package main

import (
	"context"

	"github.com/Rinai-R/ApexLecture/server/cmd/user/dao"
	"github.com/Rinai-R/ApexLecture/server/cmd/user/model"
	user "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/user"
)

// UserServiceImpl implements the last service interface defined in the IDL.
type UserServiceImpl struct {
	MysqlManager
}

type MysqlManager interface {
	CreateUser(ctx context.Context, user *model.User) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
}

var _ MysqlManager = (*dao.DM)(nil)

// Register implements the UserServiceImpl interface.
func (s *UserServiceImpl) Register(ctx context.Context, request *user.RegisterRequest) (resp *user.RegisterResponse, err error) {
	// TODO: Your code here...
	return
}

// Login implements the UserServiceImpl interface.
func (s *UserServiceImpl) Login(ctx context.Context, request *user.LoginRequest) (resp *user.LoginResponse, err error) {
	// TODO: Your code here...
	return
}
