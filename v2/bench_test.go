package market

import (
	"fmt"
	"testing"
	"time"
)

func createProcesses(bookKeeper *Market, quantity Decimal, namePrefix string) {
	for i := 50; i < 100; i = i + 10 {
		_, _, _, err := bookKeeper.ProcessBuyOrder(fmt.Sprintf("%sbuy-%d", namePrefix, i), quantity, NewDecimalValue(int64(i)))
		if err != nil {
			panic("error : " + err.Error())
		}
	}

	for i := 100; i < 150; i = i + 10 {
		_, _, _, err := bookKeeper.ProcessSellOrder(fmt.Sprintf("%ssell-%d", namePrefix, i), quantity, NewDecimalValue(int64(i)))
		if err != nil {
			panic("error : " + err.Error())
		}
	}
}

func BenchmarkBroker(b *testing.B) {
	manager := NewBroker()
	stopwatch := time.Now()

	var o *Order
	for i := 0; i < b.N; i++ {
		o = NewBuy(
			fmt.Sprintf("order-%d", i),
			NewDecimalValue(10),
			NewDecimalValue(int64(i)),
			stopwatch,
		)
		manager.Add(o)
	}
	elapsed := time.Since(stopwatch)
	fmt.Printf("elapsed: %s -\t\t\t %d runs added %f to manager per second\n", elapsed, b.N, float64(b.N)/elapsed.Seconds())
}

func BenchmarkMarket(b *testing.B) {
	market := NewMarket()
	stopwatch := time.Now()
	for i := 0; i < b.N; i++ {
		createProcesses(market, NewDecimalValue(5), "005-")
		createProcesses(market, NewDecimalValue(10), "010-")
		createProcesses(market, NewDecimalValue(15), "015-")
		_, _, _, _ = market.ProcessBuyOrder("buy-order-150", NewDecimalValue(160), NewDecimalValue(150))
		_, _, _, _, _ = market.ProcessSell(NewDecimalValue(200))
	}
	elapsed := time.Since(stopwatch)
	fmt.Printf("elapsed: %s -\t\t\t %d runs produced %f (avg) transactions per second\n", elapsed, b.N, float64(b.N*32)/elapsed.Seconds())
}

func BenchmarkQueue(b *testing.B) {
	forPrice := NewDecimalValue(100)
	queue := NewQueue(forPrice)
	stopwatch := time.Now()

	var order *Order
	for i := 0; i < b.N; i++ {
		orderName := fmt.Sprintf("buy-order-%d", i)
		order = NewBuy(
			orderName,
			NewDecimalValue(100),
			NewDecimalValue(int64(i)),
			stopwatch,
		)
		queue.Add(order)
	}
	elapsed := time.Since(stopwatch)
	fmt.Printf("elapsed: %s -\t\t\t %d run(s) added %f orders to queue per second\n", elapsed, b.N, float64(b.N)/elapsed.Seconds())
}
