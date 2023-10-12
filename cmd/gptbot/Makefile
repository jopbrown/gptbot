
BUILD_NAME := $(notdir $(shell pwd))

WORKSPACE := $(dir $(shell go env GOMOD))

TARGET := linux-amd64

RESOURCE = gptbot.yaml.example

CLEAN_PATH := tmp

ENABLE_UPX := true

ENABLE_ARCHIVE := false

include $(WORKSPACE)/script/build.mk

.PHONY: test
test:
	## $(BUILD_NAME) test pass
