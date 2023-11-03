package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oplancelot/Home-Mortgage-Loans-In-China/api/route"
	"github.com/oplancelot/Home-Mortgage-Loans-In-China/bootstrap"
)

func main() {
	gin := gin.Default()
	env := bootstrap.NewEnv()
	timeout := time.Duration(env.ContextTimeout) * time.Second
	route.Setup(env, timeout, gin)
	gin.Run(env.ServerAddress)
}
