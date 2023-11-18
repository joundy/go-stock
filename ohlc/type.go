package ohlc

type OrderTx struct {
	DayDate  string
	Type     string
	Stock    string
	Quantity int
	Price    int
}

type StockItem struct {
	Stock   string
	DayDate string
}

type StockSummary struct {
	PreviousPrice int `redis:"previous_price"`
	OpenPrice     int `redis:"open_price"`
	HighestPrice  int `redis:"highest_price"`
	LowestPrice   int `redis:"lowest_price"`
	ClosePrice    int `redis:"close_price"`
	AveragePrice  int `redis:"average_price"`
	Volume        int `redis:"volume"`
	Value         int `redis:"value"`
}
