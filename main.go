package main

import (
	"log"
	"net/http"
	"os"
	"io"
    "io/ioutil"
	"time"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
)


type School struct{
	Name string
	Id int
	Total int
	ActivetTotal int
	ActiveStaff int
	ActiveStudents int
	TotalStaff int
	TotalStudents int
	Notes string
}

type Data struct{
	Schools map[string]School
	Date time.Time
}

var currentData Data

func main(){

	currentData = Data{Date:time.Now()}

	defer SerializeData()
	
	response, err := http.Get("https://docs.google.com/spreadsheets/d/1-dt0LlQaP-yA-koWz3tJzR0urkPvlzbEjmqVwWXA5W8/htmlembed/sheet?gid=0")
    if err != nil {
        log.Fatal(err)
    }

    // Copy data from the response to standard output
    
   /* _ , err = io.Copy(os.Stdout, response.Body)
    if err != nil {
        log.Fatal(err)
    }*/

	doc, err := goquery.NewDocumentFromReader(response.Body)
	response.Body.Close()
	
    if err != nil {
        log.Fatal(err)
    }
    
     doc.Is("#lc_searchresult > table > tbody > tr")
}

func SerializeData(){

	jsonString, err := json.Marshal(currentData)
  	if err != nil {
        log.Fatal(err)
    }

	
    err = ioutil.WriteFile("./data/"+time.Now().Format("Jan-2-2006"),jsonString, 0644)
    if err != nil {
        log.Fatal(err)
    }
}
