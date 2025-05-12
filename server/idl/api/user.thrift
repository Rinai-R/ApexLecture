namespace go user

include "../base/base.thrift"

struct RegisterRequest {
    1: required string name (api.query="name", api.vd="len($) >= 5 && len($) <= 20");
    2: required string password (api.query="password", api.vd="len($) >= 8 && len($) <= 20");
}

struct RegisterResponse {
    1: base.BaseResponse base;
}

struct LoginRequest {
    1: required string name (api.query="name", api.vd="len($) >= 5 && len($) <= 20");
    2: required string password (api.query="password", api.vd="len($) >= 8 && len($) <= 20");
}

struct LoginResponse {
    1: base.BaseResponse base;
    2: string token;
}

service UserService {
    RegisterResponse register(1: RegisterRequest request) (api.post="/user/register");
    LoginResponse    login   (1: LoginRequest   request) (api.post="/user/login");
}
