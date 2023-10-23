package main

import (
	"lona"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var db = make(map[string]string)

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.GET("/lona", func(c *gin.Context) {

		// 解析用户输入的数据
		principal, _ := strconv.ParseFloat(c.DefaultQuery("principal", "0"), 64)
		loanTerm, _ := strconv.Atoi(c.DefaultQuery("loanTerm", "12"))
		startDate := c.DefaultQuery("startDate", "2022-05-25")
		plusSpread, _ := strconv.ParseFloat(c.DefaultQuery("plusSpread", "0"), 64)
		paymentDueDay, _ := strconv.Atoi(c.DefaultQuery("paymentDueDay", "1"))

		// 获取提前还款信息的值
		earlyRepayment1Amount, _ := strconv.ParseFloat(c.DefaultQuery("earlyRepayment1Amount", "0"), 64)
		earlyRepayment1Date := c.DefaultQuery("earlyRepayment1Date", "")

		earlyRepayment2Amount, _ := strconv.ParseFloat(c.DefaultQuery("earlyRepayment2Amount", "0"), 64)
		earlyRepayment2Date := c.DefaultQuery("earlyRepayment2Date", "")

		earlyRepayment3Amount, _ := strconv.ParseFloat(c.DefaultQuery("earlyRepayment3Amount", "0"), 64)
		earlyRepayment3Date := c.DefaultQuery("earlyRepayment3Date", "")

		// 调用 LonaPrintReport 函数生成报表
		report := lona.LonaPrintReport(principal, loanTerm, startDate, plusSpread, paymentDueDay, []float64{earlyRepayment1Amount, earlyRepayment2Amount, earlyRepayment3Amount}, []string{earlyRepayment1Date, earlyRepayment2Date, earlyRepayment3Date})

		// 将结果传递给模板进行渲染
		c.HTML(http.StatusOK, "lona.tmpl", gin.H{
			"Principal":             principal,             // 从用户输入中获取的初始本金
			"LoanTerm":              loanTerm,              // 从用户输入中获取的贷款期限
			"StartDate":             startDate,             // 从用户输入中获取的放款日期
			"PlusSpread":            plusSpread,            // 从用户输入中获取的上浮点数
			"PaymentDueDay":         paymentDueDay,         // 从用户输入中获取的还款日
			"earlyRepayment1Amount": earlyRepayment1Amount, // 从用户输入中获取的提前还款金额
			"earlyRepayment1Date":   earlyRepayment1Date,   // 从用户输入中获取的提前还款日期
			"earlyRepayment2Amount": earlyRepayment2Amount, //	从用户输入中获取的提前还款金额
			"earlyRepayment2Date":   earlyRepayment2Date,   // 从用户输入中获取的提前还款日期
			"earlyRepayment3Amount": earlyRepayment3Amount, // 从用户输入中获取的提前还款金额
			"earlyRepayment3Date":   earlyRepayment3Date,   // 从用户输入中获取的提前还款日期
			"Report":                report,                // 报表结果
		})

	})
	return r
}

func main() {
	r := setupRouter()

	r.LoadHTMLGlob("templates/*")

	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}

// http://127.0.0.1:8080/lona?principal=920000&loanTerm=360&startDate=2022-05-25&plusSpread=0.6&paymentDueDay=18&earlyRepayment1Amount=200000&earlyRepayment1Date=2023-08-19&earlyRepayment2Amount=0&earlyRepayment2Date=&earlyRepayment3Amount=0&earlyRepayment3Date=
