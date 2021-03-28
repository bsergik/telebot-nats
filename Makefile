export GO111MODULE := on
export GOOS=linux
export GOARCH=amd64
export DOCKER_BUILDKIT=1

TARGET ?= telebot-nats
APP_VERSION ?= 1.0.0
# SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

OUTPUT_PATH ?= ./build

DOCKER_TAG_PREFIX := github.com/bsergik
PROD_SERVER := prod
PROD_FOLDER := /raid/docker-images

################################################################################
### Builds
### 

build-sources:
	go build -ldflags "-X main.buildVersion=$(BUILD_VERSION)" -v -o $(OUTPUT_PATH)/$(TARGET) ./cmd/telebot-nats

build-docker:
	docker build --tag=$(DOCKER_TAG_PREFIX)/$(TARGET):$(APP_VERSION) \
		--build-arg build_version=$(git rev-parse HEAD) --build-arg COMMON_VERSION=$(COMMON_VERSION) \
		-f ./deployments/docker/Dockerfile .

################################################################################
### Generators
### 

# generate-sources:
# 	mockgen -destination=internal/generated/mock_storage.go -source=internal/models/storage.go -package=generated
# 	mockgen -destination=internal/generated/mock_grpc_client.go -source=internal/generated/counter_grpc.pb.go -package=generated

# gen-protoc:
# 	mkdir -p \
# 	# 	./internal/generated/auth-abac/

# 	buf beta mod update
# 	buf generate --output ./internal/generated/

################################################################################
### Deployments
### 

upload-to-prod: make-build-path
	docker image save -o $(OUTPUT_PATH)/$(TARGET)-$(APP_VERSION).image.docker $(DOCKER_TAG_PREFIX)/$(TARGET):$(APP_VERSION)
	# scp $(OUTPUT_PATH)/$(TARGET)-$(APP_VERSION).image.docker $(PROD_SERVER):$(PROD_FOLDER)
	rsync -aP --rsh=ssh $(OUTPUT_PATH)/$(TARGET)-$(APP_VERSION).image.docker $(PROD_SERVER):$(PROD_FOLDER)

# push-helm-chart: $(HELM_USERNAME) $(HELM_PASSWORD) $(HELM_VERSION)
# 	helm push --username=$(HELM_USERNAME) --password=$(HELM_PASSWORD) --version=$(HELM_VERSION) helm/$(TARGET) $(HELM_REPOSITORY)

# push-docker: build-docker
# 	docker push $(DOCKER_TAG_PREFIX)/$(TARGET):$(APP_VERSION)

docker-compose-up:
	docker-compose -f ./deployments/docker-compose.yml up -d database
	docker-compose -f deployments/docker-compose.yml exec -e PGPASSWORD="postgres" -T database \
		psql --user postgres --no-password -c 'create database stan'
	docker-compose -f deployments/docker-compose.yml exec -e PGPASSWORD="postgres" -T database \
		psql --user postgres --no-password -c 'create database telebot'
	curl https://raw.githubusercontent.com/nats-io/nats-streaming-server/master/scripts/postgres.db.sql | \
		docker-compose -f deployments/docker-compose.yml exec -e PGPASSWORD="postgres" -T database \
		psql --user postgres --no-password --dbname=stan
	docker-compose -f ./deployments/docker-compose.yml up -d $(TARGET)

################################################################################
### Tests
### 

test-sources: mod-download
	go test -v -race -cover ./...

bench-sources: mod-download
	go test -v ./... -bench=.

################################################################################
### Linters
### 

lint:
	find -type f -name "*.go" | grep -v '.*\.pb\.go' | grep -v '\/[a-z_]*.go' && echo "Files should be named in snake case" && exit 1 || echo "All files named in snake case"
	golangci-lint version
	golangci-lint -v run

structslop: $(SOURCES) mod-download
	structslop ./...

################################################################################
### Golang related
### 

tidy:
	go mod tidy

modup:
	go get -u ./...
	go mod tidy

################################################################################
### Helpers
### 

make-build-path:
	mkdir -p $(OUTPUT_PATH)

mod-download:
	go mod download -x

download-structslop:
	go get github.com/orijtech/structslop/cmd/structslop

load-image-on-prod:
	ssh -t $(PROD_SERVER) docker image load -i $(PROD_FOLDER)/$(TARGET)-$(APP_VERSION).image.docker

docker-compose-up-d: load-image-on-prod
	ssh -t $(PROD_SERVER) docker-compose -f /raid/cluster/docker-compose.yml up -d $(TARGET)

# clean:
# 	rm -rf $(OUTPUT_PATH)

# .PHONY: build-sources build-docker generate-sources upload-to-prod push-helm-chart push-docker test-sources bench-sources lint structslop tidy modup make-build-path mod-download clean

# all: build-source test-sources structslop lint build-docker push-docker push-helm-chart
