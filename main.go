package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

//Fees A collection of Fee
type Fees []Fee

//NewFees read fees from file
func NewFees(file string) Fees {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Can not open %s", file)
	}
	defer f.Close()

	fees := Fees{}

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "#") {
			continue
		}
		strs := strings.Split(scanner.Text(), ",")
		fee := Fee{}
		yy, _ := strconv.Atoi(strs[0])
		mm, _ := strconv.Atoi(strs[1])
		fee.Date = time.Date(yy, time.Month(mm), 1, 0, 0, 0, 0, time.UTC)
		paymentPercent, _ := strconv.ParseFloat(strs[3], 64)
		institutionPercent, _ := strconv.ParseFloat(strs[4], 64)
		base, _ := strconv.ParseFloat(strs[2], 64)
		fee.Fee = base * paymentPercent * institutionPercent
		fees = append(fees, fee)
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
	Start       time.Time
	Stop        time.Time
	PaymentDate time.Time
}

//NewDuration A struct of the start & stop date of assumed premium term, and the actual payment date. 费款所属期间和真正缴费日期
func NewDuration(start, stop, payment string) Duration {
	return Duration{
		time.Date(toInt(start[:4]), time.Month(toInt(start[4:6])), toInt(start[6:]), 0, 0, 0, 0, time.UTC),
		time.Date(toInt(stop[:4]), time.Month(toInt(stop[4:6])), toInt(stop[6:]), 24, 0, 0, 0, time.UTC),
		time.Date(toInt(payment[:4]), time.Month(toInt(payment[4:6])), toInt(payment[6:]), 24, 0, 0, 0, time.UTC),
	}
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
func NewRates(file string) Rates {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Can not open %s", file)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	rates := Rates{}

	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "#") {
			continue
		}
		strs := strings.Split(scanner.Text(), ",")
		r := Rate{}
		r.Rate, _ = strconv.ParseFloat(strs[0], 64)
		r.StartDate = time.Date(toInt(strs[1][:4]), time.Month(toInt(strs[1][4:6])), toInt(strs[1][6:]), 0, 0, 0, 0, time.UTC)
		r.StopDate = time.Date(toInt(strs[2][:4]), time.Month(toInt(strs[2][4:6])), toInt(strs[2][6:]), 24, 0, 0, 0, time.UTC)
		rates = append(rates, r)
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
func genOverdueDays(overdue, paid string) float64 {
	if overdue == paid {
		return 0.00
	}
	tOverdue := time.Date(toInt(overdue[:4]), time.Month(toInt(overdue[4:6])), toInt(overdue[6:]), 0, 0, 0, 0, time.UTC)
	tPaid := time.Date(toInt(paid[:4]), time.Month(toInt(paid[4:6])), toInt(paid[6:]), 24, 0, 0, 0, time.UTC)

	duration := tPaid.Sub(tOverdue).Hours() / 24
	gracePeriod := 15

	// same month
	if tPaid.Year() == tOverdue.Year() && tPaid.Month() == tOverdue.Month() {
		if tPaid.Day() >= gracePeriod {
			return float64(tPaid.Day() - gracePeriod)
		}
		return 0.00
	}

	return duration - math.Abs(float64(gracePeriod-tOverdue.Day())) - 1.00
}

func genOverdueFines(durations Durations, fees Fees, rates Rates) float64 {
	fines := 0.00
	for _, duration := range durations {
		fine := 0.00
		for i := duration.Start; i.Before(duration.Stop); i = i.AddDate(0, 1, 0) {
			a := genMonthlyOverdueFine(i, duration.PaymentDate, fees, rates)
			fmt.Println(a)
			fine += a
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
		return fee * r[0].Rate * genOverdueDays(overdueDate.Format("20060102"), paymentDate.Format("20060102"))
	}
	// segment
	d := 0.00
	for i, rate := range r {
		if i == 0 {
			days := genOverdueDays(overdueDate.Format("20060102"), rate.StopDate.Format("20060102"))
			fine += fee * rate.Rate * days
			d = days
			continue
		}
		stop := time.Time{}
		if i == len(r)-1 {
			stop = paymentDate
		} else {
			stop = rate.StopDate
		}
		days := genOverdueDays(overdueDate.Format("20060102"), stop.Format("20060102"))
		fine += fee * rate.Rate * (days - d)
		d = days
	}
	return fine
}

func main() {
	d := Durations{NewDuration("20090501", "20110430", "20180831")}
	fees := NewFees("base.txt")
	rates := NewRates("rate.txt")
	fmt.Println(genOverdueFines(d, fees, rates))
}
