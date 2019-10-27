package main

import (
	"log"
	"math"
	"sort"
	"strconv"
	"time"
	"unicode"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

//Fees A collection of Fee
type Fees []Fee

//NewFees read fees from file
func NewFees(file *excelize.File) Fees {
	fees := Fees{}
	rows := file.GetRows("每月缴费数据")

	if len(rows) == 0 {
		log.Fatal("Can not find Sheet \"每月缴费数据\" in config.xlsx.")
	}

	for _, row := range rows {
		if unicode.IsDigit(rune(row[0][0])) {
			var fee Fee
			fee.Date = time.Date(toInt(row[0]), time.Month(toInt(row[1])), 1, 0, 0, 0, 0, time.UTC)
			payment, _ := strconv.ParseFloat(row[3], 64)
			institution, _ := strconv.ParseFloat(row[4], 64)
			base, _ := strconv.ParseFloat(row[2], 64)
			fee.Fee = base * payment * institution
			fees = append(fees, fee)
		}
	}

	return fees
}

//GetFeeForMonth get fee for a month
func (fees Fees) GetFeeForMonth(d time.Time) float64 {
	// detect tax or social insurance
	m := make(map[int]int, 0)
	for _, v := range fees {
		if i, ok := m[v.Date.Year()]; ok {
			m[v.Date.Year()] = i + 1
		} else {
			m[v.Date.Year()] = 1
		}
	}

	multiply := 1.00
	month := d.Month()

	// for social insurance 针对社保费
	if m[fees[0].Date.Year()] < 3 {
		multiply = 12.00
		if month > 6 {
			month = time.Month(7)
		} else {
			month = time.Month(6)
		}
	}

	for _, v := range fees {
		if v.Date.Year() == d.Year() && v.Date.Month() == month {
			return v.Fee / multiply
		}
	}
	return 0.00
}

//Fee How much you should pay for a single month 月缴费额
type Fee struct {
	Date time.Time
	Fee  float64
}

//Durations premium terms
type Durations []Duration

//Duration premium term and payment date 缴费期间和真正缴费时间
type Duration struct {
	StartDate   time.Time
	StopDate    time.Time
	PaymentDate time.Time
}

func toInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

//Rates overdue fine rates
type Rates []Rate

func (rates Rates) Len() int {
	return len(rates)
}

func (rates Rates) Swap(i, j int) {
	rates[i], rates[j] = rates[j], rates[i]
}

func (rates Rates) Less(i, j int) bool {
	return rates[i].StopDate.Before(rates[j].StopDate)
}

//Rate overdue fine rate, execute time, stop time
type Rate struct {
	Rate      float64
	StartDate time.Time
	StopDate  time.Time
}

//NewRates read overdue fine rates from file
func NewRates(file *excelize.File) Rates {
	var rates Rates

	rows := file.GetRows("滞纳金费率")

	if len(rows) == 0 {
		log.Fatal("Can not find Sheet \"滞纳金费率\" in config.xlsx.")
	}

	for _, row := range rows {
		if unicode.IsDigit(rune(row[0][0])) {
			var rate Rate
			rate.Rate, _ = strconv.ParseFloat(row[0], 64)
			rate.StartDate = time.Date(toInt(row[1][6:]), time.Month(toInt(row[1][:2])), toInt(row[1][3:5]), 0, 0, 0, 0, time.UTC)
			rate.StopDate = time.Date(toInt(row[2][6:]), time.Month(toInt(row[2][:2])), toInt(row[2][3:5]), 24, 0, 0, 0, time.UTC)
			rates = append(rates, rate)
		}
	}

	return rates
}

//FilterByDuration filter rates by start and stop date
func (rates Rates) FilterByDuration(start, stop time.Time) Rates {
	r := Rates{}
	for _, v := range rates {
		if !(v.StopDate.Before(start) || stop.Before(v.StartDate)) {
			r = append(r, v)
		}
	}
	sort.Sort(r)
	return r
}

//genOverdueDays param: date starting to overdue, date the overdue fine is paid
func genOverdueDays(overdue, paid time.Time) float64 {
	// same day
	if overdue.Year() == paid.Year() && overdue.Month() == paid.Month() && overdue.Day() == paid.Day() {
		return 0.00
	}

	duration := paid.Sub(overdue).Hours() / 24
	gracePeriod := 15

	// same month
	if paid.Year() == overdue.Year() && paid.Month() == overdue.Month() {
		if paid.Day() >= gracePeriod {
			return float64(paid.Day() - gracePeriod)
		}
		return 0.00
	}

	return duration - math.Abs(float64(gracePeriod-overdue.Day())) - 1.00
}

func genOverdueFines(durations Durations, fees Fees, rates Rates) float64 {
	fines := 0.00
	for _, duration := range durations {
		fine := 0.00
		for i := duration.StartDate; i.Before(duration.StopDate); i = i.AddDate(0, 1, 0) {
			fine += genMonthlyOverdueFine(i, duration.PaymentDate, fees, rates)
		}
		fines += fine
	}
	return fines
}

func genMonthlyOverdueFine(overdueDate, paymentDate time.Time, fees Fees, rates Rates) float64 {
	fine := 0.00
	// get the fee of the overdueDate
	fee := fees.GetFeeForMonth(overdueDate)
	r := rates.FilterByDuration(overdueDate, paymentDate)
	if len(r) < 2 {
		return fee * r[0].Rate * genOverdueDays(overdueDate, paymentDate)
	}
	// segment
	d := 0.00
	for i, rate := range r {
		if i == 0 {
			days := genOverdueDays(overdueDate, rate.StopDate)
			fine += fee * rate.Rate * days
			d = days
			continue
		}
		var stop time.Time
		if i == len(r)-1 {
			stop = paymentDate
		} else {
			stop = rate.StopDate
		}
		days := genOverdueDays(overdueDate, stop)
		fine += fee * rate.Rate * (days - d)
		d = days
	}
	return fine
}

func main() {
	f, err := excelize.OpenFile("config.xlsx")
	if err != nil {
		log.Fatalf("Can not open config.xlsx.")
	}
	fees := NewFees(f)
	rates := NewRates(f)

	var start, stop, paid time.Time

	var startDate, stopDate, paidDate *walk.DateEdit
	var display, result *walk.TextEdit
	MainWindow{
		Title:  "税费款滞纳金计算程序",
		Size:   Size{600, 400},
		Layout: VBox{},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					TextLabel{Text: "税费款所属期起"},
					DateEdit{AssignTo: &startDate,
						OnDateChanged: func() {
							start = startDate.Date()
						},
					},
					TextLabel{Text: "税费款所属期止"},
					DateEdit{AssignTo: &stopDate,
						OnDateChanged: func() {
							stop = stopDate.Date().AddDate(0, 0, 1)
						},
					},
					TextLabel{Text: "入库日期"},
					DateEdit{AssignTo: &paidDate,
						OnDateChanged: func() {
							paid = paidDate.Date().AddDate(0, 0, 1)
						},
					},
				},
			},
			TextLabel{Text: "过程界面"},
			TextEdit{AssignTo: &display, ReadOnly: true},
			TextLabel{Text: "结果界面"},
			TextEdit{AssignTo: &result, ReadOnly: true},
			PushButton{
				Text: "计算",
				OnClicked: func() {
					display.SetText("税费款所属期起:\t" + start.Format("20060102") + "\t\t税费款所属期止:\t" + stop.AddDate(0, 0, -1).Format("20060102") + "\t\t入库日期:\t" + paid.AddDate(0, 0, -1).Format("20060102"))
					duration := Durations{Duration{start, stop, paid}}
					result.SetText("滞纳天数:\t" + strconv.FormatFloat(genOverdueDays(start, paid), 'f', -1, 64) + "\t\t累计滞纳金:\t" + strconv.FormatFloat(genOverdueFines(duration, fees, rates), 'f', -1, 64))
				},
			},
			TextLabel{Text: "Copyright by Tse Leung Chaw (122165974@qq.com), 2019"},
		},
	}.Run()
}
