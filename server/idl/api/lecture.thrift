namespace go lecture

include "../base/base.thrift"

struct Lecture {
    1: required i64 roomid,
    2: required string title,
    3: required string description,
    4: required string speaker,
    5: required string date,
    6: required string status,
}

struct OfferRequest {
    1: required string type,
    2: required string sdp,
}

struct OfferResponse {
    1: required string answer,
}
    
struct AttendRequest {
    1: required string answer,
}

service LectureService {
    base.NilResponse createLecture(1: base.NilResponse response) (api.post = "lecture/"),
    OfferResponse offerLecture(1: OfferRequest request) (api.post = "lecture/:roomid/offer"),
    base.BaseResponse attendLecture(1: AttendRequest request) (api.post = "lecture/:roomid/attend"),
    base.NilResponse inroom (1: base.NilResponse response) (api.get = "lecture/:roomid/ws"),
}