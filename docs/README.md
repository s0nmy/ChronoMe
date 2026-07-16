# ChronoMe ドキュメント

ChronoMe（一人用タイムカード Web アプリ）の全ドキュメントへのナビゲーションです。

## 📋 プロダクト仕様

プロダクトの要件定義と全体的な技術設計を記載しています。

- [PRD.md](product/PRD.md) - プロダクト要件定義書
  - 概要、目的、ユーザーストーリー、提案する解決策
- [DesignDoc.md](product/DesignDoc.md) - 技術設計書
  - システムアーキテクチャ、技術スタック、実装方針、テスト設計

## 🏗️ アーキテクチャ設計

システムの内部構造とデータ設計に関する詳細ドキュメントです。

- [CleanArchitecture.md](architecture/CleanArchitecture.md) - クリーンアーキテクチャ実装ガイド
  - 層ごとの責務、依存性注入、テスト方針、コード例
- [APIDesign.md](architecture/APIDesign.md) - API 設計書
  - REST API 仕様、リクエスト/レスポンス詳細
- [DBDesign.md](architecture/DBDesign.md) - データベース設計書
  - テーブル定義、制約、インデックス方針
- [cloud-run-supabase-architecture.md](architecture/cloud-run-supabase-architecture.md) - Cloud Run + Supabase アーキテクチャ
  - クラウド環境での構成設計

## 🚀 デプロイメント

本番環境へのデプロイに関するガイドです。

- [gcp-cd-setup-guide.md](deployment/gcp-cd-setup-guide.md) - GCP CD環境セットアップガイド（Supabase対応）
- [deployment-guide.md](deployment/deployment-guide.md) - デプロイ運用ガイド
- [deploy-gcp-cloud-run.md](deployment/deploy-gcp-cloud-run.md) - GCP Cloud Run デプロイガイド（参考）
- [github-actions-deployment-guide.md](deployment/github-actions-deployment-guide.md) - GitHub Actions デプロイガイド（参考）

## 💻 開発ガイドライン

開発時の規約とテスト戦略です。

- [CommitGuidelines.md](development/CommitGuidelines.md) - コミットメッセージ規約
- [TestStrategy.md](development/TestStrategy.md) - テスト戦略と実装例

## 🔌 API 仕様

特定の API に関する詳細仕様です。

- [TimeAllocationAPI.md](api/TimeAllocationAPI.md) - 時間配分 API 仕様

## 📱 iOS マイグレーション

Web 版から iOS アプリへの移行計画です。

- [ios-migration/README.md](ios-migration/README.md) - iOS マイグレーション概要
- [ios-migration/TechStack.md](ios-migration/TechStack.md) - iOS 技術スタック
- [ios-migration/Architecture.md](ios-migration/Architecture.md) - iOS アーキテクチャ設計
- [ios-migration/FeatureParity.md](ios-migration/FeatureParity.md) - 機能対応表
- [ios-migration/PhasesPlan.md](ios-migration/PhasesPlan.md) - 実装計画

## 🎨 WordPress 展示サイト

ポートフォリオ用の展示サイト素材です。

- [wordpress/exhibition-site.html](wordpress/exhibition-site.html) - 展示サイト HTML
- [wordpress/exhibition-site.css](wordpress/exhibition-site.css) - 展示サイト CSS

## 📝 ADR (Architecture Decision Records)

アーキテクチャ上の重要な意思決定の記録です。

- [adr/0001-use-supabase-postgresql-for-portfolio-database.md](adr/0001-use-supabase-postgresql-for-portfolio-database.md) - Supabase PostgreSQL 採用の決定
- [adr/0003-adopt-github-actions-cd-with-staging-and-production.md](adr/0002-adopt-github-actions-cd-with-staging-and-production.md) - GitHub Actions によるステージング・本番の自動デプロイ環境構築の決定

## 🔧 ワークフロー

Claude Code のカスタムスキル定義です。

- [workflows/create-pull-request/](workflows/create-pull-request/) - PR 作成スキル
- [workflows/suggest-branch-name/](workflows/suggest-branch-name/) - ブランチ名提案スキル

---

## ドキュメント構造

```
docs/
├── README.md                          # このファイル
│
├── product/                           # プロダクト仕様
│   ├── PRD.md
│   └── DesignDoc.md
│
├── architecture/                      # アーキテクチャ設計
│   ├── CleanArchitecture.md
│   ├── APIDesign.md
│   ├── DBDesign.md
│   └── cloud-run-supabase-architecture.md
│
├── deployment/                        # デプロイメント
│   ├── gcp-cd-setup-guide.md
│   ├── deployment-guide.md
│   ├── deploy-gcp-cloud-run.md
│   └── github-actions-deployment-guide.md
│
├── development/                       # 開発ガイドライン
│   ├── CommitGuidelines.md
│   └── TestStrategy.md
│
├── api/                              # API 仕様
│   └── TimeAllocationAPI.md
│
├── wordpress/                         # WordPress 展示サイト
│   ├── exhibition-site.html
│   └── exhibition-site.css
│
├── adr/                              # Architecture Decision Records
├── ios-migration/                    # iOS マイグレーション
└── workflows/                         # Claude Code スキル
```

## 閲覧推奨順序

初めて ChronoMe のドキュメントを読む場合、以下の順序をおすすめします。

1. [PRD.md](product/PRD.md) - プロダクトの目的と要件を理解
2. [DesignDoc.md](product/DesignDoc.md) - システム全体の設計を把握
3. [CleanArchitecture.md](architecture/CleanArchitecture.md) - 実装方針の詳細
4. [APIDesign.md](architecture/APIDesign.md) / [DBDesign.md](architecture/DBDesign.md) - API/DB の詳細仕様
5. [CommitGuidelines.md](development/CommitGuidelines.md) / [TestStrategy.md](development/TestStrategy.md) - 開発規約

---

最終更新: 2026-07-16
