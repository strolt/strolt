.PHONY: build
build:
	cd ./ui && yarn build && rm -rf ../internal/ui/build/dist && cp -R ./dist ../internal/ui/build/dist && cd ../ && go build

.PHONY: run-debug
run-debug:
	go build -ldflags "-X github.com/strolt/strolt/apps/stroltm/internal/ui.mode=proxy" && STROLTM_LOG_LEVEL=DEBUG STROLTM_PORT=8090 ./stroltm --config=./testdata/stroltm.yml

.PHONY: run
run:
	go build -ldflags "-X github.com/strolt/strolt/apps/stroltm/internal/ui.mode=proxy" && STROLTM_PORT=8090 ./stroltm --config=./testdata/stroltm.yml
