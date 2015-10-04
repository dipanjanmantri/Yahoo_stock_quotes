package main

import (
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	// "net/rpc"
	"os"
	"strconv"
	"strings"

	"github.com/bakins/net-http-recover"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"github.com/justinas/alice"
)

type stock_accounts struct {
	stockPortfolio map[int](*portfolio)
}

type portfolio struct {
	stocks           map[string](*Share)
	uninvestedAmount float32
}

type Share struct {
	boughtPrice float32
	sharenum    int
}

type BuyRequest struct {
	StockSymbolAndPercentage string
	Budget                   float32
}

type BuyResponse struct {
	TradeNum         int
	Stocks           []string
	UninvestedAmount float32
}

type CheckResponse struct {
	Stocks           []string
	UninvestedAmount float32
	TotalMarketValue float32
}

type CheckRequest struct {
	TradeId string
}


var st stock_accounts


var tradeId int


func (st *stock_accounts) Buy(httpRq *http.Request, rq *BuyRequest, rsp *BuyResponse) error {

	
	tradeId++
	rsp.TradeNum = tradeId


	if st.stockportfolio == nil {

		st.stockPortfolio = make(map[int](*portfolio))

		st.stockPortfolio[tradeId] = new(Portfolio)
		st.stockPortfolio[tradeId].stocks = make(map[string]*Share)

	}

	
	symbolAndPercentages := strings.Split(rq.StockSymbolAndPercentage, ",")
	newbudget := float32(rq.Budget)
	var spent float32

	for _, stk := range symbolAndPercentages {

		//parse how many shares and their separate budget
		split := strings.Split(stk, ":")
		stockQuote := splited[0]
		percentage := splited[1]
		strPercentage := strings.TrimSuffix(percentage, "%")
		floatPercentage64, _ := strconv.ParseFloat(strPercentage, 32)
		floatPercentage := float32(floatPercentage64 / 100.00)
		currentPrice := checkQuote(stkQuote)

		shares := int(math.Floor(float64(newbudget * floatPercentage / currentPrice)))
		sharesFloat := float32(shares)
		spent += sharesFloat * currentPrice

	
		if _, ok := st.stockPortfolio[tradeId]; !ok {

			newPortfolio := new(Portfolio)
			newPortfolio.stocks = make(map[string]*Share)
			st.stockPortfolio[tradeId] = newPortfolio
		}
		if _, ok := st.stockPortfolio[tradeId].stocks[stkQuote]; !ok {

			newShare := new(Share)
			newShare.boughtPrice = currentPrice
			newShare.shareNum = shares
			st.stockPortfolio[tradeId].stocks[stkQuote] = newShare
		} else {

			total := float32(sharesFloat*currentPrice) + float32(st.stockPortfolio[tradeId].stocks[stkQuote].shareNum)*st.stockPortfolio[tradeId].stocks[stkQuote].boughtPrice
			st.stockPortfolio[tradeId].stocks[stkQuote].boughtPrice = total / float32(shares+st.stockPortfolio[tradeId].stocks[stkQuote].shareNum)
			st.stockPortfolio[tradeId].stocks[stkQuote].shareNum += shares
		}

		stockBought := stkQuote + ":" + strconv.Itoa(shares) + ":$" + strconv.FormatFloat(float64(currentPrice), 'f', 2, 32)

		rsp.Stocks = append(rsp.Stocks, stockBought)
	}

	
	leftOver := newbudget - spent
	rsp.UninvestedAmount = leftOver
	st.stockPortfolio[tradeId].uninvestedAmount += leftOver

	return nil
}

//check account with trade number
func (st *StockAccounts) Check(httpRq *http.Request, checkRq *CheckRequest, checkResp *CheckResponse) error {

	if st.stockPortfolio == nil {
		return errors.New("No account set up yet.")
	}

	//parse argument into a tradeId
	tradeNum64, err := strconv.ParseInt(checkRq.TradeId, 10, 64)

	if err != nil {
		return errors.New("Illegal Trade ID. ")
	}
	tradeNum := int(tradeNum64)

	if pocket, ok := st.stockPortfolio[tradeNum]; ok {

		var currentMarketVal float32
		for stockquote, sh := range pocket.stocks {
			//obtain current price
			currentPrice := checkQuote(stockquote)

			//obtain price when bought,and compare with current price to determine up or down
			var str string
			if sh.boughtPrice < currentPrice {
				str = "+$" + strconv.FormatFloat(float64(currentPrice), 'f', 2, 32)
			} else if sh.boughtPrice > currentPrice {
				str = "-$" + strconv.FormatFloat(float64(currentPrice), 'f', 2, 32)
			} else {
				str = "$" + strconv.FormatFloat(float64(currentPrice), 'f', 2, 32)
			}

			//setup object to response back
			entry := stockquote + ":" + strconv.Itoa(sh.shareNum) + ":" + str

			checkResp.Stocks = append(checkResp.Stocks, entry)

			currentMarketVal += float32(sh.shareNum) * currentPrice
		}

		//calculated uninvested amount
		checkResp.UninvestedAmount = pocket.uninvestedAmount

		//calculate total market value of holding shares
		checkResp.TotalMarketValue = currentMarketVal
	} else {
		return errors.New("No such trade ID. ")
	}

	return nil
}

func main() {

	
	var st = (new(stock_accounts))

	
	tradeId = rand.Intn(10000) + 1


	
	

	
	router := mux.NewRouter()
	server := rpc.NewServer()
	server.RegisterCodec(json.NewCodec(), "application/json")
	server.RegisterService(st, "")

	chain := alice.New(
		func(h http.Handler) http.Handler {
			return handlers.CombinedLoggingHandler(os.Stdout, h)
		},
		handlers.CompressHandler,
		func(h http.Handler) http.Handler {
			return recovery.Handler(os.Stderr, h, true)
		})

	router.Handle("/rpc", chain.Then(server))
	log.Fatal(http.ListenAndServe(":1234", server))

	

}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func checkQuote(stockName string) float32 {
	//yahoo api, to simplify, query only one stock each time
	baseUrlLeft := "https://query.yahooapis.com/v1/public/yql?q=select%20LastTradePriceOnly%20from%20yahoo.finance%0A.quotes%20where%20symbol%20%3D%20%22"
	baseUrlRight := "%22%0A%09%09&format=json&env=http%3A%2F%2Fdatatables.org%2Falltables.env"

	//request http api
	resp, err := http.Get(baseUrlLeft + stockName + baseUrlRight)

	if err != nil {
		log.Fatal(err)
	}

	
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		log.Fatal(err)
	}

	
	if resp.StatusCode != 200 {
		log.Fatal("Query failure, possibly no network connection or illegal stock quote ")
	}

	
	newjson, err := simplejson.NewJson(body)
	if err != nil {
		fmt.Println(err)
	}

	
	price, _ := newjson.Get("query").Get("results").Get("quote").Get("LastTradePriceOnly").String()
	floatPrice, err := strconv.ParseFloat(price, 32)

	
	return float32(floatPrice)
}