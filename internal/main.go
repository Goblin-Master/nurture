package main

import (
	"nurture/internal/config"
	"nurture/internal/global"
	"nurture/internal/router"
)

func main() {
	config.LoadConfig() //加载配置
	global.Init()       //初始化全局中间件
	router.RunServer()  //启动服务端
}
