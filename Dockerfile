# マルチステージビルド: フロントエンドビルド
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

# 依存関係をコピーしてインストール
COPY frontend/package*.json ./
RUN npm ci

# ソースコードをコピーしてビルド
COPY frontend/ ./
RUN npm run build

# マルチステージビルド: バックエンドビルド
FROM golang:1.25-alpine AS backend-builder

WORKDIR /app/backend

# sqliteドライバがCGOを使うため、ビルド用のC toolchainを入れる
RUN apk add --no-cache gcc musl-dev

# 依存関係をコピー
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# ソースコードをコピーしてビルド
COPY backend/ ./
RUN CGO_ENABLED=1 GOOS=linux go build -o /chronome-server ./cmd/server

# 最終イメージ: Nginx + バックエンド
FROM nginx:alpine

# Nginxの設定ファイルをコピー
COPY nginx.conf /etc/nginx/nginx.conf

# フロントエンドのビルド成果物をNginxの公開ディレクトリにコピー
COPY --from=frontend-builder /app/frontend/build /usr/share/nginx/html

# バックエンドのバイナリをコピー
COPY --from=backend-builder /chronome-server /usr/local/bin/chronome-server

# 起動スクリプトをコピー
COPY start.sh /start.sh
RUN chmod +x /start.sh

# Cloud Runはポート8080を期待
EXPOSE 8080

# Nginxとバックエンドを起動
CMD ["/start.sh"]