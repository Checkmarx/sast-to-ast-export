BUILD = ./build
LD_FLAGS = -ldflags="-s -w"

lint:
	go fmt ./...
	golangci-lint run

build: windows_amd64 windows_386

unit_test:
	go test -short ./...

clean:
	rm -r $(BUILD)

windows_amd64:
	env GOOS=windows GOARCH=amd64 go build -o $(BUILD)/windows/amd64/sast-export.exe $(LD_FLAGS)

windows_386:
	env GOOS=windows GOARCH=386 go build -o $(BUILD)/windows/386/sast-export.exe $(LD_FLAGS)
