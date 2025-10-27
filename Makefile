MODULE ?= echo
BINARY ?= ~/go/bin/$(MODULE)

MAELSTROM_CMD_echo = maelstrom/maelstrom test -w echo --bin $(BINARY) --node-count 1 --time-limit 10

MAELSTROM_RUN_CMD = $(MAELSTROM_CMD_$(MODULE))

run: build
	@$(MAELSTROM_RUN_CMD)

build:
	go install ./$(MODULE)

debug:
	maelstrom/maelstrom serve
