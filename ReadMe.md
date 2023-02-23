# Matching Engine
---
This repository is a Go language exercise, for someone who asked to create an exchange.

The original requirements didn't specify how many transactions per minute should be supported, but the person who told
me about this exercise also gave me a number. Which I was trying to beat... and probably succeeded in version 2.

## First Iteration
---
I've used a heap to keep the prices.

The orders are filled immediately when the order is taken.

See tests for poor man's benchmark.

## Second iteration
---
Can be found in v2 folder and consist of a red black tree to keep the prices.

Each market has two brokers (broker being the manager) and a map in which orders can be found by their IDs.

The brokers consist of a red black tree, a prices map (price value being the key) to a queue of orders.

Of course, more features can be added, but for the purpose of this exercise, I'm gonna stop here.

## Some Explanations
---

In trading market, exchanges maintains an order book for every commodity or stock which is traded on their exchange.
There are mainly two kinds of orders customers can send, a buy order, and a sell order.

A technique called "limit price", which means buy order with a limit price of EUR 1.50 can be executed, if it was found
a sell order having the same value EUR 1.50 or an order of a lower price says EUR 1.49, but it can NOT be executed with
a sell order of price EUR 1.51.

In the same way, a sell order can be honoured(executed) for a price, which is either equal to or higher than the limit
price.
In general, a "limit" order executes if it found a specified price or even a better price. As you can see, the rules are
very simple : lower in the case of a buy, and higher in the case of sell.

Orders are executed on a first come first serve (FIFO) basis, therefor exchange also maintains a time priority.
The flow would be like this : an order comes to exchange market, bookkeeper looks order book for that symbol /
commodity, if it found a match it honours (executes) order and transaction is being made.
Otherwise, bookkeeper adds that order at the end of a price queue - which represents time priority - where head of the
queue represents the order with highest time priority.

The bookkeeper's goal is to perform the above operations as quickly as possible. As you can expect, it involves finding
the pair order which matches the price, equal, less or greater. Removing an order from the books in the case bookkeeper
finds a match, or if that order gets cancelled. If no match was found, bookkeeper will add that order for later
matching.

All these operations makes traversing the key, which strongly suggests the usage of a binary search tree data structure.
Using a binary tree for storing orders, a matching order can be found by using binary search which is of order O(log2N),
not quite fast as O(1) but still a decent one.

Adding and removing orders in a binary tree will cost the same amount of time, because they involve the same mechanism
of traversal. Having different symbols, since the order of one symbol can not match to order of another symbol, the
books must be associated with just a symbol.

A Queue data structure to maintain time priority of the orders with the same price is being used, because the order
which comes first, should execute first (FIFO) if price matches.

Of course, the binary tree can take multiple forms (particularities) which can improve in the end the overall
performance.

