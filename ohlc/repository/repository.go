package repository

import (
	"fmt"
	"log"
	"stock/ohlc"
	"strings"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

type OHLCRepository interface {
	GetFeederStatus(ctx context.Context) (string, error)
	SetFeederStatus(ctx context.Context, status string) error

	FlushStockList(ctx context.Context) error

	GetStockList(ctx context.Context) ([]*ohlc.StockItem, error)
	SetStockSummary(ctx context.Context, stock string, dayDate string, stockSummary *ohlc.StockSummary) error
	GetStockSummary(ctx context.Context, stock string, dayDate string) (*ohlc.StockSummary, error)
}

type ohlcRepository struct {
	rdb *redis.Client
}

func NewOHLCRepository(rdb *redis.Client) OHLCRepository {
	return &ohlcRepository{rdb}
}

func (o *ohlcRepository) GetFeederStatus(ctx context.Context) (string, error) {
	status := o.rdb.Get(ctx, "feeder_status")
	err := status.Err()
	if err != nil {
		return "", nil
	}
	return status.Val(), nil
}

func (o *ohlcRepository) SetFeederStatus(ctx context.Context, status string) error {
	err := o.rdb.Set(ctx, "feeder_status", status, -1).Err()
	if err != nil {
		log.Println("errors.failed to set feeder status")
		return err
	}
	return nil
}

func (o *ohlcRepository) FlushStockList(ctx context.Context) error {
	stockList, err := o.GetStockList(ctx)
	if err != nil {
		log.Println("errors.failed to get stock list")
		return err
	}

	for _, v := range stockList {
		err = o.rdb.Del(ctx, fmt.Sprintf("stock_summary:%s:%s", v.Stock, v.DayDate)).Err()
		if err != nil {
			log.Println("errors.failed to delete list")
			return err
		}
	}

	return nil
}

func (o *ohlcRepository) GetStockList(ctx context.Context) ([]*ohlc.StockItem, error) {
	stockItems := []*ohlc.StockItem{}

	scan := o.rdb.Scan(ctx, 0, "stock_summary:*", 0).Iterator()
	for scan.Next(ctx) {
		valParse := strings.Split(scan.Val(), ":")
		stockItems = append(stockItems, &ohlc.StockItem{
			Stock:   valParse[1],
			DayDate: valParse[2],
		})
	}

	err := scan.Err()
	if err != nil {
		log.Println("errors.get stock list")
		return []*ohlc.StockItem{}, err
	}

	return stockItems, nil
}

func (o *ohlcRepository) SetStockSummary(ctx context.Context, stock string, dayDate string, stockSummary *ohlc.StockSummary) error {
	err := o.rdb.HSet(ctx, fmt.Sprintf("stock_summary:%s:%s", stock, dayDate), stockSummary).Err()
	if err != nil {
		log.Println("errors.set stock summary")
		return err
	}

	return nil
}

func (o *ohlcRepository) GetStockSummary(ctx context.Context, stock string, dayDate string) (*ohlc.StockSummary, error) {
	data := o.rdb.HGetAll(ctx, fmt.Sprintf("stock_summary:%s:%s", stock, dayDate))
	err := data.Err()
	if err != nil {
		log.Println("errors.get stock summary")
		return nil, err
	}

	stockSummary := ohlc.StockSummary{}
	err = data.Scan(&stockSummary)
	if err != nil {
		log.Println("errors.parsing stock summary from redis")
		return nil, err
	}

	return &stockSummary, nil
}
