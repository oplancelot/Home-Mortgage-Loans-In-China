package route

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oplancelot/Home-Mortgage-Loans-In-China/api/controller"
	"github.com/oplancelot/Home-Mortgage-Loans-In-China/bootstrap"
	"github.com/oplancelot/Home-Mortgage-Loans-In-China/internal/loan"
)

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
	validator := controller.InputValidator{}
	inputData, err, _ := validator.Validate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	renderTemplate(c, inputData, "")
}

func handlePOSTRequest(c *gin.Context) {
	validator := controller.InputValidator{}
	inputData, err, action := validator.Validate(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	report := loan.LoanPrintTable(inputData, action)

	// 使用 renderTemplate 函数渲染模板
	renderTemplate(c, inputData, report)
}

func LoanRoute(env *bootstrap.Env, timeout time.Duration, group *gin.RouterGroup) {
	// 计算还款计划

	group.GET("/loan", handleGETRequest)
	group.POST("/loan", handlePOSTRequest)

}
