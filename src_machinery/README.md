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





## 分析worker

运行worker（消费者）：`go run example/machinery.go -c config.yml worker`



启动消费者，其实就是调用了`worker()`函数。

```go
func worker() error {
	server, err := startServer()
	if err != nil {
		return err
	}

	// The second argument is a consumer tag
	// Ideally, each worker should have a unique tag (worker1, worker2 etc)
	worker := server.NewWorker("machinery_worker", 0)

	if err := worker.Launch(); err != nil {
		return err
	}

	return nil
}
```

该函数做了两个操作：

1. 启动一个server
2. 创建并启动一个worker



**一、启动server：startServer()的过程**



1、首先看看Server的定义：保存了所有的配置。所有的task的worker都注册在server上。

```go
// Server is the main Machinery object and stores all configuration
// All the tasks workers process are registered against the server
type Server struct {
	config          *config.Config
	registeredTasks map[string]interface{}
	broker          brokers.Interface
	backend         backends.Interface
}
```

2、startServer()中做了下面操作，

```go
func startServer() (server *machinery.Server, err error) {
	// Create server instance
	server, err = machinery.NewServer(loadConfig())
	// ...

	// Register tasks
	tasks := map[string]interface{}{
		"add":        exampletasks.Add,
		"multiply":   exampletasks.Multiply,
		"panic_task": exampletasks.PanicTask,
	}

	err = server.RegisterTasks(tasks)
	return
}
```



- 2.1、使用`NewServer()`创建一个Server对象；

  - 创建Broker：使用`BrokerFactory()`进行创建不同的Broker，在当前版本中可以支持amqp、redis、redis+socket、eager协议/规范的broker。我们当前使用的是RabbitMQ，也就是AMQP协议。
  - 创建Backend：Backend用于保存task结果，目前支持AMQP、memcache、redis、redis+socket、mongodb、eager协议/规范。我们当前配置使用的是redis。
  - 创建好broker和backend后，也即完成了Server对象的创建。

- 2.2、注册任务给Server；通过`server.RegisterTasks(tasks)`进行task的注册。当前默认注册的task有3个。

  ```go
  	// Register tasks
  	tasks := map[string]interface{}{
  		"add":        exampletasks.Add,
  		"multiply":   exampletasks.Multiply,
  		"panic_task": exampletasks.PanicTask,
  	}
  ```

  RegisterTasks函数：

  ```go
  // RegisterTasks registers all tasks at once
  func (server *Server) RegisterTasks(namedTaskFuncs map[string]interface{}) error {
  	for _, task := range namedTaskFuncs {
  		if err := tasks.ValidateTask(task); err != nil {
  			return err
  		}
  	}
  	server.registeredTasks = namedTaskFuncs
  	server.broker.SetRegisteredTaskNames(server.GetRegisteredTaskNames())
  	return nil
  }
  ```

  把相应的task的处理函数注册到server上面；把任务的名字注册到broker里面。

**二、启动worker的过程**

启动worker有两步操作，

1. 创建一个worker：在上一步创建的server上面创建一个worker；
2. 启动一个worker：worker.Launch()函数。





























































