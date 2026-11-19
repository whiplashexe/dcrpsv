PS5_PAYLOAD_SDK ?= /opt/ps5-payload-sdk
include $(PS5_PAYLOAD_SDK)/toolchain/prospero.mk

CFLAGS := -Wall -O2
# Unverified: the SDK sample passes no -l flags; if linking fails, try/remove this.
LDLIBS := -lSceSystemService
ELF := ps5-rpc.elf

all: $(ELF)

$(ELF): main.c
	$(CC) $(CFLAGS) -o $@ $< $(LDLIBS)

clean:
	rm -f $(ELF)
