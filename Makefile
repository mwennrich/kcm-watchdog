GO111MODULE := on
DOCKER_TAG := $(or ${GITHUB_TAG_NAME}, latest)

all: shoot-watchdog

.PHONY: shoot-watchdog
shoot-watchdog:
	go build -o bin/shoot-watchdog
	strip bin/shoot-watchdog

.PHONY: dockerimages
dockerimages:
	docker build -t mwennrich/shoot-watchdog:${DOCKER_TAG} .

.PHONY: dockerpush
dockerpush:
	docker push mwennrich/shoot-watchdog:${DOCKER_TAG}

.PHONY: clean
clean:
	rm -f bin/*

