package main

import (
	"fmt"
	"time"
)

// Loan represents the loan details.
type Loan struct {
	Principal    float64 // 初始本金
	InterestRate float64 // 年利率
	TermInMonths int     // 贷款期限（月）
	StartDate    string  // 放款日期
}

// Payment represents the details of each monthly payment.
type Payment struct {
	Month              int     // 期数
	Principal          float64 // 本金部分（固定为每月还款金额）
	Interest           float64 // 利息部分
	MonthTotalAmount   float64 // 本月还款总金额
	RemainingPrincipal float64 // 剩余本金
	TotalInterestPaid  float64 // 已支付总利息
	InterestRate       float64 // 利率
}

// CalculateAmortizationSchedule calculates the amortization schedule for the loan.
func (loan *Loan) CalculateAmortizationSchedule() []Payment {
	monthlyInterestRate := loan.InterestRate / 12 / 100
	monthlyPayment := loan.Principal / float64(loan.TermInMonths)
	payments := make([]Payment, 0)

	remainingPrincipal := loan.Principal
	totalInterestPaid := 0.0

	// 解析放款日期
	startDate, _ := parseDate(loan.StartDate)
	daysInFirstMonth := float64(daysUntilEndOfMonth(startDate))

	for month := 1; month <= loan.TermInMonths; month++ {
		interestPayment := remainingPrincipal * monthlyInterestRate
		// 第一个月的利息根据天数计算
		if month == 1 {
			interestPayment = remainingPrincipal * monthlyInterestRate * (daysInFirstMonth / 31)
		}
		// 固定本金支付为每月还款金额
		principalPayment := monthlyPayment
		remainingPrincipal -= principalPayment
		totalInterestPaid += interestPayment

		payment := Payment{
			Month:              month,
			Principal:          principalPayment,
			Interest:           interestPayment,
			MonthTotalAmount:   monthlyPayment + interestPayment,
			RemainingPrincipal: remainingPrincipal,
			TotalInterestPaid:  totalInterestPaid,
			InterestRate:       loan.InterestRate,
		}
		payments = append(payments, payment)
	}

	return payments
}

// 解析日期字符串并返回时间
func parseDate(dateString string) (time.Time, error) {
	return time.Parse("20060102", dateString)
}

// 计算给定日期距离下月的天数
func daysUntilEndOfMonth(date time.Time) int {
	year, month, _ := date.Date()
	nextMonth := time.Date(year, month+1, 18, 0, 0, 0, 0, date.Location())
	daysUntilEnd := nextMonth.Sub(date).Hours() / 24
	fmt.Println(daysUntilEnd)
	return int(daysUntilEnd)
}

func main() {
	// 输入贷款信息
	initialPrincipal := 920000.0 // 初始本金
	interestRate := 5.05         // 年利率（百分比）
	loanTerm := 360              // 贷款期限（月）
	startDate := "20220525"      // 放款日期

	// 创建 Loan 结构
	loan := Loan{
		Principal:    initialPrincipal,
		InterestRate: interestRate,
		TermInMonths: loanTerm,
		StartDate:    startDate,
	}

	// 计算等额本金还款计划
	payments := loan.CalculateAmortizationSchedule()

	// 输出更详细的还款计划
	fmt.Println("期数\t本金\t利息\t本月还款\t剩余本金\t已支付总利息\t本月利率")
	for _, payment := range payments {
		fmt.Printf("%d\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f%%\n", payment.Month, payment.Principal, payment.Interest, payment.MonthTotalAmount, payment.RemainingPrincipal, payment.TotalInterestPaid, payment.InterestRate)
	}
}
