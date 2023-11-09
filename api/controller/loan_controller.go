package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/oplancelot/Home-Mortgage-Loans-In-China/internal/loan"
	"github.com/shopspring/decimal"
)

// InputValidator 是一个实现了 Validator 接口的结构体
type InputValidator struct{}

func (v InputValidator) Validate(c *gin.Context) (inputData loan.Input, err error, action string) {
	// 获取表单数据
	action = c.PostForm("action")
	principal, err := strconv.ParseFloat(c.DefaultPostForm("principal", "0"), 64)
	if err != nil {
		return loan.Input{}, errors.New("Invalid principal: it should be a number"), action
	}
	if principal < 0 || principal > 100000000 {
		return loan.Input{}, errors.New("Invalid InitialPrincipal: it should be between 0 and 100000000"), action
	}

	loanTerm, err := strconv.Atoi(c.DefaultPostForm("loanTerm", "12"))
	if err != nil {
		return loan.Input{}, errors.New("Invalid loanTerm: it should be an integer"), action
	}
	if loanTerm < 12 || loanTerm > 360 {
		return loan.Input{}, errors.New("Invalid loanTerm: it should be between 12 and 360"), action
	}

	startDate := c.DefaultPostForm("startDate", "2022-05-25")
	if startDate == "" {
		return loan.Input{}, errors.New("Invalid startDate: it cannot be empty"), action
	}

	plusSpread, err := strconv.ParseFloat(c.DefaultPostForm("plusSpread", "0"), 64)
	if err != nil {
		return loan.Input{}, errors.New("Invalid plusSpread: it should be a number"), action
	}
	if plusSpread >= 1 {
		return loan.Input{}, errors.New("Invalid plusSpread: it should be between 0 and 1"), action
	}
	paymentDueDay, err := strconv.Atoi(c.DefaultPostForm("paymentDueDay", "1"))
	if err != nil {
		return loan.Input{}, errors.New("Invalid paymentDueDay: it should be an integer"), action
	}

	if paymentDueDay < 1 || paymentDueDay > 31 {
		return loan.Input{}, errors.New("Invalid PaymentDueDay: it should be between 1 and 31"), action
	}
	// 获取提前还款信息的值
	earlyRepayment1Amount, err := strconv.ParseFloat(c.DefaultPostForm("earlyRepayment1Amount", "0"), 64)
	if err != nil {
		return loan.Input{}, errors.New("Invalid earlyRepaymentAmount: it should be a number"), action
	}
	earlyRepayment1Date := c.DefaultPostForm("earlyRepayment1Date", "2023-08-19")

	earlyRepayment2Amount, _ := strconv.ParseFloat(c.DefaultPostForm("earlyRepayment2Amount", "0"), 64)
	if err != nil {
		return loan.Input{}, errors.New("Invalid earlyRepaymentAmount: it should be a number"), action
	}
	earlyRepayment2Date := c.DefaultPostForm("earlyRepayment2Date", "2099-05-25")

	earlyRepayment3Amount, _ := strconv.ParseFloat(c.DefaultPostForm("earlyRepayment3Amount", "0"), 64)
	if err != nil {
		return loan.Input{}, errors.New("Invalid earlyRepaymentAmount: it should be a number"), action
	}
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
	return inputData, nil, action
}
