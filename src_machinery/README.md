# 源码分析



[RichardKnop/machinery](https://github.com/RichardKnop/machinery)

Machinery is an asynchronous task queue/job queue based on distributed message passing.



## 环境准备

- 版本：v1.0.0

  ```bash
  git clone https://github.com/RichardKnop/machinery/tree/master
  git checkout v1.0.0
  ```


* 依赖服务1：RabbitMQ。使用docker-compose搭建rabbitmq做为broker backend。新建`docker-compose.yml`后，运行`docker-compose up -d`启动。

  > ```yaml
  > version: '2'
  > 
  > services:
  >   rabbitmq:
  >     image: 'tutum/rabbitmq'
  >     ports:
  >       - '4369:4369'
  >       - '5672:5672'
  >       - '25672:25672'
  >       - '15672:15672'
  > ```

  启动后查看日志，获取rabbitmq的账号密码：`docker-compose logs`。

  ```bash
  ➜  rabbitmq_bitnami docker-compose logs
  ...
  rabbitmq_1  |     curl --user admin:SB0gzhcUjJVD http://<host>:<port>/api/vhosts
  ...
  ```

  使用`curl --user admin:SB0gzhcUjJVD http://localhost:15672/api/vhosts`检查服务的状态。

- 依赖服务2：redis。使用redis做为保存结果的服务。使用docker启动redis服务。

  ```bash
  docker run --name test-redis -p 6379:6379 -d redis
  ```

- 为machinery测试创建`config.yml`作为测试的配置文件。注意修改rabbitmq对应的账号密码。使用rabbitmq作为任务分发的服务，redis作为获取结果的服务。

  ```yaml
  ---
  broker: 'amqp://admin:SB0gzhcUjJVD@127.0.0.1:5672'
  default_queue: machinery_tasks
  result_backend: 'redis://127.0.0.1:6379'
  results_expire_in: 3600000
  amqp:
    binding_key: machinery_task
    exchange: machinery_exchange
    exchange_type: direct
    prefetch_count: 3
  ```

- 运行测试：

  - 运行worker（消费者）：`go run example/machinery.go -c config.yml worker`
  - 运行send（生产者）：`go run example/machinery.go -c config.yml send`

