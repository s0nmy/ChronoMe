# ChronoMe GCP デプロイ

ChronoMe は `Cloud Run + Cloud SQL (PostgreSQL) + Artifact Registry` 構成でデプロイできます。

## 前提

- `gcloud` CLI が使える
- 対象 GCP プロジェクトを作成済み
- `scripts/setup-gcp.sh` で基本リソースを作成済み
- Secret Manager に以下のシークレットがある
  - `chronome-db-dsn`
  - `chronome-session-secret`

## 1. 初期セットアップ

```bash
export GCP_PROJECT_ID="your-project-id"
export GITHUB_REPO="owner/repo"

./scripts/setup-gcp.sh
```

`chronome-db-dsn` は Cloud SQL 接続名を使って登録します。

```bash
CONNECTION_NAME="$(gcloud sql instances describe chronome-db --format='value(connectionName)')"

gcloud secrets create chronome-db-dsn --replication-policy=automatic
echo -n "host=/cloudsql/${CONNECTION_NAME} user=chronome password=YOUR_PASSWORD dbname=chronome sslmode=disable" \
  | gcloud secrets versions add chronome-db-dsn --data-file=-
```

`chronome-session-secret` が未作成なら `setup-gcp.sh` が生成します。

## 2. デプロイ

ステージング:

```bash
export GCP_PROJECT_ID="your-project-id"
./scripts/deploy-cloud-run.sh staging
```

本番:

```bash
export GCP_PROJECT_ID="your-project-id"
./scripts/deploy-cloud-run.sh production
```

必要に応じて以下も上書きできます。

- `GCP_REGION` 既定値は `asia-northeast1`
- `ARTIFACT_REGISTRY_REPO` 既定値は `chronome`
- `CLOUD_SQL_INSTANCE` 既定値は `chronome-db`
- `BACKEND_SERVICE_NAME`
- `FRONTEND_SERVICE_NAME`

## 3. デプロイ時の流れ

1. バックエンドを Cloud Build でビルド
2. Cloud Run にバックエンドをデプロイ
3. バックエンド URL を使ってフロントエンドをビルド
4. Cloud Run にフロントエンドをデプロイ
5. フロントエンド URL を使ってバックエンドの `ALLOWED_ORIGIN` を更新

## 4. 注意点

- フロントエンドは Cloud Run の `PORT` 環境変数に追従するようにしてあります
- バックエンドは起動時に GORM の `AutoMigrate` を実行します
- `APP_ENV=production` でなくても `SESSION_COOKIE_SECURE=true` を付けているため HTTPS 前提の Cookie 設定になります
- 現在の GitHub Actions は実運用デプロイではなくプレースホルダを含んでいます。まずはこのスクリプトでの手動デプロイを前提にしてください
