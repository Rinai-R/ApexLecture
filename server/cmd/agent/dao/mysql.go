package dao

import (
	"context"

	"github.com/Rinai-R/ApexLecture/server/cmd/agent/model"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"gorm.io/gorm"
)

type MysqlManagerImpl struct {
	db *gorm.DB
}

func NewMysqlManager(db *gorm.DB) *MysqlManagerImpl {
	return &MysqlManagerImpl{db: db}
}

// 这里有点绕，如果是 false 并且 err == nil， 表示我们是第一次来，需要创建结构体
// 但是如果是 true 并且 err == nil， 表示总结就结束了
// 如果是 false 并且 err != nil 并且 false 则是其他错误， 需要返回
func (m *MysqlManagerImpl) IsSummaried(ctx context.Context, RoomId int64) int8 {
	summary := model.Summary{}

	result := m.db.First(&summary, "roomid = ?", RoomId)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 这里表示我们是第一次来，需要在表里面创建
			return consts.NotCreate
		}
		// 其他错误，返回
		return consts.OtherError
	}

	if summary.Status {
		// 已经总结完了，直接返回
		return consts.Summarized
	}
	// 这里需要继续调用 bot 进行总结
	return consts.NoSummary
}

func (m *MysqlManagerImpl) CreateSummary(ctx context.Context, summary *model.Summary) error {
	return m.db.Create(summary).Error
}

func (m *MysqlManagerImpl) GetSummary(ctx context.Context, RoomId int64) (*model.Summary, error) {
	summary := model.Summary{}

	result := m.db.First(&summary, "roomid = ?", RoomId)

	if result.Error != nil {
		return nil, result.Error
	}

	return &summary, nil
}
