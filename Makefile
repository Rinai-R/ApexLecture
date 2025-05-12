IDL_PATH = ./server/idl
MODULE_NAME = github.com/Rinai-R/ApexLecture
CMD_PATH = ./server/cmd
KITEX_GEN = ./server/shared/kitex_gen

# 运行 user 程序
user-run:
	go run $(CMD_PATH)/user/


# 关于 api 相关的脚本
hz-new:
	cd 	./server/cmd/api && \
	hz new -idl ../../idl/api/$(service).thrift \

hz-update:
	cd 	./server/cmd/api && \
	hz update -idl ../../idl/api/$(service).thrift \

hz-user:
	make hz-update service=user

hz-all:
	make hz-user

# user rpc 相关脚本
user-rpc:
	cd 	./server/cmd/user && \
	kitex -module $(MODULE_NAME) -service user \
	-use github.com/Rinai-R/ApexLecture/server/shared/kitex_gen \
	../../idl/rpc/user.thrift
user-gen:
	cd 	./server/shared && \
	kitex -module github.com/Rinai-R/ApexLecture ../idl/rpc/user.thrift

user-conf:
	go run $(CMD_PATH)/user/script/preprocess.go


# 一站式服务
rpc-all:
	make user-gen
	make user-rpc
