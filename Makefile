.PHONY: image

TODAY=$(shell date "+%Y%m%d%H%M")

image:
	docker buildx build --platform linux/amd64,linux/arm -t bpoetzschke/rtlamr_psql_collect:$(TODAY)-dev -t bpoetzschke/rtlamr_psql_collect:latest-dev --push .