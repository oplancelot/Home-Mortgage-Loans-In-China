package main

import (
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/shopspring/decimal"
)

// Loan represents the loan details.

type Loan struct {
	InitialPrincipal decimal.Decimal // 初始本金
	InitialLPR       decimal.Decimal // 默认利率
	PlusSpread       decimal.Decimal // 加点
	InitialTermI     int             // 贷款期限（月）
	InitialDate      time.Time       // 放款年月日
	LPR              []LPR           // 日期与利率的条目列表
	PaymentDueDay    int             // 还款日 (1-31)
}

// LPR represents the date and interest LPR entry.
type LPR struct {
	Date time.Time       // 日期
	LPR  decimal.Decimal // 利率
}

// Payment represents the details of each monthly payment.
type MonthlyPayment struct {
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
	Amount             decimal.Decimal // 提前还款金额
	Date               time.Time       // 提前还款日期
	DueDateRate        decimal.Decimal // 当月利率
	Principal          decimal.Decimal // 本金部分
	Interest           decimal.Decimal // 利息部分
	RemainingPrincipal decimal.Decimal // 剩余本金
}

// 创建一个新的结构体，包含 payments、loan 和 earlyRepayments 的字段
type Report struct {
	Index              int
	LoanTerm           int             // 期数
	Purpose            string          // 明细性质
	Principal          decimal.Decimal // 本金部分
	Interest           decimal.Decimal // 利息部分
	MonthTotalAmount   decimal.Decimal // 当月还款总金额
	RemainingPrincipal decimal.Decimal // 剩余本金
	TotalInterestPaid  decimal.Decimal // 已支付总利息
	DueDateRate        decimal.Decimal // 当月利率=lpr+加点
	DueDate            time.Time       // 当月还款日期
}

func loan2Report(loan Loan, report []Report) []Report {
	newReport := make([]Report, len(report)+1)
	copy(newReport, report)
	newReport[len(report)] = Report{
		Index:              0,
		LoanTerm:           0,
		Purpose:            "贷款发放",
		Principal:          loan.InitialPrincipal,
		Interest:           decimal.Zero,
		MonthTotalAmount:   decimal.Zero,
		RemainingPrincipal: loan.InitialPrincipal,
		TotalInterestPaid:  decimal.Zero,
		DueDateRate:        loan.getClosestLPRForYear(loan.InitialDate).Add(loan.PlusSpread),
		DueDate:            loan.InitialDate,
	}
	return newReport
}
func earlyRepayment2Report(earlyRepayments []EarlyRepayment, report []Report) []Report {
	newReport := make([]Report, len(report)+len(earlyRepayments))
	copy(newReport, report)
	for i, early := range earlyRepayments {
		newReport[len(report)+i] = Report{
			Index:              i,
			Purpose:            "提前还款",
			Principal:          early.Principal,
			Interest:           early.Interest,
			MonthTotalAmount:   early.Principal.Add(early.Interest),
			RemainingPrincipal: early.RemainingPrincipal,
			TotalInterestPaid:  decimal.Zero,
			DueDateRate:        early.DueDateRate,
			DueDate:            early.Date,
		}
		// fmt.Println(early.Date, early.Amount, early.Principal)
	}
	return newReport
}

func monthlyPayment2Report(payments []MonthlyPayment, report []Report) []Report {
	newReport := make([]Report, len(report)+len(payments))
	copy(newReport, report)
	for i, payment := range payments {
		newReport[len(report)+i] = Report{
			Index:              i,
			LoanTerm:           payment.LoanTerm,
			Purpose:            "分期",
			Principal:          payment.Principal,
			Interest:           payment.Interest,
			MonthTotalAmount:   payment.MonthTotalAmount,
			RemainingPrincipal: payment.RemainingPrincipal,
			TotalInterestPaid:  payment.TotalInterestPaid,
			DueDateRate:        payment.DueDateRate,
			DueDate:            payment.DueDate,
		}

	}
	return newReport
}
func sortReport(reports []Report) {
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].DueDate.Before(reports[j].DueDate)
	})

	for i := range reports {
		reports[i].Index = i
	}

}

func updateReport(reports []Report) {
	for i := 1; i < len(reports); i++ {
		reports[i].RemainingPrincipal = reports[i-1].RemainingPrincipal.Sub(reports[i].Principal)
		reports[i].TotalInterestPaid = reports[i-1].TotalInterestPaid.Add(reports[i].Interest)

	}
}

func printReport(reports []Report) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"序号", "期数", "明细", "日期", "本金", "利息", "本月还款", "剩余本金", "已支付总利息", "本月利率"})

	for _, row := range reports {
		table.Append([]string{
			strconv.Itoa(row.Index),
			strconv.Itoa(row.LoanTerm),
			row.Purpose,
			row.DueDate.Format("2006-01-02"),
			row.Principal.String(),
			row.Interest.String(),
			row.MonthTotalAmount.String(),
			row.RemainingPrincipal.String(),
			row.TotalInterestPaid.String(),
			row.DueDateRate.String(),
		})
	}

	table.Render()
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

func (loan *Loan) previousDueDate(dueDate time.Time) time.Time {
	previousDueDate := dueDate.AddDate(0, -1, 0)
	return previousDueDate
}

func (loan *Loan) daysDiff(startDate, endDate time.Time) decimal.Decimal {
	// Calculate the difference in days
	daysDiff := int(endDate.Sub(startDate).Hours() / 24)
	return decimal.NewFromInt(int64(daysDiff))
}

func (loan *Loan) currentYearLPRUpdate(dueDate time.Time) (decimal.Decimal, decimal.Decimal) {
	currentYearLPRUpdateDate := time.Date(dueDate.Year(), loan.InitialDate.Month(), loan.InitialDate.Day(), 0, 0, 0, 0, loan.InitialDate.Location())
	previousDueDate := loan.previousDueDate(dueDate)
	// lpr变更前的天数
	daysBefore := loan.daysDiff(previousDueDate, currentYearLPRUpdateDate)
	// fmt.Println(daysBefore)
	daysAfter := decimal.NewFromInt(30).Sub(daysBefore)
	return daysBefore, daysAfter
}

// 获取离指定日期最近的LPR

func (loan *Loan) getClosestLPRForYear(dueDate time.Time) (selectedRate decimal.Decimal) {
	var lprUpdateDate time.Time

	// 获取LPR规则如下
	// 1.如果日期在变更日之前则取离上一年变更日最近的LPR
	// 2.如果日期在变更日或者之后，则取离当年变更日最近的LPR
	if dueDate.Month() < loan.InitialDate.Month() || (dueDate.Month() == loan.InitialDate.Month() && dueDate.Day() < loan.InitialDate.Day()) {
		lprUpdateDate = time.Date(dueDate.Year()-1, loan.InitialDate.Month(), loan.InitialDate.Day(), 0, 0, 0, 0, loan.InitialDate.Location())
	} else {
		lprUpdateDate = time.Date(dueDate.Year(), loan.InitialDate.Month(), loan.InitialDate.Day(), 0, 0, 0, 0, loan.InitialDate.Location())
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

func (loan *Loan) makeEarlyRepayment(remainingPrincipal decimal.Decimal, earlyRepayments []EarlyRepayment, dueDate time.Time) (decimal.Decimal, decimal.Decimal) {
	previousDueDate := loan.previousDueDate(dueDate)
	for i, early := range earlyRepayments {
		if early.Date.After(previousDueDate) && early.Date.Before(dueDate) {
			currentYearLPR := loan.getClosestLPRForYear(dueDate)
			currentYearRate := currentYearLPR.Add(loan.PlusSpread).Div(decimal.NewFromInt(100))
			daysDiff := loan.daysDiff(previousDueDate, early.Date)
			earlyInterest := remainingPrincipal.Mul(currentYearRate).Div(decimal.NewFromInt(360)).Mul(daysDiff)
			amount := remainingPrincipal.Add(earlyInterest).Sub(early.Amount)
			// fmt.Println(dueDate, currentYearRate, earlyInterest, remainingPrincipal, daysDiff)
			// 更新本金利息和利率
			early.Principal = early.Amount.Sub(earlyInterest).Round(2)
			early.Interest = earlyInterest.Round(2)
			early.RemainingPrincipal = remainingPrincipal.Round(2)
			early.DueDateRate = currentYearRate.Mul(decimal.NewFromInt(100))
			// 将更新后的 early 对象存储回 earlyRepayments 切片中
			earlyRepayments[i] = early

			return amount, daysDiff
		}
	}
	return remainingPrincipal, decimal.Decimal{}
}

func (loan *Loan) EqualPrincipalPaymentPlan(earlyRepayment []EarlyRepayment) []MonthlyPayment {
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
	principalPayment := remainingPrincipal.Div(decimal.NewFromInt(int64(loan.InitialTermI))).Round(2)
	lastPrincipalPayment := principalPayment.Add(remainingPrincipal.Sub(principalPayment.Mul(decimal.NewFromInt(int64(loan.InitialTermI)))))
	dueDate := time.Date(loan.InitialDate.Year(), loan.InitialDate.Month()+1, loan.PaymentDueDay, 0, 0, 0, 0, loan.InitialDate.Location())

	for loanTerm := 1; loanTerm <= loan.InitialTermI; loanTerm++ {

		// 计算如果有提前还款则需要减去提前还款的本金
		amount, daysDiff := loan.makeEarlyRepayment(remainingPrincipal, earlyRepayment, dueDate)
		// 只在提前还款后,重新计算每月应还本金;否则多次计算会有小数点导致的差异
		if amount.Cmp(remainingPrincipal) == -1 {
			remainTerm := loan.InitialTermI - loanTerm + 1
			principalPayment = amount.Div(decimal.NewFromInt(int64(remainTerm))).Round(2)
			lastPrincipalPayment = principalPayment.Add(amount.Sub(principalPayment.Mul(decimal.NewFromInt(int64(remainTerm)))))
		}
		// 以下处理每月正常还款
		// 计算利率利息
		interestPayment := decimal.Zero
		currentYearRate := loan.getClosestLPRForYear(dueDate).Add(loan.PlusSpread)
		days := decimal.NewFromInt(int64(30))

		var repaymentStatus string
		if loanTerm == 1 { // 第一期
			repaymentStatus = "A"
		} else if loanTerm%12 == 1 { // lpr变更月
			repaymentStatus = "B"
		} else if loanTerm == loan.InitialTermI { // 最后一期
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
			// fmt.Println(days, loan.previousDueDate(dueDate), dueDate)
			principalPayment = lastPrincipalPayment.Round(2)
			dueDate = lastDueDate
			interestPayment = lastPrincipalPayment.Mul(currentYearRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360)).Mul(days).Round(2)

		default:
			remainDay := days.Sub(daysDiff)
			interestPayment = remainingPrincipal.Mul(currentYearRate).Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(360)).Mul(remainDay).Round(2)

		}

		remainingPrincipal = remainingPrincipal.Sub(principalPayment).Round(2)

		totalInterestPaid := decimal.Zero
		totalInterestPaid = totalInterestPaid.Add(interestPayment).Round(2)

		payment := MonthlyPayment{
			LoanTerm:           loanTerm,
			Principal:          principalPayment,
			Interest:           interestPayment,
			MonthTotalAmount:   principalPayment.Add(interestPayment),
			RemainingPrincipal: remainingPrincipal,
			TotalInterestPaid:  totalInterestPaid,
			DueDateRate:        currentYearRate,
			DueDate:            dueDate,
		}
		monthlypayments = append(monthlypayments, payment)

		// 下一个还款日期
		dueDate = dueDate.AddDate(0, 1, 0)

	}

	return monthlypayments
}

func main() {
	// LPR
	lpr := []LPR{
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
		InitialPrincipal: initialPrincipal,
		InitialLPR:       defaultLPR,
		InitialTermI:     loanTerm,
		InitialDate:      startDate,
		LPR:              lpr,
		PlusSpread:       plusSpread,
		PaymentDueDay:    paymentDueDay,
	}

	// 输入提前还款信息
	earlyRepayments := []EarlyRepayment{
		{Amount: decimal.NewFromFloat(200000), Date: parseDate("2023-08-19")},
	}

	// 计算等额本金还款计划
	payments := loan.EqualPrincipalPaymentPlan(earlyRepayments)

	// 整理数据
	report := []Report{}
	report = loan2Report(loan, report)
	report = monthlyPayment2Report(payments, report)
	report = earlyRepayment2Report(earlyRepayments, report)
	sortReport(report)
	updateReport(report)

	// printReport(report)
	printReport(report)
}
