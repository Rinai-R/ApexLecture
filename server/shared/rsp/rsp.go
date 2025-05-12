package rsp

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

const (
	Success                  = 20000
	UsernameOrPasswordLength = 40001
	UsernameNotExists        = 40002
	PasswordError            = 40003
	UnAuthorized             = 40004
)

func OK(data interface{}) Response {
	return Response{
		Code:    Success,
		Message: "OK",
		Data:    data,
	}
}

func ErrorUsernameOrPasswordLength() Response {
	return Response{
		Code:    UsernameOrPasswordLength,
		Message: "Username or password length should be between 4 and 20 characters",
	}
}

func ErrorUsernameNotExists() Response {
	return Response{
		Code:    UsernameNotExists,
		Message: "Username not exists",
	}
}

func ErrorPasswordError() Response {
	return Response{
		Code:    PasswordError,
		Message: "Password error",
	}
}

func ErrorUnAuthorized() Response {
	return Response{
		Code:    UnAuthorized,
		Message: "Unauthorized",
	}
}
