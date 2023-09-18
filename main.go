package main

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// NewFromFloat(5.45).StringFixed(2) // output: "5.45"
// Loan represents the loan details.
// 贷款输出项2位小数,利率6位小数(因为%是2位小数)

type Loan struct {
	Principal     decimal.Decimal // 初始本金
	DefaultLPR    decimal.Decimal // 默认利率
	PlusSpread    decimal.Decimal // 加点
	TermInMonths  int             // 贷款期限（月）
	StartDate     time.Time       // 放款年月日
	LPR           []LPR           // 日期与利率的条目列表
	PaymentDueDay int             // 还款日 (1-31)
}

// LPR represents the date and interest LPR entry.
type LPR struct {
	Date time.Time       // 日期
	LPR  decimal.Decimal // 利率
}

// Payment represents the details of each monthly payment.
type Payment struct {
	LoanTerm           int             // 期数
	Principal          decimal.Decimal // 本金部分（固定为每月还款金额）
	Interest           decimal.Decimal // 利息部分
	MonthTotalAmount   decimal.Decimal // 当月还款总金额
	RemainingPrincipal decimal.Decimal // 剩余本金
	TotalInterestPaid  decimal.Decimal // 已支付总利息
	DueDateRate        decimal.Decimal // 当月利率=lpr+加点
	DueDate            time.Time       // 当月还款日期
}

// 提前还款
type EarlyRepayment struct {
	Amount decimal.Decimal // 提前还款金额
	Date   time.Time       // 提前还款日期
}

// PaymentWithIndex
type PaymentWithIndex struct {
	Index   int
	Payment Payment
}

// parseDate 解析日期字符串并返回时间。如果出现错误，将返回一个零值时间。
func parseDate(dateString string) time.Time {
	layout := "2006-01-02" // 统一的日期布局字符串
	parsedTime, err := time.Parse(layout, dateString)
	if err != nil {
		// 返回零值时间
		return time.Time{}
	}
	return parsedTime
}

// // 取两位小数
// func round4(value float64) float64 {
// 	return math.Round(value*10000) / 10000
// }

// func round2(value float64) float64 {
// 	return math.Round(value*100) / 100
// }

// 计算LPR变更当月执行不同利率的天数
// 放款日	6.14		5.25		6.20
// 还款日	18			18			18
// 变更月	6			6			7
// 第一段	5.18~6.13	5.18~5.24	6.18~6.19
// 第二段	6.14~6.17	5.25~6.17	6.20~7.17

func previousDueDate(dueDate time.Time) time.Time {
	previousMonth := dueDate.AddDate(0, -1, 0)
	previousDueDate := time.Date(previousMonth.Year(), previousMonth.Month(), dueDate.Day(), 0, 0, 0, 0, dueDate.Location())
	return previousDueDate
}

func daysDiff(startDate, endDate time.Time) decimal.Decimal {
	// Calculate the difference in days
	daysDiff := int(endDate.Sub(startDate).Hours() / 24)
	return decimal.NewFromInt(int64(daysDiff))
}

func (loan *Loan) currentYearLPRUpdate(dueDate time.Time) (decimal.Decimal, decimal.Decimal) {
	dueYear := dueDate.Year()
	startMonth := loan.StartDate.Month()
	startDay := loan.StartDate.Day()

	currentYearLPRUpdateDate := time.Date(dueYear, startMonth, startDay, 0, 0, 0, 0, loan.StartDate.Location())
	previousDueDate := previousDueDate(dueDate)

	// lpr变更前的天数
	daysBefore := daysDiff(previousDueDate, currentYearLPRUpdateDate)
	// fmt.Println(daysBefore)
	daysAfter := decimal.NewFromInt(30).Sub(daysBefore)
	return daysBefore, daysAfter
}

// Calculate the due date based on the start date and month.
func (loan *Loan) dueDate(startDate time.Time, paymentDueDay int) time.Time {
	// 创建一个新日期，月份和年份会随着时间增加，日期为贷款合同中指定的还款日
	dueDate := time.Date(startDate.Year(), startDate.Month()+1, paymentDueDay, 0, 0, 0, 0, startDate.Location())
	return dueDate
}

// 获取离指定日期最近的LPR
func (loan *Loan) getClosestLPRForYear(dueDate time.Time) decimal.Decimal {
	var lprUpdateDate time.Time
	var selectedRate decimal.Decimal

	// 获取LPR规则如下
	// 1.如果日期在变更日之前则取离上一年变更日最近的LPR
	// 2.如果日期在变更日或者之后，则取离当年变更日最近的LPR
	dueYear := dueDate.Year()
	dueMonth := dueDate.Month()
	dueDay := dueDate.Day()
	startMonth := loan.StartDate.Month()
	startDay := loan.StartDate.Day()

	if dueMonth < startMonth || (dueMonth == startMonth && dueDay < startDay) {
		lprUpdateDate = time.Date(dueYear-1, startMonth, startDay, 0, 0, 0, 0, loan.StartDate.Location())
	} else {
		lprUpdateDate = time.Date(dueYear, startMonth, startDay, 0, 0, 0, 0, loan.StartDate.Location())
	}
	// 找到小于或等于给定年份的最近的RateEntry日期
	for _, entry := range loan.LPR {
		// 如果找到比 lprUpdateDate 之前的日期，且比当前选定的日期更接近，则更新 selectedRate 和 lprUpdateDate
		if entry.Date.Before(lprUpdateDate) && (selectedRate == decimal.Decimal{} || lprUpdateDate.Sub(entry.Date) < lprUpdateDate.Sub(lprUpdateDate)) {
			selectedRate = entry.LPR
			lprUpdateDate = entry.Date
		}
	}
	return selectedRate
}

// Add contents of payments to IPaymentWithIndex slice
func paymentsWithIndex(payments []Payment) []PaymentWithIndex {
	ipayments := make([]PaymentWithIndex, len(payments))
	for i, payment := range payments {
		ipayment := PaymentWithIndex{
			Index:   i,
			Payment: payment,
		}
		ipayments[i] = ipayment
	}
	return ipayments
}

func (loan *Loan) makeEarlyRepayment(remainingPrincipal decimal.Decimal, earlyRepayments []EarlyRepayment, dueDate time.Time) (decimal.Decimal, decimal.Decimal) {
	previousDueDate := previousDueDate(dueDate)
	for _, early := range earlyRepayments {
		if early.Date.After(previousDueDate) && early.Date.Before(dueDate) {
			currentYearLPR := loan.getClosestLPRForYear(dueDate)
			currentYearRate := currentYearLPR.Add(loan.PlusSpread).Div(decimal.NewFromInt(100))
			daysDiff := daysDiff(previousDueDate, early.Date)
			earlyInterest := remainingPrincipal.Mul(currentYearRate).Div(decimal.NewFromInt(360)).Mul(daysDiff)
			remainingPrincipal = remainingPrincipal.Add(earlyInterest).Sub(early.Amount)
			// fmt.Println(dueDate, currentYearRate, earlyInterest, remainingPrincipal, daysDiff)
			return remainingPrincipal, daysDiff
		}
	}
	return remainingPrincipal, decimal.Decimal{}
}

// 计算利息的规则
// 1.天数:全年360天,12个月每月30天

// 2.年利率=(lpr+加点)/100
//	每日利率=剩余本金*年利率/360

// 3.每日利息=剩余本金*每日利率
//	还款日还的利息=上个月剩余本金*每日利息*30

// 4.如果是lpr变更的月份,分为两段计算.
//	第一段lpr为前一年lpr,天数是变更日~还款日(取头去尾)
//	第二段为当年lpr,天数是30-第一段

func (loan *Loan) loanRepaymentSchedule(earlyRepayment []EarlyRepayment) []Payment {
	// var lprUpdateDate time.Time // 计算lpr的日期
	var interestPayment decimal.Decimal
	var currentYearRate decimal.Decimal
	var totalInterestPaid decimal.Decimal
	var lprChangeMonth time.Month
	var days, earlydays decimal.Decimal
	payments := make([]Payment, 0)
	remainingPrincipal := loan.Principal
	dueDate := loan.dueDate(loan.StartDate, loan.PaymentDueDay)
	days = decimal.NewFromInt(int64(30))
	if loan.StartDate.Day() < loan.PaymentDueDay {
		lprChangeMonth = loan.StartDate.Month()
	} else {
		lprChangeMonth = loan.StartDate.Month() + 1
	}

	for loanTerm := 1; loanTerm <= loan.TermInMonths; loanTerm++ {
		// 提取startDate和nextDueDate的月份和日期
		// dueYear := dueDate.Year()
		dueMonth := dueDate.Month()
		// dueDay := dueDate.Day()

		// startMonth := loan.StartDate.Month()
		// startDay := loan.StartDate.Day()
		// previousDueDate := previousDueDate(dueDate)

		// 此处处理提前还款,在正常还款期前更新本金额
		/// 检查是否有提前还款需要处理

		remainingPrincipal, earlydays = loan.makeEarlyRepayment(remainingPrincipal, earlyRepayment, dueDate)
		// 9.18 2691.16
		// 更新每月还款本金
		// principalPayment := roundDecimalPlaces(remainingPrincipal / float64(loan.TermInMonths-loanTerm+1))
		// principalPayment := remainingPrincipal / float64(loan.TermInMonths-loanTerm+1)
		remainTerm := loan.TermInMonths - loanTerm + 1
		// principalPayment := remainingPrincipal.Div(loan.TermInMonths.Sub(decimal.NewFromInt(int64(loanTerm))).Add(decimal.NewFromInt(1)))
		principalPayment := remainingPrincipal.Div(decimal.NewFromInt(int64(remainTerm)))

		// 以下处理每月正常还款
		// 计算利息
		//  第一个月的利息根据天数计算
		if loanTerm == 1 {
			// 2968.28
			// 利率周期2022-05-18 ~ 2022-06-17 共30天
			// 实际天数2022-05-26 ~ 2022-06-17 共23天
			// days := int(dueDate.Sub(loan.StartDate).Hours() / 24)
			days := daysDiff(loan.StartDate, dueDate).Sub(decimal.NewFromInt(1))
			currentYearLPR := loan.getClosestLPRForYear(dueDate)
			currentYearRate = currentYearLPR.Add(loan.PlusSpread)
			// interestPayment = remainingPrincipal * currentYearRate / 100 / 360 * (float64(days))
			interestPayment = remainingPrincipal.Mul(currentYearRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360)).Mul(days)

		} else if dueMonth == lprChangeMonth {
			// 利率变更月特殊处理 3657.38
			// 分为两段
			// 上一年利率 2023-05-18 ~ 2023-05-24
			daysBefore, daysAfter := loan.currentYearLPRUpdate(dueDate)
			previousDueDate := previousDueDate(dueDate)
			previousYearLPR := loan.getClosestLPRForYear(previousDueDate)
			// previousYearRate := previousYearLPR + loan.PlusSpread
			previousYearRate := previousYearLPR.Add(loan.PlusSpread)
			// interestPayment = (remainingPrincipal * previousYearRate / 100 / 360) * float64(daysBefore)
			interestPayment = remainingPrincipal.Mul(previousYearRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360)).Mul(daysBefore)

			// 当年利率 2023-05-25 ~ 2023-06-18
			currentYearLPR := loan.getClosestLPRForYear(dueDate)
			// currentYearRate = currentYearLPR + loan.PlusSpread
			currentYearRate = currentYearLPR.Add(loan.PlusSpread)
			// interestPayment += (remainingPrincipal * currentYearRate / 100 / 360) * float64(daysAfter)
			interestPayment = interestPayment.Add((remainingPrincipal.Mul(currentYearRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360))).Mul(daysAfter))
		} else {

			currentYearLPR := loan.getClosestLPRForYear(dueDate)
			// currentYearRate = currentYearLPR + loan.PlusSpread
			currentYearRate = currentYearLPR.Add(loan.PlusSpread)
			// fmt.Println(edays)
			// interestPayment = round4(remainingPrincipal * currentYearRate / 100 / 360 * float64(days-earlydays))

			// interestPayment = remainingPrincipal * currentYearRate / 100 / 360 * (float64(days))
			remainDay := days.Sub(earlydays)
			interestPayment = remainingPrincipal.Mul(currentYearRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360)).Mul(remainDay)

		}

		remainingPrincipal = remainingPrincipal.Sub(principalPayment)
		// totalInterestPaid += interestPayment
		totalInterestPaid = totalInterestPaid.Add(interestPayment)

		payment := Payment{
			LoanTerm:           loanTerm,
			Principal:          principalPayment.Round(2),
			Interest:           interestPayment.Round(2),
			MonthTotalAmount:   principalPayment.Add(interestPayment).Round(2),
			RemainingPrincipal: remainingPrincipal.Round(2),
			TotalInterestPaid:  totalInterestPaid.Round(2),
			DueDateRate:        currentYearRate.Round(2),
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
		{parseDate("2023-08-21"), decimal.NewFromFloat(4.20)},
		{parseDate("2023-07-20"), decimal.NewFromFloat(4.20)},
		{parseDate("2023-06-20"), decimal.NewFromFloat(4.20)},
		{parseDate("2023-05-22"), decimal.NewFromFloat(4.30)},
		{parseDate("2023-04-20"), decimal.NewFromFloat(4.30)},
		{parseDate("2023-03-20"), decimal.NewFromFloat(4.30)},
		{parseDate("2023-02-20"), decimal.NewFromFloat(4.30)},
		{parseDate("2023-01-20"), decimal.NewFromFloat(4.30)},
		{parseDate("2022-12-20"), decimal.NewFromFloat(4.30)},
		{parseDate("2022-11-21"), decimal.NewFromFloat(4.30)},
		{parseDate("2022-10-20"), decimal.NewFromFloat(4.30)},
		{parseDate("2022-09-20"), decimal.NewFromFloat(4.30)},
		{parseDate("2022-08-22"), decimal.NewFromFloat(4.30)},
		{parseDate("2022-07-20"), decimal.NewFromFloat(4.45)},
		{parseDate("2022-06-20"), decimal.NewFromFloat(4.45)},
		{parseDate("2022-05-20"), decimal.NewFromFloat(4.45)},
		{parseDate("2022-04-20"), decimal.NewFromFloat(4.60)},
		{parseDate("2022-03-21"), decimal.NewFromFloat(4.60)},
		{parseDate("2022-02-21"), decimal.NewFromFloat(4.60)},
		{parseDate("2022-01-20"), decimal.NewFromFloat(4.60)},
		{parseDate("2021-12-20"), decimal.NewFromFloat(4.65)},
		{parseDate("2021-11-22"), decimal.NewFromFloat(4.65)},
		{parseDate("2021-10-20"), decimal.NewFromFloat(4.65)},
		{parseDate("2021-09-22"), decimal.NewFromFloat(4.65)},
		{parseDate("2021-08-20"), decimal.NewFromFloat(4.65)},
		{parseDate("2021-07-20"), decimal.NewFromFloat(4.65)},
		{parseDate("2021-06-21"), decimal.NewFromFloat(4.65)},
		{parseDate("2021-05-20"), decimal.NewFromFloat(4.65)},
		{parseDate("2021-04-20"), decimal.NewFromFloat(4.65)},
		{parseDate("2021-03-22"), decimal.NewFromFloat(4.65)},
		{parseDate("2021-02-22"), decimal.NewFromFloat(4.65)},
		{parseDate("2021-01-20"), decimal.NewFromFloat(4.65)},
		{parseDate("2020-12-21"), decimal.NewFromFloat(4.65)},
		{parseDate("2020-11-20"), decimal.NewFromFloat(4.65)},
		{parseDate("2020-10-20"), decimal.NewFromFloat(4.65)},
		{parseDate("2020-09-21"), decimal.NewFromFloat(4.65)},
		{parseDate("2020-08-20"), decimal.NewFromFloat(4.65)},
		{parseDate("2020-07-20"), decimal.NewFromFloat(4.65)},
		{parseDate("2020-06-22"), decimal.NewFromFloat(4.65)},
		{parseDate("2020-05-20"), decimal.NewFromFloat(4.65)},
		{parseDate("2020-04-20"), decimal.NewFromFloat(4.65)},
		{parseDate("2020-03-20"), decimal.NewFromFloat(4.75)},
		{parseDate("2020-02-20"), decimal.NewFromFloat(4.75)},
		{parseDate("2020-01-20"), decimal.NewFromFloat(4.80)},
		{parseDate("2019-12-20"), decimal.NewFromFloat(4.80)},
		{parseDate("2019-11-20"), decimal.NewFromFloat(4.80)},
		{parseDate("2019-10-21"), decimal.NewFromFloat(4.85)},
		{parseDate("2019-09-20"), decimal.NewFromFloat(4.85)},
		{parseDate("2019-08-20"), decimal.NewFromFloat(4.85)},
	}

	// 输入贷款信息
	initialPrincipal := decimal.NewFromFloat(920000.0) // 初始本金
	defaultLPR := decimal.NewFromFloat(8.05)           // 年利率（百分比）
	loanTerm := 360                                    // 贷款期限（月）
	startDate := parseDate("2022-05-25")               // 放款日期
	plusSpread := decimal.NewFromFloat(0.60)           // 上浮点数
	paymentDueDay := 18                                // 还款日
	// 创建 Loan 结构
	loan := Loan{
		Principal:     initialPrincipal,
		DefaultLPR:    defaultLPR,
		TermInMonths:  loanTerm,
		StartDate:     startDate,
		LPR:           lprs,
		PlusSpread:    plusSpread,
		PaymentDueDay: paymentDueDay,
	}

	// 输入提前还款信息
	earlyRepayments := []EarlyRepayment{
		{Amount: decimal.NewFromFloat(200000), Date: parseDate("2023-08-19")},
		{Amount: decimal.NewFromFloat(200000), Date: parseDate("2024-08-19")},
	}

	// 计算等额本金还款计划
	payments := loan.loanRepaymentSchedule(earlyRepayments)

	iPaymentWithIndex := paymentsWithIndex(payments)

	// 输出iPaymentWithIndex的所有内容
	fmt.Println("序号\t期数\t还款日期\t本金\t利息\t本月还款\t剩余本金\t已支付总利息\t本月利率")
	for _, ip := range iPaymentWithIndex {
		if ip.Index <= 40 {
			fmt.Println(ip.Index, ip.Payment.LoanTerm, ip.Payment.DueDate.Format("2006-01-02"), ip.Payment.Principal, ip.Payment.Interest, ip.Payment.MonthTotalAmount, ip.Payment.RemainingPrincipal, ip.Payment.TotalInterestPaid, ip.Payment.DueDateRate)
		}
	}

}
