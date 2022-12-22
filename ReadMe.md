Matching Engine
---
This repository is a Go language exercise.

First Iteration
---
I've used a heap to keep the prices.

The orders are filled immediately when the order is taken.

See tests for poor man's benchmark.

Second iteration
---
Can be found in v2 folder and consist of a red black tree to keep the prices.

Each market has two brokers (broker being the manager) and a map in which orders can be found by their IDs.

The brokers consist of a red black tree, a prices map (price value being the key) to a queue of orders.

Of course, more features can be added, but for the purpose of this exercise, I'm gonna stop here.
