package initialize

import (
	"context"
	"time"

	"github.com/Rinai-R/ApexLecture/server/cmd/agent/model"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/kitex/pkg/klog"
)

func InitEino() compose.Runnable[*model.Ask, *schema.Message] {
	ctx := context.Background()
	chatModel, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		BaseURL: consts.AgentBaseURL,
		Region:  consts.AgentRegion,
		APIKey:  consts.AgentAPIKey,
		Model:   consts.AgentModel,
	})
	if err != nil {
		klog.Fatal("init eino failed: ", err)
	}

	template := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage("你是一个名叫 ApexLecture 的在线课堂平台的问答助手，你需要对于用户的问题进行一定的推理，并作出回答，注意，你需要扮演一名女仆，使用女仆的口吻可爱地回答用户的问题。"),
		schema.SystemMessage("虽然你需要扮演可爱女仆的角色，但是你的目的是帮助用户解决问题，散发可爱的同时也要积极帮助用户解决问题"),
		schema.MessagesPlaceholder("history_message", true),
		schema.SystemMessage("你和用户的历史消息为: {history_message}"),
		schema.SystemMessage("当前时间为: {current_time}"),
		schema.UserMessage("用户当前发出的消息为: {current_message}"),
	)

	lamda := compose.InvokableLambda(func(ctx context.Context, ask *model.Ask) (map[string]any, error) {
		output := map[string]any{
			"current_time":    time.Now().Format("2006-01-02 15:04:05"),
			"current_message": ask.Message,
			"history_message": ask.History,
		}
		return output, nil
	})
	AskChain := compose.NewChain[*model.Ask, *schema.Message]()
	AskChain.
		AppendLambda(lamda).
		AppendChatTemplate(template).
		AppendChatModel(chatModel)
	AskApp, err := AskChain.Compile(ctx)
	if err != nil {
		klog.Fatal("init eino failed: ", err)
	}

	return AskApp
}
