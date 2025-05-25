namespace go chat

include "../base/base.thrift"


struct SendMessageRequest {
    1: i8 type
    2: string message  (api.vd="len($) <= 1000")
}

struct SendMessageResponse {
    1: base.BaseResponse response
}

service ChatService {
    SendMessageResponse chat(1: SendMessageRequest msg) (api.post="/chat/:roomid")
}