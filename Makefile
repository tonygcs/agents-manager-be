IMAGE := agents-manager-be
MCP_IMAGE := agents-manager-mcp

.PHONY: build
build:
	docker build --ssh default -t $(IMAGE) .

.PHONY: build-mcp
build-mcp:
	docker build -t $(MCP_IMAGE) ./mcp
