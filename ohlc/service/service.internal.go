package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"stock/ohlc"
	"strconv"
	"strings"

	"github.com/IBM/sarama"
)

func (o *ohlcService) scanFileAndPushOrder(fileLocation, name string) error {
	readFile, err := os.Open(filepath.Clean(fileLocation))
	if err != nil {
		log.Println("errors.cannot open subset data")
		return err
	}

	dayDateParse := strings.Split(name, "-")
	dayDate := strings.Join(dayDateParse[0:3], "-")

	readFileScan := bufio.NewScanner(readFile)
	for readFileScan.Scan() {
		text := readFileScan.Text()

		var result map[string]interface{}
		_ = json.Unmarshal([]byte(text), &result)

		orderTx := &ohlc.OrderTx{}
		orderTx.DayDate = dayDate

		for key, value := range result {
			switch key {
			case "type":
				orderTx.Type = fmt.Sprint(value)
			case "stock_code":
				orderTx.Stock = fmt.Sprint(value)
			case "executed_quantity":
				v, _ := strconv.Atoi(fmt.Sprint(value))
				orderTx.Quantity = v
			case "quantity":
				v, _ := strconv.Atoi(fmt.Sprint(value))
				orderTx.Quantity = v
			case "price":
				v, _ := strconv.Atoi(fmt.Sprint(value))
				orderTx.Price = v
			case "execution_price":
				v, _ := strconv.Atoi(fmt.Sprint(value))
				orderTx.Price = v
			}
		}

		err = o.pushOrderTx(orderTx)
		if err != nil {
			log.Println("errors.to push order tx")
		}
	}

	err = readFile.Close()
	if err != nil {
		log.Println("errors.cannot close the file")
		return err
	}
	return nil
}

func (o *ohlcService) pushOrderTx(orderTx *ohlc.OrderTx) error {
	data, err := json.Marshal(orderTx)
	if err != nil {
		log.Println("errors.to marshal orderTx")
		return err
	}

	msg := &sarama.ProducerMessage{}
	msg.Topic = "ohlc"
	msg.Key = sarama.StringEncoder(orderTx.Stock)
	msg.Value = sarama.StringEncoder(data)

	_, _, err = o.kafkaProducer.SendMessage(msg)
	if err != nil {
		log.Println("errors.when sending message to kafka broker")
		return err
	}

	return nil
}

func (*ohlcService) UpdateSummary(pSummary *ohlc.StockSummary, orderTx *ohlc.OrderTx) {
	if orderTx.Type == "A" {
		if orderTx.Quantity > 0 {
			return
		}

		pSummary.PreviousPrice = orderTx.Price
	}

	if orderTx.Type == "E" || orderTx.Type == "P" {
		pSummary.ClosePrice = orderTx.Price
		pSummary.Volume += orderTx.Quantity
		pSummary.Value += orderTx.Price * orderTx.Quantity
		pSummary.AveragePrice = int(math.Round(float64(pSummary.Value) / float64(pSummary.Volume)))

		if pSummary.OpenPrice == 0 {
			pSummary.OpenPrice = orderTx.Price
		}

		if (pSummary.LowestPrice != 0 && pSummary.LowestPrice > orderTx.Price) || pSummary.LowestPrice == 0 {
			pSummary.LowestPrice = orderTx.Price
		}

		if pSummary.HighestPrice < orderTx.Price {
			pSummary.HighestPrice = orderTx.Price
		}
	}
}

func (o *ohlcService) processOrderTx(orderTx *ohlc.OrderTx) error {
	pSummary, err := o.ohlcRepository.GetStockSummary(context.Background(), orderTx.Stock, orderTx.DayDate)
	if err != nil {
		return err
	}

	o.UpdateSummary(pSummary, orderTx)

	return o.ohlcRepository.SetStockSummary(context.Background(), orderTx.Stock, orderTx.DayDate, pSummary)
}
