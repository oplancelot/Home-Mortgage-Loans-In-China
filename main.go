package main

import (
	"fmt"
	"math"
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
	LPR  float64   // 利率
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
func calculateInterestRatePeriod(startDate time.Time, dueDate time.Time) (float64, float64, float64) {
	// 计算利率周期的天数
	periodDays := dueDate.Sub(startDate).Hours() / 24

	// 利率变更日前的天数
	daysBefore := startDate.Sub(dueDate).Hours() / 24

	// 利率变更日后的天数
	daysAfter := periodDays - daysBefore

	return float64(periodDays), float64(daysBefore), float64(daysAfter)
}

// Calculate the due date based on the start date and month.
func dueDate(startDate time.Time, dueday int) time.Time {
	// 创建一个新日期，月份和年份会随着时间增加，日期为贷款合同中指定的还款日
	dueDate := time.Date(startDate.Year(), startDate.Month()+1, dueday, 0, 0, 0, 0, startDate.Location())
	return dueDate
}

// 获取离指定日期最近的LPR
func (loan *Loan) getClosestLPRForYear(LPRDate time.Time) float64 {

	// 找到小于或等于给定年份的最近的RateEntry日期
	var selectedRate float64

	for _, entry := range loan.LPRS {
		// 如果找到比 LPRDate 之前的日期，且比当前选定的日期更接近，则更新 selectedRate 和 LPRDate
		if entry.Date.Before(LPRDate) && (selectedRate == 0 || LPRDate.Sub(entry.Date) < LPRDate.Sub(LPRDate)) {
			selectedRate = entry.LPR
			LPRDate = entry.Date
		}
	}

	return selectedRate

}

// 取两位小数
func roundToTwoDecimalPlaces(value float64) float64 {
	return math.Round(value*100) / 100
}

// CalculateAmortizationSchedule calculates the amortization schedule for the loan.
func (loan *Loan) CalculateAmortizationSchedule() []Payment {
	var LPRDate time.Time
	var interestPayment float64
	var currentYearRate, currentYearLPR float64
	var changemonth int

	principalPayment := roundToTwoDecimalPlaces(loan.Principal / float64(loan.TermInMonths))
	payments := make([]Payment, 0)
	remainingPrincipal := loan.Principal
	totalInterestPaid := 0.0
	dueDate := dueDate(loan.StartDate, loan.Dueday)

	// 提取startDate和dueDate的月份和日期
	dueYear := dueDate.Year()
	dueMonth := dueDate.Month()
	dueDay := dueDate.Day()

	startMonth := loan.StartDate.Month()
	startDay := loan.StartDate.Day()

	if startDay < dueDay {
		changemonth = int(startMonth)
	} else {
		changemonth = int(startMonth) + 1
	}
	println(changemonth)

	for month := 1; month <= loan.TermInMonths; month++ {

		// var previousYearRate, previousYearLPR float64
		// var lastcompareDate time.Time

		// 3.如果是变更月，则需要返回上一年以及当年的LPR
		// 4.

		// 获取LPR规则如下
		// 1.如果日期在变更日之前则取离上一年变更日最近的LPR
		// 2.如果日期在变更日或者之后，则取离当年变更日最近的LPR
		if dueMonth < startMonth || (dueMonth == startMonth && dueDay < startDay) {
			LPRDate = time.Date(dueYear-1, startMonth, startDay, 0, 0, 0, 0, loan.StartDate.Location())
		} else {
			LPRDate = time.Date(dueYear, startMonth, startDay, 0, 0, 0, 0, loan.StartDate.Location())
		}

		// // 第一个月的利息根据天数计算
		if month == 1 {
			//2968.28
			// 利率周期2022-05-18 ~ 2022-06-17 共30天
			// 实际天数2022-05-26 ~ 2022-06-17 共23天

			days := int(dueDate.Sub(loan.StartDate).Hours() / 24)
			// fmt.Println(days)
			currentYearLPR = loan.getClosestLPRForYear(LPRDate)
			currentYearRate = currentYearLPR + loan.PlusSpread
			interestPayment = remainingPrincipal * currentYearRate / 12 / 100 * (float64(days-1) / 30)
		} else if month == changemonth {
			// 利率变更月特殊处理 3657.38
			// 分为两段
			// 上一年利率 2023-05-18 ~ 2023-06-24
			periodDays, daysBefore, daysAfter := calculateInterestRatePeriod(loan.StartDate, dueDate)
			LPRDate = time.Date(dueYear-1, startMonth, startDay, 0, 0, 0, 0, loan.StartDate.Location())
			// fmt.Println(periodDays, daysBefore, daysAfter, LPRDate)
			currentYearLPR = loan.getClosestLPRForYear(LPRDate)
			currentYearRate = currentYearLPR + loan.PlusSpread
			interestPayment = roundToTwoDecimalPlaces(remainingPrincipal * currentYearRate / 100 / 12 * daysBefore / periodDays)

			// 当年利率 2023-05-25 ~ 2023-06-18
			LPRDate = time.Date(dueYear, startMonth, startDay, 0, 0, 0, 0, loan.StartDate.Location())
			// fmt.Println(periodDays, daysBefore, daysAfter, LPRDate)

			currentYearLPR = loan.getClosestLPRForYear(LPRDate)
			currentYearRate = currentYearLPR + loan.PlusSpread
			interestPayment = interestPayment + roundToTwoDecimalPlaces(remainingPrincipal*currentYearRate/100/12*daysAfter/periodDays)

		} else {
			currentYearLPR = loan.getClosestLPRForYear(LPRDate)
			currentYearRate = currentYearLPR + loan.PlusSpread
			// daysInMonth = daysUntilLastMonthSameDay(dueDate)
			interestPayment = roundToTwoDecimalPlaces(remainingPrincipal * currentYearRate / 100 / 12)

		}

		// interestPayment = remainingPrincipal * (currentYearRate) / 12 / 100
		remainingPrincipal -= principalPayment
		totalInterestPaid += interestPayment

		payment := Payment{
			Month:              month,
			Principal:          principalPayment,
			Interest:           interestPayment,
			MonthTotalAmount:   principalPayment + interestPayment,
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
		// fmt.Printf("%d\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f%%\n", payment.Month, payment.DueDate.Format("2006-01-02"), payment.Principal, payment.Interest, payment.MonthTotalAmount, payment.RemainingPrincipal, payment.TotalInterestPaid, payment.DueDateRate)
		// 只打印前10行记录
		if payment.Month <= 20 {
			fmt.Printf("%d\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f%%\n", payment.Month, payment.DueDate.Format("2006-01-02"), payment.Principal, payment.Interest, payment.MonthTotalAmount, payment.RemainingPrincipal, payment.TotalInterestPaid, payment.DueDateRate)
		} else {
			break // 如果已经打印了前10行，就退出循环
		}
	}
}
