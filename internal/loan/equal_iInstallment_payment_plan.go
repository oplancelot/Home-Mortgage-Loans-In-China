package loan

import (
	"time"

	"github.com/shopspring/decimal"
)

func (loan *Loan) EqualInstallmentPaymentPlan(earlyRepayment []EarlyRepayment) []MonthlyPayment {
	// 计算利息的规则

	// 1.天数:全年360天,12个月每月30天
	// 2.年利率=(lpr+加点)/100
	//		每日利率=剩余本金*年利率/360
	// 3.每日利息=剩余本金*每日利率
	//		还款日还的利息=上个月剩余本金*每日利息*30
	// 4.如果是lpr变更的月份,分为两段计算.
	//		第一段lpr为前一年lpr,天数是变更日~还款日(取头去尾)
	//		第二段为当年lpr,天数是30-第一段
	// 5.提前还款会对下月的还款计算有影响
	//		暂不考虑lpr变更这个月提前还款.
	// 6.第一期利息的计算
	//		放款日当天不计算,原因是可能下午才放款?
	//		天数 = 还款日-放款日-1
	// 7.最后一期本金的计算去掉误差
	//		最后一期本金 = 贷款金额 - (贷款金额/期数).round(2)*(期数-1)
	//		最后一期还款日 = 默认是放款日,而不是还款日

	monthlypayments := make([]MonthlyPayment, 0)
	remainingPrincipal := loan.InitialPrincipal
	principalPayment := remainingPrincipal.Div(decimal.NewFromInt(int64(loan.InitialTerm))).Round(2)
	lastPrincipalPayment := principalPayment.Add(remainingPrincipal.Sub(principalPayment.Mul(decimal.NewFromInt(int64(loan.InitialTerm)))))
	dueDate := time.Date(loan.InitialDate.Year(), loan.InitialDate.Month()+1, loan.PaymentDueDay, 0, 0, 0, 0, loan.InitialDate.Location())

	for loanTerm := 1; loanTerm <= loan.InitialTerm; loanTerm++ {

		interestPayment := decimal.Zero
		days := decimal.NewFromInt(int64(30))
		// 计算如果有提前还款则需要减去提前还款的本金
		amount, daysDiff := loan.makeEarlyRepayment(remainingPrincipal, earlyRepayment, dueDate)
		// 只在提前还款后,重新计算每月应还本金;否则多次计算会有小数点导致的差异
		if amount.Cmp(remainingPrincipal) == -1 {
			remainTerm := loan.InitialTerm - loanTerm + 1
			principalPayment = amount.Div(decimal.NewFromInt(int64(remainTerm))).Round(2)
			lastPrincipalPayment = principalPayment.Add(amount.Sub(principalPayment.Mul(decimal.NewFromInt(int64(remainTerm)))))

		}

		remainingPrincipal = amount
		// 以下处理每月正常还款
		// 计算利率利息
		currentYearRate := loan.getClosestLPRForYear(dueDate).Add(loan.PlusSpread)

		var repaymentStatus string
		if loanTerm == 1 { // 第一期
			repaymentStatus = "A"
		} else if loanTerm%12 == 1 { // lpr变更月
			repaymentStatus = "B"
		} else if loanTerm == loan.InitialTerm { // 最后一期
			repaymentStatus = "C"

		} else {
			repaymentStatus = "D" // 默认
		}

		switch repaymentStatus {
		case "A": // 第一期
			// 2968.28
			// 利率周期2022-05-18 ~ 2022-06-17 共30天
			// 实际天数2022-05-26 ~ 2022-06-17 共23天
			// days := int(dueDate.Sub(loan.InitialDate).Hours() / 24)
			days = loan.daysDiff(loan.InitialDate, dueDate).Sub(decimal.NewFromInt(1))
			interestPayment = remainingPrincipal.Mul(currentYearRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360)).Mul(days).Round(2)
			// fmt.Printf("interest")
		case "B": // lpr变更月
			// 分为两段
			daysBefore, daysAfter := loan.currentYearLPRUpdate(dueDate)

			// 上一年利率 2023-05-18 ~ 2023-05-24
			previousDueDate := loan.previousDueDate(dueDate)
			previousYearRate := loan.getClosestLPRForYear(previousDueDate).Add(loan.PlusSpread)
			interestPayment = remainingPrincipal.Mul(previousYearRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360)).Mul(daysBefore).Round(4)

			// 当年利率 2023-05-25 ~ 2023-06-18
			interestPayment = interestPayment.Add((remainingPrincipal.Mul(currentYearRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360))).Mul(daysAfter)).Round(2)

		case "C": // 最后一期
			lastDueDate := loan.InitialDate.AddDate(0, loanTerm, 0)
			days = loan.daysDiff(loan.previousDueDate(dueDate), lastDueDate)
			principalPayment = lastPrincipalPayment.Round(2)
			dueDate = lastDueDate
			interestPayment = lastPrincipalPayment.Mul(currentYearRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360)).Mul(days).Round(2)

		default:
			remainDay := days.Sub(daysDiff)
			interestPayment = remainingPrincipal.Mul(currentYearRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360)).Mul(remainDay).Round(2)

		}

		remainingPrincipal = remainingPrincipal.Sub(principalPayment).Round(2)

		payment := MonthlyPayment{
			LoanTerm:           loanTerm,
			Principal:          principalPayment,
			Interest:           interestPayment,
			MonthTotalAmount:   principalPayment.Add(interestPayment),
			RemainingPrincipal: remainingPrincipal,
			DueDateRate:        currentYearRate,
			DueDate:            dueDate,
		}
		monthlypayments = append(monthlypayments, payment)

		// 下一个还款日期
		dueDate = dueDate.AddDate(0, 1, 0)

	}

	return monthlypayments
}