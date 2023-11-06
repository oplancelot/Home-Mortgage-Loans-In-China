package loan

import (
	"time"

	"github.com/shopspring/decimal"
)

// Loan represents the loan details.

type Loan struct {
	InitialPrincipal decimal.Decimal // 初始本金
	PlusSpread       decimal.Decimal // 加点
	InitialTerm      int             // 贷款期限（月）
	InitialDate      time.Time       // 放款年月日
	LPR              []LPR           // 日期与利率的条目列表
	PaymentDueDay    int             // 还款日 (1-31)
}

// LPR represents the date and interest LPR entry.
type LPR struct {
	Date time.Time       // 日期
	LPR  decimal.Decimal // 利率
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

func (loan *Loan) LPRChangeDateOffset(dueDate time.Time) (decimal.Decimal, decimal.Decimal) {
	LPRUpdateDate := time.Date(dueDate.Year(), loan.InitialDate.Month(), loan.InitialDate.Day(), 0, 0, 0, 0, loan.InitialDate.Location())
	previousDueDate := loan.previousDueDate(dueDate)
	// lpr变更前的天数
	daysBefore := loan.daysDiff(previousDueDate, LPRUpdateDate)
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

var Lprs = []LPR{
	{ParseDate("2023-08-21"), decimal.NewFromFloat(4.20)},
	{ParseDate("2023-07-20"), decimal.NewFromFloat(4.20)},
	{ParseDate("2023-06-20"), decimal.NewFromFloat(4.20)},
	{ParseDate("2023-05-22"), decimal.NewFromFloat(4.30)},
	{ParseDate("2023-04-20"), decimal.NewFromFloat(4.30)},
	{ParseDate("2023-03-20"), decimal.NewFromFloat(4.30)},
	{ParseDate("2023-02-20"), decimal.NewFromFloat(4.30)},
	{ParseDate("2023-01-20"), decimal.NewFromFloat(4.30)},
	{ParseDate("2022-12-20"), decimal.NewFromFloat(4.30)},
	{ParseDate("2022-11-21"), decimal.NewFromFloat(4.30)},
	{ParseDate("2022-10-20"), decimal.NewFromFloat(4.30)},
	{ParseDate("2022-09-20"), decimal.NewFromFloat(4.30)},
	{ParseDate("2022-08-22"), decimal.NewFromFloat(4.30)},
	{ParseDate("2022-07-20"), decimal.NewFromFloat(4.45)},
	{ParseDate("2022-06-20"), decimal.NewFromFloat(4.45)},
	{ParseDate("2022-05-20"), decimal.NewFromFloat(4.45)},
	{ParseDate("2022-04-20"), decimal.NewFromFloat(4.60)},
	{ParseDate("2022-03-21"), decimal.NewFromFloat(4.60)},
	{ParseDate("2022-02-21"), decimal.NewFromFloat(4.60)},
	{ParseDate("2022-01-20"), decimal.NewFromFloat(4.60)},
	{ParseDate("2021-12-20"), decimal.NewFromFloat(4.65)},
	{ParseDate("2021-11-22"), decimal.NewFromFloat(4.65)},
	{ParseDate("2021-10-20"), decimal.NewFromFloat(4.65)},
	{ParseDate("2021-09-22"), decimal.NewFromFloat(4.65)},
	{ParseDate("2021-08-20"), decimal.NewFromFloat(4.65)},
	{ParseDate("2021-07-20"), decimal.NewFromFloat(4.65)},
	{ParseDate("2021-06-21"), decimal.NewFromFloat(4.65)},
	{ParseDate("2021-05-20"), decimal.NewFromFloat(4.65)},
	{ParseDate("2021-04-20"), decimal.NewFromFloat(4.65)},
	{ParseDate("2021-03-22"), decimal.NewFromFloat(4.65)},
	{ParseDate("2021-02-22"), decimal.NewFromFloat(4.65)},
	{ParseDate("2021-01-20"), decimal.NewFromFloat(4.65)},
	{ParseDate("2020-12-21"), decimal.NewFromFloat(4.65)},
	{ParseDate("2020-11-20"), decimal.NewFromFloat(4.65)},
	{ParseDate("2020-10-20"), decimal.NewFromFloat(4.65)},
	{ParseDate("2020-09-21"), decimal.NewFromFloat(4.65)},
	{ParseDate("2020-08-20"), decimal.NewFromFloat(4.65)},
	{ParseDate("2020-07-20"), decimal.NewFromFloat(4.65)},
	{ParseDate("2020-06-22"), decimal.NewFromFloat(4.65)},
	{ParseDate("2020-05-20"), decimal.NewFromFloat(4.65)},
	{ParseDate("2020-04-20"), decimal.NewFromFloat(4.65)},
	{ParseDate("2020-03-20"), decimal.NewFromFloat(4.75)},
	{ParseDate("2020-02-20"), decimal.NewFromFloat(4.75)},
	{ParseDate("2020-01-20"), decimal.NewFromFloat(4.80)},
	{ParseDate("2019-12-20"), decimal.NewFromFloat(4.80)},
	{ParseDate("2019-11-20"), decimal.NewFromFloat(4.80)},
	{ParseDate("2019-10-21"), decimal.NewFromFloat(4.85)},
	{ParseDate("2019-09-20"), decimal.NewFromFloat(4.85)},
	{ParseDate("2019-08-20"), decimal.NewFromFloat(4.85)},
}
