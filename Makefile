GO111MODULE := on
DOCKER_TAG := $(or ${GITHUB_TAG_NAME}, latest)

all: kcm-watchdog

.PHONY: kcm-watchdog
kcm-watchdog:
	go build -o bin/kcm-watchdog
	strip bin/kcm-watchdog

.PHONY: dockerimages
dockerimages:
	docker build -t mwennrich/kcm-watchdog:${DOCKER_TAG} .

.PHONY: dockerpush
dockerpush:
	docker push mwennrich/kcm-watchdog:${DOCKER_TAG}

.PHONY: clean
clean:
	rm -f bin/*

