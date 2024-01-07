# Golang Rest Server

+ go의 feign 패키지를 구현해 보고 RestServer를 만들어 실제 Eureka Application통신에 적용해 보았다.

```go
func Init(port int, eurekaConfig EurekaConfig) error {
	quitServ = make(chan bool)

	logrus.SetLevel(logrus.DebugLevel)
	gin.SetMode(gin.DebugMode)

	listenPort = fmt.Sprint(port)

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan

		quitServ <- true
	}()

	eureka.Init(eurekaConfig.Url)
	err := eureka.NewInstance(eurekaConfig.ServiceName, eurekaConfig.HostName, eurekaConfig.Port)
	if err != nil {
		return err
	}

	auth, err := eureka.GetApplication("AUTH-SERVER")
	if err != nil {
		return err
	}
	feign.Append(*auth)

	return nil
}
```

종료시 eureka server에 Application 제거를 위해 Gracefull shutdown을 구현했다. 

```go
<-quitServ

_ = eureka.Unregister()

ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
if err := server.Shutdown(context.Background()); err != nil {
  logrus.Error("Error during server shutdown:", err)
} else {
  quitServ <- true
  close(quitServ)
}

select {
case <-ctx.Done():
  logrus.Info("timeout of 3 seconds.")
}

logrus.Info("Server shutdown complete.")
```

signal 응답시 Eureka 서버에 DELETE메소드를 호출한다.


```go
header := map[string]string{
  "Content-Type": "application/json",
}

feignResponse, err := feign.Request("AUTH-SERVER", feign.RequeustOption{
  Path:   "/auth",
  Method: "POST",
  Body:   string(accountInfo),
  Header: header,
})

if err != nil {
  errorHandler(ctx, 500, vo.AuhResponse{Code: "99999", Result: "Server Error"})
}

logrus.Info("status:" + feignResponse.Response.Status + "body:" + string(feignResponse.Body))
```

RequestOption 구조체에 Requset 설정 값을 설정하고 feign패키지의 Request 함수를 호출하니

응답 결과를 얻어와 정상 동작한다.

