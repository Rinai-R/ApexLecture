namespace go base

struct BaseResponse {
    1: i64 code;
    2: string message;
}

struct NilResponse {}

struct NilRequest {}


// Redis 内部通信结构体

enum InternalMessageType {
    CHAT_MESSAGE = 1,
    CONTRAL_MESSAGE = 2,
    QUIZ_CHOICE = 3,
    QUIZ_JUDGE = 4,
    QUIZ_STATUS = 5,
}

struct InternalMessage {
    1: required InternalMessageType type;
    2: required InternalPayload payload;
}

union InternalPayload {
    1: optional InternalChatMessage chatMessage;
    2: optional InternalControlMessage controlMessage;
    3: optional InternalQuizChoice quizChoice;
    4: optional InternalQuizJudge quizJudge;
    5: optional InternalQuizStatus quizStatus;
}

struct InternalChatMessage {
    1: required i64 roomId;
    2: required i64 userId;
    3: required string message;
}

struct InternalControlMessage {
    1: required string operation;
}

struct InternalQuizChoice {
    1: required i64 roomId;
    2: required i64 userId;
    3: required i64 questionId;
    4: required string title,
    5: required list<string> options,
    6: required list<i8> answers,
}

struct InternalQuizJudge {
    1: required i64 roomId;
    2: required i64 userId;
    3: required i64 questionId;
    4: required i64 answer;
}

struct InternalQuizStatus {
    1: required i64 roomId;
    2: required i64 requiredNum; // 课堂的人数
    3: required i64 currentNum; // 当前参与答题人数
    4: required i64 AcceptRate; // 正确率（AC率）
}