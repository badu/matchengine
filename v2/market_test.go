package market

import (
	"fmt"
	"testing"
	"time"
)

func TestBroker(t *testing.T) {
	broker := NewBroker()
	firstOrder := NewBuy(
		"buy-order-1",
		NewDecimalValue(10),
		NewDecimalValue(10),
		time.Now().UTC(),
	)
	secondOrder := NewBuy(
		"buy-order-2",
		NewDecimalValue(10),
		NewDecimalValue(20),
		time.Now().UTC(),
	)

	if broker.MinPriceQueue() != nil || broker.MaxPriceQueue() != nil {
		t.Fatal("invalid broker price levels")
	}

	firstOrderEl := broker.Add(firstOrder)
	if broker.MinPriceQueue() != broker.MaxPriceQueue() {
		t.Fatal("invalid broker price levels")
	}

	secondOrderEl := broker.Add(secondOrder)
	if broker.Depth != 2 {
		t.Fatal("invalid broker depth (should be two)")
	}
	if broker.Len != 2 {
		t.Fatal("invalid broker orders count")
	}
	if broker.MinPriceQueue().Head() != firstOrderEl || broker.MinPriceQueue().Tail() != firstOrderEl ||
		broker.MaxPriceQueue().Head() != secondOrderEl || broker.MaxPriceQueue().Tail() != secondOrderEl {
		t.Fatal("invalid broker price levels")
	}
	// removal test
	if order := broker.Remove(firstOrderEl); order != firstOrder {
		t.Fatal("invalid order on remove (should be first order)")
	}

	if broker.MinPriceQueue() != broker.MaxPriceQueue() {
		t.Fatal("invalid broker price levels")
	}
}

func TestBrokerPrices(t *testing.T) {
	broker := NewBroker()

	broker.Add(NewSell("sell-1", NewDecimalValue(5), NewDecimalValue(170), time.Now().UTC()))
	broker.Add(NewSell("sell-2", NewDecimalValue(5), NewDecimalValue(160), time.Now().UTC()))
	broker.Add(NewSell("sell-3", NewDecimalValue(5), NewDecimalValue(150), time.Now().UTC()))
	broker.Add(NewSell("sell-4", NewDecimalValue(5), NewDecimalValue(140), time.Now().UTC()))
	broker.Add(NewSell("sell-5", NewDecimalValue(5), NewDecimalValue(130), time.Now().UTC()))
	broker.Add(NewSell("sell-6", NewDecimalValue(5), NewDecimalValue(120), time.Now().UTC()))
	broker.Add(NewSell("sell-7", NewDecimalValue(5), NewDecimalValue(110), time.Now().UTC()))
	broker.Add(NewSell("sell-8", NewDecimalValue(5), NewDecimalValue(100), time.Now().UTC()))

	if !broker.Volume.Equal(NewDecimalValue(40)) {
		t.Fatal("invalid volume (should be 40)")
	}

	if !broker.LessThan(NewDecimalValue(101)).Price.Equal(NewDecimalValue(100)) ||
		!broker.LessThan(NewDecimalValue(150)).Price.Equal(NewDecimalValue(140)) ||
		broker.LessThan(NewDecimalValue(100)) != nil {
		t.Fatal("invalid price on less than")
	}

	if !broker.GreaterThan(NewDecimalValue(169)).Price.Equal(NewDecimalValue(170)) ||
		!broker.GreaterThan(NewDecimalValue(150)).Price.Equal(NewDecimalValue(160)) ||
		broker.GreaterThan(NewDecimalValue(170)) != nil {
		t.Fatal("invalid price on greater than")
	}

	less := broker.LessThan(NewDecimalValue(101))
	t.Logf("less : $%s %s pcs", less.Price, less.Volume)
	more := broker.GreaterThan(NewDecimalValue(169))
	t.Logf("more : $%s %s pcs", more.Price, more.Volume)

}

func TestBuySell(t *testing.T) {
	market := NewMarket()
	quantity := NewDecimalValue(2)
	for i := 50; i < 100; i = i + 10 {
		done, partial, partialVolume, err := market.ProcessBuyOrder(fmt.Sprintf("buy-order-%d", i), quantity, NewDecimalValue(int64(i)))
		if len(done) != 0 {
			t.Fatal("market failed to process buy order (done is not empty)")
		}
		if partial != nil {
			t.Fatal("market failed to process buy order (partial is not empty)")
		}
		if partialVolume.Sign() != 0 {
			t.Fatal("market failed to process buy order (partialVolume is not zero)")
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	for i := 100; i < 150; i = i + 10 {
		done, partial, partialVolume, err := market.ProcessSellOrder(fmt.Sprintf("sell-order-%d", i), quantity, NewDecimalValue(int64(i)))
		if len(done) != 0 {
			t.Fatal("market failed to process sell order (done is not empty)")
		}
		if partial != nil {
			t.Fatal("market failed to process sell order (partial is not empty)")
		}
		if partialVolume.Sign() != 0 {
			t.Fatal("market failed to process sell order (partialVolume is not zero)")
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	if market.Order("undeclared-order") != nil {
		t.Fatal("should not have this order")
	}

	if market.Order("sell-order-100") == nil {
		t.Fatal("can't get real order")
	}

	sales, buys := market.Depth()
	for _, sale := range sales {
		t.Logf("TestBuySell sale : $%v = %v pcs", sale.Price, sale.Volume)
	}
	for _, buy := range buys {
		t.Logf("TestBuySell buy : $%v = %v pcs", buy.Price, buy.Volume)
	}
}

func TestPartialBuySell(t *testing.T) {
	market := NewMarket()
	createProcesses(market, NewDecimalValue(2), "")

	done, partial, partialVolume, err := market.ProcessBuyOrder("buy-order-100", NewDecimalValue(1), NewDecimalValue(100))
	if err != nil {
		t.Fatal(err)
	}

	for _, order := range done {
		t.Logf("done order %s $%s %s pcs", order.ID, order.Price, order.Volume)
	}

	if done[0].ID != "buy-order-100" {
		t.Fatal("wrong order id")
	}

	t.Logf("partial order : %s $%s %s pcs", partial.ID, partial.Price, partial.Volume)
	if partial.ID != "sell-100" {
		t.Fatal("wrong partial order id")
	}

	if !partialVolume.Equal(NewDecimalValue(1)) {
		t.Fatal("wrong partial volume processed")
	}

	done, partial, partialVolume, err = market.ProcessBuyOrder("buy-order-150", NewDecimalValue(10), NewDecimalValue(150))
	if err != nil {
		t.Fatal(err)
	}

	for _, order := range done {
		t.Logf("TestPartialBuySell : done order %s $%s %s pcs", order.ID, order.Price, order.Volume)
	}

	if len(done) != 5 {
		t.Fatal("wrong done len")
	}

	t.Logf("partial order : %s $%s %s pcs", partial.ID, partial.Price, partial.Volume)
	if partial.ID != "buy-order-150" {
		t.Fatal("wrong partial id")
	}

	if !partialVolume.Equal(NewDecimalValue(9)) {
		t.Fatal("wrong partial volume processed", partialVolume)
	}

	if _, _, _, err := market.ProcessSellOrder("buy-70", NewDecimalValue(11), NewDecimalValue(40)); err == nil {
		t.Fatal("should not be possible to process existing order")
	}

	if _, _, _, err := market.ProcessSellOrder("empty-volume-70", NewDecimalValue(0), NewDecimalValue(40)); err == nil {
		t.Fatal("should not be possible to add empty volume")
	}

	if _, _, _, err := market.ProcessSellOrder("zero-price-70", NewDecimalValue(10), NewDecimalValue(0)); err == nil {
		t.Fatal("should not be possible to add zero prices")
	}

	if o := market.CancelOrder("buy-order-100"); o != nil {
		t.Fatal("should not be possible to cancel order that is done")
	}

	done, partial, partialVolume, err = market.ProcessSellOrder("order-s40", NewDecimalValue(11), NewDecimalValue(40))
	if err != nil {
		t.Fatal(err)
	}

	if len(done) != 7 {
		t.Fatal("wrong done volume")
	}

	if partial != nil {
		t.Fatal("wrong partial")
	}

	if partialVolume.Sign() != 0 {
		t.Fatal("wrong partialVolume")
	}

	for _, order := range done {
		t.Logf("TestPartialBuySell : done order %s $%s %s pcs", order.ID, order.Price, order.Volume)
	}
}

func TestPrices(t *testing.T) {
	market := NewMarket()
	createProcesses(market, NewDecimalValue(5), "05-")
	createProcesses(market, NewDecimalValue(10), "10-")
	createProcesses(market, NewDecimalValue(15), "15-")

	price, err := market.MakeBuyPrice(NewDecimalValue(115))
	if err != nil {
		t.Fatal(err)
	}

	if !price.Equal(NewDecimalValue(13150)) {
		t.Fatal("invalid price", price)
	}

	price, err = market.MakeBuyPrice(NewDecimalValue(200))
	if err == nil {
		t.Fatal("invalid prices count")
	}

	if !price.Equal(NewDecimalValue(18000)) {
		t.Fatal("invalid price", price)
	}

	price, err = market.MakeSellPrice(NewDecimalValue(115))
	if err != nil {
		t.Fatal(err)
	}

	if !price.Equal(NewDecimalValue(8700)) {
		t.Fatal("invalid price", price)
	}

	price, err = market.MakeSellPrice(NewDecimalValue(200))
	if err == nil {
		t.Fatal("invalid quantity count")
	}

	if !price.Equal(NewDecimalValue(10500)) {
		t.Fatal("invalid price", price)
	}
}

func TestMarketProcess(t *testing.T) {
	market := NewMarket()
	createProcesses(market, NewDecimalValue(2), "")

	done, partial, partialVolume, left, err := market.ProcessBuy(NewDecimalValue(3))
	if err != nil {
		t.Fatal(err)
	}

	if left.Sign() > 0 {
		t.Fatal("wrong volume left")
	}

	if !partialVolume.Equal(NewDecimalValue(1)) {
		t.Fatal("wrong partial volume left")
	}

	if _, _, _, _, err := market.ProcessBuy(NewDecimalValue(0)); err == nil {
		t.Fatal("should not be possible to add zero volume")
	}
	for _, order := range done {
		t.Logf("TestMarketProcess : done order %s $%s %s pcs", order.ID, order.Price, order.Volume)
	}

	done, partial, partialVolume, left, err = market.ProcessSell(NewDecimalValue(12))
	if err != nil {
		t.Fatal(err)
	}

	if partial != nil {
		t.Fatal("partial should be nil")
	}

	if partialVolume.Sign() != 0 {
		t.Fatal("partial volume should be zero")
	}

	if len(done) != 5 {
		t.Fatal("invalid done length")
	}

	if !left.Equal(NewDecimalValue(2)) {
		t.Fatal("invalid left value", left)
	}
	for _, order := range done {
		t.Logf("TestMarketProcess : done order %s $%s %s pcs", order.ID, order.Price, order.Volume)
	}
}
