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

func parseFormData(c *gin.Context) (inputData loan.Input, err error) {
	principal, _ := strconv.ParseFloat(c.DefaultPostForm("principal", "0"), 64)
	loanTerm, _ := strconv.Atoi(c.DefaultPostForm("loanTerm", "12"))
	startDate := c.DefaultPostForm("startDate", "2022-05-25")
	plusSpread, _ := strconv.ParseFloat(c.DefaultPostForm("plusSpread", "0"), 64)
	paymentDueDay, _ := strconv.Atoi(c.DefaultPostForm("paymentDueDay", "1"))

	// 获取提前还款信息的值
	earlyRepayment1Amount, _ := strconv.ParseFloat(c.DefaultPostForm("earlyRepayment1Amount", "0"), 64)
	earlyRepayment1Date := c.DefaultPostForm("earlyRepayment1Date", "2023-08-19")

	earlyRepayment2Amount, _ := strconv.ParseFloat(c.DefaultPostForm("earlyRepayment2Amount", "0"), 64)
	earlyRepayment2Date := c.DefaultPostForm("earlyRepayment2Date", "2099-05-25")

	earlyRepayment3Amount, _ := strconv.ParseFloat(c.DefaultPostForm("earlyRepayment3Amount", "0"), 64)
	earlyRepayment3Date := c.DefaultPostForm("earlyRepayment3Date", "2099-05-25")

	inputData = loan.Input{
		Loan: loan.Loan{
			InitialPrincipal: decimal.NewFromFloat(principal),
			InitialTerm:      loanTerm,
			InitialDate:      loan.ParseDate(startDate),
			LPR:              loan.Lprs, // 常量
			PlusSpread:       decimal.NewFromFloat(plusSpread),
			PaymentDueDay:    paymentDueDay,
		},
		EarlyRepayment: []loan.EarlyRepayment{
			{Amount: decimal.NewFromFloat(earlyRepayment1Amount), Date: loan.ParseDate(earlyRepayment1Date)},
			{Amount: decimal.NewFromFloat(earlyRepayment2Amount), Date: loan.ParseDate(earlyRepayment2Date)},
			{Amount: decimal.NewFromFloat(earlyRepayment3Amount), Date: loan.ParseDate(earlyRepayment3Date)},
		},
	}

	return inputData, nil
}
func renderTemplate(c *gin.Context, inputData loan.Input, report string) {
	c.HTML(http.StatusOK, "loan.tmpl", gin.H{
		"Principal":             inputData.Loan.InitialPrincipal,
		"LoanTerm":              inputData.Loan.InitialTerm,
		"StartDate":             inputData.Loan.InitialDate.Format("2006-01-02"),
		"PlusSpread":            inputData.Loan.PlusSpread,
		"PaymentDueDay":         inputData.Loan.PaymentDueDay,
		"earlyRepayment1Amount": inputData.EarlyRepayment[0].Amount,
		"earlyRepayment1Date":   inputData.EarlyRepayment[0].Date.Format("2006-01-02"),
		"earlyRepayment2Amount": inputData.EarlyRepayment[1].Amount,
		"earlyRepayment2Date":   inputData.EarlyRepayment[1].Date.Format("2006-01-02"),
		"earlyRepayment3Amount": inputData.EarlyRepayment[2].Amount,
		"earlyRepayment3Date":   inputData.EarlyRepayment[2].Date.Format("2006-01-02"),
		"Report":                report, // 报表结果
	})
}

func handleGETRequest(c *gin.Context) {
	inputData, err := parseFormData(c)
	if err != nil {
		// 处理错误...
		return
	}
	renderTemplate(c, inputData, "")
}

func handlePOSTRequest(c *gin.Context) {
	inputData, err := parseFormData(c)
	if err != nil {
		// 处理错误...
		return
	}

	// 计算还款计划
	action := c.Query("action")
	report := loan.LoanPrintReport(inputData, action)

	// 使用 renderTemplate 函数渲染模板
	renderTemplate(c, inputData, report)
}

func LoanRoute(env *bootstrap.Env, timeout time.Duration, group *gin.RouterGroup) {
	group.GET("/loan", handleGETRequest)
	group.POST("/loan", handlePOSTRequest)

}
