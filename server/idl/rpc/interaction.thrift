namespace go interaction

include "../base/base.thrift"


// ======================== 创建交互房间的请求 ====================

struct CreateRoomRequest {
    1: required i64 teacher_id
    2: required i64 room_id
}

struct CreateRoomResponse {
    1: base.BaseResponse response
}


// ======================== 发送消息的请求 ========================

struct SendMessageRequest {
    1: string message
    2: i64 userId
    3: i64 roomId
}

struct SendMessageResponse {
    1: base.BaseResponse response
}

// ======================== 创建问题的请求 ========================

struct CreateQuestionRequest {
    1: required i64 room_id
    2: required i64 teacher_id
    3: required i8 type
    4: required string content
    5: optional ExtraContent extra
}

struct ExtraContent {
    1: optional list<string> options
    2: optional double score
    3: optional string answer_text
    4: optional bool answer_true_false
    5: optional list<i8> answer_choice
}

struct CreateQuestionResponse {
    1: base.BaseResponse response
}

// ======================== 提交答案的请求 ========================

struct SubmitAnswerRequest {
    1: required i64 question_id
    2: required i64 student_id
    3: required string type
    4: required AnswerContent content
}

// 使用union区分不同类型的答案内容
struct AnswerContent {
    1: optional list<i8> choice_answer
    2: optional bool true_false_answer
    3: optional string text_answer
}

struct SubmitAnswerResponse {
    1: base.BaseResponse response
    // 这里仅仅对于选择题和判断题而言才有分数。
    2: optional double score
}

struct ReceiveRequest {
    1: required i64 room_id
    2: required i64 user_id
}

struct Msg {
    1: required string msg_id
    2: i64 sender_id
    3: string content
}

struct ReceiveResponse {

}

struct MessageInfo {
    
}

service InteractionService {
    CreateRoomResponse createRoom(1: CreateRoomRequest request)
    SendMessageResponse sendMessage(1: SendMessageRequest request)
    CreateQuestionResponse createQuestion(1: CreateQuestionRequest request)
    SubmitAnswerResponse submitAnswer(1: SubmitAnswerRequest request) 
    ReceiveResponse receive(1: ReceiveRequest request) (streaming.mode = "server")
}
