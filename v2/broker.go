package market

type Broker struct {
	tree   *RedBlackTree
	prices map[string]*OrderQueue
	Volume Decimal
	Len    int
	Depth  int
}

func NewBroker() *Broker {
	return &Broker{
		tree:   &RedBlackTree{},
		prices: map[string]*OrderQueue{},
		Volume: NewZeroDecimal(),
	}
}

// Add appends order to definite price level
func (m *Broker) Add(order *Order) *LinkedListElement {
	price := order.Price
	strPrice := price.String()

	queue, ok := m.prices[strPrice]
	if !ok {
		queue = NewQueue(order.Price)
		m.prices[strPrice] = queue
		m.tree.Put(price, queue)
		m.Depth++
	}

	m.Len++
	m.Volume = m.Volume.Add(order.Volume)
	return queue.Add(order)
}

// Remove removes order from definite price level
func (m *Broker) Remove(e *LinkedListElement) *Order {
	price := e.Order.Price
	strPrice := price.String()

	queue := m.prices[strPrice]
	o := queue.Remove(e)

	if queue.Len() == 0 {
		delete(m.prices, strPrice)
		m.tree.Remove(price)
		m.Depth--
	}

	m.Len--
	m.Volume = m.Volume.Sub(o.Volume)
	return o
}

// MaxPriceQueue returns maximal level of price
func (m *Broker) MaxPriceQueue() *OrderQueue {
	if m.Depth <= 0 {
		return nil
	}

	if value, found := m.tree.Max(); found {
		return value
	}

	return nil
}

// MinPriceQueue returns maximal level of price
func (m *Broker) MinPriceQueue() *OrderQueue {
	if m.Depth <= 0 {
		return nil
	}

	if value, found := m.tree.Min(); found {
		return value
	}

	return nil
}

// LessThan returns the nearest OrderQueue with price less than given
func (m *Broker) LessThan(price Decimal) *OrderQueue {
	tree := m.tree
	node := tree.Root

	var bottom *TreeNode
	for node != nil {
		if price.Cmp(node.Price) > 0 {
			bottom = node
			node = node.Right
		} else {
			node = node.Left
		}
	}

	if bottom != nil {
		return bottom.Queue
	}

	return nil
}

// GreaterThan returns the nearest OrderQueue with price greater than given
func (m *Broker) GreaterThan(price Decimal) *OrderQueue {
	tree := m.tree
	node := tree.Root

	var up *TreeNode
	for node != nil {
		if price.Cmp(node.Price) < 0 {
			up = node
			node = node.Left
		} else {
			node = node.Right
		}
	}

	if up != nil {
		return up.Queue
	}

	return nil
}
