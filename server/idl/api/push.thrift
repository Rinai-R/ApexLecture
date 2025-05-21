namespace go push

include "../base/base.thrift"

enum MessageType {
    CHAT = 1,
}

struct ChatMessage {
    1: required string room_id,
    2: required string user_id,
    3: required string text,
}

union Payload {
    1: ChatMessage chat_message,
}

struct PushMessageResponse {
    1: required i8 type,
    2: required Payload payload,
}

struct PushQuestionRequest {}


service PushService {
    PushMessageResponse Receive(1: PushQuestionRequest request) (api.get="/receive/:roomid")
}