namespace go base

struct BaseResponse {
    1: i64 code;
    2: string message;
}

struct NilResponse {}

struct NilRequest {}


// 队列内部通信结构体

enum InternalMessageType {
    CHAT_MESSAGE = 1,
    CONTRAL_MESSAGE = 2,
}

struct InternalMessage {
    1: required InternalMessageType type;
    2: required InternalPayload payload;
}

union InternalPayload {
    1: optional InternalChatMessage chatMessage;
    2: optional InternalControlMessage controlMessage;
}

struct InternalChatMessage {
    1: required i64 roomId;
    2: required i64 userId;
    3: required string message;
}

struct InternalControlMessage {
    1: required string operation;
}