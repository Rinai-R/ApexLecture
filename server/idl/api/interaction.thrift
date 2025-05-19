namespace go interaction

include "../base/base.thrift"

// ======================== 发送消息的请求 ========================

struct SendMessageRequest {
    1: string message
}

struct SendMessageResponse {
    1: base.BaseResponse response
}

// ======================== 创建问题的请求 ========================

struct QuestionBase {
    1: required i64 id
    2: required i64 room_id
    3: required i64 teacher_id
    4: required string type
    5: required string title
    6: required i32 score
    7: required i64 create_time
    8: required QuestionContent content
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

struct TrueFalseQuestion {
    1: required bool answer
}

struct TextQuestion {
    1: required string reference_answer
    2: optional list<string> keywords
}
// 选择题的选项
struct Option {
    1: required string id
    2: required string content
}

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


// ======================== 提交答案的请求 ========================

struct SubmitAnswerRequest {
    1: required i64 question_id
    2: required i64 student_id
    3: required i64 room_id
    4: required string type
    5: required AnswerContent content
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

struct ReceiveRequest {}

struct ReceiveResponse {}

service InteractionService {
    SendMessageResponse sendMessage(1: SendMessageRequest request) (api.post = "/interaction/:roomid/send"),
    CreateQuestionResponse createQuestion(1: CreateQuestionRequest request) (api.post = "/interaction/:roomid/create"),
    SubmitAnswerResponse submitAnswer(1: SubmitAnswerRequest request) (api.post = "/interaction/:roomid/submit"),
    ReceiveResponse receive(1: ReceiveRequest request) (api.get = "/interaction/:roomid/receive"),
}



