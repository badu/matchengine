package matchengine

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Status int

const (
	NEW Status = iota
	OPEN
	BUY
	SELL
	PARTIAL
	FILLED
	PARTIALLY_FILLED
	CANCEL
	CANCELLED
	DONE
)

func (s Status) String() string {
	switch s {
	case NEW:
		return "new"
	case OPEN:
		return "open"
	case BUY:
		return "buy"
	case SELL:
		return "sell"
	case PARTIAL:
		return "partial"
	case FILLED:
		return "filled"
	case PARTIALLY_FILLED:
		return "partially filled"
	case CANCEL:
		return "cancel"
	case CANCELLED:
		return "cancelled"
	case DONE:
		return "done"
	default:
		return "unknown status"
	}
}

type Action struct {
	Type         Status
	OrderId      uint32
	BuyerOrderId uint32
	Volume       uint32
	Price        uint32
}

func (a Action) String() string {
	var sb strings.Builder
	sb.WriteString("Action : ")
	sb.WriteString(a.Type.String())
	sb.WriteRune(' ')

	sb.WriteString("order : ")
	sb.WriteString(strconv.Itoa(int(a.OrderId)))
	sb.WriteRune(' ')

	if a.BuyerOrderId > 0 {
		sb.WriteString("from order : ")
		sb.WriteString(strconv.Itoa(int(a.BuyerOrderId)))
		sb.WriteRune(' ')
	}

	sb.WriteString("volume : ")
	sb.WriteString(strconv.Itoa(int(a.Volume)))
	sb.WriteString(" at ")
	sb.WriteRune('$')
	sb.WriteString(strconv.Itoa(int(a.Price)))
	return sb.String()
}

func NewBuyAction(order *Order) *Action {
	return &Action{Type: BUY, OrderId: order.Id, Volume: order.Volume, Price: order.Price}
}

func NewSellAction(order *Order) *Action {
	return &Action{Type: SELL, OrderId: order.Id, Volume: order.Volume, Price: order.Price}
}

func NewCancelAction(id uint32) *Action {
	return &Action{Type: CANCEL, OrderId: id}
}

func NewCancelledAction(id uint32) *Action {
	return &Action{Type: CANCELLED, OrderId: id}
}

func NewPartialFilledAction(order, fromOrder *Order) *Action {
	return &Action{Type: PARTIALLY_FILLED, OrderId: order.Id, BuyerOrderId: fromOrder.Id, Volume: fromOrder.Volume, Price: fromOrder.Price}
}

func NewFilledAction(order, fromOrder *Order) *Action {
	return &Action{Type: FILLED, OrderId: order.Id, BuyerOrderId: fromOrder.Id, Volume: order.Volume, Price: fromOrder.Price}
}

func NewDoneAction() *Action {
	return &Action{Type: DONE}
}

func PrintActionListener(actionsCh <-chan *Action, cancelFunc context.CancelFunc) {
	for {
		switch a := <-actionsCh; a.Type {
		case BUY, SELL, CANCEL, CANCELLED, PARTIALLY_FILLED, FILLED:
			fmt.Printf("%s\n", a)
		case DONE:
			fmt.Printf("%s!\n", a.Type)
			cancelFunc()
			return
		default:
			panic("Unknown action type")
		}
	}
}

type heap struct {
	head *Order
	tail *Order
}

func (p *heap) Insert(order *Order) {
	if p.head == nil {
		p.head = order
		p.tail = order
		return
	}

	p.tail.next = order
	p.tail = order
}

type Order struct {
	next   *Order
	Id     uint32
	Price  uint32
	Volume uint32
	Status Status
	IsBuy  bool
}

func (o Order) String() string {
	var sb strings.Builder
	sb.WriteString("Order ID :")
	sb.WriteString(strconv.Itoa(int(o.Id)))
	sb.WriteRune(' ')
	if o.IsBuy {
		sb.WriteString("BUY")
	} else {
		sb.WriteString("SELL")
	}
	sb.WriteRune(' ')
	sb.WriteString("volume : ")
	sb.WriteString(strconv.Itoa(int(o.Volume)))
	sb.WriteRune(' ')
	sb.WriteRune('$')
	sb.WriteString(strconv.Itoa(int(o.Price)))
	return sb.String()
}

func NewSell(id, price, amount uint32) *Order {
	return &Order{Id: id, Price: price, Volume: amount, Status: NEW}
}

func NewBuy(id, price, amount uint32) *Order {
	return &Order{Id: id, IsBuy: true, Price: price, Volume: amount, Status: NEW}
}

type Bookkeeping struct {
	prices    []*heap           //  linked list
	orders    map[uint32]*Order // keep a map of orders to find them for cancel
	actionsCh chan<- *Action    // main communication channel
	asks      uint32            // estimates of the lowest ask price and the highest bid price currently in the order book
	bids      uint32            //
}

func NewBookkeeper(maximumPrice uint32, actionsCh chan<- *Action) *Bookkeeping {
	result := Bookkeeping{
		asks:      maximumPrice,
		actionsCh: actionsCh,
		orders:    make(map[uint32]*Order),
	}
	result.prices = make([]*heap, maximumPrice)
	for i := int(maximumPrice) - 1; i >= 0; i-- {
		result.prices[uint32(i)] = &heap{}
	}
	return &result
}

func (b *Bookkeeping) TakeOrder(order *Order) error {
	if order.Volume <= 0 {
		return errors.New("volume cannot be less or equal to zero")
	}

	// attempt to fill immediately
	if order.IsBuy {
		b.actionsCh <- NewBuyAction(order)
		b.Buy(order)
	} else {
		b.actionsCh <- NewSellAction(order)
		b.Sell(order)
	}

	b.open(order)
	return nil
}

func (b *Bookkeeping) open(order *Order) {
	prices := b.prices[order.Price]
	prices.Insert(order)
	order.Status = OPEN
	b.orders[order.Id] = order

	if order.IsBuy && order.Price > b.bids {
		b.bids = order.Price
		return
	}

	if !order.IsBuy && order.Price < b.asks {
		b.asks = order.Price
	}
}

func (b *Bookkeeping) Buy(order *Order) {
	for b.asks < order.Price && order.Volume > 0 {
		prices := b.prices[b.asks]
		head := prices.head
		for head != nil {
			b.fill(order, head)
			head = head.next
			prices.head = head
		}
		b.asks++
	}
}

func (b *Bookkeeping) Sell(order *Order) {
	for b.bids >= order.Price && order.Volume > 0 {
		prices := b.prices[b.bids]
		head := prices.head
		for head != nil {
			b.fill(order, head)
			head = head.next
			prices.head = head
		}
		b.bids--
	}
}

func (b *Bookkeeping) fill(order, fromOrder *Order) {
	if fromOrder.Volume >= order.Volume {
		b.actionsCh <- NewFilledAction(order, fromOrder)
		fromOrder.Volume -= order.Volume
		order.Volume = 0
		order.Status = FILLED
		return
	}

	// partial fill
	if fromOrder.Volume > 0 {
		b.actionsCh <- NewPartialFilledAction(order, fromOrder)
		order.Volume -= fromOrder.Volume
		order.Status = PARTIAL
		fromOrder.Volume = 0
	}
}

func (b *Bookkeeping) Cancel(orderID uint32) {
	b.actionsCh <- NewCancelAction(orderID)
	if o, ok := b.orders[orderID]; ok {
		// If this is the last order at a particular price point we need to update the bid / ask!
		o.Volume = 0
		o.Status = CANCELLED
		// TODO : check if has already a match and cancel that as well
		delete(b.orders, orderID)
	}
	b.actionsCh <- NewCancelledAction(orderID)
}

func (b *Bookkeeping) Finish() {
	b.actionsCh <- NewDoneAction()
}
