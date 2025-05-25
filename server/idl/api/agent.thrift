namespace go agent


include "../base/base.thrift"

// 问问题请求
struct AskRequest {
    3: required string content (api.vd="len($) <= 1000")
}


struct AskResponse {
    1: required base.BaseResponse response
    2: required string content
}

// 请求课程智能纪要
struct StartSummaryRequest {
}

struct StartSummaryResponse {
    1: required base.BaseResponse response
}

// 获取课程智能纪要
struct GetSummaryRequest {
}

struct GetSummaryResponse {
    1: required base.BaseResponse response
    2: required string summary
}


service AgentService {
    AskResponse ask(1: AskRequest askRequest) (api.post="/agent/:roomid/ask")
    StartSummaryResponse startSummary(1: StartSummaryRequest summaryRequest) (api.post="/agent/:roomid/summary")
    GetSummaryResponse getSummary(1: GetSummaryRequest summaryRequest) (api.get="/agent/:roomid/summary")
}