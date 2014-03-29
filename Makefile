PID     = "../.pid"
SOURCE  = $(wildcard *.go)
TAG     = $(shell git describe --tags)
GOBUILD = go build -ldflags '-w'

serve:
	@clear
	@WITH_PID=$(PID) go run $(SOURCE) &
	@fswatch . "make restart"

restart:
	@clear
	@kill `cat $(PID)` || echo
	@WITH_PID=$(PID) go run $(SOURCE) &

# $(tag) here will contain either `-1.0-` or just `-`
ALL = \
	$(foreach arch,32 64,\
    $(foreach tag,-$(TAG)- -,\
	$(foreach suffix,win.exe linux osx,\
		build/gostatic$(tag)$(arch)-$(suffix))))

all: $(ALL)

# os is determined as thus: if variable of suffix exists, it's taken, if not, then
# suffix itself is taken
win.exe = windows
osx = darwin
build/gostatic-$(TAG)-64-%: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=amd64 $(GOBUILD) -o $@

build/gostatic-$(TAG)-32-%: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=386 $(GOBUILD) -o $@

build/gostatic-%: build/gostatic-$(TAG)-%
	@mkdir -p $(@D)
	cd $(@D) && ln -sf $(<F) $(@F)

# upload: $(ALL)
# ifndef UPLOAD_PATH
# 	@echo "Define UPLOAD_PATH to determine where files should be uploaded"
# else
# 	rsync -l -P $(ALL) $(UPLOAD_PATH)
# endif
