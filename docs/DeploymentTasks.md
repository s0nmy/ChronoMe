# ChronoMe デプロイタスクリスト

## 概要

GCP (Google Cloud Platform) を使用した本番環境デプロイのタスクリスト。
Cloud Run + Cloud SQL (PostgreSQL) の構成を採用。

---

## フェーズ1: 前提条件の整理

### 1.1 CI/CDの修正
- [x] `.github/workflows/ci.yml` から存在しない `node-backend` ジョブを削除
- [ ] CIが正常に動作することを確認

### 1.2 GCPプロジェクトの準備
- [ ] GCPプロジェクトの作成または選択
- [ ] 必要なAPIの有効化
  - Cloud Run API
  - Cloud SQL Admin API
  - Container Registry / Artifact Registry API
  - Cloud Build API
  - Secret Manager API
- [ ] サービスアカウントの作成と権限設定

---

## フェーズ2: コンテナ化

### 2.1 Backend Dockerfile
- [x] `backend/Dockerfile` の作成（マルチステージビルド）
- [x] `.dockerignore` の作成
- [ ] ローカルでのビルド・動作確認

### 2.2 Frontend Dockerfile
- [x] `frontend/Dockerfile` の作成（ビルド + nginx）
- [x] `frontend/nginx.conf` の作成（SPA用設定 + ヘルスチェック）
- [x] `.dockerignore` の作成
- [ ] ローカルでのビルド・動作確認

### 2.3 docker-compose.yml
- [x] ローカル本番相当環境用の `docker-compose.yml` 作成
- [ ] PostgreSQL + Backend + Frontend の連携確認

---

## フェーズ3: 環境変数・シークレット管理

### 3.1 環境変数の整理
- [x] `.env.example` の作成
- [x] 環境変数の文書化

### 3.2 GCP Secret Manager
- [ ] SESSION_SECRET の登録
- [ ] DB接続情報の登録
- [ ] Cloud Runからのアクセス設定

---

## フェーズ4: データベース (Cloud SQL)

### 4.1 Cloud SQL インスタンス
- [ ] PostgreSQL インスタンスの作成
- [ ] データベースの作成
- [ ] ユーザーの作成
- [ ] プライベートIP / Cloud SQL Proxy の設定

### 4.2 マイグレーション
- [ ] マイグレーション戦略の決定
- [ ] 初期スキーマの適用

---

## フェーズ5: Cloud Run デプロイ

### 5.1 Artifact Registry
- [ ] リポジトリの作成
- [ ] イメージのプッシュ設定

### 5.2 Backend サービス
- [ ] Cloud Run サービスの作成
- [ ] 環境変数・シークレットの設定
- [ ] Cloud SQL 接続の設定
- [ ] ヘルスチェックの設定

### 5.3 Frontend サービス
- [ ] Cloud Run サービスの作成
- [ ] API プロキシ設定（または直接バックエンドを呼び出す設定）

---

## フェーズ6: CI/CD パイプライン完成

### 6.1 GitHub Actions 更新
- [x] GCP認証の設定 (Workload Identity Federation)
- [x] Artifact Registry へのプッシュ
- [x] Cloud Run へのデプロイ
- [x] staging 環境デプロイの実装
- [x] production 環境デプロイの実装
- [x] `scripts/setup-gcp.sh` GCPセットアップスクリプト作成

### 6.2 GitHub Secrets
- [ ] `GCP_PROJECT_ID`
- [ ] `GCP_PROJECT_NUMBER`
- [ ] `STAGING_BACKEND_URL`
- [ ] `STAGING_FRONTEND_URL`
- [ ] `PRODUCTION_BACKEND_URL`
- [ ] `PRODUCTION_FRONTEND_URL`

---

## フェーズ7: ネットワーク・セキュリティ

### 7.1 ドメイン・HTTPS
- [ ] カスタムドメインの設定（必要に応じて）
- [ ] Cloud Run のマネージドSSL

### 7.2 セキュリティ設定
- [ ] CORS の本番設定
- [ ] Cloud Armor（必要に応じて）

---

## フェーズ8: 監視・運用

### 8.1 ロギング
- [ ] Cloud Logging の設定
- [ ] 構造化ログの実装

### 8.2 監視
- [ ] Cloud Monitoring アラートの設定
- [ ] アップタイムチェック

### 8.3 バックアップ
- [ ] Cloud SQL 自動バックアップの設定

---

## 構成図

```
┌─────────────────────────────────────────────────────────────┐
│                        Internet                              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Cloud Load Balancer                       │
│                    (マネージドSSL)                            │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
┌─────────────────────────┐     ┌─────────────────────────┐
│      Cloud Run          │     │      Cloud Run          │
│      (Frontend)         │────▶│      (Backend)          │
│      nginx + static     │     │      Go API             │
└─────────────────────────┘     └─────────────────────────┘
                                              │
                                              ▼
                              ┌─────────────────────────┐
                              │      Cloud SQL          │
                              │      (PostgreSQL)       │
                              └─────────────────────────┘
```

---

## 環境変数一覧

| 変数名 | 説明 | Staging | Production |
|--------|------|---------|------------|
| `APP_ENV` | 環境識別子 | staging | production |
| `SERVER_ADDRESS` | サーバーアドレス | :8080 | :8080 |
| `DB_DRIVER` | DBドライバ | postgres | postgres |
| `DB_DSN` | DB接続文字列 | (Secret) | (Secret) |
| `SESSION_SECRET` | セッション秘密鍵 | (Secret) | (Secret) |
| `SESSION_COOKIE_SECURE` | Secure Cookie | true | true |
| `ALLOWED_ORIGIN` | 許可オリジン | ステージングURL | 本番URL |

---

## 進捗状況

- 開始日: 2026-03-01
- 現在のフェーズ: フェーズ2完了、GCPリソース作成待ち
- 完了済み:
  - CI/CDからnode-backendジョブを削除
  - Backend/Frontend Dockerfile作成
  - docker-compose.yml作成
  - .env.example作成
  - CI/CDワークフローをGCPデプロイ用に更新
  - GCPセットアップスクリプト作成
