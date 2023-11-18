package handler

import (
	"context"
	"log"
	"stock/ohlc/service"
	"stock/pb"
	"strconv"
)

func NewOHLCgRPCHandler(ohlcService service.OHLCService) pb.OHLCServer {
	return &ohlcGrpcHandler{
		ohlcService: ohlcService,
	}
}

type ohlcGrpcHandler struct {
	pb.UnimplementedOHLCServer

	ohlcService service.OHLCService
}

func (o *ohlcGrpcHandler) GetFeederStatus(ctx context.Context, _ *pb.GetFeederStatusRequest) (*pb.GetFeederStatusResponse, error) {
	status, err := o.ohlcService.GetFeederStatus(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GetFeederStatusResponse{
		Status: status,
	}, nil
}

func (o *ohlcGrpcHandler) GetStockList(ctx context.Context, _ *pb.GetStockListRequest) (*pb.GetStockListResponse, error) {
	stockItems, err := o.ohlcService.GetStockList(ctx)
	if err != nil {
		return nil, err
	}

	pbStockItems := []*pb.StockList{}
	for _, v := range stockItems {
		pbStockItems = append(pbStockItems, &pb.StockList{
			Stock:   v.Stock,
			DayDate: v.DayDate,
		})
	}

	return &pb.GetStockListResponse{
		List: pbStockItems,
	}, nil
}

func (o *ohlcGrpcHandler) GetStockSummary(ctx context.Context, request *pb.GetStockSummaryRequest) (*pb.GetStockSummaryResponse, error) {
	stock := request.Stock
	dayDate := request.DayDate

	stockSummary, err := o.ohlcService.GetStockSummary(ctx, stock, dayDate)
	if err != nil {
		return nil, err
	}

	return &pb.GetStockSummaryResponse{
		PreviousPrice: strconv.Itoa(stockSummary.PreviousPrice),
		OpenPrice:     strconv.Itoa(stockSummary.OpenPrice),
		HighestPrice:  strconv.Itoa(stockSummary.HighestPrice),
		LowestPrice:   strconv.Itoa(stockSummary.LowestPrice),
		ClosePrice:    strconv.Itoa(stockSummary.ClosePrice),
		AveragePrice:  strconv.Itoa(stockSummary.AveragePrice),
		Volume:        strconv.Itoa(stockSummary.Volume),
		Value:         strconv.Itoa(stockSummary.Value),
	}, nil
}

func (o *ohlcGrpcHandler) Feed(ctx context.Context, request *pb.FeedRequest) (*pb.FeedResponse, error) {
	go func() {
		err := o.ohlcService.Feeder()
		if err != nil {
			log.Println("errors.when feeding data")
		}
	}()

	response := &pb.FeedResponse{}
	return response, nil
}
