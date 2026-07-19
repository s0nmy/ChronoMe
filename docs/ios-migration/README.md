# ChronoMe iOS移植計画

## 概要

ChronoMeは個人向けタイムカードWebアプリケーションです。本ドキュメントでは、既存のWeb版（React + TypeScript）をiOSネイティブアプリケーション（Swift）に移植する計画をまとめています。

## 移植の目的

- **モバイルファースト体験**: iOSネイティブアプリとして、より快適なモバイル体験を提供
- **オフライン対応**: ローカルデータベースを活用し、オフライン環境でも作業時間を記録可能
- **プラットフォーム固有機能の活用**: ウィジェット、通知、Face ID/Touch IDなどのiOS固有機能の統合
- **パフォーマンス向上**: ネイティブアプリならではの高速なUI/UXの実現

## ドキュメント構成

本ディレクトリには以下のドキュメントが含まれています：

- **[TechStack.md](TechStack.md)** - iOS技術スタックの選定と理由
- **[Architecture.md](Architecture.md)** - アーキテクチャ設計とレイヤー構成
- **[CurrentArchitecture.md](CurrentArchitecture.md)** - 現在のiOS実装に基づくアーキテクチャとデータフロー
- **[FeatureParity.md](FeatureParity.md)** - Web版との機能対応表
- **[PhasesPlan.md](PhasesPlan.md)** - 段階的な実装計画とマイルストーン

## 移植スコープ

### 含まれる機能

- 認証（サインアップ/ログイン/ログアウト）
- タイムエントリの作成・編集・削除
- プロジェクト管理
- タグ付けと分類
- レポート表示（日次/週次/月次）
- データエクスポート（CSV/JSON）

### iOS固有の追加機能

- ホーム画面ウィジェット（今日の作業時間表示）
- プッシュ通知（作業開始/終了のリマインダー）
- Face ID/Touch ID認証
- Apple Watch対応（将来的な拡張）
- ショートカットアプリ連携

## 前提条件

- **開発環境**: Xcode 26以上
- **対象iOS**: iOS 17以上
- **バックエンド**: 既存のGo製APIサーバーを継続利用
- **API互換性**: REST API (`/api/*`) は現行仕様を維持

## Web版との関係

- Web版とiOS版は同一のバックエンドAPIを共有
- ユーザーデータは同一サーバー上で管理され、Web/iOS間で同期
- 認証セッションはプラットフォーム間で独立
- UIデザインはiOS Human Interface Guidelinesに準拠し、Web版とは独立したデザイン

## 次のステップ

1. [技術スタック](TechStack.md)の確認と承認
2. [アーキテクチャ設計](Architecture.md)のレビュー
3. [実装計画](PhasesPlan.md)に基づいた段階的な開発
4. 各フェーズでのテストとレビュー

## 参考

- [Web版PRD](../product/PRD.md)
- [API設計](../architecture/APIDesign.md)
- [DB設計](../architecture/DBDesign.md)
