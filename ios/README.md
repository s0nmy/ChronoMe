# ChronoMe for iOS

## Requirements

- Xcode 26 or later
- iOS 17 or later

## Run

1. Open `ChronoMe.xcodeproj` in Xcode.
2. Select the `ChronoMe` scheme and an iPhone simulator.
3. Press Run (`Command-R`).

The initial runnable target currently uses only Apple frameworks, so no dependency
resolution is required (Swift Package Manager is not used yet).

The app currently skips authentication for local iOS verification and opens the
home screen as a local development user. Auth screens and API clients remain in
the codebase, but `ChronoMeApp` passes `skipsAuthentication: true` while this
temporary mode is active.

In this temporary mode, project/tag creation, entry editing/deletion, and timer
recording run locally without calling the backend API.

When authentication is enabled again, the app restores an existing session with
`GET /api/auth/me` on launch. If no session exists, it shows a minimal
login/signup screen. Auth requests use the backend cookie session and copy the
`chronome_csrf` cookie into the `X-CSRF-Token` header for mutating requests.

After login, the timer can be associated with a project, tags, and notes.
Stopping the timer persists a local work entry with SwiftData, then attempts to
sync it with `POST /api/entries/`. If sync fails, the local entry remains visible
as unsynced. The sync button retries unsynced local entries and imports the
current month's remote entries into SwiftData.

The signed-in home screen also fetches and displays projects and tags from the
backend with:

- `GET /api/projects/`
- `GET /api/tags/`
- `POST /api/projects/`
- `PATCH /api/projects/{id}`
- `POST /api/tags/`
- `PATCH /api/tags/{id}`
- `GET /api/entries/`
- `POST /api/entries/`
- `PATCH /api/entries/{id}`
- `DELETE /api/entries/{id}`

Recent entries can be tapped to edit project, tags, and notes. Swipe an entry to
delete it.
Projects and tags can be added or edited from their home sections.
The home screen includes a date picker to filter entries by day and displays the
selected day's total duration. It also shows a daily summary with project and
tag breakdown bars.

## Test

```sh
xcodebuild -quiet -project ios/ChronoMe.xcodeproj -scheme ChronoMe -sdk iphonesimulator -destination 'platform=iOS Simulator,name=iPhone 16,OS=18.6,arch=arm64' -derivedDataPath ios/DerivedData CODE_SIGNING_ALLOWED=NO test
```
