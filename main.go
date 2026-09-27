package main

import (
	"flag"
	"fmt"

	"shortlink/internal/config"
	"shortlink/internal/handler"
	"shortlink/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)
//声明flag参数
var configFile = flag.String("f", "etc/shorturl-api.yaml", "the config file")

func main() {
	//解析flag参数
	flag.Parse()
	//加载配置文件
	var c config.Config
	conf.MustLoad(*configFile, &c)
	//创建rest server
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()
	//创建service context
	ctx := svc.NewServiceContext(c)
	//注册handler
	handler.RegisterHandlers(server, ctx)
	//启动server
	//打印启动信息
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
