package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bitly/go-simplejson"
	"os"
	"strconv"
	"strings"
)

type buy_request struct {
	StockSymbolAndPercentage string
	Budget                   float32
}

type buy_response struct {
	TradeNum         int
	Stocks           []string
	UninvestedAmount float32
}

type check_response struct {
	Stocks           []string
	UninvestedAmount float32
	TotalMarketValue float32
}

func main() {

	
	if len(os.Args) > 4 || len(os.Args) < 2 {
		fmt.Println("Wrong number of arguments!")
		usage()
		return

	} else if len(os.Args) == 2 { 

		_, err := strconv.ParseInt(os.Args[1], 10, 64)
		if err != nil {
			fmt.Println("Illegal argument!")
			usage()
			return
		}

		

		data, err := json.Marshal(map[string]interface{}{
			"method": "StockAccounts.Check",
			"id":     1,
			"params": []map[string]interface{}{map[string]interface{}{"TradeId": os.Args[1]}},
		})

		if err != nil {
			log.Fatal("Marshal : %v", err)
		}

		resp, err := http.Post("http://127.0.0.1:8888/rpc", "application/json", strings.NewReader(string(data)))

		if err != nil {
			log.Fatalf("Post: %v", err)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Fatal("ReadAll: %v", err)
		}

		newjson, err := simplejson.NewJson(body)

		checkError(err)

		fmt.Println("stocks: ")
		stocks := newjson.Get("result").Get("Stocks")
		fmt.Println(*stocks)

		fmt.Print("uninvested amount: ")
		uninvestedAmount, _ := newjson.Get("result").Get("UninvestedAmount").Float64()
		fmt.Print("$")
		fmt.Println(uninvestedAmount)

		fmt.Println("market value: ")
		marketvalue, _ := newjson.Get("result").Get("marketvalue").Float64()
		fmt.Print("$")
		fmt.Println(marketvalue)

	} else if len(os.Args) == 3 { 
		budget, err := strconv.ParseFloat(os.Args[2], 64)
		if err != nil {
			fmt.Println("Wrong budget argument.")
			usage()
			return
		}

		data, err := json.Marshal(map[string]interface{}{
			"method": "stock_accounts.Buy",
			"id":     2,
			"params": []map[string]interface{}{map[string]interface{}{"StockSymbolAndPercentage": os.Args[1], "Budget": float32(budget)}},
		})

		if err != nil {
			log.Fatal("Marshal : %v", err)
		}

		resp, err := http.Post("http://127.0.0.1:8888/rpc", "application/json", strings.NewReader(string(data)))

		if err != nil {
			log.Fatalf("Post: %v", err)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Fatal("ReadAll: %v", err)
		}

		newjson, err := simplejson.NewJson(body)

		checkError(err)

		fmt.Print("Trade Num: ")
		tradenum, _ := newjson.Get("result").Get("TradeNum").Int()
		fmt.Println(tradenum)

		fmt.Print("stocks: ")
		stocks := newjson.Get("result").Get("Stocks")
		fmt.Println(*stocks)

		fmt.Print("uninvested amount: ")
		uninvestedAmount, _ := newjson.Get("result").Get("UninvestedAmount").Float64()
		fmt.Print("$")
		fmt.Println(uninvestedAmount)

	} else {
		fmt.Println("Unknown error.")
		usage()
		return
	}

}


func check_error(err error) {
	if err != nil {
		fmt.println(os.Stderr, "Fatal error: %s\n", err.Error())
		log.Fatal("error: ", err)
		os.Exit(2)
	}

}


func usage() {

	fmt.Println("Usage: ", os.Args[0], "tradeId")
	fmt.Println("or")
	fmt.Println(os.Args[0], "“GOOG:50%,YHOO:50%” 10000(your budget)")
}