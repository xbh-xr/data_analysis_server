package router

import (
	"os"

	"github.com/gin-gonic/gin"
	log "github.com/go-admin-team/go-admin-core/logger"
	"github.com/go-admin-team/go-admin-core/sdk"
	"github.com/go-admin-team/go-admin-core/sdk/config"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	common "go-admin/common/middleware"
	_ "go-admin/docs/report"
)

var (
	routerNoCheckRole = make([]func(*gin.RouterGroup), 0)
	routerCheckRole   = make([]func(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware), 0)
)

// InitRouter 报表模块路由初始化
func InitRouter() {
	var r *gin.Engine
	h := sdk.Runtime.GetEngine()
	if h == nil {
		h = gin.New()
		sdk.Runtime.SetEngine(h)
	}
	switch h.(type) {
	case *gin.Engine:
		r = h.(*gin.Engine)
	default:
		log.Fatal("not support other engine")
		os.Exit(-1)
	}

	authMiddleware, err := common.AuthInit()
	if err != nil {
		log.Fatalf("JWT Init Error, %s", err.Error())
	}

	if config.ApplicationConfig.Mode != "prod" {
		r.GET("/swagger/report/*any", ginSwagger.WrapHandler(swaggerfiles.NewHandler(), ginSwagger.InstanceName("report")))
	}

	initRouter(r, authMiddleware)
}

func initRouter(r *gin.Engine, authMiddleware *jwt.GinJWTMiddleware) {
	noCheckRoleRouter(r)
	checkRoleRouter(r, authMiddleware)
}

func noCheckRoleRouter(r *gin.Engine) {
	v := r.Group("/api/v1")
	for _, f := range routerNoCheckRole {
		f(v)
	}
}

func checkRoleRouter(r *gin.Engine, authMiddleware *jwt.GinJWTMiddleware) {
	v := r.Group("/api/v1")
	for _, f := range routerCheckRole {
		f(v, authMiddleware)
	}
}
