package service

import (
	"reflect"
	"stock/ohlc"
	ohlc_mocks "stock/ohlc/mocks"
	"testing"

	sarama_mocks "github.com/IBM/sarama/mocks"
)

func TestUpdateSummaryFromExample(t *testing.T) {
	syncProcedure := sarama_mocks.NewSyncProducer(t, nil)
	defer func() {
		if err := syncProcedure.Close(); err != nil {
			t.Error(err)
		}
	}()

	ohlcRepository := ohlc_mocks.NewOHLCRepository(t)
	ohlcService := NewOHLCService(ohlcRepository, syncProcedure)

	pSummary := ohlc.StockSummary{}

	orderTxs := []ohlc.OrderTx{
		{
			Type:     "A",
			Stock:    "BBCA",
			Quantity: 0,
			Price:    8000,
		},
		{
			Type:     "P",
			Stock:    "BBCA",
			Quantity: 100,
			Price:    8050,
		},
		{
			Type:     "P",
			Stock:    "BBCA",
			Quantity: 500,
			Price:    7950,
		},
		{
			Type:     "A",
			Stock:    "BBCA",
			Quantity: 200,
			Price:    8150,
		},
		{
			Type:     "E",
			Stock:    "BBCA",
			Quantity: 300,
			Price:    8100,
		},
		{
			Type:     "A",
			Stock:    "BBCA",
			Quantity: 100,
			Price:    8200,
		},
	}

	for _, v := range orderTxs {
		ohlcService.UpdateSummary(&pSummary, &v)
	}

	wantSummary := ohlc.StockSummary{
		PreviousPrice: 8000,
		OpenPrice:     8050,
		HighestPrice:  8100,
		LowestPrice:   7950,
		ClosePrice:    8100,
		Volume:        900,
		Value:         7210000,
		AveragePrice:  8011,
	}

	if !reflect.DeepEqual(pSummary, wantSummary) {
		t.Errorf("errors.final summary is not correct, want: %v, not: %v", wantSummary, pSummary)
	}
}
