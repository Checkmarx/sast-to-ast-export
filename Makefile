BUILD = ./build
LD_FLAGS = -ldflags="-s -w"

lint:
	go fmt ./...
	golangci-lint run

build: windows_amd64 windows_386 linux_amd64 linux_386 darwin_amd64

unit_test:
	go test -short ./...

clean:
	rm -r $(BUILD)

windows_amd64:
	env GOOS=windows GOARCH=amd64 go build -o $(BUILD)/windows/amd64/sast-export.exe $(LD_FLAGS)

windows_386:
	env GOOS=windows GOARCH=386 go build -o $(BUILD)/windows/386/sast-export.exe $(LD_FLAGS)

linux_amd64:
	env GOOS=linux GOARCH=amd64 go build -o $(BUILD)/linux/amd64/sast-export $(LD_FLAGS)

linux_386:
	env GOOS=linux GOARCH=386 go build -o $(BUILD)/linux/386/sast-export $(LD_FLAGS)

darwin_amd64:
	env GOOS=darwin GOARCH=amd64 go build -o $(BUILD)/darwin/amd64/sast-export $(LD_FLAGS)
