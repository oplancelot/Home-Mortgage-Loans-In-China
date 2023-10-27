package route

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oplancelot/Home-Mortgage-Loans-In-China/bootstrap"
)

func Setup(env *bootstrap.Env, timeout time.Duration, gin *gin.Engine) {

	gin.Static("/static", ".assets/static")
	gin.LoadHTMLGlob("assets/templates/*")
	publicRouter := gin.Group("")
	// All Public APIs
	PingRoute(env, timeout, publicRouter)
	LonaRoute(env, timeout, publicRouter)
}
