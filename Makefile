MODULE = broadcast-3d
BINARY = ~/go/bin/maelstrom-$(MODULE)

WORKLOAD = broadcast-3de

MAELSTROM_CMD_echo = maelstrom/maelstrom test -w echo --bin $(BINARY) --node-count 1 --time-limit 10
MAELSTROM_CMD_unique-ids = maelstrom/maelstrom test -w unique-ids --bin $(BINARY) --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition
MAELSTROM_CMD_broadcast-3a = maelstrom/maelstrom test -w broadcast --bin $(BINARY) --node-count 1 --time-limit 20 --rate 10
MAELSTROM_CMD_broadcast-3b = maelstrom/maelstrom test -w broadcast --bin $(BINARY) --node-count 5 --time-limit 20 --rate 10
MAELSTROM_CMD_broadcast-3c = maelstrom/maelstrom test -w broadcast --bin $(BINARY) --node-count 5 --time-limit 20 --rate 10 --nemesis partition
MAELSTROM_CMD_broadcast-3de = maelstrom/maelstrom test -w broadcast --bin $(BINARY) --node-count 25 --time-limit 20 --rate 100 --latency 100

MAELSTROM_RUN_CMD = $(MAELSTROM_CMD_$(WORKLOAD))

run: build
	@$(MAELSTROM_RUN_CMD)

build:
	go build -o $(BINARY) ./$(MODULE)

debug:
	maelstrom/maelstrom serve
