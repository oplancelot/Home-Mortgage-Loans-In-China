package route

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oplancelot/Home-Mortgage-Loans-In-China/bootstrap"
)

func Setup(env *bootstrap.Env, timeout time.Duration, gin *gin.Engine) {
	// 载入assets
	gin.Static("/static", "assets/static")
	gin.LoadHTMLGlob("assets/templates/*")
	publicRouter := gin.Group("")
	// All Public APIs
	PingRoute(env, timeout, publicRouter)
	LoanRoute(env, timeout, publicRouter)
}
