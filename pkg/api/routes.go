package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bitmex-api/pkg/model"
)

func configureRouter(api *api) *gin.Engine {
	router := gin.Default()

	router.Use(CORSMiddleware())

	router.GET("/connect", api.UserWebSocket().Connect)

	public := router.Group("api/v1")

	public.POST("/login", api.Auth().Login)
	public.POST("/refresh", api.Auth().Refresh)
	public.PATCH("/change-password", api.Auth().ChangePassword)

	private := router.Group("api/v1")

	private.Use(api.auth.Authorize)

	private.POST("/registration", api.Auth().Register)

	privateUser := private.Group("/user")

	privateUser.PATCH("/update-info", api.User().UpdateInfo)
	privateUser.GET("/", api.User().Get)

	privateBitMex := private.Group("/bit-mex")

	privateBitMex.PATCH("/subscription", api.UserWebSocket().SubscribeAction)

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, model.ErrRecordNotFound)
	})

	return router
}
