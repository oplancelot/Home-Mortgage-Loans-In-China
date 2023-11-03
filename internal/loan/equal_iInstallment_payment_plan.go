package loan

import (
	"time"

	"github.com/shopspring/decimal"
)

// todo: 等额本息未完成,逻辑不对
func (loan *Loan) EqualInstallmentPaymentPlan(earlyRepayment []EarlyRepayment) []MonthlyPayment {
	monthlyPayments := make([]MonthlyPayment, 0)
	remainingPrincipal := loan.InitialPrincipal
	dueDate := time.Date(loan.InitialDate.Year(), loan.InitialDate.Month()+1, loan.PaymentDueDay, 0, 0, 0, 0, loan.InitialDate.Location())
	lastRemainingPrincipal := decimal.Zero
	principalPayment := decimal.Zero
	interestPayment := decimal.Zero

	for loanTerm := 1; loanTerm <= loan.InitialTerm; loanTerm++ {
		// for loanTerm := 1; loanTerm <= 2; loanTerm++ {

		amount, _ := loan.makeEarlyRepayment(remainingPrincipal, earlyRepayment, dueDate)
		// 只在提前还款后,重新计算每月应还本金;否则多次计算会有小数点导致的差异
		if amount.Cmp(remainingPrincipal) == -1 {
			remainTerm := loan.InitialTerm - loanTerm + 1
			principalPayment = amount.Div(decimal.NewFromInt(int64(remainTerm))).Round(2)

		}

		remainingPrincipal = amount
		// 以下处理每月正常还款

		// 本月还款总额
		currentYearRate := loan.getClosestLPRForYear(dueDate).Add(loan.PlusSpread)
		monthlyInterestRate := currentYearRate.Div(decimal.NewFromInt(1200))
		onePlusMonthlyInterestRate := decimal.NewFromFloat(1).Add(monthlyInterestRate)
		onePlusMonthlyInterestRatePow := onePlusMonthlyInterestRate.Pow(decimal.NewFromInt(int64(loan.InitialTerm)))
		monthlyPayment := loan.InitialPrincipal.Mul(monthlyInterestRate).Mul(onePlusMonthlyInterestRatePow).Div(onePlusMonthlyInterestRatePow.Sub(decimal.NewFromInt(1))).Round(2)

		var repaymentStatus string
		if loanTerm == 1 { // 第一期
			repaymentStatus = "A"
		} else if loanTerm%12 == 1 { // lpr变更月
			repaymentStatus = "B"
		} else if loanTerm == loan.InitialTerm-1 { // 最后一期
			repaymentStatus = "C"

		} else if loanTerm == loan.InitialTerm { // 最后一期
			repaymentStatus = "D"
		}

		switch repaymentStatus {
		case "A": // 第一期
			// 2968.28
			// 利率周期2022-05-18 ~ 2022-06-17 共30天
			// 实际天数2022-05-26 ~ 2022-06-17 共23天
			// days := int(dueDate.Sub(loan.InitialDate).Hours() / 24)

			// 利息
			days := loan.daysDiff(loan.InitialDate, dueDate).Sub(decimal.NewFromInt(1))
			interestPayment = remainingPrincipal.Mul(currentYearRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360)).Mul(days).Round(2)
			// 本金
			principalPayment := monthlyPayment.Sub(interestPayment)
			// 剩余
			remainingPrincipal = remainingPrincipal.Sub(principalPayment)
			// fmt.Printf("principal")
		case "B": // lpr变更月
			// 分为两段
			daysBefore, daysAfter := loan.currentYearLPRUpdate(dueDate)
			// 上一年利率 2023-05-18 ~ 2023-05-24
			previousDueDate := loan.previousDueDate(dueDate)
			previousYearRate := loan.getClosestLPRForYear(previousDueDate).Add(loan.PlusSpread)
			interestPayment = remainingPrincipal.Mul(previousYearRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360)).Mul(daysBefore).Round(4)
			// 当年利率 2023-05-25 ~ 2023-06-18
			interestPayment = interestPayment.Add((remainingPrincipal.Mul(currentYearRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360))).Mul(daysAfter)).Round(2)

			principalPayment := monthlyPayment.Sub(interestPayment)
			remainingPrincipal = remainingPrincipal.Sub(principalPayment)
		case "C": // 倒数第二期
			lastRemainingPrincipal = remainingPrincipal
		case "D": // 最后一期
			// 最后一期还款日变更
			lastDueDate := loan.InitialDate.AddDate(0, loanTerm, 0)
			dueDate = lastDueDate

			// 最后一期本金和利息
			principalPayment = lastRemainingPrincipal
			days := loan.daysDiff(loan.previousDueDate(dueDate), lastDueDate)
			// 利息=剩余本金*每天利率*天数
			interestPayment = principalPayment.Mul(currentYearRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360)).Mul(days).Round(2)
			monthlyPayment = principalPayment.Add(interestPayment)
			// 剩余本金
			remainingPrincipal = decimal.Zero

		default:
			interestPayment := remainingPrincipal.Mul(monthlyInterestRate).Round(2)
			principalPayment := monthlyPayment.Sub(interestPayment)
			remainingPrincipal = remainingPrincipal.Sub(principalPayment)
		}

		payment := MonthlyPayment{
			LoanTerm:           loanTerm,
			Principal:          principalPayment,
			Interest:           interestPayment,
			MonthTotalAmount:   monthlyPayment,
			RemainingPrincipal: remainingPrincipal,
			DueDateRate:        currentYearRate,
			DueDate:            dueDate,
		}

		monthlyPayments = append(monthlyPayments, payment)

		dueDate = dueDate.AddDate(0, 1, 0)
	}

	return monthlyPayments
}
