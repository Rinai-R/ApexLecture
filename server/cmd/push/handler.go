package main

import (
	"context"
	push "github.com/Rinai-R/ApexLecture/server/shared/kitex_gen/push"
)

// PushServiceImpl implements the last service interface defined in the IDL.
type PushServiceImpl struct{}

// Receive implements the PushServiceImpl interface.
func (s *PushServiceImpl) Receive(ctx context.Context, request *push.PushQuestionRequest) (resp *push.PushMessageResponse, err error) {
	// TODO: Your code here...
	return
}
