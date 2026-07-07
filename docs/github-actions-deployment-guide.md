# GitHub Actions 自動デプロイ実装ガイド

## 概要

ChronoMeアプリケーションに、ステージング環境と本番環境の自動デプロイを実装します。

**環境構成:**
- **ステージング**: `main`ブランチへのpush時に自動デプロイ
- **本番**: Gitタグ（`v*.*.*`）作成時に自動デプロイ
- **データベース**: Cloud SQL 1インスタンスを共有（コスト削減）
- **認証**: Workload Identity Federation（安全なキーレス認証）

---

## 実装手順

### フェーズ1: GCP初期セットアップ

#### 1.1 環境変数の設定

```bash
export PROJECT_ID="your-gcp-project-id"  # 実際のプロジェクトIDに置き換え
export GITHUB_REPO_OWNER="your-github-username"  # GitHubユーザー名
export GITHUB_REPO_NAME="ChronoMe"
export REGION="asia-northeast1"
```

#### 1.2 必要なAPIの有効化

```bash
gcloud config set project $PROJECT_ID

gcloud services enable run.googleapis.com
gcloud services enable sqladmin.googleapis.com
gcloud services enable artifactregistry.googleapis.com
gcloud services enable secretmanager.googleapis.com
```

#### 1.3 Artifact Registryリポジトリ作成

```bash
# Dockerイメージ保存用のリポジトリ作成
gcloud artifacts repositories create chronome-repo \
  --repository-format=docker \
  --location=$REGION \
  --description="ChronoMe application container images"

# Docker認証設定
gcloud auth configure-docker ${REGION}-docker.pkg.dev
```

#### 1.4 Cloud SQLの確認

既存のCloud SQLインスタンスを確認：

```bash
# インスタンス情報を確認
gcloud sql instances describe chronome-db

# 接続名を取得（後で使用）
export CLOUD_SQL_INSTANCE_CONNECTION_NAME=$(gcloud sql instances describe chronome-db --format="value(connectionName)")
echo "Cloud SQL接続名: $CLOUD_SQL_INSTANCE_CONNECTION_NAME"
```

---

### フェーズ2: Workload Identity Federation設定

#### 2.1 Workload Identity Poolの作成

```bash
# プール作成
gcloud iam workload-identity-pools create "github-actions-pool" \
  --location="global" \
  --display-name="GitHub Actions Pool" \
  --project="${PROJECT_ID}"
```

#### 2.2 OIDC Providerの作成

```bash
# GitHub用のOIDCプロバイダー作成
gcloud iam workload-identity-pools providers create-oidc "github-provider" \
  --workload-identity-pool="github-actions-pool" \
  --location="global" \
  --issuer-uri="https://token.actions.githubusercontent.com" \
  --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository,attribute.repository_owner=assertion.repository_owner" \
  --attribute-condition="assertion.repository_owner=='${GITHUB_REPO_OWNER}'" \
  --project="${PROJECT_ID}"
```

#### 2.3 サービスアカウント作成

```bash
# デプロイ用サービスアカウント作成
gcloud iam service-accounts create "github-actions-deploy" \
  --display-name="GitHub Actions Deployment SA" \
  --project="${PROJECT_ID}"
```

#### 2.4 IAM権限の付与

```bash
# Cloud Run管理者権限
gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --member="serviceAccount:github-actions-deploy@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/run.admin"

# Artifact Registry書き込み権限
gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --member="serviceAccount:github-actions-deploy@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

# サービスアカウント使用権限
gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --member="serviceAccount:github-actions-deploy@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"

# Secret Manager読み取り権限
gcloud projects add-iam-policy-binding "${PROJECT_ID}" \
  --member="serviceAccount:github-actions-deploy@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

#### 2.5 Workload Identity Bindingの設定

```bash
# プロジェクト番号を取得
export PROJECT_NUMBER=$(gcloud projects describe "${PROJECT_ID}" --format="value(projectNumber)")

# GitHub Actionsがサービスアカウントになりすますことを許可
gcloud iam service-accounts add-iam-policy-binding \
  "github-actions-deploy@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/github-actions-pool/attribute.repository/${GITHUB_REPO_OWNER}/${GITHUB_REPO_NAME}" \
  --project="${PROJECT_ID}"
```

#### 2.6 Workload Identity Provider情報の取得

```bash
# WIFプロバイダーのリソース名を取得（GitHub Secretsで使用）
export WIF_PROVIDER=$(gcloud iam workload-identity-pools providers describe "github-provider" \
  --workload-identity-pool="github-actions-pool" \
  --location="global" \
  --format="value(name)" \
  --project="${PROJECT_ID}")

echo "WIF Provider: $WIF_PROVIDER"
```

**この値を控えておいてください。**

---

### フェーズ3: Secret Manager設定

#### 3.1 セッションシークレットの生成

```bash
# ステージング用セッションシークレット生成（64文字）
export STAGING_SESSION_SECRET=$(openssl rand -base64 48)

# 本番用セッションシークレット生成（64文字）
export PRODUCTION_SESSION_SECRET=$(openssl rand -base64 48)

# 値を確認（後で使用）
echo "Staging Session Secret: $STAGING_SESSION_SECRET"
echo "Production Session Secret: $PRODUCTION_SESSION_SECRET"
```

**これらの値を安全な場所に控えておいてください。**

#### 3.2 Secret Managerへの保存

```bash
# ステージング用セッションシークレット
echo -n "$STAGING_SESSION_SECRET" | gcloud secrets create chronome-staging-session-secret \
  --data-file=- \
  --replication-policy="automatic" \
  --project="${PROJECT_ID}"

# 本番用セッションシークレット
echo -n "$PRODUCTION_SESSION_SECRET" | gcloud secrets create chronome-production-session-secret \
  --data-file=- \
  --replication-policy="automatic" \
  --project="${PROJECT_ID}"
```

#### 3.3 Cloud Runサービスアカウントへの権限付与

```bash
# Cloud Runのデフォルトサービスアカウントに権限を付与
gcloud secrets add-iam-policy-binding chronome-staging-session-secret \
  --member="serviceAccount:${PROJECT_NUMBER}-compute@developer.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor" \
  --project="${PROJECT_ID}"

gcloud secrets add-iam-policy-binding chronome-production-session-secret \
  --member="serviceAccount:${PROJECT_NUMBER}-compute@developer.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor" \
  --project="${PROJECT_ID}"
```

---

### フェーズ4: GitHub Secrets設定

GitHubリポジトリの **Settings > Secrets and variables > Actions** から、以下の7つのSecretsを追加してください。

| Secret名 | 値 | 取得方法 |
|---------|-----|---------|
| `GCP_PROJECT_ID` | your-gcp-project-id | 実際のプロジェクトID |
| `GCP_WORKLOAD_IDENTITY_PROVIDER` | projects/123.../providers/github-provider | フェーズ2.6で取得した`$WIF_PROVIDER`の値 |
| `GCP_SERVICE_ACCOUNT` | github-actions-deploy@PROJECT_ID.iam.gserviceaccount.com | `github-actions-deploy@${PROJECT_ID}.iam.gserviceaccount.com` |
| `CLOUD_SQL_INSTANCE_CONNECTION_NAME` | project-id:region:instance-name | フェーズ1.4で取得した`$CLOUD_SQL_INSTANCE_CONNECTION_NAME`の値 |
| `DB_PASSWORD` | データベースパスワード | Cloud SQLユーザー作成時に設定したパスワード |
| `STAGING_SESSION_SECRET` | （64文字のランダム文字列） | フェーズ3.1で生成した`$STAGING_SESSION_SECRET`の値 |
| `PRODUCTION_SESSION_SECRET` | （64文字のランダム文字列） | フェーズ3.1で生成した`$PRODUCTION_SESSION_SECRET`の値 |

**設定手順:**
1. GitHubリポジトリページを開く
2. Settings > Secrets and variables > Actions をクリック
3. "New repository secret" をクリック
4. Name と Secret を入力して "Add secret" をクリック
5. 上記7つすべてを追加

**確認:** Secrets一覧に7つすべてが表示されることを確認してください。

---

### フェーズ5: GitHub Actionsワークフロー修正

#### 5.1 .github/workflows/ci.yml の修正

**ファイル:** `.github/workflows/ci.yml`

**変更内容:**

##### 変更1: トリガーにタグを追加（5行目付近）

```yaml
on:
  push:
    branches: ["main"]
    tags: ["v*.*.*"]  # ← この行を追加
  pull_request:
    branches: ["main"]
  workflow_dispatch:
```

##### 変更2: deploy-staging ジョブの置き換え（111-148行目）

既存の`deploy-staging`ジョブ全体を以下で置き換えてください：

```yaml
  deploy-staging:
    name: Deploy · Staging
    runs-on: ubuntu-latest
    needs: [backend-tests, frontend-build]
    if: ${{ github.ref == 'refs/heads/main' && github.event_name != 'pull_request' }}

    permissions:
      contents: read
      id-token: write

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Authenticate to Google Cloud
        uses: google-github-actions/auth@v2
        with:
          workload_identity_provider: ${{ secrets.GCP_WORKLOAD_IDENTITY_PROVIDER }}
          service_account: ${{ secrets.GCP_SERVICE_ACCOUNT }}

      - name: Set up Cloud SDK
        uses: google-github-actions/setup-gcloud@v2

      - name: Configure Docker for Artifact Registry
        run: gcloud auth configure-docker asia-northeast1-docker.pkg.dev

      - name: Build Docker image
        run: |
          docker build -t asia-northeast1-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/chronome-repo/chronome-staging:${{ github.sha }} .
          docker tag asia-northeast1-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/chronome-repo/chronome-staging:${{ github.sha }} \
                     asia-northeast1-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/chronome-repo/chronome-staging:latest

      - name: Push Docker image
        run: |
          docker push asia-northeast1-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/chronome-repo/chronome-staging:${{ github.sha }}
          docker push asia-northeast1-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/chronome-repo/chronome-staging:latest

      - name: Deploy to Cloud Run
        run: |
          gcloud run deploy chronome-staging \
            --image=asia-northeast1-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/chronome-repo/chronome-staging:${{ github.sha }} \
            --platform=managed \
            --region=asia-northeast1 \
            --allow-unauthenticated \
            --set-env-vars="APP_ENV=staging,SERVER_ADDRESS=:8081,DB_DRIVER=postgres,SESSION_COOKIE_SECURE=true,SESSION_TTL=12h,DEFAULT_PROJECT_COLOR=#3B82F6,DB_DSN=host=/cloudsql/${{ secrets.CLOUD_SQL_INSTANCE_CONNECTION_NAME }} user=chronome_user password=${{ secrets.DB_PASSWORD }} dbname=chronome_db sslmode=disable" \
            --update-secrets="SESSION_SECRET=chronome-staging-session-secret:latest" \
            --set-cloudsql-instances="${{ secrets.CLOUD_SQL_INSTANCE_CONNECTION_NAME }}" \
            --memory=512Mi \
            --cpu=1 \
            --min-instances=0 \
            --max-instances=10 \
            --timeout=300

      - name: Get Service URL
        run: |
          SERVICE_URL=$(gcloud run services describe chronome-staging \
            --region=asia-northeast1 \
            --format="value(status.url)")
          echo "::notice::Staging deployed to: $SERVICE_URL"
          echo "::warning::初回デプロイ後、ALLOWED_ORIGINを更新してください: gcloud run services update chronome-staging --region=asia-northeast1 --update-env-vars=\"ALLOWED_ORIGIN=\$SERVICE_URL\""
```

##### 変更3: deploy-production ジョブの置き換え（149-186行目）

既存の`deploy-production`ジョブ全体を以下で置き換えてください：

```yaml
  deploy-production:
    name: Deploy · Production
    runs-on: ubuntu-latest
    needs: [backend-tests, frontend-build]
    if: ${{ startsWith(github.ref, 'refs/tags/v') }}

    permissions:
      contents: read
      id-token: write

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Authenticate to Google Cloud
        uses: google-github-actions/auth@v2
        with:
          workload_identity_provider: ${{ secrets.GCP_WORKLOAD_IDENTITY_PROVIDER }}
          service_account: ${{ secrets.GCP_SERVICE_ACCOUNT }}

      - name: Set up Cloud SDK
        uses: google-github-actions/setup-gcloud@v2

      - name: Configure Docker for Artifact Registry
        run: gcloud auth configure-docker asia-northeast1-docker.pkg.dev

      - name: Extract version
        id: version
        run: echo "VERSION=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT

      - name: Build Docker image
        run: |
          docker build -t asia-northeast1-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/chronome-repo/chronome-production:${{ github.sha }} .
          docker tag asia-northeast1-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/chronome-repo/chronome-production:${{ github.sha }} \
                     asia-northeast1-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/chronome-repo/chronome-production:${{ steps.version.outputs.VERSION }}
          docker tag asia-northeast1-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/chronome-repo/chronome-production:${{ github.sha }} \
                     asia-northeast1-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/chronome-repo/chronome-production:latest

      - name: Push Docker image
        run: |
          docker push asia-northeast1-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/chronome-repo/chronome-production:${{ github.sha }}
          docker push asia-northeast1-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/chronome-repo/chronome-production:${{ steps.version.outputs.VERSION }}
          docker push asia-northeast1-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/chronome-repo/chronome-production:latest

      - name: Deploy to Cloud Run
        run: |
          gcloud run deploy chronome-production \
            --image=asia-northeast1-docker.pkg.dev/${{ secrets.GCP_PROJECT_ID }}/chronome-repo/chronome-production:${{ github.sha }} \
            --platform=managed \
            --region=asia-northeast1 \
            --allow-unauthenticated \
            --set-env-vars="APP_ENV=production,SERVER_ADDRESS=:8081,DB_DRIVER=postgres,SESSION_COOKIE_SECURE=true,SESSION_TTL=12h,DEFAULT_PROJECT_COLOR=#3B82F6,DB_DSN=host=/cloudsql/${{ secrets.CLOUD_SQL_INSTANCE_CONNECTION_NAME }} user=chronome_user password=${{ secrets.DB_PASSWORD }} dbname=chronome_db sslmode=disable" \
            --update-secrets="SESSION_SECRET=chronome-production-session-secret:latest" \
            --set-cloudsql-instances="${{ secrets.CLOUD_SQL_INSTANCE_CONNECTION_NAME }}" \
            --memory=512Mi \
            --cpu=1 \
            --min-instances=0 \
            --max-instances=10 \
            --timeout=300

      - name: Get Service URL
        run: |
          SERVICE_URL=$(gcloud run services describe chronome-production \
            --region=asia-northeast1 \
            --format="value(status.url)")
          echo "::notice::Production (${{ steps.version.outputs.VERSION }}) deployed to: $SERVICE_URL"
          echo "::warning::初回デプロイ後、ALLOWED_ORIGINを更新してください: gcloud run services update chronome-production --region=asia-northeast1 --update-env-vars=\"ALLOWED_ORIGIN=\$SERVICE_URL\""
```

**完了後の確認:**
- YAMLの構文エラーがないか確認（インデントに注意）
- `backend-tests`と`frontend-build`ジョブは変更しない

---

### フェーズ6: developブランチのセットアップ（オプション）

developブランチをまだ作成していない場合は、以下の手順で作成します：

```bash
# developブランチを作成
git checkout -b develop

# リモートにプッシュ
git push -u origin develop

# mainブランチに戻る
git checkout main
```

今後の開発作業は基本的にdevelopブランチで行い、mainにマージすることでステージング環境にデプロイされます。

### フェーズ7: 初回ステージングデプロイ

#### 7.1 ワークフロー変更のコミット・プッシュ

```bash
# 変更をコミット
git add .github/workflows/ci.yml
git commit -m "feat(ci): GitHub Actions自動デプロイを追加"

# mainブランチにプッシュ（ステージングデプロイがトリガーされる）
git push origin main
```

#### 7.2 デプロイ監視

1. GitHubリポジトリの **Actions** タブを開く
2. 実行中のワークフローをクリック
3. `deploy-staging`ジョブのログを確認
4. エラーが発生した場合は、トラブルシューティングセクションを参照

#### 7.3 デプロイ成功後の確認

```bash
# ステージングサービスURLを取得
gcloud run services describe chronome-staging \
  --region=asia-northeast1 \
  --format="value(status.url)"
```

URLをブラウザで開いて、アプリケーションが表示されることを確認してください。

#### 7.4 ALLOWED_ORIGINの更新（重要）

初回デプロイ後、CORS設定のために`ALLOWED_ORIGIN`環境変数を更新する必要があります：

```bash
# サービスURLを取得
STAGING_URL=$(gcloud run services describe chronome-staging \
  --region=asia-northeast1 \
  --format="value(status.url)")

# ALLOWED_ORIGINを更新
gcloud run services update chronome-staging \
  --region=asia-northeast1 \
  --update-env-vars="ALLOWED_ORIGIN=${STAGING_URL}"
```

**再度ブラウザで確認:** ログイン・ユーザー登録が正しく動作することを確認してください。

---

### フェーズ8: 初回本番デプロイ

#### 8.1 Gitタグの作成とプッシュ

```bash
# タグを作成（バージョン番号は適宜変更）
git tag v0.1.0

# タグをプッシュ（本番デプロイがトリガーされる）
git push origin v0.1.0
```

#### 8.2 デプロイ監視

1. GitHubリポジトリの **Actions** タブを開く
2. タグ名で実行されたワークフローをクリック
3. `deploy-production`ジョブのログを確認

#### 8.3 デプロイ成功後の確認

```bash
# 本番サービスURLを取得
gcloud run services describe chronome-production \
  --region=asia-northeast1 \
  --format="value(status.url)"
```

#### 8.4 ALLOWED_ORIGINの更新（重要）

```bash
# サービスURLを取得
PRODUCTION_URL=$(gcloud run services describe chronome-production \
  --region=asia-northeast1 \
  --format="value(status.url)")

# ALLOWED_ORIGINを更新
gcloud run services update chronome-production \
  --region=asia-northeast1 \
  --update-env-vars="ALLOWED_ORIGIN=${PRODUCTION_URL}"
```

**本番環境の動作確認:** ブラウザでアクセスして、すべての機能が正常に動作することを確認してください。

---

## トラブルシューティング

### エラー1: WIF認証失敗

**エラーメッセージ:** `failed to generate Google Cloud access token`

**解決策:**
1. GitHub Secretsの`GCP_WORKLOAD_IDENTITY_PROVIDER`と`GCP_SERVICE_ACCOUNT`が正しいか確認
2. フェーズ2.5のWorkload Identity Bindingが正しく設定されているか確認
3. GitHubリポジトリのownerが正しいか確認

```bash
# Workload Identity Bindingの確認
gcloud iam service-accounts get-iam-policy \
  github-actions-deploy@${PROJECT_ID}.iam.gserviceaccount.com
```

### エラー2: Dockerプッシュ失敗

**エラーメッセージ:** `401 Unauthorized`

**解決策:**
1. Artifact Registryリポジトリが存在するか確認
2. サービスアカウントに`artifactregistry.writer`ロールがあるか確認

```bash
# リポジトリの確認
gcloud artifacts repositories describe chronome-repo --location=asia-northeast1

# IAMロールの確認
gcloud projects get-iam-policy $PROJECT_ID \
  --flatten="bindings[].members" \
  --filter="bindings.members:github-actions-deploy@"
```

### エラー3: Cloud Runデプロイ失敗

**エラーメッセージ:** `PERMISSION_DENIED`

**解決策:**
1. サービスアカウントに`roles/run.admin`があるか確認
2. Cloud Run APIが有効化されているか確認

```bash
# APIの有効化状態を確認
gcloud services list --enabled | grep run.googleapis.com
```

### エラー4: SESSION_SECRETエラー

**エラーメッセージ:** `SESSION_SECRET must be provided and at least 32 characters long`

**解決策:**
1. Secret Managerにシークレットが存在するか確認
2. Cloud Runサービスアカウントに`secretAccessor`ロールがあるか確認

```bash
# Secretの確認
gcloud secrets versions access latest --secret=chronome-staging-session-secret

# IAMポリシーの確認
gcloud secrets get-iam-policy chronome-staging-session-secret
```

### エラー5: データベース接続失敗

**エラーメッセージ:** `dial unix /cloudsql/...: no such file or directory`

**解決策:**
1. `--set-cloudsql-instances`が正しく設定されているか確認
2. Cloud SQL接続名の形式が正しいか確認（`project:region:instance`）

```bash
# Cloud SQL接続名の確認
gcloud sql instances describe chronome-db --format="value(connectionName)"
```

### エラー6: CORSエラー

**症状:** ブラウザコンソールに `blocked by CORS policy`

**解決策:**
フェーズ7.4またはフェーズ8.4の`ALLOWED_ORIGIN`更新を実行してください。

---

## 日常運用

### ブランチ戦略

このプロジェクトでは以下のブランチ戦略を使用します：

- **develop**: 開発作業用ブランチ（日常的な開発はこちらで行う）
- **main**: 本番準備完了コード（developからマージされた時点でステージングに自動デプロイ）
- **タグ (v*.*.*)**: 本番リリース（タグ作成時に本番環境に自動デプロイ）

### ステージング環境へのデプロイ

```bash
# 1. developブランチで開発作業
git checkout develop
# ... 開発作業 ...
git add .
git commit -m "feat: 新機能を追加"
git push origin develop

# 2. developからmainにマージ（ステージングデプロイがトリガーされる）
git checkout main
git merge develop
git push origin main

# または、GitHubでPull Requestを作成してマージ
```

### 本番環境へのデプロイ

```bash
# 1. バージョンタグを作成
git tag v1.0.0

# 2. タグをプッシュ
git push origin v1.0.0

# 自動的に本番環境にデプロイされます
```

### ロールバック手順

#### 方法1: 前のリビジョンに戻す

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

#### 方法2: 前のタグを再デプロイ

```bash
# 新しいタグを前のコミットから作成
git tag v1.0.1 <previous-commit-sha>
git push origin v1.0.1
```

---

## 検証項目

### フェーズ1-4完了後の確認

- [ ] GCPプロジェクトが正しく設定されている
- [ ] 必要なAPIがすべて有効化されている
- [ ] Artifact Registryリポジトリが作成されている
- [ ] Workload Identity Federationが設定されている
- [ ] Secret Managerに2つのセッションシークレットが保存されている
- [ ] GitHub Secretsに7つすべてが設定されている

### ステージングデプロイ完了後の確認

- [ ] GitHub Actionsワークフローが成功している
- [ ] Cloud Runサービス`chronome-staging`が作成されている
- [ ] ブラウザでステージングURLにアクセスできる
- [ ] `ALLOWED_ORIGIN`が更新されている
- [ ] ユーザー登録・ログインが動作する
- [ ] プロジェクト作成・時間記録が動作する
- [ ] データがCloud SQLに保存されている

### 本番デプロイ完了後の確認

- [ ] タグプッシュでワークフローがトリガーされる
- [ ] Cloud Runサービス`chronome-production`が作成されている
- [ ] ブラウザで本番URLにアクセスできる
- [ ] `ALLOWED_ORIGIN`が更新されている
- [ ] すべての機能が正常に動作する
- [ ] ステージングと本番のセッションが分離されている

### 運用フロー確認

- [ ] mainブランチへのpushでステージングが自動デプロイされる
- [ ] タグ作成で本番が自動デプロイされる
- [ ] ロールバック手順を理解している
- [ ] ログの確認方法を理解している

---

## コスト見積もり

| サービス | 月額費用 |
|---------|---------|
| Cloud Run（ステージング） | $0-2 |
| Cloud Run（本番） | $2-10 |
| Cloud SQL（共有） | $15-20 |
| Artifact Registry | $0.20-0.30 |
| Secret Manager | $0（無料枠） |
| **合計** | **$17-33/月** |

---

## セキュリティのベストプラクティス

- ✅ Workload Identity Federation使用（キーレス認証）
- ✅ Secret Managerでシークレット管理
- ✅ 環境ごとに異なるSESSION_SECRET
- ✅ HTTPSのみでCookie送信
- ✅ 最小権限の原則でIAMロール設定

---

**作成日:** 2026-07-07
