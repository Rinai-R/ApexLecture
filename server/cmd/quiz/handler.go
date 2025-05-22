package main

import (
	"context"
	quiz "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/quiz"
)

// QuizServiceImpl implements the last service interface defined in the IDL.
type QuizServiceImpl struct{}

// SubmitQuestion implements the QuizServiceImpl interface.
func (s *QuizServiceImpl) SubmitQuestion(ctx context.Context, request *quiz.SubmitQuestionRequest) (resp *quiz.SubmitQuestionResponse, err error) {
	// TODO: Your code here...
	return
}

// SubmitAnswer implements the QuizServiceImpl interface.
func (s *QuizServiceImpl) SubmitAnswer(ctx context.Context, request *quiz.SubmitAnswerRequest) (resp *quiz.SubmitAnswerResponse, err error) {
	// TODO: Your code here...
	return
}
