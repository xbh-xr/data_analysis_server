package api

import "go-admin/app/report/router"

func init() {
	AppRouters = append(AppRouters, router.InitRouter)
}
