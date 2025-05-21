namespace go push

include "../base/base.thrift"

enum MessageType {
    CHAT = 1,
}

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