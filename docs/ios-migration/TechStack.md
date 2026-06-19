# iOS技術スタック

## 概要

ChronoMe iOSアプリの開発に使用する技術スタックを定義します。モダンなSwift開発のベストプラクティスに従い、保守性と拡張性を重視した選定を行っています。

## 言語・フレームワーク

### Swift

- **バージョン**: Swift 5.9以上
- **理由**:
  - Appleの公式言語で、最新のiOS機能へのフルアクセス
  - 型安全性と現代的な言語機能（async/await、Actorなど）
  - 豊富なコミュニティとライブラリ

### SwiftUI

- **用途**: UIフレームワーク
- **理由**:
  - 宣言的UIで開発効率が高い
  - iOS 17以降の最新機能をフルサポート
  - プレビュー機能による高速な開発サイクル
  - ウィジェット、Apple Watch対応が容易

## アーキテクチャパターン

### The Composable Architecture (TCA)

- **ライブラリ**: [pointfreeco/swift-composable-architecture](https://github.com/pointfreeco/swift-composable-architecture)
- **理由**:
  - 状態管理が明確で、複雑なビジネスロジックを整理しやすい
  - テスタビリティが高い
  - 副作用の管理が容易（API通信、データベースアクセスなど）
  - Web版のReduxライクな状態管理と概念が近い

### 代替案検討

- **MVVM**: シンプルだがテストの難しさと状態管理の複雑化
- **Clean Architecture**: 過度な抽象化によるボイラープレート増加
- **決定**: TCAを採用（規模に対して適切なバランス）

## ネットワーキング

### URLSession + async/await

- **用途**: HTTP通信
- **理由**:
  - 標準ライブラリで追加依存なし
  - Swift Concurrency（async/await）による可読性の高いコード
  - Web版と同じREST APIを利用

### 補助ライブラリ（検討中）

- **Alamofire**: より高度なネットワーク機能が必要な場合に検討
- **現時点**: 標準URLSessionで開始し、必要に応じて追加

## データ永続化

### SwiftData

- **用途**: ローカルデータベース
- **バージョン**: iOS 17以上
- **理由**:
  - SwiftUIとのネイティブ統合
  - シンプルなAPI（@Modelマクロ）
  - CloudKitとの連携が容易
  - Core Dataの後継として推奨

### 保存データ

- ユーザー認証情報（Keychain）
- タイムエントリのキャッシュ（オフライン対応）
- プロジェクト・タグのローカルコピー
- ユーザー設定

## 認証・セキュリティ

### Keychain Services

- **用途**: 認証トークン、機密情報の保存
- **理由**: iOSの標準的なセキュアストレージ

### LocalAuthentication

- **用途**: Face ID / Touch ID
- **理由**: 生体認証による安全で快適なログイン体験

## 通知

### UserNotifications Framework

- **用途**: ローカル通知、リモート通知
- **ユースケース**:
  - 作業開始リマインダー
  - 作業終了リマインダー
  - 週次レポートの通知

## ウィジェット

### WidgetKit

- **用途**: ホーム画面ウィジェット
- **表示内容**:
  - 今日の合計作業時間
  - 現在進行中のタイムエントリ
  - 週間サマリー

## テスト

### XCTest

- **用途**: ユニットテスト、UIテスト
- **理由**: Xcode標準のテストフレームワーク

### TCAのテストサポート

- **用途**: Reducerのロジックテスト
- **理由**: TCA組み込みのテストツールで副作用を含むテストが容易

## CI/CD

### Xcode Cloud（検討中）

- **用途**: ビルド、テスト、配布の自動化
- **代替**: GitHub Actions（コスト面で有利な場合）

## 依存関係管理

### Swift Package Manager (SPM)

- **理由**:
  - Xcodeネイティブサポート
  - CocoaPodsやCarthageよりもモダン
  - ビルドが高速

### 主要な依存関係

```swift
dependencies: [
    .package(url: "https://github.com/pointfreeco/swift-composable-architecture", from: "1.0.0"),
    // 必要に応じて追加
]
```

## デザイン・UI

### SF Symbols

- **用途**: アイコン
- **理由**: iOSネイティブで統一感のあるデザイン

### カスタムデザインシステム

- **カラー**: Asset Catalogでダークモード対応
- **フォント**: システムフォント（San Francisco）をベースに統一
- **コンポーネント**: 再利用可能なSwiftUIビューを構築

## 開発ツール

### Xcode 15以上

- **理由**: 最新のSwift・SwiftUIサポート

### SwiftLint

- **用途**: コードスタイルの統一
- **理由**: チーム開発でのコード品質維持

### SwiftFormat

- **用途**: コードフォーマット自動化
- **理由**: フォーマットの一貫性を保証

## 対象OS

### iOS 17.0以上

- **理由**:
  - SwiftDataの活用
  - 最新のSwiftUI機能
  - ウィジェット・通知の最新API
  - 市場シェア: iOS 17以上のユーザーが大多数（2025年時点）

## デバイスサポート

- **iPhone**: iPhone XS以降（iOS 17対応機種）
- **iPad**: サポート予定（ユニバーサルアプリとして）
- **Apple Watch**: Phase 2以降で検討

## パフォーマンス・モニタリング

### Instruments（Xcode組み込み）

- **用途**: パフォーマンス分析、メモリリーク検出

### OSLog

- **用途**: 構造化ロギング
- **理由**: Appleの推奨ログフレームワーク

## まとめ

| カテゴリ | 選定技術 | 理由 |
|---------|---------|------|
| 言語 | Swift 5.9+ | 型安全性、最新機能 |
| UI | SwiftUI | 宣言的UI、開発効率 |
| アーキテクチャ | TCA | 状態管理、テスタビリティ |
| ネットワーク | URLSession | 標準ライブラリ、async/await |
| DB | SwiftData | iOS 17ネイティブ、シンプル |
| 認証 | Keychain | セキュア、標準 |
| テスト | XCTest + TCA | 統合されたテスト環境 |
| 依存管理 | SPM | Xcodeネイティブ |
| 最小OS | iOS 17.0 | 最新機能の活用 |

この技術スタックにより、モダンで保守性の高いiOSアプリケーションを構築します。
