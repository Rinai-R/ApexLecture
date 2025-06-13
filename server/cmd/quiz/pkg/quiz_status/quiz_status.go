package service

import (
	"context"
	"time"

	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/dao"
	"github.com/cloudwego/kitex/pkg/klog"
)

type QuizStatusHanlderImpl struct {
	RedisManager *dao.RedisManagerImpl
}

func NewQuizStatusHanlder(RedisManager *dao.RedisManagerImpl) *QuizStatusHanlderImpl {
	return &QuizStatusHanlderImpl{
		RedisManager: RedisManager,
	}
}

// 这里如果 mysql 数据插入失败不会及时停止，可以加个定期检查 mysql 库存。
func (q *QuizStatusHanlderImpl) HandleStatus(ctx context.Context, questionId int64, roomId int64) error {
	// 固定三秒推送一次。
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if _, err := q.RedisManager.GetAnswer(ctx, questionId); err != nil {
				klog.Info("Question Has Ended", err)
				return nil
			}
			status, err := q.RedisManager.GetQuizStatus(ctx, questionId, roomId)
			if err != nil {
				klog.Error("Get Quiz Status Error: ", err)
				continue
			}
			if err := q.RedisManager.SendQuizStatus(ctx, status); err != nil {
				klog.Error("Send Quiz Status Error: ", err)
				continue
			}
		case <-ctx.Done():
			return nil
		}
	}
}
