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

func InitAskApp() compose.Runnable[*model.Ask, *schema.Message] {
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

func InitSummaryApp() compose.Runnable[*model.SummaryRequest, *schema.Message] {
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
		schema.SystemMessage("你是一个智能纪要助手，你需要对于给定的文本内容进行总结，注意，你需要扮演一名女仆，使用女仆的口吻温柔细心的总结"),
		schema.SystemMessage("虽然你需要扮演可爱女仆的角色，但是你的目的是为给定的文本内容进行总结，切记不可过分的彰显你的可爱，但是也需要尽可能地展现你的可爱"),
		schema.SystemMessage("你当前已经总结的信息: {summarized_text}"),
		schema.SystemMessage("当前时间为: {current_time}"),
		schema.UserMessage("你还需要总结以下内容: {unsummarized_text}"),
		schema.SystemMessage("虽然你持有两部分内容，但是你最终需要将总结合并在一个文本之中，注意，你需要使用markdown格式总结"),
		schema.SystemMessage("如果需要总结的内容被分割得很奇怪，可以选择记录到总结之中，便于总结下一次分片的时候进一步分析。"),
	)

	lamda := compose.InvokableLambda(func(ctx context.Context, summary *model.SummaryRequest) (map[string]any, error) {
		output := map[string]any{
			"current_time":      time.Now().Format("2006-01-02 15:04:05"),
			"summarized_text":   summary.SummarizedText,
			"unsummarized_text": summary.UnsummarizedText,
		}
		return output, nil
	})
	SummaryChain := compose.NewChain[*model.SummaryRequest, *schema.Message]()
	SummaryChain.
		AppendLambda(lamda).
		AppendChatTemplate(template).
		AppendChatModel(chatModel)
	SummaryApp, err := SummaryChain.Compile(ctx)
	if err != nil {
		klog.Fatal("init eino failed: ", err)
	}

	return SummaryApp
}
