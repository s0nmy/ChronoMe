# ChronoMe デプロイ運用ガイド

## 概要

このガイドでは、ChronoMeアプリケーションの日常的なデプロイ運用について説明します。

**前提条件:**
- GCP側のセットアップが完了していること（`gcp-cd-setup-guide.md`参照）
- GitHub Secretsが設定されていること
- developブランチが作成されていること

---

## ブランチ戦略

ChronoMeでは以下のブランチ戦略を採用しています：

```
develop → main → タグ (v*.*.*) → 本番デプロイ
   ↓        ↓
開発作業  ステージングデプロイ
```

### ブランチの役割

- **`develop`**: 日常的な開発作業用ブランチ
- **`main`**: 本番準備完了コード（mainへのマージでステージングに自動デプロイ）
- **タグ (`v*.*.*`)**: 本番リリース（タグ作成で本番環境に自動デプロイ）

---

## 開発フロー

### 1. 新機能の開発

```bash
# developブランチに移動
git checkout develop

# 最新のコードを取得
git pull origin develop

# （オプション）フィーチャーブランチを作成
git checkout -b feature/new-feature

# ... 開発作業 ...

# コミット
git add .
git commit -m "feat: 新機能を追加"

# プッシュ
git push origin feature/new-feature
# または
git push origin develop
```

### 2. ステージング環境へのデプロイ

developブランチの開発が完了したら、mainブランチにマージしてステージング環境にデプロイします。

#### 方法1: Pull Requestを使用（推奨）

1. GitHubでdevelopブランチからmainブランチへのPull Requestを作成
2. レビュー・承認
3. マージ（自動的にステージング環境へデプロイされます）

```bash
# ブラウザでGitHubを開く
# New Pull Request をクリック
# base: main ← compare: develop を選択
# Create Pull Request をクリック
# レビュー後、Merge Pull Request をクリック
```

#### 方法2: コマンドラインで直接マージ

```bash
# mainブランチに移動
git checkout main

# 最新のコードを取得
git pull origin main

# developブランチをマージ
git merge develop

# プッシュ（自動的にステージング環境へデプロイされます）
git push origin main
```

#### デプロイの確認

1. GitHubリポジトリの **Actions** タブを開く
2. 実行中のワークフロー（CI）をクリック
3. `deploy-staging` ジョブのログを確認
4. デプロイ完了後、ステージング環境のURLにアクセスして動作確認

```bash
# ステージング環境のURLを取得
gcloud run services describe chronome-staging \
  --region=asia-northeast1 \
  --format="value(status.url)"
```

### 3. 本番環境へのデプロイ

ステージング環境での動作確認が完了したら、本番環境にデプロイします。

```bash
# mainブランチに移動
git checkout main

# 最新のコードを取得
git pull origin main

# バージョンタグを作成（セマンティックバージョニング）
git tag v1.0.0

# タグをプッシュ（自動的に本番環境へデプロイされます）
git push origin v1.0.0
```

#### デプロイの確認

1. GitHubリポジトリの **Actions** タブを開く
2. タグ名で実行されたワークフローをクリック
3. `deploy-production` ジョブのログを確認
4. デプロイ完了後、本番環境のURLにアクセスして動作確認

```bash
# 本番環境のURLを取得
gcloud run services describe chronome-production \
  --region=asia-northeast1 \
  --format="value(status.url)"
```

---

## バージョン管理

### セマンティックバージョニング

ChronoMeでは[セマンティックバージョニング](https://semver.org/)を採用しています：

```
v{MAJOR}.{MINOR}.{PATCH}
```

- **MAJOR**: 互換性のない大きな変更
- **MINOR**: 後方互換性のある機能追加
- **PATCH**: 後方互換性のあるバグ修正

#### 例

- `v1.0.0`: 最初の安定版リリース
- `v1.1.0`: 新機能追加
- `v1.1.1`: バグ修正
- `v2.0.0`: 破壊的変更を含むメジャーアップデート

### タグの命名規則

- 必ず `v` で始める（例: `v1.0.0`）
- 3つの数字をドットで区切る
- プレリリース版の場合は `-alpha`, `-beta`, `-rc1` などのサフィックスを付ける（例: `v1.0.0-beta.1`）

---

## ロールバック手順

### ステージング環境のロールバック

#### 方法1: 前のコミットに戻す

```bash
# mainブランチで前のコミットに戻す
git checkout main
git reset --hard <previous-commit-sha>
git push origin main --force

# または、revertを使用（履歴を残す）
git checkout main
git revert <commit-sha>
git push origin main
```

#### 方法2: Cloud Runのリビジョンを切り替え

```bash
# リビジョン一覧を確認
gcloud run revisions list \
  --service=chronome-staging \
  --region=asia-northeast1

# 特定のリビジョンにトラフィックを切り替え
gcloud run services update-traffic chronome-staging \
  --region=asia-northeast1 \
  --to-revisions=chronome-staging-00042-abc=100
```

### 本番環境のロールバック

#### 方法1: 前のタグを再プッシュ

```bash
# 前のタグを確認
git tag -l

# 新しいタグを前のコミットから作成
git tag v1.0.1 <previous-commit-sha>
git push origin v1.0.1
```

#### 方法2: Cloud Runのリビジョンを切り替え

```bash
# リビジョン一覧を確認
gcloud run revisions list \
  --service=chronome-production \
  --region=asia-northeast1

# 特定のリビジョンにトラフィックを切り替え
gcloud run services update-traffic chronome-production \
  --region=asia-northeast1 \
  --to-revisions=chronome-production-00042-abc=100
```

---

## トラブルシューティング

### 問題1: デプロイが失敗する

#### WIF認証エラー

**エラーメッセージ:** `failed to generate Google Cloud access token`

**解決策:**
1. GitHub Secretsの`GCP_WORKLOAD_IDENTITY_PROVIDER`と`GCP_SERVICE_ACCOUNT`が正しいか確認
2. Workload Identity Bindingが正しく設定されているか確認

```bash
# Workload Identity Bindingの確認
gcloud iam service-accounts get-iam-policy \
  github-actions-deploy@chronome-488908.iam.gserviceaccount.com
```

#### Dockerプッシュエラー

**エラーメッセージ:** `401 Unauthorized` または `403 Permission denied`

**解決策:**
1. Artifact Registryリポジトリが存在するか確認
2. サービスアカウントに`artifactregistry.writer`ロールがあるか確認

```bash
# リポジトリの確認
gcloud artifacts repositories describe chronome-repo \
  --location=asia-northeast1

# IAMロールの確認
gcloud projects get-iam-policy chronome-488908 \
  --flatten="bindings[].members" \
  --filter="bindings.members:github-actions-deploy@"
```

### 問題2: アプリケーションエラー

#### データベース接続エラー

**症状:** Cloud Runログに `dial tcp: lookup ... no such host` や接続エラーが表示される

**解決策:**
1. GitHub Secretsの`SUPABASE_DB_DSN`が正しいか確認
2. Supabaseプロジェクトが起動しているか確認
3. パスワードがURLエンコードされているか確認

```bash
# Cloud Runの環境変数を確認
gcloud run services describe chronome-staging \
  --region=asia-northeast1 \
  --format="yaml(spec.template.spec.containers[0].env)"
```

#### SESSION_SECRETエラー

**エラーメッセージ:** `SESSION_SECRET must be provided and at least 32 characters long`

**解決策:**
1. Secret Managerにシークレットが存在するか確認
2. Cloud Runサービスアカウントに`secretAccessor`ロールがあるか確認

```bash
# Secretの確認
gcloud secrets versions access latest \
  --secret=chronome-staging-session-secret

# IAMポリシーの確認
gcloud secrets get-iam-policy chronome-staging-session-secret
```

#### CORSエラー

**症状:** ブラウザコンソールに `blocked by CORS policy` が表示される

**解決策:**
`ALLOWED_ORIGIN`環境変数が正しく設定されているか確認してください。デプロイワークフローの`Update ALLOWED_ORIGIN`ステップで自動的に設定されますが、手動で確認する場合は以下のコマンドを使用します。

```bash
# 現在のALLOWED_ORIGINを確認
gcloud run services describe chronome-staging \
  --region=asia-northeast1 \
  --format="yaml(spec.template.spec.containers[0].env)" | grep ALLOWED_ORIGIN

# 手動で更新する場合
SERVICE_URL=$(gcloud run services describe chronome-staging \
  --region=asia-northeast1 \
  --format="value(status.url)")
gcloud run services update chronome-staging \
  --region=asia-northeast1 \
  --update-env-vars="ALLOWED_ORIGIN=${SERVICE_URL}"
```

### 問題3: ログの確認

```bash
# ステージング環境のログを表示
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=chronome-staging" \
  --limit=50 \
  --format="table(timestamp,severity,textPayload)"

# エラーログのみ表示
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=chronome-staging AND severity>=ERROR" \
  --limit=20

# 本番環境のログを表示
gcloud logging read "resource.type=cloud_run_revision AND resource.labels.service_name=chronome-production" \
  --limit=50 \
  --format="table(timestamp,severity,textPayload)"
```

---

## 環境変数の更新

### デプロイ済み環境の環境変数を更新する

```bash
# ステージング環境の環境変数を更新
gcloud run services update chronome-staging \
  --region=asia-northeast1 \
  --update-env-vars="KEY1=value1,KEY2=value2"

# 本番環境の環境変数を更新
gcloud run services update chronome-production \
  --region=asia-northeast1 \
  --update-env-vars="KEY1=value1,KEY2=value2"
```

### Secret Managerのシークレットを更新する

```bash
# 新しいバージョンを追加
echo -n "new-secret-value" | gcloud secrets versions add chronome-staging-session-secret \
  --data-file=-

# Cloud Runサービスを再起動して新しいシークレットを読み込む
gcloud run services update chronome-staging \
  --region=asia-northeast1 \
  --update-secrets="SESSION_SECRET=chronome-staging-session-secret:latest"
```

---

## モニタリング

### Cloud Runのメトリクス確認

```bash
# サービスの状態確認
gcloud run services describe chronome-staging \
  --region=asia-northeast1

gcloud run services describe chronome-production \
  --region=asia-northeast1
```

### GCPコンソールでのモニタリング

1. [Cloud Run コンソール](https://console.cloud.google.com/run)を開く
2. サービス（`chronome-staging` または `chronome-production`）をクリック
3. **メトリクス** タブで以下を確認：
   - リクエスト数
   - レスポンスタイム
   - エラー率
   - インスタンス数
   - メモリ使用率
   - CPU使用率

---

## ベストプラクティス

### 開発フロー

1. **developブランチで開発**: すべての開発作業はdevelopブランチで行う
2. **Pull Requestでレビュー**: mainブランチへのマージは必ずPull Requestを経由する
3. **ステージング環境で確認**: 本番デプロイ前に必ずステージング環境で動作確認する
4. **セマンティックバージョニング**: タグは必ずセマンティックバージョニングに従う

### デプロイタイミング

- **ステージング**: mainブランチへのマージ時（自動）
- **本番**: タグ作成時（手動）
  - 重要な変更は業務時間外にデプロイ
  - 緊急時以外は本番デプロイ後、しばらく監視

### セキュリティ

- **シークレットの管理**: パスワードやAPIキーは必ずGitHub SecretsまたはSecret Managerで管理
- **環境分離**: ステージングと本番で異なるSESSION_SECRETを使用
- **アクセス制御**: GCPプロジェクトへのアクセスは最小権限の原則に従う

### ロールバック準備

- **タグの履歴管理**: すべての本番リリースはタグで管理
- **リビジョンの保持**: Cloud Runのリビジョンは自動的に保持されるが、重要なリビジョンは明示的に保持
- **バックアップ**: Supabaseのバックアップを定期的に確認

---

## よくある質問（FAQ）

### Q: developブランチがない場合は？

```bash
# developブランチを作成
git checkout -b develop
git push -u origin develop

# mainブランチに戻る
git checkout main
```

### Q: mainブランチに直接プッシュしてしまった場合は？

問題ありません。mainブランチへのプッシュで自動的にステージング環境にデプロイされます。

### Q: タグを間違えて作成してしまった場合は？

```bash
# ローカルのタグを削除
git tag -d v1.0.0

# リモートのタグを削除
git push origin :refs/tags/v1.0.0
```

### Q: ステージング環境と本番環境で異なる設定を使いたい

環境変数を使用してください。GitHub Actionsワークフローで、環境ごとに異なる環境変数を設定できます。

### Q: デプロイ履歴を確認したい

```bash
# Cloud Runのリビジョン一覧を確認
gcloud run revisions list \
  --service=chronome-staging \
  --region=asia-northeast1

# Gitタグ一覧を確認
git tag -l
```

---

## 次のステップ

1. **カスタムドメインの設定**
   - Cloud Runにカスタムドメインをマッピング
   - SSL証明書の自動更新

2. **モニタリングの強化**
   - Cloud Monitoringでアラート設定
   - エラー率、レスポンスタイムの監視

3. **継続的改善**
   - デプロイ頻度の向上
   - テストカバレッジの向上
   - パフォーマンスの最適化

---

**作成日:** 2026-07-15
**バージョン:** 1.0
