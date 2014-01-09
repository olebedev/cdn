GOFILES = $(wildcard *.go)
PID_PREFIX=.cdn

all:
	@go build

serve:
	@clear
	@WITH_PID=$(PID_PREFIX) go run $(GOFILES) &
	@fswatch . "make restart"

restart:
	@clear
	@kill `cat ../$(PID_PREFIX)` || echo
	@WITH_PID=$(PID_PREFIX) go run $(GOFILES) &
