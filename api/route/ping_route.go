package route

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oplancelot/Home-Mortgage-Loans-In-China/bootstrap"
)

func PingRoute(env *bootstrap.Env, timeout time.Duration, group *gin.RouterGroup) {

	group.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
}
