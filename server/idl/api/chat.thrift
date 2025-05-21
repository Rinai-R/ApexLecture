namespace go chat

include "../base/base.thrift"


struct SendMessageRequest {
    1: string message
}

struct SendMessageResponse {
    1: base.BaseResponse response
}

service ChatService {
    SendMessageResponse chat(1: SendMessageRequest msg) (api.post="/chat/:roomid")
}