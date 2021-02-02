package main

import (
	"log"
	"net/http"
    "io/ioutil"
	"time"
	"encoding/json"
	"strconv"
	"strings"
	"fmt"
	"os"
	"github.com/PuerkitoBio/goquery"
	 "github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)


type School struct{
	Name string `json:Name`
	Population int `json:population`
	ActiveTotal int `json:activeTotal`
	ActiveStaff int `json:activeStaff`
	ActiveStudents int `json:activeStudents`
	Proportional float64 `json:Proportional`
	Notes string `json:notes`
}

type Data struct{
	Schools map[string]School `json:schools`
	Date string `json:date`
}

var currentData Data

func main(){

	currentData = Data{Date:time.Now().Format("Jan-2-2006")}
	currentData.Schools = make(map[string]School)
	
	response, err := http.Get("https://docs.google.com/spreadsheets/d/1-dt0LlQaP-yA-koWz3tJzR0urkPvlzbEjmqVwWXA5W8/htmlembed/sheet?gid=0")
    if err != nil {
        log.Fatal(err)
    }


	doc, err := goquery.NewDocumentFromReader(response.Body)
	response.Body.Close()
	
    if err != nil {
        log.Fatal(err)
    }

	names := make([]string,43)
	total := make([]string,43)
	staff := make([]string,43)
	students := make([]string,43)
	proportional := make([]string,43)
    
    doc.Find("td").Each(func (i int, s *goquery.Selection) {
      //fmt.Printf("Content of cell %d: %s\n", i, s.Text())

		text := s.Text()
		if i >= 13 && i < 400{
			if (i - 13) % 9 == 0{
				names[(i-13)/9] = text
			//	fmt.Printf("%s ",text)
				return
			}
			if (i - 15) % 9 == 0{
				if text=="NA"{text = "0"}

				total[(i-15)/9] = text
		//		fmt.Printf("%s ",text)

				return
			}
			if (i - 17) % 9 == 0{
				text = strings.TrimSuffix(text,"^^")
				text = strings.TrimSuffix(text,"^")

				if text=="NA"{text = "0"}
				
				staff[(i-17)/9] = text
			//	fmt.Printf("%s ",text)

				return
			}
			if (i - 19) % 9 == 0{
				if text=="NA"{text = "0" }
				text = strings.TrimSuffix(text,"^^")
				text = strings.TrimSuffix(text,"^")

				students[(i-19)/9] = text
							//	fmt.Printf("%s ",text)

				return
			}
			if (i - 21) % 9 == 0{
				if text=="NA"{text = "0" }

				proportional[(i-21)/9] = text
			//	fmt.Printf("%s \n",text)

				return
			}
		}
      
    })


    for i, s := range names{
    		school := School{Name:s}
    		fmt.Printf("School:%s ",s)
    		fmt.Printf("Total:%s ",total[i])
    		school.Population,err = strconv.Atoi(total[i])
    		if err != nil {
    		        log.Fatal(err)

    		}
    		fmt.Printf("Staff:%s ",staff[i])

    		school.ActiveStaff,err = strconv.Atoi(staff[i])
    		if err != nil {
    		        log.Fatal(err)

    		}
    		fmt.Printf("Students:%s ",students[i])

    		school.ActiveStudents,err = strconv.Atoi(students[i])
    		if err != nil {
    		        log.Fatal(err)

    		}
    	     fmt.Printf("Proportional:%s\n",proportional[i])

    		school.ActiveTotal = school.ActiveStudents + school.ActiveStaff
    		school.Proportional,err = strconv.ParseFloat(proportional[i],8)
    		if err != nil {
    		        log.Fatal(err)
    		}
			currentData.Schools[s] = school
    	
    }

    SerializeData()
    MakeChart()
}

func SerializeData(){

	// jsonString, err := json.Marshal(currentData)
	jsonString, err := json.MarshalIndent(currentData,""," ")
  	if err != nil {
        log.Fatal(err)
    }
	
    err = ioutil.WriteFile("./data/"+time.Now().Format("Jan-2-2006")+".json",jsonString, 0644)
    if err != nil {
        log.Fatal(err)
    }
}


func MakeChart(){

	var days []Data = []Data{}
	
    files, err := ioutil.ReadDir("./data/")
    if err != nil {
        log.Fatal(err)
    }

	var daystrings = []string{}
 
    for _, f := range files {

    		file, err := os.Open("./data/"+f.Name())
 			if err != nil {
        		log.Fatal(err)
    		}

			fileData := Data{}

			jsonParser := json.NewDecoder(file)
			if err = jsonParser.Decode(&fileData); err != nil {
				log.Fatal(err)
			}

			days = append(days, fileData)
			daystrings = append(daystrings, fileData.Date)

			
			//fmt.Printf("%s",fileData.Date)
            fmt.Println(f.Name(),"\n")
    }	

    for i,s := range days[len(days)-1].Schools{

		total := make([]opts.LineData, 0)
		staff := make([]opts.LineData, 0)
		student := make([]opts.LineData, 0)
		proportional := make([]opts.LineData, 0)
			
		for _,t := range days{
			total = append(total, opts.LineData{Value: t.Schools[i].ActiveTotal})
			staff = append(staff, opts.LineData{Value: t.Schools[i].ActiveStaff})
			student = append(student, opts.LineData{Value: t.Schools[i].ActiveStudents})
			proportional = append(proportional, opts.LineData{Value: t.Schools[i].Proportional})
		}
    	
    	line := charts.NewLine()

    	line.SetGlobalOptions(
    			charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
    			charts.WithTitleOpts(opts.Title{
    				Title:    i,
    				Subtitle: "Total In-Person Population: " + strconv.Itoa(s.Population) + "   Staff Cases: "  + strconv.Itoa(s.ActiveStaff) + "  Student Cases: "   + strconv.Itoa(s.ActiveStudents) + "  Proportional: "  + strconv.FormatFloat(s.Proportional, 'E', -1, 64),
    				}))

    	line.SetXAxis(daystrings).
    			AddSeries("Total Active Cases", total).
    			AddSeries("Total Active Staff", staff).
    			AddSeries("Total Active Students", student).
    			SetSeriesOptions(
    				charts.WithLineChartOpts(opts.LineChart{Smooth: true}),
    				charts.WithLabelOpts(opts.Label{Show: true}),
    				)
    					
    	f, _ := os.Create("./charts/"+i+".html")
    	line.Render(f)
    }
}
