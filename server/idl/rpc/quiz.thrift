namespace go quiz

include "../base/base.thrift"

// 把题目的服务独立出来了，主要是怕 lecture 压力过大。


// 上传题目部分。
struct SubmitQuestionRequest {
    1: required i64 roomId,
    2: required i64 userId,
    3: required i8 type,
    4: required Payload Payload,
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
    1: required i64 roomId,
    2: required i64 userId,
    3: required i64 questionId,
    4: required i8 type,
    5: required AnswerPayload payload,
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
    SubmitQuestionResponse submitQuestion(1: SubmitQuestionRequest request)
    SubmitAnswerResponse submitAnswer(1: SubmitAnswerRequest request)
}