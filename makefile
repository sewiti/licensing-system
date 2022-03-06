GO := $(shell test -x /usr/local/go/bin/go && echo /usr/local/go/bin/go || echo go)

.PHONY: build test clean install

build:
	mkdir -p ./build
	$(GO) build -o ./build/licensing-server ./cmd/server

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

	[ -s /opt/licensing-server/.env ] || {                             \
	    echo -n "LICENSING_SERVER_KEY=" > /opt/licensing-server/.env;  \
	    /opt/licensing-server/licensing-server generate-keys -base64 | \
	        grep ^key | cut -d: -f3 >> /opt/licensing-server/.env;     \
	}
	chmod 600 /opt/licensing-server/.env