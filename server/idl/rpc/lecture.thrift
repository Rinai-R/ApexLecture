namespace go lecture

include "../base/base.thrift"

struct CreareLectureRequest {
    1: required i64 hostId,
    2: required string title,
    3: required string description,
    4: required string speaker,
    5: required string date,
    6: required string sdp,
}

struct CreareLectureResponse {
    1: required base.BaseResponse response,
    2: required i64 roomId,
    3: required string answer,
}

struct AttendRequest {
    1: required i64 roomId,
    2: required i64 userId,
    3: required string sdp,
}

struct AttendResponse {
    1: required base.BaseResponse response,
    2: required string answer,
}

service LectureService {
    CreareLectureResponse createLecture(1: CreareLectureRequest request),
    AttendResponse attend(1: AttendRequest request),
}