package loan

import (
	"sort"
	"strconv"
	"strings"
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
			TotalInterestPaid:  decimal.Zero,
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

func CalculateTotalInterest(reports []Report) {
	for i := 1; i < len(reports); i++ {

		reports[i].TotalInterestPaid = reports[i-1].TotalInterestPaid.Add(reports[i].Interest)

	}
}

func Report2table(reports []Report) string {

	// 创建 tablewriter 实例
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

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
	// 设置表格内容，可以调用 table.SetHeader()、table.Append() 等方法
	table.SetHeader([]string{"序号", "期数", "明细", "日期", "本金", "利息", "本月还款   ", "剩余本金", "利息合计", "APR"})
	table.SetAutoWrapText(true)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("") // 表头和内容之间的分隔符
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetTablePadding("\t") // 列与列之间的分隔符
	table.SetNoWhiteSpace(true)
	table.SetColMinWidth(2, 10) // 设置列index的最小宽度,0表示第一列 和autoWrapText共同作用
	table.SetColMinWidth(4, 12)
	table.SetColMinWidth(5, 12)
	table.Render()

	// 输出为字符串
	return tableString.String()
}

type Input struct {
	Loan           Loan
	EarlyRepayment []EarlyRepayment
}

func LoanPrintTable(inputdata Input, action string) string {
	var payments []MonthlyPayment

	switch action {
	case "epp":
		// 计算等额本金还款计划
		payments = inputdata.Loan.EqualPrincipalPayment(inputdata.EarlyRepayment)
	default:
		// 计算等额本息还款计划
		payments = inputdata.Loan.EqualMonthlyInstallment(inputdata.EarlyRepayment)

	}

	// 整理数据
	report := []Report{}
	report = loan2Report(inputdata.Loan, report)
	report = monthlyPayment2Report(payments, report)
	report = earlyRepayment2Report(inputdata.EarlyRepayment, report)
	sortReport(report)
	CalculateTotalInterest(report)

	// printReport(report)
	p := Report2table(report)

	return p
}
