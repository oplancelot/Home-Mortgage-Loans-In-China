package main

import (
	"fmt"
	"time"
)

// Loan represents the loan details.
type Loan struct {
	Principal        float64     // 初始本金
	InterestRate     float64     // 年利率
	TermInMonths     int         // 贷款期限（月）
	StartDate        string      // 放款日期
	RateEntries      []RateEntry // 日期与利率的条目列表
	RateChangeDate   time.Time   // 年利率变更日期
	NewInterestRate  float64     // 变更后的年利率
}

// RateEntry represents the date and interest rate entry.
type RateEntry struct {
	Date time.Time // 日期
	Rate float64   // 利率
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
	monthlyPayment := loan.Principal / float64(loan.TermInMonths)
	payments := make([]Payment, 0)

	remainingPrincipal := loan.Principal
	totalInterestPaid := 0.0

	// 解析放款日期
	startDate := parseDate(loan.StartDate)

	for month := 1; month <= loan.TermInMonths; month++ {
		// 计算当月的利率
		currentMonth := startDate.AddDate(0, month, 0)

		
		currentRate := loan.getInterestRate(currentMonth)

		interestPayment := remainingPrincipal * currentRate / 12 / 100
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
			InterestRate:       currentRate,
		}
		payments = append(payments, payment)
	}

	return payments
}

// 获取给定日期的利率
func (loan *Loan) getInterestRate(date time.Time) float64 {
	if date.Before(loan.RateChangeDate) || date.Equal(loan.RateChangeDate) {
		return loan.InterestRate // 在利率变更日期之前使用原始利率
	}
	return loan.NewInterestRate // 在利率变更日期之后使用新的利率
}

// 解析日期字符串并返回时间
func parseDate(dateString string) time.Time {
    parsedTime, err := time.Parse("2006-01-02", dateString)
    if err != nil {
        panic(err)
    }
    return parsedTime
}



func main() {
	// 输入贷款信息
	initialPrincipal := 920000.0 // 初始本金
	interestRate := 5.05         // 年利率（百分比）
	loanTerm := 360              // 贷款期限（月）
	startDate := "2022-05-25"      // 放款日期

// 创建 Loan 结构
loan := Loan{
    Principal:    initialPrincipal,
    InterestRate: interestRate,
    TermInMonths: loanTerm,
    StartDate:    startDate,
    RateEntries: []RateEntry{
        {parseDate("2023-06-18"), 4.5}, // 2023年6月18日变更为4.5%的利率
        {parseDate("2024-06-18"), 4.0}, // 2024年6月18日变更为4.0%的利率
        // 添加其他日期和利率条目
    },
    RateChangeDate:  parseDate("2023-06-18"), // 第一个利率变更日期
    NewInterestRate: 4.5,                   // 第一个利率变更后的利率
}


	// 计算等额本金还款计划
	payments := loan.CalculateAmortizationSchedule()

	// 输出更详细的还款计划
	fmt.Println("期数\t本金\t利息\t本月还款\t剩余本金\t已支付总利息\t本月利率")
	for _, payment := range payments {
		fmt.Printf("%d\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f%%\n", payment.Month, payment.Principal, payment.Interest, payment.MonthTotalAmount, payment.RemainingPrincipal, payment.TotalInterestPaid, payment.InterestRate)
	}
}
