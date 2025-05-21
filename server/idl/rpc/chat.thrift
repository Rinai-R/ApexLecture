namespace go chat

include "../base/base.thrift"


namespace go chat

struct ChatMessage {
    1: i64 roomId,
    2: i64 userId,
    3: string text,
}

struct ChatMessageResponse {
    1: base.BaseResponse response,
}

service ChatService {
    ChatMessageResponse SendChat(1: ChatMessage msg)
}
