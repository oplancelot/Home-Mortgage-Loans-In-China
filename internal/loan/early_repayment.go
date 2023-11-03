package loan

import (
	"time"

	"github.com/shopspring/decimal"
)

// 提前还款
type EarlyRepayment struct {
	Amount             decimal.Decimal // 提前还款金额
	Date               time.Time       // 提前还款日期
	DueDateRate        decimal.Decimal // 当月利率
	Principal          decimal.Decimal // 本金部分
	Interest           decimal.Decimal // 利息部分
	RemainingPrincipal decimal.Decimal // 剩余本金
}

func (loan *Loan) makeEarlyRepayment(remainingPrincipal decimal.Decimal, earlyRepayments []EarlyRepayment, dueDate time.Time) (amount, daysDiff decimal.Decimal) {
	previousDueDate := loan.previousDueDate(dueDate)
	for i, early := range earlyRepayments {
		if early.Date.After(previousDueDate) && early.Date.Before(dueDate) {
			currentYearLPR := loan.getClosestLPRForYear(dueDate)
			currentYearRate := currentYearLPR.Add(loan.PlusSpread).Div(decimal.NewFromInt(100))
			daysDiff = loan.daysDiff(previousDueDate, early.Date)
			earlyInterest := remainingPrincipal.Mul(currentYearRate).Div(decimal.NewFromInt(360)).Mul(daysDiff)

			amount = remainingPrincipal.Add(earlyInterest).Sub(early.Amount)
			// fmt.Println(dueDate, currentYearRate, earlyInterest, remainingPrincipal, daysDiff)

			// 更新本金利息和利率
			early.Principal = early.Amount.Sub(earlyInterest).Round(2)
			early.Interest = earlyInterest.Round(2)
			early.RemainingPrincipal = remainingPrincipal.Sub(early.Principal).Round(2)
			early.DueDateRate = currentYearRate.Mul(decimal.NewFromInt(100))
			// 将更新后的 early 对象存储回 earlyRepayments 切片中
			earlyRepayments[i] = early
			return amount, daysDiff

		}
	}
	// 如果本期没有提前还款没有则返回原值
	return remainingPrincipal, decimal.Zero
}
