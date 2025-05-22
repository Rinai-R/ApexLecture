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

struct RecordResponse {
    1: required base.BaseResponse response,
}

struct GetHistoryLectureRequest {
    1: required string offer,
}

struct GetHistoryLectureResponse {
    1: required base.BaseResponse response,
    2: required string answer,
}

struct RandomSelectRequest {}

struct RandomSelectResponse {
    1: required base.BaseResponse response,
    2: required i64 selectedId,
}

service LectureService {
    StartResponse startLecture(1: StartRequest request) (api.post = "lecture/"),
    AttendResponse attendLecture(1: AttendRequest request) (api.post = "lecture/:roomid/attend"),
    RecordResponse recordLecture(1: base.NilRequest request) (api.post = "lecture/:roomid/record"),
    GetHistoryLectureResponse getHistoryLecture(1: GetHistoryLectureRequest request) (api.get = "lecture/:roomid/history"),
    RandomSelectResponse randomSelect(1: RandomSelectRequest request) (api.get = "lecture/:roomid/randomselect"),
}