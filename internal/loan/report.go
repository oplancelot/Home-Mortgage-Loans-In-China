package loan

import (
	"bytes"
	"sort"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/shopspring/decimal"
)

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

func printReport(reports []Report) string {

	// 创建一个 buffer 用于保存表格内容
	var buffer bytes.Buffer

	// 创建 tablewriter 实例
	table := tablewriter.NewWriter(&buffer)
	// 设置表格内容，可以调用 table.SetHeader()、table.Append() 等方法
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
	// 渲染表格到 buffer 中
	table.Render()

	// 将 buffer 中的内容转换为字符串并返回
	return buffer.String()
}

func LoanPrintReport(initialPrincipal float64, loanTerm int, startDate string, plusSpread float64, paymentDueDay int, earlyRepaymentsAmount []float64, earlyRepaymentsDate []string) string {

	// 创建 Loan 结构
	loan := Loan{
		InitialPrincipal: decimal.NewFromFloat(initialPrincipal),
		InitialLPR:       decimal.NewFromFloat(4.45),
		InitialTerm:     loanTerm,
		InitialDate:      parseDate(startDate),
		LPR:              Lprs, // 常量
		PlusSpread:       decimal.NewFromFloat(plusSpread),
		PaymentDueDay:    paymentDueDay,
	}

	// 输入提前还款信息
	// earlyRepayments := []EarlyRepayment{
	// 	{Amount: decimal.NewFromFloat(200000), Date: parseDate("2023-08-19")},
	// }
	earlyRepayments := make([]EarlyRepayment, len(earlyRepaymentsAmount))
	for i := range earlyRepaymentsAmount {
		earlyRepayments[i] = EarlyRepayment{
			Amount: decimal.NewFromFloat(earlyRepaymentsAmount[i]),
			Date:   parseDate(earlyRepaymentsDate[i]),
		}
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
	p := printReport(report)

	return p
}
