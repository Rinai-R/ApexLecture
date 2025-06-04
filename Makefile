IDL_PATH = ./server/idl
MODULE_NAME = github.com/Rinai-R/ApexLecture
CMD_PATH = ./server/cmd
KITEX_GEN = ./server/shared/kitex_gen

# ============================== docker-compose =========================
re:
	make down
	make up

up:
	docker-compose up -d
	sleep 5
	docker-compose down
	sudo chmod -R 0777 ./data
	sudo chmod 400 ./data/rabbitmq/.erlang.cookie
	docker-compose up -d
	make conf

down:
	docker-compose down

clear:
	sudo aa-remove-unknown

# ================================== 运行程序 ============================
hz-run:
	go run $(CMD_PATH)/api/

user-run:
	go run $(CMD_PATH)/user/

lecture-run:
	go run $(CMD_PATH)/lecture/

chat-run:
	go run $(CMD_PATH)/chat/

push-run:
	go run $(CMD_PATH)/push/

quiz-run:
	go run $(CMD_PATH)/quiz/

agent-run:
	go run $(CMD_PATH)/agent/

# ============================= 更新 API 脚手架文件 ========================
hz-update:
	make api-update service=user
	make api-update service=lecture
	make api-update service=chat
	make api-update service=push
	make api-update service=quiz
	make api-update service=agent

hz-new:
	make api-new service=user
	make api-new service=lecture
	make api-new service=push
	make api-new service=chat
	make api-new service=quiz
	make api-new service=agent

# ============================= 更新 rpc 脚手架文件 ========================
rpc-all:
	make user-gen
	make user-rpc
	make lecture-gen
	make lecture-rpc
	make chat-gen
	make chat-rpc
	make push-gen
	make push-rpc
	make quiz-gen
	make quiz-rpc
	make agent-gen
	make agent-rpc

user-rpc:
	cd 	./server/cmd/user && \
	kitex -module $(MODULE_NAME) -service user \
	-use github.com/Rinai-R/ApexLecture/server/shared/kitex_gen \
	../../idl/rpc/user.thrift

user-gen:
	cd 	./server/shared && \
	kitex -module github.com/Rinai-R/ApexLecture ../idl/rpc/user.thrift

lecture-rpc:
	cd 	./server/cmd/lecture && \
	kitex -module $(MODULE_NAME) -service lecture \
	-use github.com/Rinai-R/ApexLecture/server/shared/kitex_gen \
	../../idl/rpc/lecture.thrift

lecture-gen:
	cd 	./server/shared && \
	kitex -module github.com/Rinai-R/ApexLecture ../idl/rpc/lecture.thrift

push-rpc:
	cd 	./server/cmd/push && \
	kitex -streamx -module $(MODULE_NAME) -service push \
	-use github.com/Rinai-R/ApexLecture/server/shared/kitex_gen \
	../../idl/rpc/push.thrift

push-gen:
	cd 	./server/shared && \
	kitex -streamx  -module github.com/Rinai-R/ApexLecture ../idl/rpc/push.thrift

chat-rpc:
	cd 	./server/cmd/chat && \
	kitex -module $(MODULE_NAME) -service chat \
	-use github.com/Rinai-R/ApexLecture/server/shared/kitex_gen \
	../../idl/rpc/chat.thrift

chat-gen:
	cd 	./server/shared && \
	kitex -module github.com/Rinai-R/ApexLecture ../idl/rpc/chat.thrift

quiz-rpc:
	cd 	./server/cmd/quiz && \
	kitex -module $(MODULE_NAME) -service quiz \
	-use github.com/Rinai-R/ApexLecture/server/shared/kitex_gen \
	../../idl/rpc/quiz.thrift

quiz-gen:
	cd 	./server/shared && \
	kitex -module github.com/Rinai-R/ApexLecture ../idl/rpc/quiz.thrift

agent-rpc:
	cd 	./server/cmd/agent && \
	kitex -module $(MODULE_NAME) -service agent \
	-use github.com/Rinai-R/ApexLecture/server/shared/kitex_gen \
	../../idl/rpc/agent.thrift

agent-gen:
	cd 	./server/shared && \
	kitex -module github.com/Rinai-R/ApexLecture ../idl/rpc/agent.thrift

# =================================== 更新配置文件 ==================================
hz-conf:
	go run $(CMD_PATH)/api/script/preprocess.go

user-conf:
	go run $(CMD_PATH)/user/script/preprocess.go

lecture-conf:
	go run $(CMD_PATH)/lecture/script/preprocess.go

push-conf:
	go run $(CMD_PATH)/push/script/preprocess.go

chat-conf:
	go run $(CMD_PATH)/chat/script/preprocess.go

quiz-conf:
	go run $(CMD_PATH)/quiz/script/preprocess.go

agent-conf:
	go run $(CMD_PATH)/agent/script/preprocess.go

conf:
	make hz-conf
	make user-conf
	make lecture-conf
	make push-conf
	make chat-conf
	make quiz-conf
	make agent-conf
	
# ============================ 杂项，给其他 make 指令用的 ==========================
api-update:
	cd 	./server/cmd/api && \
	hz update -idl ../../idl/api/$(service).thrift \

api-new:
	cd 	./server/cmd/api && \
	hz new -idl ../../idl/api/$(service).thrift \