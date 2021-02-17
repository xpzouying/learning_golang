BUILDTIME=$(shell date '+%Y%m%d')
COMMIT_ID=$(shell git rev-parse --short HEAD)

build:
	CGO_ENABLED=0 go build -ldflags '-X "main.BuildTime=$(BUILDTIME)"  \
		    -X "main.BuildVersion=$(COMMIT_ID)" '

		    
clean:
	go clean
