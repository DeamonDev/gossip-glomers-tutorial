MODULE = echo
BINARY = ~/go/bin/maelstrom-$(MODULE)

MAELSTROM_CMD_echo = maelstrom/maelstrom test -w echo --bin $(BINARY) --node-count 1 --time-limit 10

MAELSTROM_RUN_CMD = $(MAELSTROM_CMD_$(MODULE))

run: build
	@$(MAELSTROM_RUN_CMD)

build:
	go build -o $(BINARY) ./$(MODULE)

debug:
	maelstrom/maelstrom serve
