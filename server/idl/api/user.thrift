namespace go user

include "../base/base.thrift"

struct RegisterRequest {
    1: required string username (api.query="username", api.vd="len($) >= 5 && len($) <= 25");
    2: required string password (api.query="password", api.vd="len($) >= 8 && len($) <= 25");
}


struct LoginRequest {
    1: required string username (api.query="username", api.vd="len($) >= 5 && len($) <= 25");
    2: required string password (api.query="password", api.vd="len($) >= 8 && len($) <= 25");
}


service UserService {
    base.NilResponse   register(1: RegisterRequest request) (api.post="/user/register");
    base.NilResponse   login   (1: LoginRequest   request) (api.post="/user/login");
}
