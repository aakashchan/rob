package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type MockSale struct {
	Id            int    `json:"id"`
	Title         string `json:"title"`
	Brand         string `json:"brand"`
	ProductSku    string `json:"sku"`
	Description   string `json:"desc"`
	SaleStartTime int64  `json:"start"`
	SaleEndTime   int64  `json:"end"`
	Src           int64  `json:"src"`
}

func salesHandler(w http.ResponseWriter, r *http.Request) {
	// do mocking here
	var saleValues [10]MockSale

	for i := 0; i < 10; i += 1 {
		saleValues[i].Id = i
		saleValues[i].Title = "Just a mock value" + strconv.Itoa(i)
		saleValues[i].Brand = "Test Brand" + strconv.Itoa(i)
		saleValues[i].ProductSku = "sdlfjds1312312"
		saleValues[i].Description = "Just for testing purpose"
		saleValues[i].SaleStartTime = time.Now().UTC().UnixNano()
		saleValues[i].SaleEndTime = time.Now().UTC().UnixNano()
	}

	j, _ := json.Marshal(saleValues)

	w.Write(j)

}

func main() {

	http.HandleFunc("/sales", salesHandler)

	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		fmt.Println(err)
	}
}
