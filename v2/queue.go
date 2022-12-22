package market

import (
	"time"
)

type LinkedListElement struct {
	next   *LinkedListElement
	prev   *LinkedListElement
	parent *LinkedList
	Order  *Order
}

// Next returns the next list element or nil.
func (e *LinkedListElement) Next() *LinkedListElement {
	if p := e.next; e.parent != nil && p != &e.parent.root {
		return p
	}
	return nil
}

// Prev returns the previous list element or nil.
func (e *LinkedListElement) Prev() *LinkedListElement {
	if p := e.prev; e.parent != nil && p != &e.parent.root {
		return p
	}
	return nil
}

// LinkedList represents a doubly linked list.
// The zero value for LinkedList is an empty list ready to use.
type LinkedList struct {
	root LinkedListElement // sentinel list element, only &root, root.prev, and root.next are used
	Len  int               // current list length excluding (this) sentinel element
}

// Init initializes or clears list l.
func (l *LinkedList) Init() *LinkedList {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.Len = 0
	return l
}

// NewList returns an initialized list.
func NewList() *LinkedList {
	return new(LinkedList).Init()
}

// Front returns the first element of list l or nil if the list is empty.
func (l *LinkedList) Front() *LinkedListElement {
	if l.Len == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
func (l *LinkedList) Back() *LinkedListElement {
	if l.Len == 0 {
		return nil
	}
	return l.root.prev
}

// lazyInit lazily initializes a zero LinkedList value.
func (l *LinkedList) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

// insert inserts e after at, increments l.Len, and returns e.
func (l *LinkedList) insert(e, at *LinkedListElement) *LinkedListElement {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	e.parent = l
	l.Len++
	return e
}

// insertValue is a convenience wrapper for insert(&LinkedListElement{OrderQueue: v}, at).
func (l *LinkedList) insertValue(v *Order, at *LinkedListElement) *LinkedListElement {
	return l.insert(&LinkedListElement{Order: v}, at)
}

// remove removes e from its list, decrements l.Len
func (l *LinkedList) remove(e *LinkedListElement) {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	e.parent = nil
	l.Len--
}

// Remove removes e from l if e is an element of list l.
// It returns the element value e.OrderQueue.
// The element must not be nil.
func (l *LinkedList) Remove(e *LinkedListElement) any {
	if e.parent == l {
		// if e.list == l, l must have been initialized when e was inserted
		// in l or l == nil (e is a zero LinkedListElement) and l.remove will crash
		l.remove(e)
	}
	return e.Order
}

// Append inserts a new element e with value v at the back of list l and returns e.
func (l *LinkedList) Append(v *Order) *LinkedListElement {
	l.lazyInit()
	return l.insertValue(v, l.root.prev)
}

// OrderQueue stores and manage chain of orders
type OrderQueue struct {
	Volume Decimal
	Price  Decimal
	orders *LinkedList
}

// NewQueue creates and initialize OrderQueue object
func NewQueue(price Decimal) *OrderQueue {
	return &OrderQueue{
		Price:  price,
		Volume: NewZeroDecimal(),
		orders: NewList(),
	}
}

// Len returns amount of orders in queue
func (q *OrderQueue) Len() int {
	return q.orders.Len
}

// Head returns top order in queue
func (q *OrderQueue) Head() *LinkedListElement {
	return q.orders.Front()
}

// Tail returns bottom order in queue
func (q *OrderQueue) Tail() *LinkedListElement {
	return q.orders.Back()
}

// Add adds order to tail of the queue
func (q *OrderQueue) Add(order *Order) *LinkedListElement {
	q.Volume = q.Volume.Add(order.Volume)
	return q.orders.Append(order)
}

// Update sets up new order to list value
func (q *OrderQueue) Update(element *LinkedListElement, order *Order) *LinkedListElement {
	q.Volume = q.Volume.Sub(element.Order.Volume)
	q.Volume = q.Volume.Add(order.Volume)
	element.Order = order
	return element
}

// Remove removes order from the queue and link order chain
func (q *OrderQueue) Remove(e *LinkedListElement) *Order {
	q.Volume = q.Volume.Sub(e.Order.Volume)
	return q.orders.Remove(e).(*Order)
}

type Kind int

const (
	Sell Kind = iota
	Buy
)

type Order struct {
	Time   time.Time
	ID     string
	Volume Decimal
	Price  Decimal
	Kind   Kind
}

func NewBuy(orderID string, quantity, price Decimal, timestamp time.Time) *Order {
	return &Order{
		ID:     orderID,
		Kind:   Buy,
		Volume: quantity,
		Price:  price,
		Time:   timestamp,
	}
}

func NewSell(orderID string, quantity, price Decimal, timestamp time.Time) *Order {
	return &Order{
		ID:     orderID,
		Kind:   Sell,
		Volume: quantity,
		Price:  price,
		Time:   timestamp,
	}
}
