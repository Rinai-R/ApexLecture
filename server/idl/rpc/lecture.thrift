namespace go lecture

include "../base/base.thrift"


// 防止直播间压力比较大，把很多服务都独立出去了。

struct StartRequest {
    1: required i64 hostId,
    2: required string title,
    3: required string description,
    4: required string speaker,
    5: required string offer,
}

struct StartResponse {
    1: required base.BaseResponse response,
    2: required i64 roomId,
    3: required string answer,
}

struct AttendRequest {
    1: required i64 roomId,
    2: required i64 userId,
    3: required string offer,
}

struct AttendResponse {
    1: required base.BaseResponse response,
    2: required string answer,
}

struct RecordRequest {
    1: required i64 roomId,
}

struct RecordResponse {
    1: required base.BaseResponse response,
}

struct GetHistoryLectureRequest {
    1: required i64 roomId,
    2: required string offer,
}

struct GetHistoryLectureResponse {
    1: required base.BaseResponse response,
    2: required string answer,
}

struct RandomSelectRequest {
    1: required i64 roomId,
    2: required i64 userId,
    3: required i64 number,
}

struct RandomSelectResponse {
    1: required base.BaseResponse response,
    2: required list<i64> selectedIds,
}

service LectureService {
    StartResponse start(1: StartRequest request)
    AttendResponse attend(1: AttendRequest request)
    RecordResponse record(1: RecordRequest request)
    GetHistoryLectureResponse getHistoryLecture(1: GetHistoryLectureRequest request)
    RandomSelectResponse randomSelect(1: RandomSelectRequest request)
}