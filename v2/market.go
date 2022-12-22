package market

import (
	"errors"
	"time"
)

type Market struct {
	orders map[string]*LinkedListElement // orderID -> *Order (via *LinkedListElement.Order)
	sales  *Broker                       // sales (ask) manager
	buys   *Broker                       // buys (bids) manager
}

func NewMarket() *Market {
	return &Market{
		orders: map[string]*LinkedListElement{},
		buys:   NewBroker(),
		sales:  NewBroker(),
	}
}

func (m *Market) Order(orderID string) *Order {
	result, ok := m.orders[orderID]
	if !ok {
		return nil
	}

	return result.Order
}

func (m *Market) CancelOrder(orderID string) *Order {
	order, ok := m.orders[orderID]
	if !ok {
		return nil
	}

	delete(m.orders, orderID)

	if order.Order.Kind == Buy {
		return m.buys.Remove(order)
	}

	return m.sales.Remove(order)
}

type PriceVolume struct {
	Price  Decimal
	Volume Decimal
}

// Depth returns price levels and volumes at price level
func (m *Market) Depth() ([]PriceVolume, []PriceVolume) {
	var sales []PriceVolume
	var buys []PriceVolume

	level := m.sales.MaxPriceQueue()
	for level != nil {
		sales = append(sales, PriceVolume{Price: level.Price, Volume: level.Volume})
		level = m.sales.LessThan(level.Price)
	}

	level = m.buys.MaxPriceQueue()
	for level != nil {
		buys = append(buys, PriceVolume{Price: level.Price, Volume: level.Volume})
		level = m.buys.LessThan(level.Price)
	}

	return sales, buys
}

// MakeBuyPrice returns total market buy price for requested volume
func (m *Market) MakeBuyPrice(volume Decimal) (Decimal, error) {
	price := NewZeroDecimal()
	level := m.sales.MinPriceQueue()

	for volume.Sign() > 0 && level != nil {
		if volume.GreaterThanOrEqual(level.Volume) {
			price = price.Add(level.Price.Mul(level.Volume))
			volume = volume.Sub(level.Volume)
			level = m.sales.GreaterThan(level.Price)
			continue
		}

		price = price.Add(level.Price.Mul(volume))
		volume = NewZeroDecimal()

	}

	if volume.Sign() > 0 {
		return price, errors.New("insufficient volume to calculate buy price")
	}

	return price, nil
}

// MakeSellPrice returns total market sell price for requested volume
func (m *Market) MakeSellPrice(volume Decimal) (Decimal, error) {
	price := NewZeroDecimal()
	level := m.buys.MaxPriceQueue()

	for volume.Sign() > 0 && level != nil {
		if volume.GreaterThanOrEqual(level.Volume) {
			price = price.Add(level.Price.Mul(level.Volume))
			volume = volume.Sub(level.Volume)
			level = m.buys.LessThan(level.Price)
			continue
		}

		price = price.Add(level.Price.Mul(volume))
		volume = NewZeroDecimal()

	}

	if volume.Sign() > 0 {
		return price, errors.New("insufficient volume to calculate sell price")
	}

	return price, nil
}

type Processed struct {
	PartialVolume Decimal
	VolumeLeft    Decimal
	Partial       *Order
	Done          []*Order
}

// processQueue processes the indicated queue for volume value
func (m *Market) processQueue(queue *OrderQueue, volume Decimal) Processed {
	result := Processed{VolumeLeft: volume}

	for queue.Len() > 0 && result.VolumeLeft.Sign() > 0 {
		headOrderEl := queue.Head()
		headOrder := headOrderEl.Order

		if result.VolumeLeft.LessThan(headOrder.Volume) {
			if headOrder.Kind == Buy {
				result.Partial = NewBuy(headOrder.ID, headOrder.Volume.Sub(result.VolumeLeft), headOrder.Price, headOrder.Time)
			} else {
				result.Partial = NewSell(headOrder.ID, headOrder.Volume.Sub(result.VolumeLeft), headOrder.Price, headOrder.Time)
			}
			result.PartialVolume = result.VolumeLeft
			queue.Update(headOrderEl, result.Partial)
			result.VolumeLeft = NewZeroDecimal()
			continue
		}

		// done offers
		result.VolumeLeft = result.VolumeLeft.Sub(headOrder.Volume)
		result.Done = append(result.Done, m.CancelOrder(headOrder.ID))
	}

	return result
}

// ProcessBuyOrder places new buy order to the Market
//
//	orderID - unique order ID
//	volume - how much volume you want to buy
//	price  - no more expensive this price
//
// Result :
//
//	 A slice of 'done' orders - if your order satisfies another order, these orders will be added to this slice.
//	 If your own order is 'done' too, it will be placed into this slice as well
//		partial - if your order has been 'done' but the top order is not fully done, or if your order is
//		          'partial done' and placed to the market without full volume - partial will contain your order with volume left
//		partialVolume - if partial order is not nil this result contains processed volume from partial order
//		error   - not nil if volume (or price) is less or equal 0. Or if order with given ID is exists
func (m *Market) ProcessBuyOrder(orderID string, volume, price Decimal) ([]*Order, *Order, Decimal, error) {
	if _, ok := m.orders[orderID]; ok {
		return nil, nil, NewZeroDecimal(), errors.New("order already exists")
	}

	if volume.Sign() <= 0 {
		return nil, nil, NewZeroDecimal(), errors.New("invalid order volume")
	}

	if price.Sign() <= 0 {
		return nil, nil, NewZeroDecimal(), errors.New("invalid order price")
	}

	var (
		done          []*Order
		partial       *Order
		partialVolume Decimal
	)

	volumeLeft := volume
	bestPrice := m.sales.MinPriceQueue()
	for volumeLeft.Sign() > 0 && m.sales.Len > 0 && price.GreaterThanOrEqual(bestPrice.Price) {
		processed := m.processQueue(bestPrice, volumeLeft)
		done = append(done, processed.Done...)
		partial = processed.Partial
		partialVolume = processed.PartialVolume
		volumeLeft = processed.VolumeLeft
		bestPrice = m.sales.MinPriceQueue()
	}

	if volumeLeft.Sign() > 0 {
		order := NewBuy(orderID, volumeLeft, price, time.Now().UTC())

		if len(done) > 0 {
			partialVolume = volume.Sub(volumeLeft)
			partial = order
		}
		m.orders[orderID] = m.buys.Add(order)
		return done, partial, partialVolume, nil
	}

	totalVolume := NewZeroDecimal()
	totalPrice := NewZeroDecimal()

	for _, order := range done {
		totalVolume = totalVolume.Add(order.Volume)
		totalPrice = totalPrice.Add(order.Price.Mul(order.Volume))
	}

	if partialVolume.Sign() > 0 {
		totalVolume = totalVolume.Add(partialVolume)
		totalPrice = totalPrice.Add(partial.Price.Mul(partialVolume))
	}

	done = append(done, NewBuy(orderID, volume, totalPrice.Div(totalVolume), time.Now().UTC()))
	return done, partial, partialVolume, nil
}

// ProcessSellOrder places new sell order to the Market
//
//	orderID - unique order ID
//	volume - how much volume you want to sell
//	price - no less cheap than this price
//
// Result :
//
//	 A slice of 'done' orders - if your order satisfies another order, these orders will be added to this slice.
//	 If your own order is 'done' too, it will be placed into this slice as well
//		partial - if your order has been 'done' but the top order is not fully done, or if your order is
//		          'partial done' and placed to the market without full volume - partial will contain your order with volume left
//		partialVolume - if partial order is not nil this result contains processed volume from partial order
//		error   - not nil if volume (or price) is less or equal 0. Or if order with given ID is exists
func (m *Market) ProcessSellOrder(orderID string, volume, price Decimal) ([]*Order, *Order, Decimal, error) {
	if _, ok := m.orders[orderID]; ok {
		return nil, nil, NewZeroDecimal(), errors.New("order already exists")
	}

	if volume.Sign() <= 0 {
		return nil, nil, NewZeroDecimal(), errors.New("invalid order volume")
	}

	if price.Sign() <= 0 {
		return nil, nil, NewZeroDecimal(), errors.New("invalid order price")
	}

	var (
		done          []*Order
		partial       *Order
		partialVolume Decimal
	)

	volumeLeft := volume
	bestPrice := m.buys.MaxPriceQueue()
	for volumeLeft.Sign() > 0 && m.buys.Len > 0 && price.LessThanOrEqual(bestPrice.Price) {
		processed := m.processQueue(bestPrice, volumeLeft)
		done = append(done, processed.Done...)
		partial = processed.Partial
		partialVolume = processed.PartialVolume
		volumeLeft = processed.VolumeLeft
		bestPrice = m.buys.MaxPriceQueue()
	}

	if volumeLeft.Sign() > 0 {
		order := NewSell(orderID, volumeLeft, price, time.Now().UTC())

		if len(done) > 0 {
			partialVolume = volume.Sub(volumeLeft)
			partial = order
		}
		m.orders[orderID] = m.sales.Add(order)
		return done, partial, partialVolume, nil
	}

	totalVolume := NewZeroDecimal()
	totalPrice := NewZeroDecimal()

	for _, order := range done {
		totalVolume = totalVolume.Add(order.Volume)
		totalPrice = totalPrice.Add(order.Price.Mul(order.Volume))
	}

	if partialVolume.Sign() > 0 {
		totalVolume = totalVolume.Add(partialVolume)
		totalPrice = totalPrice.Add(partial.Price.Mul(partialVolume))
	}

	done = append(done, NewSell(orderID, volume, totalPrice.Div(totalVolume), time.Now().UTC()))
	return done, partial, partialVolume, nil
}

// ProcessSell - sells a volume
func (m *Market) ProcessSell(volume Decimal) ([]*Order, *Order, Decimal, Decimal, error) {

	if volume.Sign() <= 0 {
		return nil, nil, NewZeroDecimal(), NewZeroDecimal(), errors.New("invalid volume")
	}

	var done []*Order
	var partial *Order
	var partialVolume, volumeLeft Decimal

	for volume.Sign() > 0 && m.buys.Len > 0 {
		bestPrice := m.buys.MaxPriceQueue()
		processed := m.processQueue(bestPrice, volume)
		done = append(done, processed.Done...)
		partial = processed.Partial
		partialVolume = processed.PartialVolume
		volume = processed.VolumeLeft
	}

	volumeLeft = volume
	return done, partial, partialVolume, volumeLeft, nil
}

// ProcessBuy - buys for a volume
func (m *Market) ProcessBuy(volume Decimal) ([]*Order, *Order, Decimal, Decimal, error) {
	if volume.Sign() <= 0 {
		return nil, nil, NewZeroDecimal(), NewZeroDecimal(), errors.New("invalid volume")
	}

	var done []*Order
	var partial *Order
	var partialVolume, volumeLeft Decimal

	for volume.Sign() > 0 && m.sales.Len > 0 {
		bestPrice := m.sales.MinPriceQueue()
		processed := m.processQueue(bestPrice, volume)
		done = append(done, processed.Done...)
		partial = processed.Partial
		partialVolume = processed.PartialVolume
		volume = processed.VolumeLeft
	}

	volumeLeft = volume
	return done, partial, partialVolume, volumeLeft, nil
}
