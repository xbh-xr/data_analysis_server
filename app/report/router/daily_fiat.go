package router

import (
	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"

	"go-admin/app/report/apis"
)

func init() {
	routerCheckRole = append(routerCheckRole, registerDailyFiatRouter)
}

func registerDailyFiatRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.DailyFiat{}
	r := v1.Group("/report").Use(authMiddleware.MiddlewareFunc())
	{
		r.GET("/daily-fiat", api.Get)
	}
}
