package main

import (
	"os"

	"go-admin/cmd"
	"go-admin/cmd/api"
)

//go:generate swag init --parseDependency --parseDepth=6 --instanceName admin -o ./docs/admin

// @title go-admin API
// @version 2.0.0
// @description 基于Gin + Vue + Element UI的前后端分离权限管理系统的接口文档
// @description 添加qq群: 521386980 进入技术交流群 请先star，谢谢！
// @license.name MIT
// @license.url https://github.com/go-admin-team/go-admin/blob/master/LICENSE.md

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func main() {
	// 无参数时直接启动 API 主进程，便于 IDE 一键运行 / 调试
	if len(os.Args) <= 1 {
		if err := api.Start(); err != nil {
			os.Exit(-1)
		}
		return
	}
	cmd.Execute()
}
