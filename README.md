# Gossip Glomers Challenges

These are my solutions for distributed systems problems: https://fly.io/dist-sys/ known as "Gossip Glomers". 

## How to run it? 

To run these, you should have maelstrom binary on your `$PATH`. Then, when you adjust the `Makefile` you're ready to run
challenges after setting proper `MODULE` and `WORKLOAD` make's variables. For example: 

```shell
MODULE=broadcast-3b
WORKLOAD=broadcast-3c
make run
```

will run code under `./broadcast-3b` with workload prescribed in: https://fly.io/dist-sys/3c/
