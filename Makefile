GOPATH?=`pwd`/../../../../
GOBIN=$(GOPATH)/bin/
GO=GOPATH=$(GOPATH) GOBIN=$(GOBIN) go
OK_COLOR=\033[32;01m
NO_COLOR=\033[0m

APPS=mineshaft mineshaft-bench

all: $(APPS)

$(APPS): %: $(GOBIN)/%

$(GOBIN)/%: cmd/%.go
	$(GO) get -d -v ./...
	$(GO) install -v $<

run: $(GOBIN)/mineshaft
	@echo "$(OK_COLOR)==>$(NO_COLOR) Running"
	$< -f=mineshaft.conf

test:
	$(GO) test -v ./...

clean:
	rm -rf $(GOBIN)/{$(shell echo $(APPS) | sed -e "s/ /,/")}

.PHONY: $(APPS) all run test clean
