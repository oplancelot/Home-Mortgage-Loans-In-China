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


// 获取给定日期的利率

func (loan *Loan) getInterestRate(date time.Time) float64 {
    // 默认使用原始利率
    interestRate := loan.InterestRate
    for _, entry := range loan.RateEntries {
        if date.After(entry.Date) || date.Equal(entry.Date) {
            // 如果日期在或之后 RateEntry 中的日期，则更新利率和 RateChangeDate
            interestRate = entry.Rate
            loan.RateChangeDate = entry.Date
        }
    }
    return interestRate
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
	DueDate            time.Time // 还款日期
}


// 计算给定日期距离下月的天数
func daysUntilEndOfMonth(date time.Time) int {
	year, month, _ := date.Date()
	nextMonth := time.Date(year, month+1, 18, 0, 0, 0, 0, date.Location())
	daysUntilEnd := nextMonth.Sub(date).Hours() / 24
	// fmt.Println(daysUntilEnd)
	return int(daysUntilEnd)
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
		dueDate := calculateDueDate(startDate.AddDate(0, month, 0), month)
		dueDateRate := loan.getInterestRate(dueDate)

		interestPayment := remainingPrincipal * dueDateRate / 12 / 100
		// 第一个月的利息根据天数计算
		if month == 1 {
			daysInFirstMonth := float64(daysUntilEndOfMonth(startDate))
			interestPayment = remainingPrincipal * dueDateRate / 12 / 100 * (daysInFirstMonth / 31)
		}
		//日期在5.25日汇率变化，则6.18这一期，利息由不同的税率组成。


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
			InterestRate:       dueDateRate,
			DueDate:            dueDate,
		}
		payments = append(payments, payment)
	}

	return payments
}

// Calculate the due date based on the start date and month.
func calculateDueDate(startDate time.Time, month int) time.Time {
	// 创建一个新日期，月份和年份会随着时间增加，但日期始终为18号
	dueDate := time.Date(startDate.Year(), startDate.Month(), 18, 0, 0, 0, 0, startDate.Location())
	// if month ==1 {dueDate = time.Date(startDate.Year(), startDate.Month(), 25, 0, 0, 0, 0, startDate.Location())}	
	

	return dueDate
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
        {parseDate("2023-06-18"), 4.9}, // 2023年6月18日变更为4.5%的利率
        {parseDate("2024-06-18"), 4.8}, // 2024年6月18日变更为4.0%的利率
        // 添加其他日期和利率条目
    },
    RateChangeDate:  parseDate("2023-06-18"), // 第一个利率变更日期
    NewInterestRate: 4.5,                   // 第一个利率变更后的利率
}

	// 计算等额本金还款计划
	payments := loan.CalculateAmortizationSchedule()

	// 输出更详细的还款计划
	fmt.Println("期数\t还款日期\t本金\t利息\t本月还款\t剩余本金\t已支付总利息\t本月利率")
	for _, payment := range payments {
		fmt.Printf("%d\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f%%\n", payment.Month, payment.DueDate.Format("2006-01-02"), payment.Principal, payment.Interest, payment.MonthTotalAmount, payment.RemainingPrincipal, payment.TotalInterestPaid, payment.InterestRate)
	}
}
