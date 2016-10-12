PROGRAM=localdns
VERSION=0.2.0
LDFLAGS="-X main.programVersion=$(VERSION)"

all: test

deps:
	go get ./...

install: deps
	go install -a -v -ldflags $(LDFLAGS)

test: deps
	go test -v ./...

qa:
	go vet
	golint
	go test -coverprofile=.cover~
	go tool cover -html=.cover~

dist: linux windows

linux:
	@for os in linux darwin; do \
		for arch in 386 amd64; do \
			target=$(PROGRAM)-$$os-$$arch-$(VERSION); \
			echo Building $$target; \
			GOOS=$$os GOARCH=$$arch go build -ldflags $(LDFLAGS) -o $$target/$(PROGRAM) ; \
			cp -r ./ext/$$os/* ./README.md ./LICENSE $$target; \
			tar -zcf $$target.tar.gz $$target; \
			rm -rf $$target;                   \
		done                                 \
	done

windows:
	@for os in windows; do \
		for arch in 386 amd64; do \
			target=$(PROGRAM)-$$os-$$arch-$(VERSION); \
			echo Building $$target; \
			GOOS=$$os GOARCH=$$arch go build -ldflags $(LDFLAGS) -o $$target/$(PROGRAM).exe ; \
			cp -r ./README.md ./LICENSE $$target; \
			tar -zcf $$target.tar.gz $$target; \
			rm -rf $$target;                   \
		done                                 \
	done

clean:
	rm -rf *.tar.gz
