TAG = latest
ORGANIZATION = palchukovsky
PRODUCT = wallet
GO_VER = 1.12
NODE_OS_NAME = alpine
NODE_OS_TAG = 3.9
DB_TAG = 11.2-alpine

IMAGE_TAG_REST = $(ORGANIZATION)/$(PRODUCT).rest:$(TAG)
IMAGE_TAG_DB = $(ORGANIZATION)/$(PRODUCT).db:$(TAG)

THIS_FILE := $(lastword $(MAKEFILE_LIST))

.DEFAULT_GOAL := help

define build_docker_cmd_builder_image
	$(eval BUILDER_SOURCE_TAG = ${GO_VER}-${NODE_OS_NAME}${NODE_OS_TAG})
	$(eval BUILDER_TAG = $(ORGANIZATION)/builder.golang:$(BUILDER_SOURCE_TAG))
	docker build \
		--rm \
		--build-arg TAG=$(BUILDER_SOURCE_TAG) \
		--file "$(CURDIR)/build/cmdbuilder/Dockerfile" \
		--tag $(BUILDER_TAG) \
		./
endef

define build_docker_cmd_image
	$(if $(BUILDER_TAG),, $(call build_docker_cmd_builder_image))
	docker build \
		--rm \
		--build-arg NODE_OS_NAME=$(NODE_OS_NAME) \
		--build-arg NODE_OS_TAG=$(NODE_OS_TAG) \
		--build-arg BUILDER=$(BUILDER_TAG) \
		--file "$(CURDIR)/cmd/$(1)/Dockerfile" \
		--tag $(2) \
		./
endef

define build_docker_db_image
	docker build \
		--rm \
		--build-arg TAG=$(DB_TAG) \
		--file "$(CURDIR)/build/db/$(1)/Dockerfile" \
		--tag $(2) \
		./
endef

define push_docker_image
	docker push $(1)
endef

define get_mock
	mockgen -source=$(1).go -destination=mock/$(1).go $(2)
endef

define make_target
	$(MAKE) -f $(THIS_FILE) $(1)
endef


.PHONY: help build release mock


help: ## Show this help.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

build: ## Build docker images from actual local sources.
	@$(call build_docker_cmd_image,rest-server,$(IMAGE_TAG_REST))
	@$(call build_docker_db_image,,$(IMAGE_TAG_DB))

release: ## Push docker images to the hub.
	@$(call push_docker_image,$(IMAGE_TAG_DB))
	@$(call push_docker_image,$(IMAGE_TAG_REST))

mock: ## Generate mock interfaces for unit-tests.
	@$(call get_mock,db,DB)
	@$(call get_mock,executor,Executor)
	@$(call get_mock,repo,RepoTrans Repo)
	@$(call get_mock,service,Service)
	@$(call get_mock,cmd/rest-server/protocol,Protocol)