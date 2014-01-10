GOFILES = $(wildcard *.go)
PID="../.pid"

all:
	@go build

serve:
	@clear
	@WITH_PID=$(PID) go run $(GOFILES) &
	@fswatch . "make restart"

restart:
	@clear
	@kill `cat $(PID)` || echo
	@WITH_PID=$(PID) go run $(GOFILES) &
