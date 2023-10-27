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

// http://127.0.0.1:8080/lona?principal=920000&loanTerm=360&startDate=2022-05-25&plusSpread=0.6&paymentDueDay=18&earlyRepayment1Amount=200000&earlyRepayment1Date=2023-08-19&earlyRepayment2Amount=0&earlyRepayment2Date=&earlyRepayment3Amount=0&earlyRepayment3Date=
