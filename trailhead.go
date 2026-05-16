package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

type logDetails struct {
  Amount float64 `json:"amount"`
  Description string `json:"description"`
  transactionType string `json:"transactionType"`
  Currency string `json:"currency"`
  TransactionDate string `json:"transactionDate"`
}


 var reg= regexp.MustCompile(`^[a-zA-Z]+$`)

func main(){
 var amount float64
 var description string
 var transactionType string
 var currency string
 var transactionDate string
 
 flag.Float64Var(&amount, "amount", 0.00,"The amount of the transaction")
 flag.StringVar(&description, "description","","Description of the transaction")
 flag.StringVar(&transactionType, "transactionType","expense", "income or expense")
 flag.StringVar(&currency, "currency","ngn","currency of the transaction")
 flag.StringVar(&transactionDate, "date",time.Now().Format("2006-01-02"),"Date of the transaction (YYYY-MM-DD)")

 flag.Parse()

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

 transactionType= strings.ToLower(transactionType)

 if transactionType != "expense" && transactionType != "income" {
	log.Fatalln("Invalid transaction type")
 }

 //validate the time
_, err := time.Parse("2006-01-02", transactionDate)
if err != nil {
	log.Fatalln("Invalid date, date should be in the format YYYY-MM-DD")
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
	transactionType: transactionType,
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