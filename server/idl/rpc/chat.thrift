namespace go chat

include "../base/base.thrift"


namespace go chat

struct ChatMessage {
    1: string room_id,
    2: string user_id,
    3: string text,
}

struct ChatMessageResponse {
    1: base.BaseResponse response,
}

service ChatService {
    ChatMessageResponse SendChat(1: ChatMessage msg)
}
