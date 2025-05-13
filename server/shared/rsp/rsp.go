package rsp

import "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/base"

const (
	Success                  = 20000
	UsernameOrPasswordLength = 40001
	UsernameNotExists        = 40002
	PasswordError            = 40003
	UnAuthorized             = 40004
)

func OK(data interface{}) base.BaseResponse {
	return base.BaseResponse{
		Code:    Success,
		Message: "OK",
	}
}

func ErrorUsernameOrPasswordLength() base.BaseResponse {
	return base.BaseResponse{
		Code:    UsernameOrPasswordLength,
		Message: "Username or password length should be between 4 and 20 characters",
	}
}

func ErrorUsernameNotExists() base.BaseResponse {
	return base.BaseResponse{
		Code:    UsernameNotExists,
		Message: "Username not exists",
	}
}

func ErrorPasswordError() base.BaseResponse {
	return base.BaseResponse{
		Code:    PasswordError,
		Message: "Password error",
	}
}

func ErrorUnAuthorized() base.BaseResponse {
	return base.BaseResponse{
		Code:    UnAuthorized,
		Message: "Unauthorized",
	}
}
