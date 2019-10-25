package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type MonthlyFees []MonthlyFee

type MonthlyFee struct {
	Year        int
	Month       int
	PercentInst float64
	PercentPaid float64
	Fee         float64
}

func genFeesFromBaseFile(file string) MonthlyFees {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Can not open %s", file)
	}
	defer f.Close()

	fees := MonthlyFees{}

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "#") {
			continue
		}
		strs := strings.Split(scanner.Text(), ",")
		fee := MonthlyFee{}
		fee.Year, _ = strconv.Atoi(strs[0])
		fee.Month, _ = strconv.Atoi(strs[1])
		fee.PercentPaid, _ = strconv.ParseFloat(strs[3], 64)
		fee.PercentInst, _ = strconv.ParseFloat(strs[4], 64)
		base, _ := strconv.ParseFloat(strs[2], 64)
		fee.Fee = base * fee.PercentPaid * fee.PercentInst / 12.00
		fees = append(fees, fee)
	}
	return fees
}

//Date date struct
type Date struct {
	Year  int
	Month int
	Day   int
}

//NewDate generiate new Date from a date string
func NewDate(date string) Date {
	d := Date{}
	d.Year, _ = strconv.Atoi(date[:4])
	d.Month, _ = strconv.Atoi(date[4:6])
	d.Day, _ = strconv.Atoi(date[6:])
	return d
}

func genOverdueDays(start, stop string) int {
	// same day
	if start == stop {
		return 0
	}
	startDate := NewDate(start)
	stopDate := NewDate(stop)

	days := 0

	// start day to the end of that month
	daysThisMonth := getDaysFromMonth(startDate.Year, startDate.Month)
	overdueThisMonth := daysThisMonth - startDate.Day + 1
	days += overdueThisMonth

	// stop at the same year
	if stopDate.Year == startDate.Year {
		// stop at the same month
		if stopDate.Month == startDate.Month {
			return stopDate.Day - startDate.Day + 1
		}
		days += stopDate.Day
		if stopDate.Month-startDate.Month > 1 {
			for i := startDate.Month + 1; i < stopDate.Month; i++ {
				days += getDaysFromMonth(startDate.Year, i)
			}
		}
		return days
	}
	// next month to the end of that year
	for i := startDate.Month + 1; i <= 12; i++ {
		days += getDaysFromMonth(startDate.Year, i)
	}
	// stop year months
	for i := 1; i < stopDate.Month; i++ {
		days += getDaysFromMonth(stopDate.Year, i)
	}
	// stop year days
	days += stopDate.Day
	if stopDate.Year-startDate.Year > 1 {
		// from next year of start year to the last year of stop year
		for i := startDate.Year + 1; i < stopDate.Year; i++ {
			days += getDaysFromYear(i)
		}
	}
	return days
}

func getDaysFromYear(year int) int {
	if year%4 == 0 {
		return 366
	}
	return 365
}

func getDaysFromMonth(year, month int) int {
	// m : how many days a month contains
	m := map[int]int{1: 31, 2: 28, 3: 31, 4: 30, 5: 31, 6: 30, 7: 31, 8: 31, 9: 30, 10: 31, 11: 30, 12: 31}
	days := m[month]
	if year%4 == 0 {
		days++
	}
	return days
}

func main() {
	fees := genFeesFromBaseFile("base.txt")
	fmt.Println(fees)
	fmt.Println(genOverdueDays("20090503", "20090503"))
	fmt.Println(genOverdueDays("20090503", "20090504"))
	fmt.Println(genOverdueDays("20090503", "20090804"))
	fmt.Println(genOverdueDays("20090503", "20100502"))
	fmt.Println(genOverdueDays("20090503", "20110504"))
}
