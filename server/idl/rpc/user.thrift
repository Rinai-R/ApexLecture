namespace go user

include "../base/base.thrift"

struct RegisterRequest {
  1: required string username,
  2: required string password,
}

struct RegisterResponse {
  1: required base.BaseResponse base,
  2: required i64 id,
}

struct LoginRequest {
  1: required string username,
  2: required string password,
}

struct LoginResponse {
  1: required base.BaseResponse base,
  2: required string token,
}

struct GetPublicKeyRequest {}

struct GetPublicKeyResponse {
  1: required string publicKey,
}

service UserService {
    RegisterResponse register(1: RegisterRequest request)
    LoginResponse login(1: LoginRequest request)
    GetPublicKeyResponse getPublicKey(1: GetPublicKeyRequest request)
}