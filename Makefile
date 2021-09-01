BUILD_PATH = ./build
PRODUCT_NAME = cxsast_exporter
PRODUCT_VERSION = $(shell cat VERSION)
PRODUCT_BUILD = $(shell date +%Y%m%d%H%M%S)
PUBLIC_KEY = $(shell cat public.key)
LD_FLAGS = -ldflags="-s -w -X sast-export/cmd.productName=$(PRODUCT_NAME) -X sast-export/cmd.productVersion=$(PRODUCT_VERSION) -X sast-export/cmd.productBuild=$(PRODUCT_BUILD) -X sast-export/internal.buildTimeRSAPublicKey=$(PUBLIC_KEY)"

SAST_EXPORT_USER = '###########'
SAST_EXPORT_PASS = '###########'

lint:
	go fmt ./...
	golangci-lint run

build: windows_amd64 windows_386 linux_amd64 linux_386 darwin_amd64

run: windows_amd64 run_windows

debug: windows_amd64 debug_windows

unit_test:
	go test -short $(LD_FLAGS) ./... -coverprofile=coverage.out

clean:
	rm -r $(BUILD_PATH)

windows_amd64: check_public_key
	env GOOS=windows GOARCH=amd64 go build -o $(BUILD_PATH)/windows/amd64/$(PRODUCT_NAME).exe $(LD_FLAGS)

windows_386: check_public_key
	env GOOS=windows GOARCH=386 go build -o $(BUILD_PATH)/windows/386/$(PRODUCT_NAME).exe $(LD_FLAGS)

linux_amd64: check_public_key
	env GOOS=linux GOARCH=amd64 go build -o $(BUILD_PATH)/linux/amd64/$(PRODUCT_NAME) $(LD_FLAGS)

linux_386: check_public_key
	env GOOS=linux GOARCH=386 go build -o $(BUILD_PATH)/linux/386/$(PRODUCT_NAME) $(LD_FLAGS)

darwin_amd64: check_public_key
	env GOOS=darwin GOARCH=amd64 go build -o $(BUILD_PATH)/darwin/amd64/$(PRODUCT_NAME) $(LD_FLAGS)

public_key:
	if [ -z $(SAST_EXPORT_KMS_KEY_ID) ]; then echo "Please specify env var SAST_EXPORT_KMS_KEY_ID"; exit 1; fi
	aws kms get-public-key --key-id $(SAST_EXPORT_KMS_KEY_ID) | jq -r .PublicKey > public.key

check_public_key:
	if [ -z $(PUBLIC_KEY) ]; then echo "Please run: make public_key"; exit 1; fi

run_windows:
	build/windows/amd64/cxsast_exporter --user $(SAST_EXPORT_USER) --pass $(SAST_EXPORT_PASS) --url http://localhost --export users,results,teams

debug_windows:
	build/windows/amd64/cxsast_exporter --user $(SAST_EXPORT_USER) --pass $(SAST_EXPORT_PASS) --url http://localhost --export users,results,teams --debug