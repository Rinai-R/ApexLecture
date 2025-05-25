package eino

import (
	"context"

	"github.com/Rinai-R/ApexLecture/server/cmd/agent/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type BotManagerImpl struct {
	AskApp     compose.Runnable[*model.Ask, *schema.Message]
	SummaryApp compose.Runnable[*model.SummaryRequest, *schema.Message]
}

func NewBotManaer(AskApp compose.Runnable[*model.Ask, *schema.Message], SummaryApp compose.Runnable[*model.SummaryRequest, *schema.Message]) *BotManagerImpl {
	return &BotManagerImpl{
		AskApp:     AskApp,
		SummaryApp: SummaryApp,
	}
}

func (bm *BotManagerImpl) Ask(ctx context.Context, AskMsg *model.Ask) *model.AskResponse {
	output, err := bm.AskApp.Invoke(ctx, AskMsg)
	if err != nil {
		return nil
	}
	return &model.AskResponse{
		Role:    string(output.Role),
		Message: output.Content,
	}
}

func (bm *BotManagerImpl) Summary(ctx context.Context, SummaryReq *model.SummaryRequest) *model.SummaryResponse {
	output, err := bm.SummaryApp.Invoke(ctx, SummaryReq)
	if err != nil {
		return nil
	}
	return &model.SummaryResponse{
		Summary: output.Content,
	}
}
