ALL=\
	wiki\

BENCH=\
	wiki\

REQS=\
	github.com/go-sql-driver/mysql\

all: $(ALL)

%: %.go
	GOPATH="$(HOME)/go" go build $*.go

%.bench: %
	time ./$*

bench: $(addsuffix .bench, $(BENCH))

clean:
	rm -f $(ALL)

reqs:
	GOPATH="$(HOME)/go" go get $(REQS)
