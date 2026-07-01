# 現在の iOS アーキテクチャとデータフロー

このドキュメントは、現在の iOS 実装に基づくアーキテクチャ構成とデータフローを示す。

設計資料では The Composable Architecture (TCA) の採用を計画しているが、現時点の実装は
SwiftUI + `AppFeature` (`ObservableObject`) + Client/Store 抽象 + SwiftData + 既存 Go API
という構成である。

## アーキテクチャ構成

発表などで全体像を説明する場合は、現在の iOS 版を以下のように抽象化できる。

```mermaid
flowchart TB
    User["ユーザー"]

    subgraph IOS["iOSアプリ"]
        View["SwiftUI View<br/>画面表示・操作受付"]
        Feature["AppFeature<br/>状態管理・アプリロジック"]
        LocalStore["Local Store<br/>SwiftData"]
        APIClient["API Client<br/>REST通信"]
    end

    subgraph Backend["既存Web版バックエンド"]
        API["Go REST API"]
        DB["Database"]
    end

    User --> View
    View --> Feature
    Feature --> View

    Feature --> LocalStore
    LocalStore --> Feature

    Feature --> APIClient
    APIClient --> API
    API --> DB
    DB --> API
    API --> APIClient
    APIClient --> Feature
```

レイヤー構成として整理すると、以下のようになる。

```mermaid
flowchart TB
    Presentation["Presentation Layer<br/>SwiftUI Views"]
    Application["Application Layer<br/>AppFeature / State / Actions"]
    Data["Data Layer<br/>SwiftData Store / API Clients"]
    External["External Systems<br/>Go API / Database"]

    Presentation -->|"ユーザー操作"| Application
    Application -->|"状態更新"| Presentation

    Application -->|"ローカル保存・取得"| Data
    Data -->|"ローカルデータ"| Application

    Data -->|"REST API通信"| External
    External -->|"JSONレスポンス"| Data
```

詳細な実装構成は以下の通り。

```mermaid
flowchart TB
    subgraph App["iOS App"]
        ChronoMeApp["ChronoMeApp<br/>App Entry Point"]
        ContentView["ContentView<br/>認証状態で画面切替"]
        TimerHomeView["TimerHomeView<br/>TabView / ScrollView / Sheet"]
        LoginView["LoginView"]
    end

    subgraph State["State / Application Logic"]
        AppFeature["AppFeature<br/>ObservableObject<br/>状態管理・画面イベント・同期制御"]
    end

    subgraph Local["Local Persistence"]
        TimeEntryStoreProtocol["TimeEntryStoring<br/>Protocol"]
        SwiftDataStore["SwiftDataTimeEntryStore"]
        SwiftData["SwiftData<br/>TimeEntryRecord"]
    end

    subgraph API["API Layer"]
        AuthClient["AuthClient"]
        ProjectClient["ProjectClient"]
        TagClient["TagClient"]
        EntryClient["EntryClient"]
        APIClient["APIClient<br/>URLSession / Cookie / CSRF / JSON"]
    end

    subgraph Backend["Backend"]
        GoAPI["Go REST API<br/>/api/auth<br/>/api/projects<br/>/api/tags<br/>/api/entries"]
    end

    subgraph Models["Models"]
        AuthUser["AuthUser"]
        Project["Project"]
        Tag["Tag"]
        Entry["Entry"]
        TimeEntryRecord["TimeEntryRecord"]
    end

    ChronoMeApp --> ContentView
    ContentView --> LoginView
    ContentView --> TimerHomeView
    LoginView --> AppFeature
    TimerHomeView --> AppFeature

    ChronoMeApp --> AppFeature
    ChronoMeApp --> SwiftDataStore
    ChronoMeApp --> AuthClient
    ChronoMeApp --> ProjectClient
    ChronoMeApp --> TagClient
    ChronoMeApp --> EntryClient

    AppFeature --> TimeEntryStoreProtocol
    TimeEntryStoreProtocol --> SwiftDataStore
    SwiftDataStore --> SwiftData
    SwiftData --> TimeEntryRecord

    AppFeature --> AuthClient
    AppFeature --> ProjectClient
    AppFeature --> TagClient
    AppFeature --> EntryClient

    AuthClient --> APIClient
    ProjectClient --> APIClient
    TagClient --> APIClient
    EntryClient --> APIClient
    APIClient --> GoAPI

    AuthClient --> AuthUser
    ProjectClient --> Project
    TagClient --> Tag
    EntryClient --> Entry
    AppFeature --> Project
    AppFeature --> Tag
    AppFeature --> TimeEntryRecord
```

## 起動時・認証時のデータフロー

```mermaid
sequenceDiagram
    participant App as ChronoMeApp
    participant View as ContentView
    participant Feature as AppFeature
    participant Auth as AuthClient
    participant Project as ProjectClient
    participant Tag as TagClient
    participant Entry as EntryClient
    participant Store as SwiftDataTimeEntryStore
    participant API as Go API

    App->>Feature: 依存を注入して初期化
    App->>View: ContentView(feature)

    View->>Feature: restoreSession()

    alt DEBUG / skipsAuthentication = true
        Feature->>Feature: localDevelopmentUser を signedIn に設定
        Feature->>Store: fetchProjects() / fetchTags()
        Store-->>Feature: local projects / tags
        Feature->>Store: fetchRecent()
        Store-->>Feature: recentEntries
        Feature->>Store: fetchEntries(selectedDate)
        Store-->>Feature: selectedDateEntries
        Feature-->>View: TimerHomeView 表示
    else 認証有効時
        Feature->>Auth: currentUser()
        Auth->>API: GET /api/auth/me
        API-->>Auth: user or 401
        Auth-->>Feature: AuthUser?

        alt signed in
            Feature->>Project: listProjects()
            Project->>API: GET /api/projects/
            API-->>Project: projects

            Feature->>Tag: listTags()
            Tag->>API: GET /api/tags/
            API-->>Tag: tags

            Feature->>Store: fetchUnsynced()
            Store-->>Feature: unsynced entries

            Feature->>Entry: listEntries(current month)
            Entry->>API: GET /api/entries/
            API-->>Entry: entries

            Feature->>Store: upsertRemoteEntry()
            Feature-->>View: TimerHomeView 表示
        else signed out
            Feature-->>View: LoginView 表示
        end
    end
```

## タイマー記録・同期のデータフロー

```mermaid
sequenceDiagram
    participant User as User
    participant View as TimerHomeView
    participant Feature as AppFeature
    participant Store as SwiftDataTimeEntryStore
    participant Entry as EntryClient
    participant API as Go API

    User->>View: プロジェクトを選んで開始
    View->>Feature: startTimerSession(projectID, notes, tagIDs)
    Feature->>Feature: TimerSession を追加<br/>Timer開始

    User->>View: 作業を終了
    View->>Feature: stopTimerSession(sessionID)
    Feature->>Feature: TimerSession を終了
    Feature->>Store: save(syncStatus: pending)
    Store-->>Feature: TimeEntryRecord
    Feature->>Store: fetchRecent()
    Feature->>Store: fetchEntries(selectedDate)

    alt DEBUG / skipsAuthentication = true
        Feature->>Store: markSynced(remoteEntryID: local UUID)
        Store-->>Feature: synced
    else 認証有効時
        Feature->>Entry: createEntry(...)
        Entry->>API: POST /api/entries/
        alt API success
            API-->>Entry: Entry
            Entry-->>Feature: remote entry
            Feature->>Store: markSynced(remoteEntryID)
        else API failure
            Entry-->>Feature: error
            Feature->>Store: markSyncFailed()
            Feature->>Feature: errorMessage = ローカル保存したが同期失敗
        end
    end

    Feature-->>View: 作業履歴・集計を更新
```

## 画面イベントから状態更新まで

```mermaid
flowchart LR
    User["ユーザー操作"] --> View["SwiftUI View<br/>Button / Picker / DatePicker / Swipe"]
    View --> Action["AppFeature のメソッド呼び出し"]

    Action --> State["Published State 更新<br/>authState / projects / tags<br/>selectedDateEntries / errorMessage"]
    Action --> LocalOp["SwiftData 操作<br/>保存・取得・更新・削除"]
    Action --> RemoteOp["API 操作<br/>認証・同期・CRUD"]

    LocalOp --> State
    RemoteOp --> State
    State --> View

    View --> UI["UI再描画"]
```

## 要約

現在の iOS 版では、View は `AppFeature` の状態を表示し、ユーザー操作を `AppFeature`
のメソッド呼び出しとして渡す。`AppFeature` はローカル保存と API 通信をまとめて制御し、
SwiftData を先に更新してから、必要に応じてバックエンド API と同期する。

この構成により、ローカルでの素早い記録、同期失敗時の保持、後続の再同期が可能になっている。
