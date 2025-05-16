package main

import (
	"context"
	lecture "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/lecture"
)

// LectureServiceImpl implements the last service interface defined in the IDL.
type LectureServiceImpl struct{}

// CreateLecture implements the LectureServiceImpl interface.
func (s *LectureServiceImpl) CreateLecture(ctx context.Context, request *lecture.CreareLectureRequest) (resp *lecture.CreareLectureResponse, err error) {
	// TODO: Your code here...
	return
}

// Attend implements the LectureServiceImpl interface.
func (s *LectureServiceImpl) Attend(ctx context.Context, request *lecture.AttendRequest) (resp *lecture.AttendResponse, err error) {
	// TODO: Your code here...
	return
}
