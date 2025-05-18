namespace go lecture

include "../base/base.thrift"


struct StartRequest {
    1: required string title,
    2: required string description,
    3: required string speaker,
    5: required string offer,
}

struct StartResponse {
    1: required string roomid,
    2: required string answer,
}
    
struct AttendRequest {
    1: required string offer,
}

struct AttendResponse {
    1: required base.BaseResponse response,
    2: required string answer,
}

struct RecordRequest {
    2: required string offer,
}

struct RecordResponse {
    1: required base.BaseResponse response,
    2: required string answer,
}
service LectureService {
    StartResponse startLecture(1: StartRequest request) (api.post = "lecture/"),
    AttendResponse attendLecture(1: AttendRequest request) (api.post = "lecture/:roomid/attend"),
    RecordResponse recordLecture(1: RecordRequest request) (api.post = "lecture/:roomid/record"),
}