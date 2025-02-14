#
# /cmd/clickhouse_protobuflist/Makefile
#

# ldflags variables to update --version
# short commit hash
COMMIT := $(shell git describe --always)
DATE := $(shell date -u +"%Y-%m-%d-%H:%M")
VERSION_FILE := $(shell cat ../../VERSION)
LOCAL_MAJOR_VERSION := $(word 1,$(subst ., ,$(VERSION_FILE)))
LOCAL_MINOR_VERSION := $(word 2,$(subst ., ,$(VERSION_FILE)))
LOCAL_PATCH_VERSION := $(word 3,$(subst ., ,$(VERSION_FILE)))
BINARY := clickhouse_protobuflist

all: clean build

test:
	go test

clean:
	[ -f ${BINARY} ] && rm -rf ./${BINARY} || true

build:
	go build -ldflags \
		"-X main.commit=${COMMIT} -X main.date=${DATE} -X main.version=${LOCAL_MAJOR_VERSION}.${LOCAL_MINOR_VERSION}.${LOCAL_PATCH_VERSION}" \
		-o ./${BINARY} \
		./${BINARY}.go

examples:
	./clickhouse_protobuflist -filename protoBytes.value1.bin -value 1
	./clickhouse_protobuflist -filename protoBytes.value4294967295.bin -value 4294967295
	./clickhouse_protobuflist -filename protoBytes.value1.envelope.bin -value 1 -envelope
	./clickhouse_protobuflist -filename protoBytes.value4294967295.envelope.bin -value 4294967295 -envelope

sync:
	rsync -av --exclude '.git' ./ /home/das/Downloads/clickhouse_protobuflist/cmd/

# end