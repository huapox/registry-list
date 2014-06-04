
IMPORT_PATH := github.com/vbatts/registry-list


CWD := $(shell pwd)
GOPATH := $(CWD)/.gopath
export GOPATH
export GOBIN=$(GOPATH)/bin
export PATH=$(GOBIN):$(shell echo $$PATH)


all: registry-list
	@ls -l $<

registry-list: .gopath $(wildcard *.go) assets/script_js.go assets/style_css.go
	go build .

assets/script_js.go: .go-bindata script.js
	go-bindata -func=JsScript -out=./assets/script_js.go -pkg=assets script.js

assets/style_css.go: .go-bindata style.css
	go-bindata -func=CssStyle -out=./assets/style_css.go -pkg=assets style.css

.go-bindata: $(GOBIN)/go-bindata
	@touch $@

.gopath:
	mkdir -p $(GOPATH)/src/$(dir $(IMPORT_PATH)) && ln -sf $(CWD) $(GOPATH)/src/$(IMPORT_PATH) && touch $@

$(GOBIN)/go-bindata: $(GOBIN)
	which go-bindata 2>/dev/null || \
		go get github.com/jteeuwen/go-bindata && \
		ln -sf $$(readlink -f $$(which go-bindata)) $@

$(GOBIN):
	mkdir -p $@

clean:
	rm -rf .go-bindata .gopath registry-list

dist-clean: clean
	rm -rf $(GOPATH)

