package service

import (
	"context"
	"example/handler"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/siawase7179/go_eureka_fegin/eureka"
	feign "github.com/siawase7179/go_eureka_fegin/eureka/fegin"
	"github.com/sirupsen/logrus"
)

var quitServ chan bool
var listenPort string

type EurekaConfig struct {
	Url         []string
	ServiceName string
	HostName    string
	Port        int
}

func init() {

}

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

func Start() {
	r := gin.Default()
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"status":  404,
			"code":    "90001",
			"message": "No handler found for " + c.Request.Method + " " + c.Request.RequestURI,
		})
	})

	v1 := r.Group("/v1")
	if v1 != nil {
		v1.Use(func(ctx *gin.Context) {
			if ctx.GetHeader("X-Client-Id") == "" {
				ctx.AbortWithStatusJSON(404, gin.H{
					"status":  404,
					"code":    "90002",
					"message": "X-Client-Id is not set",
				})
			} else if ctx.GetHeader("X-Client-Password") == "" {
				ctx.AbortWithStatusJSON(404, gin.H{
					"status":  404,
					"code":    "90003",
					"message": "X-Client-Password is not set",
				})
			} else {
				ctx.Next()
			}
		})
	}
	v1.POST("/token", handler.TokenHandler)

	server := &http.Server{
		Addr:    ":" + listenPort,
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Error(err)
		}
		for {
			select {
			case <-quitServ:
				return
			default:
				time.Sleep(1 * time.Second)
			}
		}

	}()

	go func() {
		for {
			err := eureka.HeartBeat()
			if err != nil {
				logrus.Error(err)
			}

			select {
			case <-quitServ:
				return
			default:
				time.Sleep(10 * time.Second)
			}
		}
	}()

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

}
