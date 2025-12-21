# Gossip Glomers Challenges

These are my solutions for distributed systems problems: https://fly.io/dist-sys/ known as "Gossip Glomers". 

## How to run it? 

To run these, you should have maelstrom binary on your `$PATH`. If you want to obtain simplest solution, then 
all you need to do is to download `maelstrom` directory inside the (git) root: 

```shell
❯ curl -L https://github.com/jepsen-io/maelstrom/releases/download/v0.2.4/maelstrom.tar.bz2 | tar -xj
```

Then, when you adjust the `Makefile` (depending on where you installed `maelstrom` binary) you're ready to run
challenges after setting proper `MODULE` and `WORKLOAD` make's variables. For example: 

```shell
MODULE=broadcast-3e
WORKLOAD=broadcast-3de
make run
```

will run code under `./broadcast-3b` with workload prescribed in: https://fly.io/dist-sys/3c/
