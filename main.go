package main

import (
	"fmt"
	"math"
	"time"
)

// Loan represents the loan details.
type Loan struct {
	Principal     float64   // 初始本金
	DefaultLPR    float64   // 默认利率
	PlusSpread    float64   // 加点
	TermInMonths  int       // 贷款期限（月）
	StartDate     time.Time // 放款年月日
	LPRS          []LPR     // 日期与利率的条目列表
	PaymentDueDay int       // 还款日 (1-31)
}

// LPR represents the date and interest LPR entry.
type LPR struct {
	Date time.Time // 日期
	LPR  float64   // 利率
}

// Payment represents the details of each monthly payment.
type Payment struct {
	LoanTerm           int       // 期数
	Principal          float64   // 本金部分（固定为每月还款金额）
	Interest           float64   // 利息部分
	MonthTotalAmount   float64   // 当月还款总金额
	RemainingPrincipal float64   // 剩余本金
	TotalInterestPaid  float64   // 已支付总利息
	DueDateRate        float64   // 当月利率=lpr+加点
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

func daysUntilLastMonthSameDay(date time.Time) int {
	lastMonth := date.AddDate(0, -1, 0)
	// 计算两个日期之间的天数差异
	daysDiff := int(date.Sub(lastMonth).Hours() / 24)
	return daysDiff
}

// 取两位小数
func roundToTwoDecimalPlaces(value float64) float64 {
	return math.Round(value*100) / 100
}

// 计算LPR变更当月执行不同利率的天数
func (loan *Loan) calculateLprPeriod(nextDueDate time.Time) (periodDays, daysBefore, daysAfter int) {
	// var daysBefore ,daysAfter int
	// 计算利率周期的天数
	periodDays = daysUntilLastMonthSameDay(nextDueDate)

	if loan.StartDate.Day() < loan.PaymentDueDay {
		// 利率变更日前的天数
		daysBefore = loan.PaymentDueDay - loan.StartDate.Day()
	} else {
		daysBefore = loan.StartDate.Day() - loan.PaymentDueDay
	}

	// 利率变更日后的天数
	daysAfter = periodDays - daysBefore

	// fmt.Println(nextDueDate, periodDays, daysBefore, daysAfter)
	return
}

// Calculate the due date based on the start date and month.
func (loan *Loan) nextDueDate(startDate time.Time, paymentDueDay int) time.Time {
	// 创建一个新日期，月份和年份会随着时间增加，日期为贷款合同中指定的还款日
	nextDueDate := time.Date(startDate.Year(), startDate.Month()+1, paymentDueDay, 0, 0, 0, 0, startDate.Location())
	return nextDueDate
}

// 获取离指定日期最近的LPR
func (loan *Loan) getClosestLPRForYear(lprDate time.Time) float64 {
	// 找到小于或等于给定年份的最近的RateEntry日期
	var selectedRate float64
	for _, entry := range loan.LPRS {
		// 如果找到比 lprDate 之前的日期，且比当前选定的日期更接近，则更新 selectedRate 和 lprDate
		if entry.Date.Before(lprDate) && (selectedRate == 0 || lprDate.Sub(entry.Date) < lprDate.Sub(lprDate)) {
			selectedRate = entry.LPR
			lprDate = entry.Date
		}
	}
	return selectedRate
}

// CalculateAmortizationSchedule calculates the amortization schedule for the loan.
func (loan *Loan) CalculateAmortizationSchedule() []Payment {
	var lprDate time.Time // 计算lpr的日期
	var interestPayment float64
	var currentYearRate, currentYearLPR float64
	var previousYearRate, previousYearLPR float64
	var lprChangeMonth time.Month

	principalPayment := roundToTwoDecimalPlaces(loan.Principal / float64(loan.TermInMonths))
	payments := make([]Payment, 0)
	remainingPrincipal := loan.Principal
	totalInterestPaid := 0.0
	nextDueDate := loan.nextDueDate(loan.StartDate, loan.PaymentDueDay)

	if loan.StartDate.Day() < loan.PaymentDueDay {
		lprChangeMonth = loan.StartDate.Month()
	} else {
		lprChangeMonth = loan.StartDate.Month() + 1
	}

	for loanTerm := 1; loanTerm <= loan.TermInMonths; loanTerm++ {
		// 提取startDate和nextDueDate的月份和日期
		nextDueYear := nextDueDate.Year()
		nextDueMonth := nextDueDate.Month()
		nextDueDay := nextDueDate.Day()

		startMonth := loan.StartDate.Month()
		startDay := loan.StartDate.Day()

		// 获取LPR规则如下
		// 1.如果日期在变更日之前则取离上一年变更日最近的LPR
		// 2.如果日期在变更日或者之后，则取离当年变更日最近的LPR
		if nextDueMonth < startMonth || (nextDueMonth == startMonth && nextDueDay < startDay) {
			lprDate = time.Date(nextDueYear-1, startMonth, startDay, 0, 0, 0, 0, loan.StartDate.Location())
		} else {
			lprDate = time.Date(nextDueYear, startMonth, startDay, 0, 0, 0, 0, loan.StartDate.Location())
		}

		// // 第一个月的利息根据天数计算
		if loanTerm == 1 {
			// 2968.28
			// 利率周期2022-05-18 ~ 2022-06-17 共30天
			// 实际天数2022-05-26 ~ 2022-06-17 共23天
			days := int(nextDueDate.Sub(loan.StartDate).Hours() / 24)
			currentYearLPR = loan.getClosestLPRForYear(lprDate)
			currentYearRate = currentYearLPR + loan.PlusSpread
			interestPayment = remainingPrincipal * currentYearRate / 12 / 100 * (float64(days-1) / 30)
		} else if nextDueMonth == lprChangeMonth {
			// 利率变更月特殊处理 3657.38
			// 分为两段
			// 上一年利率 2023-05-18 ~ 2023-05-24
			periodDays, daysBefore, daysAfter := loan.calculateLprPeriod(nextDueDate)
			lprDate = time.Date(nextDueYear-1, startMonth, startDay, 0, 0, 0, 0, loan.StartDate.Location())
			previousYearLPR = loan.getClosestLPRForYear(lprDate)
			previousYearRate = previousYearLPR + loan.PlusSpread
			interestPayment = roundToTwoDecimalPlaces(remainingPrincipal * previousYearRate / 100 / 12 * float64(daysBefore) / float64(periodDays))

			// 当年利率 2023-05-25 ~ 2023-06-18
			lprDate = time.Date(nextDueYear, startMonth, startDay, 0, 0, 0, 0, loan.StartDate.Location())
			currentYearLPR = loan.getClosestLPRForYear(lprDate)
			currentYearRate = currentYearLPR + loan.PlusSpread
			interestPayment += roundToTwoDecimalPlaces(remainingPrincipal * currentYearRate / 100 / 12 * float64(daysAfter) / float64(periodDays))

		} else {
			currentYearLPR = loan.getClosestLPRForYear(lprDate)
			currentYearRate = currentYearLPR + loan.PlusSpread
			// daysInMonth = daysUntilLastMonthSameDay(nextDueDate)
			interestPayment = roundToTwoDecimalPlaces(remainingPrincipal * currentYearRate / 100 / 12)

		}

		remainingPrincipal -= principalPayment
		totalInterestPaid += interestPayment

		payment := Payment{
			LoanTerm:           loanTerm,
			Principal:          principalPayment,
			Interest:           interestPayment,
			MonthTotalAmount:   principalPayment + interestPayment,
			RemainingPrincipal: remainingPrincipal,
			TotalInterestPaid:  totalInterestPaid,
			DueDateRate:        currentYearRate,
			DueDate:            nextDueDate,
		}
		payments = append(payments, payment)

		// 计算下一个还款日期
		nextDueDate = nextDueDate.AddDate(0, 1, 0)

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
	paymentDueDay := 18                  // 还款日
	// 创建 Loan 结构
	loan := Loan{
		Principal:     initialPrincipal,
		DefaultLPR:    defaultLPR,
		TermInMonths:  loanTerm,
		StartDate:     startDate,
		LPRS:          lprs,
		PlusSpread:    plusSpread,
		PaymentDueDay: paymentDueDay,
	}

	// 计算等额本金还款计划
	payments := loan.CalculateAmortizationSchedule()

	// 输出更详细的还款计划
	fmt.Println("期数\t还款日期\t本金\t利息\t本月还款\t剩余本金\t已支付总利息\t本月利率")
	for _, payment := range payments {
		fmt.Printf("%d\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f%%\n", payment.LoanTerm, payment.DueDate.Format("2006-01-02"), payment.Principal, payment.Interest, payment.MonthTotalAmount, payment.RemainingPrincipal, payment.TotalInterestPaid, payment.DueDateRate)
		// // 只打印前10行记录
		// if payment.LoanTerm <= 20 {
		// 	fmt.Printf("%d\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f\t%.2f%%\n", payment.LoanTerm, payment.DueDate.Format("2006-01-02"), payment.Principal, payment.Interest, payment.MonthTotalAmount, payment.RemainingPrincipal, payment.TotalInterestPaid, payment.DueDateRate)
		// } else {
		// 	break // 如果已经打印了前10行，就退出循环
		// }
	}
}
