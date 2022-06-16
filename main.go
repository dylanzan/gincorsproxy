package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"net/http/httputil"
)

type (
	Config struct {
		Proxy struct{
			Addr string
			Port string
		}
		TargetHost struct{
			Scheme string
			Addr string
			Port string
		}
	}
)

var config =&Config{}

func InitConfig(){
	v:=viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("config/")
	err:=v.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("Fatal err config file: %s \n",err))
	}

	err=v.Unmarshal(config)
	if err!=nil{
		panic(err)
	}
	log.Printf(fmt.Sprintf("config: %+v \n",config))
}

var HostProxy=httputil.ReverseProxy{
	Director: func(request *http.Request) {
		host:=fmt.Sprintf("%s:%s",config.TargetHost.Addr,config.TargetHost.Port)
		request.URL.Scheme=config.TargetHost.Scheme
		request.URL.Host=host
		request.Host=host
	},
}



func main()  {

	InitConfig() // 初始化config
	proxyHost:=fmt.Sprintf("%s:%s",config.Proxy.Addr,config.Proxy.Port)
	engine:=gin.New()
	engine.Use(Cors())
	apiGroup:=engine.Group("/api")
	apiGroup.Use(gin.Logger())
	apiGroup.Any("/*action",DirectorHandler)
	err:=engine.Run(proxyHost)
	if err!=nil {
		log.Fatal(err)
	}
}

func DirectorHandler(ctx *gin.Context){
	HostProxy.ServeHTTP(ctx.Writer,ctx.Request)
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin") //请求头部
		if origin != "" {
			//接收客户端发送的origin （重要！）
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			//服务器支持的所有跨域请求的方法
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
			//允许跨域设置可以返回其他子段，可以自定义字段
			//c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept,Authentication") // 自定义跨域请求头
			// 允许浏览器（客户端）可以解析的头部 （重要）
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			//设置缓存时间
			c.Header("Access-Control-Max-Age", "172800")
			//允许客户端传递校验信息比如 cookie (重要)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		//允许类型校验
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "ok!")
		}

		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic info is: %v", err)
			}
		}()

		c.Next()
	}
}