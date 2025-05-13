IDL_PATH = ./server/idl
MODULE_NAME = github.com/Rinai-R/ApexLecture
CMD_PATH = ./server/cmd
KITEX_GEN = ./server/shared/kitex_gen

# 运行程序
hz-run:
	go run $(CMD_PATH)/api/

user-run:
	go run $(CMD_PATH)/user/


# 更新脚手架文件
hz-new:
	cd 	./server/cmd/api && \
	hz new -idl ../../idl/api/$(service).thrift \

user-rpc:
	cd 	./server/cmd/user && \
	kitex -module $(MODULE_NAME) -service user \
	-use github.com/Rinai-R/ApexLecture/server/shared/kitex_gen \
	../../idl/rpc/user.thrift

user-gen:
	cd 	./server/shared && \
	kitex -module github.com/Rinai-R/ApexLecture ../idl/rpc/user.thrift


# 更新配置文件
hz-conf:
	go run $(CMD_PATH)/api/script/preprocess.go

user-conf:
	go run $(CMD_PATH)/user/script/preprocess.go

conf:
	make hz-conf
	make user-conf


# 一站式服务
hz-all:
	make hz-user

rpc-all:
	make user-gen
	make user-rpc





# 杂项，给其他 make 指令用的
hz-update:
	cd 	./server/cmd/api && \
	hz update -idl ../../idl/api/$(service).thrift \

hz-user:
	make hz-update service=user