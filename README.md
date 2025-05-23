- ![ApexLecture](./img/ApexLecture.png)

  ## 🧾 ApexLecture 

  **ApexLecture** 是一个简单实用的在线课堂直播平台，支持老师发起直播、学生观看课程、参与答题互动，并实现了基础的消息推送和数据记录功能。

  ------

  ## 🎯 项目功能

  - ✅ **直播推流**：使用 WebRTC 实现音视频推送，支持直播和回放。
  - ✅ **学生在线状态管理**：通过连接状态记录学生是否在听课，可统计上课时长。
  - ✅ **答题互动**：支持选择题和判断题，学生作答后系统自动记录答题情况，并且会实时通过 push 将答题状态同步给老师。
  - ✅ **实时聊天**：学生和老师可进行简单文本交流。
  - ✅ **消息推送**：chat 和 quiz 消息通过 Redis 发布，由 push 服务通过 SSE 发送给前端。
  - ✅ **异步入库**：部分数据使用 Kafka 异步写入数据库，提升性能。
  - ✅ **用户管理**：支持用户注册、登录和身份校验。
  - ✅ **可观测性**：基于 ELK 集成日志系统，jaeger/Prometheus/opentelemetry 实现监控。

  ------

  ## 📚 技术选型

  - 后端框架：Hertz + Kitex
  - 推流：pion/webrtc
  - 消息队列：Redis Pub/Sub + Kafka
  - 数据库：MySQL + Redis + MinIO（对象存储）
  - 监控工具：Prometheus + Grafana + Jaeger
  - 协程控制：使用 ants 协程池限制协程数量

  ------

  ## 📎 接口与资料

  - 📄 [接口文档（Apifox）](https://apifox.com/apidoc/shared/ec05339a-ba50-46d9-9971-1d9ef2347f2c/297132962e0)
  - ✅ [功能完成情况](./completion.md)

  ## 快速开始

  1. 克隆项目到本地：

  ```bash
  git clone https://github.com/rinai/ApexLecture.git
  ```

  2. 安装依赖：

  ```bash
  go mod tidy
  ```

  3. 拉取依赖： 

  ```bash
  make up
  ```

  1. 启动服务：

  ```bash
  make hz-run
  
  make user-run
  
  make lecture-run
  
  make chat-run
  
  make push-run
  
  make quiz-run
  ```

