FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

FROM golang:1.25-alpine AS backend-builder

WORKDIR /app/backend

RUN apk add --no-cache gcc musl-dev

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=1 GOOS=linux go build -o /chronome-server ./cmd/server

FROM nginx:alpine

COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=frontend-builder /app/frontend/build /usr/share/nginx/html
COPY --from=backend-builder /chronome-server /usr/local/bin/chronome-server
COPY start.sh /start.sh

RUN chmod +x /start.sh

EXPOSE 8080

CMD ["/start.sh"]
