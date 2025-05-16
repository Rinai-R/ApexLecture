package rsp

import "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/base"

const (
	Success                   = 20000
	UsernameOrPasswordLength  = 40001
	UsernameNotExists         = 40002
	PasswordError             = 40003
	UnAuthorized              = 40004
	SnowFalkeError            = 40005
	UserCreateError           = 40006
	ParameterError            = 40007
	CreatePeerConnectionError = 40008
	AddtTransceiverError      = 40009
	CreateLectureError        = 40010
	SetRemoteDescriptionError = 40011
	CreateAnswerError         = 40012
	SetLocalDescriptionError  = 40013
	RoomNotFound              = 40014
	AddTrackError             = 40015
)

func OK() *base.BaseResponse {
	return &base.BaseResponse{
		Code:    Success,
		Message: "OK",
	}
}

func ErrorUsernameOrPasswordLength() *base.BaseResponse {
	return &base.BaseResponse{
		Code:    UsernameOrPasswordLength,
		Message: "Username or password length should be between 4 and 20 characters",
	}
}

func ErrorUsernameNotExists() *base.BaseResponse {
	return &base.BaseResponse{
		Code:    UsernameNotExists,
		Message: "Username not exists",
	}
}

func ErrorPasswordError() *base.BaseResponse {
	return &base.BaseResponse{
		Code:    PasswordError,
		Message: "Password error",
	}
}

func ErrorUnAuthorized(define string) *base.BaseResponse {
	return &base.BaseResponse{
		Code:    UnAuthorized,
		Message: define,
	}
}

func ErrorSnowFalke(define string) *base.BaseResponse {
	return &base.BaseResponse{
		Code:    SnowFalkeError,
		Message: define,
	}
}

func ErrorUserCreate(define string) *base.BaseResponse {
	return &base.BaseResponse{
		Code:    UserCreateError,
		Message: define,
	}
}

func ErrorParameter(define string) *base.BaseResponse {
	return &base.BaseResponse{
		Code:    ParameterError,
		Message: define,
	}
}

func ErrorCreatePeerConnection(define string) *base.BaseResponse {
	return &base.BaseResponse{
		Code:    CreatePeerConnectionError,
		Message: define,
	}
}

func ErrorAddtTransceiver(define string) *base.BaseResponse {
	return &base.BaseResponse{
		Code:    AddtTransceiverError,
		Message: define,
	}
}

func ErrorCreateLecture(define string) *base.BaseResponse {
	return &base.BaseResponse{
		Code:    CreateLectureError,
		Message: define,
	}
}

func ErrorSetRemoteDescription(define string) *base.BaseResponse {
	return &base.BaseResponse{
		Code:    SetRemoteDescriptionError,
		Message: define,
	}
}

func ErrorCreateAnswer(define string) *base.BaseResponse {
	return &base.BaseResponse{
		Code:    CreateAnswerError,
		Message: define,
	}
}

func ErrorSetLocalDescription(define string) *base.BaseResponse {
	return &base.BaseResponse{
		Code:    SetLocalDescriptionError,
		Message: define,
	}
}

func ErrorRoomNotFound() *base.BaseResponse {
	return &base.BaseResponse{
		Code:    RoomNotFound,
		Message: "Room not found",
	}
}

func ErrorAddTrack(define string) *base.BaseResponse {
	return &base.BaseResponse{
		Code:    AddTrackError,
		Message: define,
	}
}
