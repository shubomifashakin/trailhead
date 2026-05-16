package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

type logDetails struct {
  Amount float64 `json:"amount"`
  Description string `json:"description"`
  TransactionType string `json:"transactionType"`
  Currency string `json:"currency"`
  TransactionDate string `json:"transactionDate"`
}


 var reg= regexp.MustCompile(`^[a-zA-Z]+$`)
 var timeLayout="2006-01-02"

func main(){
 var amount float64
 var description string
 var transactionType string
 var currency string
 var transactionDate string
 
 var historySortOrder string
 var historyLimit int
 var historyStartDate string
 var historyEndDate string
 var historyTransactionType string

 historyFlag:= "history"
 insertFlag:= "insert"
 historyCommand:= flag.NewFlagSet(historyFlag,flag.ExitOnError)
 insertCommand:= flag.NewFlagSet(insertFlag,flag.ExitOnError)


 historyCommand.StringVar(&historySortOrder, "order", "desc", "Sort order (asc or desc)")
 historyCommand.IntVar(&historyLimit, "limit", 10, "Number of transactions to show")
 historyCommand.StringVar(&historyStartDate, "start", "", "Start date (YYYY-MM-DD)")
 historyCommand.StringVar(&historyEndDate, "end", "", "End date (YYYY-MM-DD)")
 historyCommand.StringVar(&historyTransactionType, "type", "all", "Transaction type (income or expense)")
 
 
 insertCommand.Float64Var(&amount, "amount", 0.00,"The amount of the transaction")
 insertCommand.StringVar(&description, "description","","Description of the transaction")
 insertCommand.StringVar(&transactionType, "transactionType","expense", "income or expense")
 insertCommand.StringVar(&currency, "currency","ngn","currency of the transaction")
 insertCommand.StringVar(&transactionDate, "date",time.Now().Format("2006-01-02"),"Date of the transaction (YYYY-MM-DD)")


 if len(os.Args)<2{
		fmt.Println("Expected 'history' or 'insert' subcommands")
		os.Exit(1)
 }

  home, err := os.UserHomeDir()
if err != nil {
    log.Fatalln("Could not find home directory:", err)
}

filePath := filepath.Join(home, ".trailhead", "trailhead.json")

// create the directory if it doesn't exist
if err = os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
    log.Fatalln("Could not create data directory:", err)
}

 switch os.Args[1] {
 case historyFlag:
	 historyCommand.Parse(os.Args[2:])


	 handleHistoryFlag(filePath, historySortOrder, historyLimit, historyStartDate, historyEndDate, historyTransactionType)
 case insertFlag:
	 insertCommand.Parse(os.Args[2:])

	 handleInsertFlag(filePath, amount, description, transactionType, currency, transactionDate)
 default:
	 log.Fatalln("Invalid command")
 }

}

func checkTransactionType(tType string) (string,error){
	tType = strings.ToLower(tType)

  if tType != "expense" && tType != "income" {
	return "", fmt.Errorf("Invalid transaction type")
 } 

 return tType,nil
}

func handleInsertFlag(filePath string, amount float64, description string, transactionType string, currency string, transactionDate string){
	
 if amount == 0.00 {
	log.Fatalln("No amount entered or amount was 0")
 }

 if description == "" {
	log.Fatalln("No description was entered")
 }


 //currency must not be more than 3 characters
 if utf8.RuneCountInString(currency) >3 {
	 log.Fatalln("Invalid currency, currency should not be more than 3 characters")
 }
 
 // currency should only contain alphabets 
 if !reg.MatchString(currency) {
	log.Fatalln("Invalid currency, currency should only contain alphabets")
 }

 currency= strings.ToUpper(currency)

 transactionType,err:=checkTransactionType(transactionType)
 
 if err != nil {
	log.Fatalln(err)
 }

 //validate the time
_, err = time.Parse(timeLayout, transactionDate)
if err != nil {
	log.Fatalln("Invalid date, date should be in the format YYYY-MM-DD")
}

 //get the previous logs, loads all logs into memory at once
previousLogs,err:= os.ReadFile(filePath)

 if err != nil && !os.IsNotExist(err) {
	 log.Fatalln(err)
 }

 var previousLogsArr []logDetails

 //check if there are previous logs
 if len(previousLogs) > 0 {
	//parse the previous logs and store them in the array
	err = json.Unmarshal(previousLogs, &previousLogsArr)

	if err != nil {
		log.Fatalln("Failed to parse previous logs", err)
	}
 }

  deets:= logDetails{
	Amount: amount,
	Description: description,
	TransactionType: transactionType,
	Currency: currency,
	TransactionDate: transactionDate,
  }

  previousLogsArr = append(previousLogsArr, deets)

 bytes,err:=json.MarshalIndent(previousLogsArr,"","  ")
 if err != nil {
	 log.Fatalln("Failed to marshal logs", err)
	}
	
 err=os.WriteFile(filePath, bytes, 0644)

 if err != nil {
	 log.Fatalln("Failed to write logs", err)
 }
 
}

func handleHistoryFlag(filePath string, sortOrder string, limit int, startDate string, endDate string, transactionType string){
 //validate the transaction type
 if transactionType != "all" {
	transactionType2,err := checkTransactionType(transactionType)
	transactionType=transactionType2
 
  if err != nil {
	log.Fatalln(err)
  }
 }

 //validate the sort order
 sortOrder= strings.ToLower(sortOrder)

 if sortOrder != "asc" && sortOrder != "desc" {
   log.Fatalln("Invalid sort order")
 }

 var parsedStartDate time.Time
 var parsedEndDate time.Time
 
 //validate the start and end date if passed
 if startDate != "" {
  parsedStartDate2,err:= time.Parse(timeLayout,startDate)
  
  if err != nil {
	log.Fatalln("Invalid start date")
  }

  parsedStartDate=parsedStartDate2
 }

 if endDate != "" {
  parsedEndDate2,err:= time.Parse(timeLayout,endDate)
  
  if err != nil {
	log.Fatalln("Invalid end date")
  }

  parsedEndDate=parsedEndDate2
 }
 
 logs, err := os.ReadFile(filePath)
 if err != nil {
	 log.Fatalln("Failed to read logs", err)
 }
 
 var logsArr []logDetails

 //converts the data from json into the specified data structure
 err = json.Unmarshal(logs, &logsArr)
 if err != nil {
	 log.Fatalln("Failed to parse logs", err)
 }

 filteredLogs:=logsArr

 //filter the logs for the specified transaction type if not all
 if transactionType != "all" {
	filteredLogs= []logDetails{}

	for _, log:= range logsArr {
		if log.TransactionType == transactionType {
          filteredLogs=append(filteredLogs, log)
		}
	}
 }

 logsBeforeEndDate:= filteredLogs

 //get all the logs that are before the end date
 if !parsedEndDate.IsZero() {
	logsBeforeEndDate=[]logDetails{}

	for _, tLog:= range filteredLogs {
		parsedDate,err:= time.Parse(timeLayout,tLog.TransactionDate)

		if err != nil {
			log.Fatalln("Invalid log date: ",err)
		}


		if parsedDate.Before(parsedEndDate){
         logsBeforeEndDate= append(logsBeforeEndDate, tLog)
		}
	}
 }

 logsBeforeEndDateButAfterStartDate:=logsBeforeEndDate


 //get all logs that are after the start date
 if !parsedStartDate.IsZero() {
	logsBeforeEndDateButAfterStartDate=[]logDetails{}

	for _, tLog:= range logsBeforeEndDate {
		parsedDate,err:= time.Parse(timeLayout,tLog.TransactionDate)

		if err != nil {
			log.Fatalln("Invalid log date: ",err)
		}


		if parsedDate.After(parsedStartDate){
         logsBeforeEndDateButAfterStartDate= append(logsBeforeEndDateButAfterStartDate, tLog)
		}
	}
 }
 
 slices.SortFunc(logsBeforeEndDateButAfterStartDate, func(a, b logDetails) int {
    dateA, _ := time.Parse(timeLayout, a.TransactionDate)
    dateB, _ := time.Parse(timeLayout, b.TransactionDate)

    if sortOrder == "desc" {
        return dateB.Compare(dateA)
    }
    return dateA.Compare(dateB)
})


result := logsBeforeEndDateButAfterStartDate
if limit < len(result) {
    result = result[:limit]
}
fmt.Println(result)
}