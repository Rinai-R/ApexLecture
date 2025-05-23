namespace go chat

include "../base/base.thrift"


namespace go chat

// 加上 type 字段，增强可扩展性
struct ChatMessage {
    1: i64 roomId,
    2: i64 userId,
    3: i8 type,
    4: string text,
}

struct ChatMessageResponse {
    1: base.BaseResponse response,
}

service ChatService {
    ChatMessageResponse SendChat(1: ChatMessage msg)
}
