namespace go push

include "../base/base.thrift"

// 推送服务
// 负责统一推送消息，包括但不限于聊天，题目，教师端查看（有身份校验）的做题情况。

struct ChatMessage {
    1: required i64 room_id,
    2: required i64 user_id,
    3: required string text,
}

union Payload {
    1: optional ChatMessage chat_message,
}

struct PushMessageResponse {
    1: required i8 type,
    2: required Payload payload,
}

struct PushMessageRequest {
    1: required i64 roomId,
    2: required i64 userId,
}


service PushService {
    PushMessageResponse Receive(1: PushMessageRequest request) (streaming.mode="server")
}