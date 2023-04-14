.PHONY: build
build:
	cd ./ui && yarn build && rm -rf ../internal/ui/build/dist && cp -R ./dist ../internal/ui/build/dist && cd ../ && go build

.PHONY: run-debug
run-debug:
	go build && STROLTP_LOG_LEVEL=DEBUG STROLTP_PORT=8091 ./stroltp --config=./testdata/stroltp.yml

.PHONY: run
run:
	go build && STROLTP_API_LOG_ENABLED=true STROLTP_PORT=8091 ./stroltp --config=./testdata/stroltp.yml
