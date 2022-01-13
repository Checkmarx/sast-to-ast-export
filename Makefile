KEYS_PATH = ./keys
EXTERNAL_PATH = ./external
BUILD_PATH = ./build
ENV ?= prod
PRODUCT_NAME = cxsast_exporter
PRODUCT_VERSION = $(shell cat VERSION)
PRODUCT_BUILD = $(shell date +%Y%m%d%H%M%S)
PUBLIC_KEY = "internal/app/encryption/public.key"
LD_FLAGS = -ldflags="-s -w -X github.com/checkmarxDev/ast-sast-export/cmd.productName=$(PRODUCT_NAME) -X github.com/checkmarxDev/ast-sast-export/cmd.productVersion=$(PRODUCT_VERSION) -X github.com/checkmarxDev/ast-sast-export/cmd.productBuild=$(PRODUCT_BUILD)"

SAST_EXPORT_USER = '###########'
SAST_EXPORT_PASS = '###########'

lint:
	go fmt ./...
	golangci-lint run

build: public_key windows_amd64

package: build
	zip -j $(BUILD_PATH)/$(PRODUCT_NAME)_$(PRODUCT_VERSION)_$(PRODUCT_BUILD)_windows_amd64.zip ./build/windows/amd64/*

run: windows_amd64 run_windows

debug: windows_amd64 debug_windows

unit_test:
	go test -short $(LD_FLAGS) ./... -coverprofile=coverage.out

clean:
	rm -r $(BUILD_PATH)

windows_amd64:
	env GOOS=windows GOARCH=amd64 go build -o $(BUILD_PATH)/windows/amd64/$(PRODUCT_NAME).exe $(LD_FLAGS)
	cp -v $(EXTERNAL_PATH)/windows/amd64/SimilarityCalculator.exe $(BUILD_PATH)/windows/amd64/

public_key:
	cp -v $(KEYS_PATH)/$(ENV).key $(PUBLIC_KEY)

download_public_key:
	aws kms get-public-key --key-id alias/sast-migration-key --region eu-west-1 --output json | jq -r .PublicKey | tr -d '\r' | tr -d '\n' > $(KEYS_PATH)/$(ENV).key

run_windows:
	build/windows/amd64/cxsast_exporter --user $(SAST_EXPORT_USER) --pass $(SAST_EXPORT_PASS) --url http://localhost --export users,results,teams --results-project-active-since 1

debug_windows:
	build/windows/amd64/cxsast_exporter --user $(SAST_EXPORT_USER) --pass $(SAST_EXPORT_PASS) --url http://localhost --export users,results,teams --results-project-active-since 10 --debug

mocks:
	rm -rf test/mocks
	mockgen -package mock_integration_rest -destination test/mocks/integration/rest/mock_client.go github.com/checkmarxDev/ast-sast-export/internal/integration/rest Client
	mockgen -package mock_integration_soap -destination test/mocks/integration/soap/mock_adapter.go github.com/checkmarxDev/ast-sast-export/internal/integration/soap Adapter
	mockgen -package mock_integration_similarity -destination test/mocks/integration/similarity/provider_mock.go github.com/checkmarxDev/ast-sast-export/internal/integration/similarity SimilarityIDProvider
	mockgen -package mock_app_ast_query_id -destination test/mocks/app/ast_query_id/mock_provider.go github.com/checkmarxDev/ast-sast-export/internal/app/interfaces ASTQueryIDRepo
	mockgen -package mock_app_source_file -destination test/mocks/app/source_file/mock_provider.go github.com/checkmarxDev/ast-sast-export/internal/app/interfaces SourceFileRepo
	mockgen -package mock_app_method_line -destination test/mocks/app/method_line/mock_provider.go github.com/checkmarxDev/ast-sast-export/internal/app/interfaces MethodLineRepo
	mockgen -package mock_app_metadata -destination test/mocks/app/metadata/mock_provider.go github.com/checkmarxDev/ast-sast-export/internal/app/metadata MetadataProvider
	mockgen -package mock_app_export -destination test/mocks/app/export/mock_exporter.go github.com/checkmarxDev/ast-sast-export/internal/app/export Exporter
