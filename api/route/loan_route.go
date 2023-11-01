package route

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oplancelot/Home-Mortgage-Loans-In-China/bootstrap"
	"github.com/oplancelot/Home-Mortgage-Loans-In-China/internal/loan"
	"github.com/shopspring/decimal"
)

func LoanRoute(env *bootstrap.Env, timeout time.Duration, group *gin.RouterGroup) {
	// println("api/route/loan_route.go")
	// group.GET("/loan", func(c *gin.Context) {
	// 	// c.String(http.StatusOK, "loan")
	// 	c.HTML(http.StatusOK, "loan.tmpl", gin.H{
	// 		"title": "Loan",
	// 	})
	// })

	group.GET("/loan", func(c *gin.Context) {

		// 解析用户输入的数据
		principal, _ := strconv.ParseFloat(c.DefaultQuery("principal", "0"), 64)
		loanTerm, _ := strconv.Atoi(c.DefaultQuery("loanTerm", "12"))
		startDate := c.DefaultQuery("startDate", "2022-05-25")
		plusSpread, _ := strconv.ParseFloat(c.DefaultQuery("plusSpread", "0"), 64)
		paymentDueDay, _ := strconv.Atoi(c.DefaultQuery("paymentDueDay", "1"))

		// 获取提前还款信息的值
		earlyRepayment1Amount, _ := strconv.ParseFloat(c.DefaultQuery("earlyRepayment1Amount", "0"), 64)
		earlyRepayment1Date := c.DefaultQuery("earlyRepayment1Date", "2023-08-19")

		earlyRepayment2Amount, _ := strconv.ParseFloat(c.DefaultQuery("earlyRepayment2Amount", "0"), 64)
		earlyRepayment2Date := c.DefaultQuery("earlyRepayment2Date", "2099-05-25")

		earlyRepayment3Amount, _ := strconv.ParseFloat(c.DefaultQuery("earlyRepayment3Amount", "0"), 64)
		earlyRepayment3Date := c.DefaultQuery("earlyRepayment3Date", "2099-05-25")

		// 创建 Loan 和 EarlyRepayment 的实例
		originialloan := loan.Loan{
			InitialPrincipal: decimal.NewFromFloat(principal),
			InitialLPR:       decimal.NewFromFloat(4.45),
			InitialTerm:      loanTerm,
			InitialDate:      loan.ParseDate(startDate),
			LPR:              loan.Lprs, // 常量
			PlusSpread:       decimal.NewFromFloat(plusSpread),
			PaymentDueDay:    paymentDueDay,
		}

		// 输入提前还款信息
		earlyRepayments := []loan.EarlyRepayment{
			{Amount: decimal.NewFromFloat(earlyRepayment1Amount), Date: loan.ParseDate(earlyRepayment1Date)},
			{Amount: decimal.NewFromFloat(earlyRepayment2Amount), Date: loan.ParseDate(earlyRepayment2Date)},
			{Amount: decimal.NewFromFloat(earlyRepayment3Amount), Date: loan.ParseDate(earlyRepayment3Date)},
		}

		// 创建 Input 结构体并赋值
		inputData := loan.Input{
			Loan:           originialloan,
			EarlyRepayment: earlyRepayments,
		}
		// fmt.Println(inputData)
		action := c.Query("action")
		report := loan.LoanPrintReport(inputData, action)

		// 将结果传递给模板进行渲染

		c.HTML(http.StatusOK, "loan.tmpl", gin.H{
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
}
