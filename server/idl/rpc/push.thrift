namespace go push

include "../base/base.thrift"

// 推送服务
// 负责统一推送消息，包括但不限于聊天，题目，教师端统计做题情况（有身份校验）的做题情况。

struct ChatMessage {
    1: required i64 roomId,
    2: required i64 userId,
    3: required string text,
}

struct ChoiceQuestion {
    1: required i64 roomId,
    2: required i64 userId,
    3: required i64 questionId,
    4: required string title,
    5: required list<string> options,
    6: required i64 ttl,
}

struct JudgeQuestion {
    1: required i64 roomId,
    2: required i64 userId,
    3: required i64 questionId,
    4: required string title,
    5: required i64 ttl,
}

struct QuizStatus {
    1: required i64 roomId,
    2: required i64 questionId,
    3: required i64 requiredNum,
    4: required i64 currentNum,
    5: required double acceptRate
}

struct ControlMessage {
    2: required i64 roomId,
    3: required string operation,
}

union Payload {
    1: optional ChatMessage chatMessage,
    2: optional ChoiceQuestion choiceQuestion,
    3: optional JudgeQuestion judgeQuestion,
    4: optional QuizStatus quizStatus,
    5: optional ControlMessage controlMessage,
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