namespace go quiz

include "../base/base.thrift"


struct SubmitQuestionRequest {
    1: required i8 type,
    2: required Payload Payload,
}

union Payload {
    1: optional Choice choice,
    2: optional Judge judge,
}

struct Choice {
    1: required string title,
    2: required list<string> options,
    3: required list<i8> answers,
}

struct Judge {
    1: required string title,
    2: required bool answer,
}


struct SubmitQuestionResponse {
  1: required base.BaseResponse response,
}


// 提交答案部分
struct SubmitAnswerRequest {
    1: required i64 questionId,
    2: required i8 type,
    3: required AnswerPayload payload,
}

union AnswerPayload {
    1: optional ChoiceAnswer choice,
    2: optional JudgeAnswer judge,
}

struct ChoiceAnswer {
    1: required list<i8> answer,
}

struct JudgeAnswer {
    1: required bool answer,
}

struct SubmitAnswerResponse {
  1: required base.BaseResponse response,
  2: required bool isCorrect,
  3: optional AnswerPayload payload,
}

service QuizService {
    SubmitQuestionResponse submitQuestion(1: SubmitQuestionRequest request) (api.post="/quiz/:roomid/question")
    SubmitAnswerResponse submitAnswer(1: SubmitAnswerRequest request) (api.post="/quiz/:roomid/answer")
}