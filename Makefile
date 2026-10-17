.PHONY: dev-server dev-admin build-admin build docker-up docker-down docker-logs reset-password

# 本地开发：需要先在另一个终端启动 MongoDB，或设置 MONGO_URI
dev-server:
	cd server && go run -tags dev ./cmd/feedbacker

dev-admin:
	cd admin && npm run dev

build-admin:
	cd admin && npm run build

# 编译后端（管理后台独立构建部署）
build:
	mkdir -p server/bin
	cd server && go build -trimpath -o bin/feedbacker ./cmd/feedbacker

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f app

# 在运行中的容器里重置管理员密码
reset-password:
	docker compose exec app ./feedbacker admin reset-password
