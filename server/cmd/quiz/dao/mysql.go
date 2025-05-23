package dao

import (
	"github.com/Rinai-R/ApexLecture/server/cmd/quiz/model"
	"gorm.io/gorm"
)

type MysqlManagerImpl struct {
	db *gorm.DB
}

func NewMysqlManager(db *gorm.DB) *MysqlManagerImpl {
	db.AutoMigrate(&model.QuizChoice{})
	db.AutoMigrate(&model.QuizJudge{})
	return &MysqlManagerImpl{db: db}
}

func (m *MysqlManagerImpl) CreateQuizChoice(quizChoice *model.QuizChoice) error {
	return m.db.Create(quizChoice).Error
}

func (m *MysqlManagerImpl) CreateQuizJudge(quizJudge *model.QuizJudge) error {
	return m.db.Create(quizJudge).Error
}
