IDL_PATH = ./server/idl
MODULE_NAME = github.com/Rinai-R/ApexLecture
CMD_PATH = ./server/cmd
KITEX_GEN = ./server/shared/kitex_gen
# api
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

# user-rpc
user-rpc:
	cd 	./server/cmd/user && \
	kitex -module $(MODULE_NAME) -service user \
	-use github.com/Rinai-R/ApexLecture/server/shared/kitex_gen \
	../../idl/rpc/user.thrift
user-gen:
	cd 	./server/shared && \
	kitex -module github.com/Rinai-R/ApexLecture ../idl/rpc/user.thrift

rpc-all:
	make user-gen
	make user-rpc