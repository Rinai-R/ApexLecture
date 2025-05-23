![ApexLecture](./img/ApexLecture.png)

# ApexLecture
An online live classroom

一个在线课堂直播平台。

## 简介

- 使用 hertz 和 kitex 实现整个服务的骨架，网关通过 sse 实现轻量的推送消息服务。
- 通过 pion/webrtc 实现直播推流功能，minio 实现对象存储，支持直播回放，可以通过 ./frontend 目录的前端文件进行基础的直播功能。
- 基于 lecture 服务的推流的连接变化实现学生观看直播的实时状态管理，可以统计学生的上课时长。
- chat 服务，用于学生和老师之间的实时聊天。
- quiz 服务，用于老师进行题目发布，学生进行答题，并且基于 lecture 服务的状态管理，统计答题情况，并提供统计数据。
- chat 和 quiz 通过统一消息结构，基于 redis 的 pub/sub 向 push 服务实现消息的实时推送，学生和老师都可以通过 push 拉取消息。
- 使用 kafka 实现部分数据的异步入库，降低数据库压力，提高服务的响应速度。
- user 实现基本的用户注册登录以及鉴权。
- 基于 ELK 实现分布式日志
- 基于 Prometheus 和 Grafana 实现服务的监控。