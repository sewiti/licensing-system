GO := $(shell test -x /usr/local/go/bin/go && echo /usr/local/go/bin/go || echo go)

.PHONY: build test clean install build-demo-client

build:
	mkdir -p ./build
	$(GO) build -o ./build/licensing-server ./cmd/server

build-demo-client:
	mkdir -p ./build
	$(GO) build -o ./build/demo-client ./cmd/demo-client

test:
	$(GO) test ./...

clean:
	rm -rf ./build

install: build
	mkdir -p /opt/licensing-server

	cp -f ./build/licensing-server /opt/licensing-server/licensing-server
	chmod 700 /opt/licensing-server/licensing-server

	cp -f ./licensing-server.service /etc/systemd/system/licensing-server.service
	systemctl daemon-reload

	[ -f /opt/licensing-server/.env ] || {                                                        \
	    /opt/licensing-server/licensing-server generate-keys > /opt/licensing-server/keys;        \
	    echo -n LICENSING_SERVER_KEY= >> /opt/licensing-server/.env;                              \
		grep ^key:base64: /opt/licensing-server/keys | cut -d: -f3 >> /opt/licensing-server/.env; \
	}
	[ ! -f /opt/licensing-server/keys ] || chmod 600 /opt/licensing-server/keys
	chmod 600 /opt/licensing-server/.env
