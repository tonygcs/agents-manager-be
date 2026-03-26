IMAGE := agents-manager-be

.PHONY: build
build:
	docker build --ssh default -t $(IMAGE) .
