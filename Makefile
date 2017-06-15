all: build

# if GOPATH isn't set then set it
ifeq (,$(strip $(GOPATH)))
GOPATH := $(shell go env | grep GOPATH | sed 's/GOPATH="\(.*\)"/\1/')
endif

# set GOPATH to its first element
GOPATH := $(word 1,$(subst :, ,$(GOPATH)))

BUILD_TAGS :=   gofig \
				pflag \
				mods

$(GOPATH)/bin/lsx-linux:
	go install -tags '$(BUILD_TAGS)' ./cli/lsx/lsx-linux

build: $(GOPATH)/bin/lsx-linux

clean:
	rm -f $(GOPATH)/bin/lsx-linux
