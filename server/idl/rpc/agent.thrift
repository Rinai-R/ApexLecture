namespace go agent


include "../base/base.thrift"

// 问问题请求
struct AskRequest {
    1: required i64 userId
    2: required i64 roomId
    3: required string content
}


struct AskResponse {
    1: required base.BaseResponse response
    2: required string content
}

// 请求课程智能纪要
struct StartSummaryRequest {
    2: required i64 roomId
}

struct StartSummaryResponse {
    1: required base.BaseResponse response
}

// 获取课程智能纪要
struct GetSummaryRequest {
    2: required i64 roomId
}

struct GetSummaryResponse {
    1: required base.BaseResponse response
    2: required string summary
}


service AgentService {
    AskResponse ask(1: AskRequest askRequest)
    StartSummaryResponse startSummary(1: StartSummaryRequest summaryRequest)
    GetSummaryResponse getSummary(1: GetSummaryRequest summaryRequest)
}