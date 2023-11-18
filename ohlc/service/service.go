package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"stock/ohlc"
	"stock/ohlc/repository"

	"github.com/IBM/sarama"
)

type OHLCService interface {
	GetStockList(ctx context.Context) ([]*ohlc.StockItem, error)
	GetStockSummary(ctx context.Context, stock string, dayDate string) (*ohlc.StockSummary, error)

	GetFeederStatus(ctx context.Context) (string, error)
	Feeder() error
	ListenOrderTx(key, value []byte) error

	UpdateSummary(pSummary *ohlc.StockSummary, orderTx *ohlc.OrderTx)
}

type ohlcService struct {
	ohlcRepository repository.OHLCRepository
	kafkaProducer  sarama.SyncProducer
}

func NewOHLCService(ohlcRepository repository.OHLCRepository, kafkaProducer sarama.SyncProducer) OHLCService {
	return &ohlcService{ohlcRepository, kafkaProducer}
}

func (o *ohlcService) GetFeederStatus(ctx context.Context) (string, error) {
	status, err := o.ohlcRepository.GetFeederStatus(ctx)
	if err != nil {
		return "", err
	}

	if status == "" {
		return "IDLE", nil
	}

	return status, nil
}

func (o *ohlcService) GetStockList(ctx context.Context) ([]*ohlc.StockItem, error) {
	stockItems, err := o.ohlcRepository.GetStockList(ctx)
	if err != nil {
		return []*ohlc.StockItem{}, err
	}
	return stockItems, nil
}

func (o *ohlcService) GetStockSummary(ctx context.Context, stock string, dayDate string) (*ohlc.StockSummary, error) {
	return o.ohlcRepository.GetStockSummary(ctx, stock, dayDate)
}

func (o *ohlcService) Feeder() error {
	log.Println("info.resseting the previous data")
	err := o.ohlcRepository.FlushStockList(context.Background())
	if err != nil {
		return err
	}

	log.Println("info.start feeding data..")

	err = o.ohlcRepository.SetFeederStatus(context.Background(), "FEEDING")
	if err != nil {
		return err
	}

	subsetDataFolder := "./subsetdata"
	subsetFileList, err := ioutil.ReadDir(subsetDataFolder)
	if err != nil {
		log.Println("errors.cannot list the subset data")
		return err
	}

	for _, v := range subsetFileList {
		err = o.scanFileAndPushOrder(fmt.Sprintf("%s/%s", subsetDataFolder, v.Name()), v.Name())
		if err != nil {
			log.Println("errors.when scanning the file")
		}
	}

	log.Println("info.feeding data completed")

	err = o.ohlcRepository.SetFeederStatus(context.Background(), "COMPLETED")
	if err != nil {
		return err
	}

	return nil
}

func (o *ohlcService) ListenOrderTx(key, value []byte) error {
	orderTx := ohlc.OrderTx{}
	err := json.Unmarshal(value, &orderTx)
	if err != nil {
		log.Println("errors.parsing value to struct")
		return err
	}

	return o.processOrderTx(&orderTx)
}
