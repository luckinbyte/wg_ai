.PHONY: all build test clean proto plugins plugin-role plugin-item plugin-soldier

all: proto build

build:
	go build -o bin/game ./cmd/game
	go build -o bin/login ./cmd/login
	go build -o bin/db ./cmd/db

# 编译所有插件
plugins: plugin-role plugin-item plugin-soldier

# 编译单个插件
plugin-role:
	go build -buildmode=plugin -o plugins/role.so ./plugin/role

plugin-item:
	go build -buildmode=plugin -o plugins/item.so ./plugin/item

plugin-soldier:
	go build -buildmode=plugin -o plugins/soldier.so ./plugin/soldier

test:
	go test -v ./...

clean:
	rm -rf bin/
	rm -rf plugins/*.so

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/cs/*.proto proto/ss/*.proto

# 热更示例 (编译 + 调用API)
hotreload-role: plugin-role
	curl -X POST http://localhost:8081/admin/hotreload \
		-H "Content-Type: application/json" \
		-d '{"module": "role", "path": "./plugins/role.so"}'

hotreload-item: plugin-item
	curl -X POST http://localhost:8081/admin/hotreload \
		-H "Content-Type: application/json" \
		-d '{"module": "item", "path": "./plugins/item.so"}'

hotreload-soldier: plugin-soldier
	curl -X POST http://localhost:8081/admin/hotreload \
		-H "Content-Type: application/json" \
		-d '{"module": "soldier", "path": "./plugins/soldier.so"}'
