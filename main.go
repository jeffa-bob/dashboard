package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

type School struct {
	Name           string  `json:Name`
	Population     int     `json:population`
	ActiveTotal    int     `json:activeTotal`
	ActiveStaff    int     `json:activeStaff`
	ActiveStudents int     `json:activeStudents`
	Proportional   float64 `json:Proportional`
	Notes          string  `json:notes`
}

type Data struct {
	Schools map[string]School `json:schools`
	Date    string            `json:date`
}

var currentData Data

func main() {

	currentData = Data{Date: time.Now().Format("Jan-2-2006")}
	currentData.Schools = make(map[string]School)

	response, err := http.Get("https://docs.google.com/spreadsheets/u/2/d/1VaaE_miJ1hQDbCNqjxUa0qFJZ6pw0hA9WkM2ikby8eQ/htmlembed/sheet?gid=0")
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	response.Body.Close()

	if err != nil {
		log.Fatal(err)
	}

	names := make([]string, 45)
	total := make([]string, 45)
	staff := make([]string, 45)
	students := make([]string, 45)
	proportional := make([]string, 45)

	doc.Find("td").Each(func(i int, s *goquery.Selection) {
		// fmt.Printf("Content of cell %d: %s\n", i, s.Text())

		text := s.Text()
		if i >= 20 && i <= 512 {
			if (i-20)%11 == 0 {
				names[(i-20)/11] = text
				//	fmt.Printf("%s ",text)
				return
			}
			if (i-22)%11 == 0 {
				if text == "NA" {
					text = "0"
				}

				total[(i-22)/11] = text
				//		fmt.Printf("%s ",text)

				return
			}
			if (i-24)%11 == 0 {
				text = strings.TrimSuffix(text, "^^")
				text = strings.TrimSuffix(text, "^")

				if text == "NA" {
					text = "0"
				}

				staff[(i-24)/11] = text
				//	fmt.Printf("%s ",text)

				return
			}
			if (i-26)%11 == 0 {
				if text == "NA" {
					text = "0"
				}
				text = strings.TrimSuffix(text, "^^")
				text = strings.TrimSuffix(text, "^")

				students[(i-26)/11] = text
				//	fmt.Printf("%s ",text)

				return
			}
			if (i-28)%11 == 0 {
				if text == "NA" {
					text = "0"
				}

				proportional[(i-28)/11] = text
				//	fmt.Printf("%s \n",text)

				return
			}
		}

	})

	for i, s := range names {
		school := School{Name: s}
		// fmt.Printf("School:%s \n", s)
		// fmt.Printf("Total:%s \n", total[i])
		school.Population, err = strconv.Atoi(total[i])
		if err != nil {
			log.Fatal(err)

		}
		// fmt.Printf("Staff:%s \n", staff[i])

		school.ActiveStaff, err = strconv.Atoi(staff[i])
		if err != nil {
			log.Fatal(err)

		}
		// fmt.Printf("Students:%s ", students[i])

		school.ActiveStudents, err = strconv.Atoi(students[i])
		if err != nil {
			log.Fatal(err)

		}
		// fmt.Printf("Proportional:%s\n", proportional[i])

		school.ActiveTotal = school.ActiveStudents + school.ActiveStaff
		school.Proportional, err = strconv.ParseFloat(proportional[i], 8)
		if err != nil {
			log.Fatal(err)
		}
		currentData.Schools[s] = school

	}

	SerializeData()
	MakeChart()
}

func SerializeData() {

	// jsonString, err := json.Marshal(currentData)
	jsonString, err := json.MarshalIndent(currentData, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("./data/"+time.Now().Format("Jan-2-2006")+".json", jsonString, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func MakeChart() {

	var days []Data = []Data{}

	files, err := ioutil.ReadDir("./data/")
	if err != nil {
		log.Fatal(err)
	}

	var daystrings = []string{}

	for _, f := range files {

		file, err := os.Open("./data/" + f.Name())
		if err != nil {
			log.Fatal(err)
		}

		fileData := Data{}

		jsonParser := json.NewDecoder(file)
		if err = jsonParser.Decode(&fileData); err != nil {
			log.Fatal(err)
		}

		days = append(days, fileData)

		//fmt.Printf("%s",fileData.Date)
		fmt.Println(f.Name())
	}

	sort.Slice(days, func(i, j int) bool {
		a, err := time.Parse("Jan-2-2006", days[i].Date)
		if err != nil {
			log.Fatal(err)
		}
		b, err := time.Parse("Jan-2-2006", days[j].Date)
		if err != nil {
			log.Fatal(err)
		}
		return a.Before(b)
	})

	for _, s := range days {
		daystrings = append(daystrings, s.Date)
	}

	for i, s := range days[len(days)-1].Schools {

		total := make([]opts.LineData, 0)
		staff := make([]opts.LineData, 0)
		student := make([]opts.LineData, 0)
		proportional := make([]opts.LineData, 0)

		for _, t := range days {
			total = append(total, opts.LineData{Value: t.Schools[i].ActiveTotal})
			staff = append(staff, opts.LineData{Value: t.Schools[i].ActiveStaff})
			student = append(student, opts.LineData{Value: t.Schools[i].ActiveStudents})
			proportional = append(proportional, opts.LineData{Value: t.Schools[i].Proportional})
		}

		line := charts.NewLine()

		line.SetGlobalOptions(
			charts.WithDataZoomOpts(opts.DataZoom{
				Start:      80,
				End:        100,
				XAxisIndex: []int{0},
			}),
			charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
			charts.WithTitleOpts(opts.Title{
				Title:    i,
				Subtitle: "Total In-Person Population: " + strconv.Itoa(s.Population) + "   Staff Cases: " + strconv.Itoa(s.ActiveStudents) + "  Student Cases: " + strconv.Itoa(s.ActiveStaff),
			}),
			charts.WithLegendOpts(opts.Legend{Show: true}),
		)

		line.SetXAxis(daystrings).
			AddSeries("Total Active Cases", total, charts.WithLabelOpts(opts.Label{Show: true})).
			AddSeries("Total Active Staff", student, charts.WithLabelOpts(opts.Label{Show: true})).
			AddSeries("Total Active Students", staff, charts.WithLabelOpts(opts.Label{Show: true})).
			SetSeriesOptions(
				//	charts.WithLineChartOpts(opts.LineChart{Smooth: true}),
				charts.WithLabelOpts(opts.Label{Show: true}),
				charts.WithMarkPointStyleOpts(
					opts.MarkPointStyle{Label: &opts.Label{Show: true}}),
			)

		f, _ := os.Create("./charts/" + i + ".html")
		line.Render(f)
	}
}
