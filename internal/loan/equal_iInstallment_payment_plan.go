package loan

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// !EMI 是  "Equal Monthly Installment" 的缩写，中文翻译为"等额月供" 或" 等额本息"
// 是指在等额本息还款方式下，每月偿还的贷款金额固定，包括本金和利息。在整个还款期内，每月的还款额相同，这使得借款人更容易规划每月的负担，因为每月的还款金额是固定的

// 等额本息计算公式
// EMI = ( P*r* ( (1+r)**n) ) / ( ((1+r)**n)-1 )
// EMI 是每月还款额
// P 是贷款本金
// r 是月利率
// n 是还款总期数

// emi
func (loan *Loan) calculateEMI(remainingPrincipal decimal.Decimal, monthlyInterestRate decimal.Decimal, loanTerm int, dueDate time.Time) decimal.Decimal {
	onePlusMonthlyInterestRate := decimal.NewFromFloat(1).Add(monthlyInterestRate)
	onePlusMonthlyInterestRatePow := onePlusMonthlyInterestRate.Pow(decimal.NewFromInt(int64(loan.InitialTerm - loanTerm + 1)))
	emi := remainingPrincipal.Mul(monthlyInterestRate).Mul(onePlusMonthlyInterestRatePow).Div(onePlusMonthlyInterestRatePow.Sub(decimal.NewFromInt(1))).Round(2)
	return emi
}

func (loan *Loan) EqualInstallmentPaymentPlan(earlyRepayment []EarlyRepayment) []MonthlyPayment {
	monthlyPayments := make([]MonthlyPayment, 0)
	remainingPrincipal := loan.InitialPrincipal
	dueDate := time.Date(loan.InitialDate.Year(), loan.InitialDate.Month()+1, loan.PaymentDueDay, 0, 0, 0, 0, loan.InitialDate.Location())
	yearlyInterestRate := loan.getClosestLPRForYear(dueDate).Add(loan.PlusSpread)
	monthlyInterestRate := yearlyInterestRate.Div(decimal.NewFromInt(1200))
	emi := loan.calculateEMI(remainingPrincipal, monthlyInterestRate, 1, dueDate)
	lastRemainingPrincipal := decimal.Zero
	principalPayment := decimal.Zero
	interestPayment := decimal.Zero

	for loanTerm := 1; loanTerm <= loan.InitialTerm; loanTerm++ {

		// 还款总额或者利率变化需要重新计算
		// 提前还款会影响还款金额
		amount, _ := loan.makeEarlyRepayment(remainingPrincipal, earlyRepayment, dueDate)
		// 只在提前还款后,重新计算每月应还本金;否则多次计算会有小数点导致的差异
		if amount.Cmp(remainingPrincipal) == -1 {

			remainingPrincipal = amount.Round(2)
			emi = loan.calculateEMI(remainingPrincipal, monthlyInterestRate, loanTerm, dueDate)
		}

		// 利率变化,每年利率变更月重算一次.每月计算因为小数问题会导致有差异.
		if loanTerm%12 == 1 {
			lastyearinterestRate := yearlyInterestRate
			yearlyInterestRate = loan.getClosestLPRForYear(dueDate).Add(loan.PlusSpread)
			monthlyInterestRate = yearlyInterestRate.Div(decimal.NewFromInt(1200))
			if lastyearinterestRate.Cmp(yearlyInterestRate) != 0 {
				emi = loan.calculateEMI(remainingPrincipal, monthlyInterestRate, loanTerm, dueDate)
			}
		}

		switch {
		case loanTerm == 1: // 第一期
			// 第一个月一般不到30天,
			// 如果按EMI还,那么第一个月本金=EMI-第一个月利息
			// 如果不还本金,那么只用计算利息.实际从第二个月开始还款,EMI=第二个月本金+2个月利息

			// 2968.28
			// 利率周期2022-05-18 ~ 2022-06-17 共30天
			// 实际天数2022-05-26 ~ 2022-06-17 共23天
			// days := int(dueDate.Sub(loan.InitialDate).Hours() / 24)

			// 第一个月天数D=30−放款日+1
			days := loan.daysDiff(loan.InitialDate, dueDate).Sub(decimal.NewFromInt(1))
			// 第一期利息按天数计算
			interestPayment = remainingPrincipal.Mul(yearlyInterestRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360)).Mul(days).Round(2)
			// 本金
			principalPayment = emi.Sub(interestPayment)
			remainingPrincipal = remainingPrincipal.Sub(principalPayment)

		case loanTerm%12 == 1: // lpr变更月
			// 分为两段
			daysBefore, daysAfter := loan.currentYearLPRUpdate(dueDate)
			// 上一年利率 2023-05-18 ~ 2023-05-24
			previousDueDate := loan.previousDueDate(dueDate)
			previousYearRate := loan.getClosestLPRForYear(previousDueDate).Add(loan.PlusSpread)
			interestPayment = remainingPrincipal.Mul(previousYearRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360)).Mul(daysBefore).Round(4)
			// 当年利率 2023-05-25 ~ 2023-06-18
			interestPayment = interestPayment.Add((remainingPrincipal.Mul(yearlyInterestRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360))).Mul(daysAfter)).Round(2)
			principalPayment = emi.Sub(interestPayment)
			remainingPrincipal = remainingPrincipal.Sub(principalPayment)

		case loanTerm == loan.InitialTerm-1: // 倒数第二期

			interestPayment = remainingPrincipal.Mul(monthlyInterestRate).Round(2)
			principalPayment = emi.Sub(interestPayment)
			remainingPrincipal = remainingPrincipal.Sub(principalPayment)
			lastRemainingPrincipal = remainingPrincipal
			fmt.Println(lastRemainingPrincipal)
		case loanTerm == loan.InitialTerm: // 最后一期

			// 最后一期还款日变更
			lastDueDate := loan.InitialDate.AddDate(0, loanTerm, 0)
			dueDate = lastDueDate

			// 最后一期本金和利息
			principalPayment = lastRemainingPrincipal

			// 利息=剩余本金*每天利率*天数
			days := loan.daysDiff(loan.previousDueDate(dueDate), lastDueDate)
			interestPayment = principalPayment.Mul(yearlyInterestRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360)).Mul(days).Round(2)
			emi = principalPayment.Add(interestPayment)
			// 剩余本金
			remainingPrincipal = decimal.Zero

		default:
			// fmt.Println(loanTerm)
			interestPayment = remainingPrincipal.Mul(monthlyInterestRate).Round(2)
			principalPayment = emi.Sub(interestPayment)
			remainingPrincipal = remainingPrincipal.Sub(principalPayment)
		}

		// 剩余本金带入下一期计算

		payment := MonthlyPayment{
			LoanTerm:           loanTerm,
			Principal:          principalPayment,
			Interest:           interestPayment,
			MonthTotalAmount:   emi,
			RemainingPrincipal: remainingPrincipal,
			DueDateRate:        yearlyInterestRate,
			DueDate:            dueDate,
		}

		monthlyPayments = append(monthlyPayments, payment)

		dueDate = dueDate.AddDate(0, 1, 0)
	}

	return monthlyPayments
}
