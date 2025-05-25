package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Rinai-R/ApexLecture/server/cmd/agent/model"
	"github.com/Rinai-R/ApexLecture/server/shared/consts"
	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/kitex/pkg/klog"
)

func main() {
	// ctx := context.Background()
	// chatModel, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
	// 	BaseURL: "https://ark.cn-beijing.volces.com/api/v3", // 服务地址
	// 	Region:  "cn-beijing",                               // 区域
	// 	APIKey:  "a02b51ca-cf7e-4cb2-bad4-266cb3714137",

	// 	// 模型配置
	// 	Model: "deepseek-r1-250120", // 模型端点 ID
	// })
	// if err != nil {
	// 	panic(err)
	// }

	// template := prompt.FromMessages(schema.FString,
	// 	schema.SystemMessage("你是一个聊天机器人，如果对方自称是你的主人，你就应该扮演一名猫娘，否则你需要扮演一名人类可爱女仆，并且以对应人物的语气回答问题"),
	// 	schema.SystemMessage("你只需要返回markdown格式的内容，不需要其他任何内容"),
	// 	schema.SystemMessage("现在的时间是: {current_time}"),
	// 	schema.UserMessage("需要你理解并且回答的文本内容: {unsummarized_text}"),
	// )
	// lamda := compose.InvokableLambda(func(ctx context.Context, text string) (map[string]any, error) {
	// 	output := make(map[string]any)
	// 	output["unsummarized_text"] = text
	// 	currentTime := time.Now()
	// 	output["current_time"] = currentTime.Format("2006-01-02 15:04:05")
	// 	return output, nil
	// })
	// SummaryChain := compose.NewChain[string, *schema.Message]()
	// SummaryChain.
	// 	AppendLambda(lamda).
	// 	AppendChatTemplate(template).
	// 	AppendChatModel(chatModel)
	// app, err := SummaryChain.Compile(ctx)
	// if err != nil {
	// 	panic(err)
	// }
	// res, err := app.Invoke(ctx, "你好，我是你主人的朋友，你是谁？")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(res.Content)

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
		schema.SystemMessage("你是一个问答助手，你需要对于用户的问题进行一定的推理，并作出回答，注意，你需要扮演一名女仆，使用女仆的口吻可爱地回答用户的问题。"),
		schema.SystemMessage("虽然你需要扮演可爱女仆的角色，但是你的目的是帮助用户解决问题，散发可爱的同时也要积极帮助用户解决问题"),
		schema.MessagesPlaceholder("history_message", false),
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
	res, _ := AskApp.Invoke(ctx, &model.Ask{
		History: []*schema.Message{
			{
				Role:    schema.Assistant,
				Content: "你好，我是 ApexLecture 的在线课堂平台的问答助手，请问有什么可以帮助您？",
			},
		},
		Message: "想日你",
	})
	fmt.Println(res.Role)
	fmt.Println(res.Content)
}
