BUILD_PATH = ./build
PRODUCT_NAME = cxsast_exporter
PRODUCT_VERSION = $(shell cat VERSION)
PRODUCT_BUILD = $(shell date +%Y%m%d%H%M%S)
PUBLIC_KEY = $(shell cat public.key)
LD_FLAGS = -ldflags="-s -w -X github.com/checkmarxDev/ast-sast-export/cmd.productName=$(PRODUCT_NAME) -X github.com/checkmarxDev/ast-sast-export/cmd.productVersion=$(PRODUCT_VERSION) -X github.com/checkmarxDev/ast-sast-export/cmd.productBuild=$(PRODUCT_BUILD) -X github.com/checkmarxDev/ast-sast-export/internal/encryption.BuildTimeRSAPublicKey=$(PUBLIC_KEY)"

SAST_EXPORT_USER = '###########'
SAST_EXPORT_PASS = '###########'

lint:
	go fmt ./...
	golangci-lint run

build: windows_amd64 #windows_386 linux_amd64 linux_386 darwin_amd64

run: windows_amd64 run_windows

debug: windows_amd64 debug_windows

unit_test:
	go test -short $(LD_FLAGS) ./... -coverprofile=coverage.out

clean:
	rm -r $(BUILD_PATH)

windows_amd64: check_public_key
	env GOOS=windows GOARCH=amd64 go build -o $(BUILD_PATH)/windows/amd64/$(PRODUCT_NAME).exe $(LD_FLAGS)
	cp -r external/similarity/windows/amd64/SimilarityCalculator.exe $(BUILD_PATH)/windows/amd64

#windows_386: check_public_key
#	env GOOS=windows GOARCH=386 go build -o $(BUILD_PATH)/windows/386/$(PRODUCT_NAME).exe $(LD_FLAGS)

#linux_amd64: check_public_key
#	env GOOS=linux GOARCH=amd64 go build -o $(BUILD_PATH)/linux/amd64/$(PRODUCT_NAME) $(LD_FLAGS)

#linux_386: check_public_key
#	env GOOS=linux GOARCH=386 go build -o $(BUILD_PATH)/linux/386/$(PRODUCT_NAME) $(LD_FLAGS)

#darwin_amd64: check_public_key
#	env GOOS=darwin GOARCH=amd64 go build -o $(BUILD_PATH)/darwin/amd64/$(PRODUCT_NAME) $(LD_FLAGS)

public_key:
	aws kms get-public-key --key-id alias/sast-migration-key --region eu-west-1 | jq -r .PublicKey > public.key

check_public_key:
	if [ -z $(PUBLIC_KEY) ]; then echo "Please run: make public_key"; exit 1; fi

run_windows:
	build/windows/amd64/cxsast_exporter --user $(SAST_EXPORT_USER) --pass $(SAST_EXPORT_PASS) --url http://localhost --export users,results,teams --results-project-active-since 1

debug_windows:
	build/windows/amd64/cxsast_exporter --user $(SAST_EXPORT_USER) --pass $(SAST_EXPORT_PASS) --url http://localhost --export users,results,teams --results-project-active-since 10 --debug

mocks:
	mockgen -destination test/mocks/sast/mock_client.go -package mock_sast github.com/checkmarxDev/ast-sast-export/internal/sast Client
	mockgen -destination test/mocks/export/mock_exporter.go -package mock_export github.com/checkmarxDev/ast-sast-export/internal/export Exporter
	#mockgen -destination test/mocks/database/store/task_scans_mock.go -package mock_store github.com/checkmarxDev/ast-sast-export/internal/database/store TaskScansStore
	#mockgen -destination test/mocks/database/store/component_configuration_mock.go -package mock_store github.com/checkmarxDev/ast-sast-export/internal/database/store CxComponentConfigurationStore
	#mockgen -destination test/mocks/database/store/node_results_mock.go -package mock_store github.com/checkmarxDev/ast-sast-export/internal/database/store NodeResultsStore
	mockgen -destination test/mocks/export/mock_metadata_provider.go -package mock_export github.com/checkmarxDev/ast-sast-export/internal/export MetadataProvider
	mockgen -destination test/mocks/ast/query_id_provider_mock.go -package mock_ast github.com/checkmarxDev/ast-sast-export/internal/ast QueryIDProvider
	mockgen -destination test/mocks/sast/similarity_id_provider_mock.go -package mock_sast github.com/checkmarxDev/ast-sast-export/internal/sast SimilarityIDProvider
	mockgen -destination test/mocks/soap/adapter_mock.go -package mock_soap github.com/checkmarxDev/ast-sast-export/internal/soap Adapter
