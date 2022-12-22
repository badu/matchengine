package matchengine_test

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"

	. "github.com/badu/matchengine"
)

const MaxPrice = 60

func createOrders(howMany uint32, meanPrice, standardPrice float64, maxAmount int32) (uint32, []*Order) {
	orders := make([]*Order, 0, howMany)
	maximumPrice := uint32(0)
	for i := uint32(0); i < howMany; i++ {
		absPrice := math.Abs(rand.NormFloat64()*standardPrice + meanPrice)

		price := uint32(absPrice)

		if price > maximumPrice {
			maximumPrice = price
		}

		volume := uint32(rand.Int31n(maxAmount))
		if volume <= 0 {
			volume = 1
		}

		if float64(price) >= meanPrice {
			orders = append(orders, NewBuy(i+1, price, volume))
			continue
		}

		orders = append(orders, NewSell(i+1, price, volume))
	}
	return maximumPrice + 1, orders
}

func performanceTest(howMany uint32, meanPrice, standardPrice float64, maxAmount int32) {
	maxPrice, orders := createOrders(howMany, meanPrice, standardPrice, maxAmount)
	actions := make(chan *Action)
	ctx, cancel := context.WithCancel(context.Background())
	start := time.Now()
	book := NewMarket(maxPrice, actions)

	fmt.Printf("%d orders took %s to construct bookkeeper.\n", howMany, time.Since(start))

	actionCount := 0
	go func() {
		for {
			action := <-actions
			if action.Type == DONE {
				cancel()
				return
			}
			actionCount++
		}
	}()

	start = time.Now()
	for _, order := range orders {
		if err := book.TakeOrder(order); err != nil {
			panic(err)
		}
	}
	book.Finish()

	<-ctx.Done()

	elapsed := time.Since(start)

	fmt.Printf("%d orders taken and matched %d actions that took %s at %d actions/second.\n", howMany, actionCount, elapsed, int(float64(actionCount)/elapsed.Seconds()))
}

func TestBookKeeping(t *testing.T) {
	actions := make(chan *Action)
	book := NewMarket(MaxPrice, actions)
	ctx, cancel := context.WithCancel(context.Background())
	got := make([]*Action, 0)
	go func() {
		for {
			select {
			case action := <-actions:
				got = append(got, action)
				if action.Type == DONE {
					cancel()
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	err := book.TakeOrder(NewSell(1, 50, 50))
	if err != nil {
		t.Fatalf("error : %#v", err)
	}
	err = book.TakeOrder(NewSell(2, 45, 25))
	if err != nil {
		t.Fatalf("error : %#v", err)
	}
	err = book.TakeOrder(NewSell(3, 45, 25))
	if err != nil {
		t.Fatalf("error : %#v", err)
	}
	// trigger three fills, two partial at 45 and one at 50
	err = book.TakeOrder(NewBuy(4, 55, 75))
	if err != nil {
		t.Fatalf("error : %#v", err)
	}

	// cancel immediately
	book.Cancel(1)

	err = book.TakeOrder(NewBuy(5, 55, 20))
	if err != nil {
		t.Fatalf("error : %#v", err)
	}
	err = book.TakeOrder(NewBuy(6, 50, 15))
	if err != nil {
		t.Fatalf("error : %#v", err)
	}

	// trigger two fills, one partial at 55 and one at 50
	err = book.TakeOrder(NewSell(7, 45, 25))
	if err != nil {
		t.Fatalf("error : %#v", err)
	}
	book.Finish()

	<-ctx.Done()

	expected := []*Action{
		{
			Type:    SELL,
			OrderId: 1,
			Volume:  50,
			Price:   50,
		},
		{
			Type:    SELL,
			OrderId: 2,
			Volume:  25,
			Price:   45,
		},
		{
			Type:    SELL,
			OrderId: 3,
			Volume:  25,
			Price:   45,
		},
		{
			Type:         BUY,
			OrderId:      4,
			BuyerOrderId: 0,
			Volume:       75,
			Price:        55,
		},
		{
			Type:         PARTIALLY_FILLED,
			OrderId:      4,
			BuyerOrderId: 2,
			Volume:       25,
			Price:        45,
		},
		{
			Type:         PARTIALLY_FILLED,
			OrderId:      4,
			BuyerOrderId: 3,
			Volume:       25,
			Price:        45,
		},
		{
			Type:         FILLED,
			OrderId:      4,
			BuyerOrderId: 1,
			Volume:       25,
			Price:        50,
		},
		{
			Type:    CANCEL,
			OrderId: 1,
		},
		{
			Type:    CANCELLED,
			OrderId: 1,
		},
		{
			Type:    BUY,
			OrderId: 5,
			Volume:  20,
			Price:   55,
		},
		{
			Type:    BUY,
			OrderId: 6,
			Volume:  15,
			Price:   50,
		},
		{
			Type:    SELL,
			OrderId: 7,
			Volume:  25,
			Price:   45,
		},
		{
			Type:         PARTIALLY_FILLED,
			OrderId:      7,
			BuyerOrderId: 5,
			Volume:       20,
			Price:        55,
		},
		{
			Type:         FILLED,
			OrderId:      7,
			BuyerOrderId: 6,
			Volume:       5,
			Price:        50,
		},
		{
			Type: DONE,
		},
	}

	if !reflect.DeepEqual(got, expected) {
		for i := range expected {
			if !reflect.DeepEqual(got[i], expected[i]) {
				t.Error("\n\nExpected at index ", i-1, ":\n\n", expected[i], "\n\nGot:\n\n", got[i], "\n\n")
			}
		}
	}
}

func TestPerformance(t *testing.T) {
	performanceTest(10_000, 5000, 10, 50)
	performanceTest(10_000, 5000, 1000, 5000)

	performanceTest(100_000, 5000, 10, 50)
	performanceTest(100_000, 5000, 1000, 5000)

	performanceTest(1_000_000, 5000, 10, 50)
	performanceTest(1_000_000, 5000, 1000, 5000)

	performanceTest(10_000_000, 5000, 10, 50)
	performanceTest(10_000_000, 5000, 1000, 5000)
}

func TestActionListener(t *testing.T) {
	actions := make(chan *Action)
	ctx, cancel := context.WithCancel(context.Background())

	go PrintActionListener(actions, cancel)

	book := NewMarket(MaxPrice, actions)
	err := book.TakeOrder(NewSell(1, 50, 50))
	if err != nil {
		t.Fatalf("error : %#v", err)
	}
	err = book.TakeOrder(NewSell(2, 45, 25))
	if err != nil {
		t.Fatalf("error : %#v", err)
	}
	err = book.TakeOrder(NewSell(3, 45, 25))
	if err != nil {
		t.Fatalf("error : %#v", err)
	}
	err = book.TakeOrder(NewBuy(4, 55, 75))
	if err != nil {
		t.Fatalf("error : %#v", err)
	}

	book.Cancel(1)

	book.Finish()

	<-ctx.Done()
}
