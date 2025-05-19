namespace go interaction

include "../base/base.thrift"


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
    3: required string type
    4: required string title
    5: required i32 score
    6: required QuestionContent content
}

struct CreateQuestionResponse {
    1: base.BaseResponse response
}


struct QuestionContent {
    1: optional ChoiceQuestion choice_question
    2: optional TrueFalseQuestion true_false_question
    3: optional TextQuestion text_question
}

// 支持三种题型：选择题、判断题、问答题
struct ChoiceQuestion {
    1: required list<Option> options
    2: required string correct_id
}

// 选择题的选项
struct Option {
    1: required string id
    2: required string content
}

struct TrueFalseQuestion {
    1: required bool answer
}

struct TextQuestion {
    1: required string reference_answer
    2: optional list<string> keywords
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
    1: optional ChoiceAnswer choice_answer
    2: optional TrueFalseAnswer true_false_answer
    3: optional TextAnswer text_answer
}

// 对于三种题型的答案
struct ChoiceAnswer {
    1: required string selected_id
}

struct TrueFalseAnswer {
    1: required bool answer
}


struct TextAnswer {
    1: required string content
}

struct SubmitAnswerResponse {
    1: base.BaseResponse response
    // 这里仅仅对于选择题和判断题而言才有分数。
    2: optional i32 score
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

struct ChoiceMsg {
    1: required string question_id
    2: required string title
    3: required string type
    4: required list<Option> options
}

struct TrueFalseMsg {
    1: required string question_id
    2: required string title
    3: required string type
}

struct TextMsg {
    1: required string question_id
    2: required string title
    3: required string type
}


struct ReceiveResponse {
    1: optional Msg msg
    2: optional ChoiceMsg choice_msg
    3: optional TrueFalseMsg true_false_msg
    4: optional TextMsg text_msg
}

service InteractionService {
    SendMessageResponse sendMessage(1: SendMessageRequest request)
    CreateQuestionResponse createQuestion(1: CreateQuestionRequest request)
    SubmitAnswerResponse submitAnswer(1: SubmitAnswerRequest request) 
    ReceiveResponse receive(1: ReceiveRequest request) (streaming.mode = "server")
}
