package main

import (
	"fmt"
	"time"
)

// Loan represents the loan details.
type Loan struct {
	Principal    float64   // 初始本金
	DefaultLPR   float64   // 默认利率
	PlusSpread   float64   // 加点
	TermInMonths int       // 贷款期限（月）
	StartDate    time.Time // 放款年月日
	LPRS         []LPR     // 日期与利率的条目列表
	Dueday       int       // 还款日 (1-31)
}

// LPR represents the date and interest LPR entry.
type LPR struct {
	Date time.Time // 日期
	Lpr  float64   // 利率
}

// Payment represents the details of each monthly payment.
type Payment struct {
	Month              int       // 期数
	Principal          float64   // 本金部分（固定为每月还款金额）
	Interest           float64   // 利息部分
	MonthTotalAmount   float64   // 当月还款总金额
	RemainingPrincipal float64   // 剩余本金
	TotalInterestPaid  float64   // 已支付总利息
	DueDateRate        float64   // 当月利率
	DueDate            time.Time // 当月还款日期
}

// 解析日期字符串并返回时间
func parseDate(dateString string) time.Time {
	parsedTime, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		panic(err)
	}
	return parsedTime
}

// 计算给定日期距离下月的天数
func daysUntilEndOfMonth(date time.Time) int {
	year, month, _ := date.Date()
	nextMonth := time.Date(year, month+1, 18, 0, 0, 0, 0, date.Location())
	daysUntilEnd := nextMonth.Sub(date).Hours() / 24
	// fmt.Println(daysUntilEnd)
	return int(daysUntilEnd)
}

// Calculate the due date based on the start date and month.
func calculateDueDate(startDate time.Time, dueday int) time.Time {
	// 创建一个新日期，月份和年份会随着时间增加，日期为贷款合同中指定的还款日
	dueDate := time.Date(startDate.Year(), startDate.Month()+1, dueday, 0, 0, 0, 0, startDate.Location())
	return dueDate
}

// 获取指定年份的最接近的LPR
func (loan *Loan) getClosestLPRForYear(dueDate time.Time) float64 {
	// 提取startDate和dueDate的月份和日期
	startMonth := loan.StartDate.Month()
	startDay := loan.StartDate.Day()
	dueYear := dueDate.Year()
	dueMonth := dueDate.Month()
	dueDay := loan.Dueday

	// 找到小于或等于给定年份的最近的RateEntry日期
	var selectedRate float64
	var compareDate time.Time

	if dueMonth < startMonth || (dueMonth == startMonth && dueDay < startDay) {
		compareDate = time.Date(dueYear-1, startMonth, startDay, 0, 0, 0, 0, loan.StartDate.Location())
	} else {
		compareDate = time.Date(dueYear, startMonth, startDay, 0, 0, 0, 0, loan.StartDate.Location())
	}
	// fmt.Println(dueDate, compareDate, selectedRate)

	for _, entry := range loan.LPRS {
		// 如果找到比 compareDate 之前的日期，且比当前选定的日期更接近，则更新 selectedRate 和 compareDate
		if entry.Date.Before(compareDate) && (selectedRate == 0 || compareDate.Sub(entry.Date) < compareDate.Sub(compareDate)) {
			selectedRate = entry.Lpr
			compareDate = entry.Date
		}
	}
	// fmt.Println(dueDate, compareDate, selectedRate)

	return selectedRate

}

// CalculateAmortizationSchedule calculates the amortization schedule for the loan.
func (loan *Loan) CalculateAmortizationSchedule() []Payment {

	monthlyPayment := loan.Principal / float64(loan.TermInMonths)
	payments := make([]Payment, 0)

	remainingPrincipal := loan.Principal
	totalInterestPaid := 0.0
	dueDate := calculateDueDate(loan.StartDate, loan.Dueday)

	// fmt.Println(dueDate)

	for month := 1; month <= loan.TermInMonths; month++ {
		currentYearLPR := loan.getClosestLPRForYear(dueDate)
		currentYearRate := currentYearLPR + loan.PlusSpread
		// previousYearRate := previousYearLPR + loan.PlusSpread

		var interestPayment float64
		// 第一个月的利息根据天数计算
		if month == 1 {
			daysInFirstMonth := float64(daysUntilEndOfMonth(loan.StartDate))
			interestPayment = remainingPrincipal * currentYearRate / 12 / 100 * (daysInFirstMonth / 31)
			// } else if dueDate.Month() == loan.StartDate.Month() {
			// 	interestPayment = remainingPrincipal * (previousYearRate) / 12 / 100 //lpr变更当月
		} else {
			interestPayment = remainingPrincipal * (currentYearRate) / 12 / 100

		}

		// interestPayment = remainingPrincipal * (currentYearRate) / 12 / 100
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
			DueDateRate:        currentYearRate,
			DueDate:            dueDate,
		}
		payments = append(payments, payment)

		// 计算下一个还款日期
		dueDate = dueDate.AddDate(0, 1, 0)

	}

	return payments
}

func main() {
	// LPR
	lprs := []LPR{
		{parseDate("2023-08-21"), 4.20},
		{parseDate("2023-07-20"), 4.20},
		{parseDate("2023-06-20"), 4.20},
		{parseDate("2023-05-22"), 4.30},
		{parseDate("2023-04-20"), 4.30},
		{parseDate("2023-03-20"), 4.30},
		{parseDate("2023-02-20"), 4.30},
		{parseDate("2023-01-20"), 4.30},
		{parseDate("2022-12-20"), 4.30},
		{parseDate("2022-11-21"), 4.30},
		{parseDate("2022-10-20"), 4.30},
		{parseDate("2022-09-20"), 4.30},
		{parseDate("2022-08-22"), 4.30},
		{parseDate("2022-07-20"), 4.45},
		{parseDate("2022-06-20"), 4.45},
		{parseDate("2022-05-20"), 4.45},
		{parseDate("2022-04-20"), 4.60},
		{parseDate("2022-03-21"), 4.60},
		{parseDate("2022-02-21"), 4.60},
		{parseDate("2022-01-20"), 4.60},
		{parseDate("2021-12-20"), 4.65},
		{parseDate("2021-11-22"), 4.65},
		{parseDate("2021-10-20"), 4.65},
		{parseDate("2021-09-22"), 4.65},
		{parseDate("2021-08-20"), 4.65},
		{parseDate("2021-07-20"), 4.65},
		{parseDate("2021-06-21"), 4.65},
		{parseDate("2021-05-20"), 4.65},
		{parseDate("2021-04-20"), 4.65},
		{parseDate("2021-03-22"), 4.65},
		{parseDate("2021-02-22"), 4.65},
		{parseDate("2021-01-20"), 4.65},
		{parseDate("2020-12-21"), 4.65},
		{parseDate("2020-11-20"), 4.65},
		{parseDate("2020-10-20"), 4.65},
		{parseDate("2020-09-21"), 4.65},
		{parseDate("2020-08-20"), 4.65},
		{parseDate("2020-07-20"), 4.65},
		{parseDate("2020-06-22"), 4.65},
		{parseDate("2020-05-20"), 4.65},
		{parseDate("2020-04-20"), 4.65},
		{parseDate("2020-03-20"), 4.75},
		{parseDate("2020-02-20"), 4.75},
		{parseDate("2020-01-20"), 4.80},
		{parseDate("2019-12-20"), 4.80},
		{parseDate("2019-11-20"), 4.80},
		{parseDate("2019-10-21"), 4.85},
		{parseDate("2019-09-20"), 4.85},
		{parseDate("2019-08-20"), 4.85},
	}

	// 输入贷款信息
	initialPrincipal := 920000.0         // 初始本金
	defaultLPR := 8.05                   // 年利率（百分比）
	loanTerm := 360                      // 贷款期限（月）
	startDate := parseDate("2022-05-25") // 放款日期
	plusSpread := 0.60                   // 上浮点数
	dueday := 18                         // 还款日
	// 创建 Loan 结构
	loan := Loan{
		Principal:    initialPrincipal,
		DefaultLPR:   defaultLPR,
		TermInMonths: loanTerm,
		StartDate:    startDate,
		LPRS:         lprs,
		PlusSpread:   plusSpread,
		Dueday:       dueday,
	}

	// 计算等额本金还款计划
	payments := loan.CalculateAmortizationSchedule()

	// 输出更详细的还款计划
	fmt.Println("期数\t还款日期\t本金\t利息\t本月还款\t剩余本金\t已支付总利息\t本月利率")
	for _, payment := range payments {
		fmt.Printf("%d\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f%%\n", payment.Month, payment.DueDate.Format("2006-01-02"), payment.Principal, payment.Interest, payment.MonthTotalAmount, payment.RemainingPrincipal, payment.TotalInterestPaid, payment.DueDateRate)
		// 	// 只打印前10行记录
		// 	if payment.Month <= 10 {
		// 		fmt.Printf("%d\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f%%\n", payment.Month, payment.DueDate.Format("2006-01-02"), payment.Principal, payment.Interest, payment.MonthTotalAmount, payment.RemainingPrincipal, payment.TotalInterestPaid, payment.DueDateRate)
		// 	} else {
		// 		break // 如果已经打印了前10行，就退出循环
		// 	}
	}
}
