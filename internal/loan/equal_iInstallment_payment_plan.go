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

	for loanTerm := 1; loanTerm <= loan.InitialTerm; loanTerm++ {
		currentYearRate := decimal.NewFromFloat(5.05)
		monthlyInterestRate := currentYearRate.Div(decimal.NewFromInt(120))
		onePlusMonthlyInterestRate := decimal.NewFromFloat(1).Add(monthlyInterestRate)
		onePlusMonthlyInterestRatePow := onePlusMonthlyInterestRate.Pow(decimal.NewFromInt(int64(loan.InitialTerm)))
		monthlyPayment := remainingPrincipal.Mul(monthlyInterestRate).Div(onePlusMonthlyInterestRatePow.Sub(decimal.NewFromInt(1))).Round(2)

		interestPayment := remainingPrincipal.Mul(monthlyInterestRate).Round(2)
		principalPayment := monthlyPayment.Sub(interestPayment)

		if loanTerm == loan.InitialTerm {
			principalPayment = remainingPrincipal
			monthlyPayment = principalPayment.Add(interestPayment)
		}

		payment := MonthlyPayment{
			LoanTerm:           loanTerm,
			Principal:          principalPayment,
			Interest:           interestPayment,
			MonthTotalAmount:   principalPayment.Add(interestPayment),
			RemainingPrincipal: remainingPrincipal,
			DueDateRate:        currentYearRate,
			DueDate:            dueDate,
		}

		monthlyPayments = append(monthlyPayments, payment)

		remainingPrincipal = remainingPrincipal.Sub(principalPayment)
		dueDate = dueDate.AddDate(0, 1, 0)
	}

	return monthlyPayments
}
