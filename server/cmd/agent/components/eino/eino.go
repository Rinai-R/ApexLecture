package eino

import (
	"context"

	"github.com/Rinai-R/ApexLecture/server/cmd/agent/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type BotManager struct {
	AskApp compose.Runnable[*model.Ask, *schema.Message]
}

func NewBotManaer(AskApp compose.Runnable[*model.Ask, *schema.Message]) *BotManager {
	return &BotManager{
		AskApp: AskApp,
	}
}

func (bm *BotManager) Ask(ctx context.Context, AskMsg *model.Ask) string {
	output, err := bm.AskApp.Invoke(ctx, AskMsg)
	if err != nil {
		return ""
	}
	return output.Content
}
