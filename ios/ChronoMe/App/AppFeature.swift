import Foundation

@MainActor
final class AppFeature: ObservableObject {
    enum AuthState: Equatable {
        case checking
        case signedOut
        case signedIn(AuthUser)
    }

    struct EntryEditDraft: Identifiable {
        let id: UUID
        let entry: TimeEntryRecord
        var projectID: String?
        var selectedTagIDs: Set<String>
        var notes: String
    }

    struct ProjectEditDraft: Equatable, Identifiable {
        let id: String
        let project: Project?
        var name: String
        var description: String
        var color: String
        var isArchived: Bool
    }

    struct TagEditDraft: Equatable, Identifiable {
        let id: String
        let tag: Tag?
        var name: String
        var color: String
    }

    struct DailySummaryItem: Equatable, Identifiable {
        let id: String
        let name: String
        let color: String
        let durationSeconds: Int
        let ratio: Double
    }

    @Published private(set) var authState: AuthState = .checking
    @Published private(set) var isAuthRequestInFlight = false
    @Published private(set) var recentEntries: [TimeEntryRecord] = []
    @Published private(set) var selectedDateEntries: [TimeEntryRecord] = []
    @Published private(set) var selectedEntryDate = Date()
    @Published private(set) var projects: [Project] = []
    @Published private(set) var tags: [Tag] = []
    @Published private(set) var isLoadingWorkspaceData = false
    @Published private(set) var isSyncingEntries = false
    @Published var entryEditDraft: EntryEditDraft?
    @Published var projectEditDraft: ProjectEditDraft?
    @Published var tagEditDraft: TagEditDraft?
    @Published private(set) var errorMessage: String?

    // 複数タイマー管理用
    @Published private(set) var timerSessions: [TimerSession] = []
    private var sessionTimer: Timer?
    private let entryStore: TimeEntryStoring
    private let authClient: AuthClientProtocol
    private let projectClient: ProjectClientProtocol
    private let tagClient: TagClientProtocol
    private let entryClient: EntryClientProtocol
    private let workspaceStore: WorkspaceStoring?
    private let skipsAuthentication: Bool

    init(
        entryStore: TimeEntryStoring,
        authClient: AuthClientProtocol,
        projectClient: ProjectClientProtocol,
        tagClient: TagClientProtocol,
        entryClient: EntryClientProtocol,
        workspaceStore: WorkspaceStoring? = nil,
        skipsAuthentication: Bool = false
    ) {
        self.entryStore = entryStore
        self.authClient = authClient
        self.projectClient = projectClient
        self.tagClient = tagClient
        self.entryClient = entryClient
        self.workspaceStore = workspaceStore
        self.skipsAuthentication = skipsAuthentication
        if skipsAuthentication {
            authState = .signedIn(Self.localDevelopmentUser())
            loadLocalWorkspaceData()
        }
        loadRecentEntries()
        loadSelectedDateEntries()
    }

    var selectedDateTotalSeconds: Int {
        selectedDateEntries.reduce(0) { $0 + $1.durationSeconds }
    }

    var selectedDateEntryCount: Int {
        selectedDateEntries.count
    }

    var selectedDateProjectSummaries: [DailySummaryItem] {
        let total = max(selectedDateTotalSeconds, 1)
        var durations: [String: Int] = [:]
        var names: [String: String] = [:]
        var colors: [String: String] = [:]

        for entry in selectedDateEntries {
            let id = entry.projectID ?? "unassigned"
            durations[id, default: 0] += entry.durationSeconds
            names[id] = entry.projectName ?? "未分類"
            colors[id] = projects.first { $0.id == entry.projectID }?.color ?? "#94A3B8"
        }

        return durations.map { id, duration in
            DailySummaryItem(
                id: id,
                name: names[id] ?? "未分類",
                color: colors[id] ?? "#94A3B8",
                durationSeconds: duration,
                ratio: Double(duration) / Double(total)
            )
        }
        .sorted { $0.durationSeconds > $1.durationSeconds }
    }

    var selectedDateTagSummaries: [DailySummaryItem] {
        let total = max(selectedDateTotalSeconds, 1)
        var durations: [String: Int] = [:]
        var names: [String: String] = [:]
        var colors: [String: String] = [:]

        for entry in selectedDateEntries {
            let tagIDs = entry.tagIDList
            let tagNames = entry.tagNameList
            if tagIDs.isEmpty && tagNames.isEmpty {
                durations["untagged", default: 0] += entry.durationSeconds
                names["untagged"] = "タグなし"
                colors["untagged"] = "#94A3B8"
                continue
            }

            for (index, tagID) in tagIDs.enumerated() {
                let tag = tags.first { $0.id == tagID }
                durations[tagID, default: 0] += entry.durationSeconds
                names[tagID] = tag?.name ?? tagNames[safe: index] ?? "タグ"
                colors[tagID] = tag?.color ?? "#94A3B8"
            }
        }

        return durations.map { id, duration in
            DailySummaryItem(
                id: id,
                name: names[id] ?? "タグ",
                color: colors[id] ?? "#94A3B8",
                durationSeconds: duration,
                ratio: Double(duration) / Double(total)
            )
        }
        .sorted { $0.durationSeconds > $1.durationSeconds }
    }

    func restoreSession() {
        if skipsAuthentication {
            authState = .signedIn(Self.localDevelopmentUser())
            loadLocalWorkspaceData()
            return
        }

        Task {
            isAuthRequestInFlight = true
            defer { isAuthRequestInFlight = false }

            do {
                if let user = try await authClient.currentUser() {
                    authState = .signedIn(user)
                    await loadWorkspaceData()
                    await syncEntries()
                } else {
                    authState = .signedOut
                }
                errorMessage = nil
            } catch {
                authState = .signedOut
                errorMessage = "ログイン状態を確認できませんでした。"
            }
        }
    }

    func loginButtonTapped(email: String, password: String) {
        Task {
            await authenticate {
                try await self.authClient.login(email: email, password: password)
            }
        }
    }

    func signupButtonTapped(email: String, password: String, displayName: String?) {
        Task {
            await authenticate {
                try await self.authClient.signup(
                    email: email,
                    password: password,
                    displayName: displayName,
                    timeZone: TimeZone.current.identifier
                )
            }
        }
    }

    func logoutButtonTapped() {
        guard !skipsAuthentication else {
            authState = .signedIn(Self.localDevelopmentUser())
            return
        }

        Task {
            isAuthRequestInFlight = true
            defer { isAuthRequestInFlight = false }

            do {
                try await authClient.logout()
                authState = .signedOut
                projects = []
                tags = []
                timerSessions = []
                sessionTimer?.invalidate()
                sessionTimer = nil
                errorMessage = nil
            } catch {
                errorMessage = "ログアウトできませんでした。"
            }
        }
    }

    func refreshWorkspaceData() {
        Task {
            await loadWorkspaceData()
            await syncEntries()
        }
    }

    func syncButtonTapped() {
        Task {
            await syncEntries(for: selectedEntryDate)
        }
    }

    func selectedEntryDateChanged(_ date: Date) {
        selectedEntryDate = date
        loadSelectedDateEntries()
        Task {
            await syncEntries(for: date)
        }
    }

    func entryTapped(_ entry: TimeEntryRecord) {
        entryEditDraft = EntryEditDraft(
            id: entry.id,
            entry: entry,
            projectID: entry.projectID,
            selectedTagIDs: Set(entry.tagIDList),
            notes: entry.notes
        )
    }

    func editProjectSelectionChanged(_ projectID: String?) {
        entryEditDraft?.projectID = projectID
    }

    func editTagSelectionToggled(_ tagID: String) {
        guard var draft = entryEditDraft else { return }
        if draft.selectedTagIDs.contains(tagID) {
            draft.selectedTagIDs.remove(tagID)
        } else {
            draft.selectedTagIDs.insert(tagID)
        }
        entryEditDraft = draft
    }

    func editNotesChanged(_ notes: String) {
        entryEditDraft?.notes = notes
    }

    func cancelEntryEdit() {
        entryEditDraft = nil
    }

    func saveEntryEdit() {
        guard let draft = entryEditDraft else { return }
        Task {
            await updateEntry(draft)
        }
    }

    func addProjectButtonTapped() {
        projectEditDraft = ProjectEditDraft(
            id: "new-project",
            project: nil,
            name: "",
            description: "",
            color: "#3B82F6",
            isArchived: false
        )
    }

    func projectTapped(_ project: Project) {
        projectEditDraft = ProjectEditDraft(
            id: project.id,
            project: project,
            name: project.name,
            description: project.description ?? "",
            color: project.color,
            isArchived: project.isArchived
        )
    }

    func projectDraftChanged(name: String? = nil, description: String? = nil, color: String? = nil, isArchived: Bool? = nil) {
        guard var draft = projectEditDraft else { return }
        if let name { draft.name = name }
        if let description { draft.description = description }
        if let color { draft.color = color }
        if let isArchived { draft.isArchived = isArchived }
        projectEditDraft = draft
    }

    func cancelProjectEdit() {
        projectEditDraft = nil
    }

    func saveProjectEdit() {
        guard let draft = projectEditDraft else { return }
        Task {
            await saveProject(draft)
        }
    }

    func addTagButtonTapped() {
        tagEditDraft = TagEditDraft(
            id: "new-tag",
            tag: nil,
            name: "",
            color: "#F97316"
        )
    }

    func tagTapped(_ tag: Tag) {
        tagEditDraft = TagEditDraft(
            id: tag.id,
            tag: tag,
            name: tag.name,
            color: tag.color
        )
    }

    func tagDraftChanged(name: String? = nil, color: String? = nil) {
        guard var draft = tagEditDraft else { return }
        if let name { draft.name = name }
        if let color { draft.color = color }
        tagEditDraft = draft
    }

    func cancelTagEdit() {
        tagEditDraft = nil
    }

    func saveTagEdit() {
        guard let draft = tagEditDraft else { return }
        Task {
            await saveTag(draft)
        }
    }

    func deleteEntry(_ entry: TimeEntryRecord) {
        Task {
            await deleteEntryFromStoreAndRemote(entry)
        }
    }

    // MARK: - 複数タイマー管理

    func startTimerSession(projectID: String?, notes: String, tagIDs: Set<String>) {
        let session = TimerSession(
            projectID: projectID,
            notes: notes,
            tagIDs: tagIDs,
            isBreak: false,
            startedAt: Date()
        )
        timerSessions.append(session)
        startSessionTimer()
    }

    func pauseTimerSession(_ sessionID: UUID) {
        guard let index = timerSessions.firstIndex(where: { $0.id == sessionID }) else { return }
        var session = timerSessions[index]

        if !session.isPaused {
            session.lastPausedAt = Date()
            timerSessions[index] = session
        }
    }

    func resumeTimerSession(_ sessionID: UUID) {
        guard let index = timerSessions.firstIndex(where: { $0.id == sessionID }) else { return }
        var session = timerSessions[index]

        if let pausedAt = session.lastPausedAt {
            let pauseDuration = Int(Date().timeIntervalSince(pausedAt))
            session.pausedDurationSeconds += pauseDuration
            session.lastPausedAt = nil
            timerSessions[index] = session
        }
    }

    func stopTimerSession(_ sessionID: UUID) {
        guard let index = timerSessions.firstIndex(where: { $0.id == sessionID }) else { return }
        let session = timerSessions[index]

        Task {
            await saveTimerSession(session)
            timerSessions.remove(at: index)

            if timerSessions.isEmpty {
                stopSessionTimer()
            }
        }
    }

    func updateTimerSession(_ sessionID: UUID, notes: String?, tagIDs: Set<String>?, isBreak: Bool?) {
        guard let index = timerSessions.firstIndex(where: { $0.id == sessionID }) else { return }
        var session = timerSessions[index]

        if let notes { session.notes = notes }
        if let tagIDs { session.tagIDs = tagIDs }
        if let isBreak { session.isBreak = isBreak }

        timerSessions[index] = session
    }

    private func startSessionTimer() {
        guard sessionTimer == nil else { return }

        sessionTimer = Timer.scheduledTimer(withTimeInterval: 1, repeats: true) { [weak self] _ in
            Task { @MainActor in
                self?.sessionTimerTicked()
            }
        }
    }

    private func stopSessionTimer() {
        sessionTimer?.invalidate()
        sessionTimer = nil
    }

    private func sessionTimerTicked() {
        // 配列の再代入で@Publishedの変更通知をトリガー
        timerSessions = timerSessions
    }

    private func saveTimerSession(_ session: TimerSession) async {
        let endedAt = Date()
        let durationSeconds = session.currentElapsedSeconds()
        let project = projects.first { $0.id == session.projectID }
        let selectedTags = tags.filter { session.tagIDs.contains($0.id) }
        let title = project?.name ?? "作業"
        let notes = session.notes.trimmingCharacters(in: .whitespacesAndNewlines)

        do {
            let localEntry = try entryStore.save(
                title: title,
                notes: notes,
                project: project,
                tags: selectedTags,
                startedAt: session.startedAt,
                endedAt: endedAt,
                durationSeconds: durationSeconds,
                remoteEntryID: nil,
                syncStatus: "pending"
            )
            loadRecentEntries()
            loadSelectedDateEntries()

            if skipsAuthentication {
                try entryStore.markSynced(localEntry, remoteEntryID: localEntry.id.uuidString)
                loadRecentEntries()
                loadSelectedDateEntries()
                errorMessage = nil
                return
            }

            do {
                let remoteEntry = try await entryClient.createEntry(
                    title: title,
                    notes: notes,
                    projectId: project?.id,
                    startedAt: session.startedAt,
                    endedAt: endedAt,
                    isBreak: session.isBreak,
                    tagIds: selectedTags.map(\.id)
                )
                try entryStore.markSynced(localEntry, remoteEntryID: remoteEntry.id)
                loadRecentEntries()
                loadSelectedDateEntries()
            } catch {
                try? entryStore.markSyncFailed(localEntry)
                loadRecentEntries()
                loadSelectedDateEntries()
                errorMessage = "作業記録をローカル保存しましたが、同期に失敗しました。"
            }
        } catch {
            errorMessage = "作業記録を保存できませんでした。"
        }
    }

    func loadRecentEntries() {
        do {
            recentEntries = try entryStore.fetchRecent(limit: 20)
            errorMessage = nil
        } catch {
            errorMessage = "作業記録を読み込めませんでした。"
        }
    }

    func loadSelectedDateEntries() {
        do {
            let interval = dayInterval(for: selectedEntryDate)
            selectedDateEntries = try entryStore.fetchEntries(from: interval.start, to: interval.end)
            errorMessage = nil
        } catch {
            errorMessage = "選択日の作業記録を読み込めませんでした。"
        }
    }

    func dismissError() {
        errorMessage = nil
    }

    private func authenticate(_ action: @escaping () async throws -> AuthUser) async {
        isAuthRequestInFlight = true
        defer { isAuthRequestInFlight = false }

        do {
            authState = .signedIn(try await action())
            await loadWorkspaceData()
            await syncEntries()
            errorMessage = nil
        } catch {
            errorMessage = "認証に失敗しました。メールアドレスとパスワードを確認してください。"
        }
    }

    private func loadWorkspaceData() async {
        guard !skipsAuthentication else {
            loadLocalWorkspaceData()
            return
        }

        isLoadingWorkspaceData = true
        defer { isLoadingWorkspaceData = false }

        do {
            async let projects = projectClient.listProjects()
            async let tags = tagClient.listTags()
            self.projects = try await projects
            self.tags = try await tags
            errorMessage = nil
        } catch {
            errorMessage = "プロジェクトとタグを読み込めませんでした。"
        }
    }

    private func loadLocalWorkspaceData() {
        guard let workspaceStore else { return }

        do {
            projects = try workspaceStore.fetchProjects()
            tags = try workspaceStore.fetchTags()
            errorMessage = nil
        } catch {
            errorMessage = "プロジェクトとタグを読み込めませんでした。"
        }
    }

    private func syncEntries(for date: Date? = nil) async {
        guard !skipsAuthentication else {
            loadRecentEntries()
            loadSelectedDateEntries()
            errorMessage = nil
            return
        }

        isSyncingEntries = true
        defer { isSyncingEntries = false }

        do {
            try await retryUnsyncedEntries()
            if let date {
                try await importRemoteEntries(for: date)
            } else {
                try await importRemoteEntriesForCurrentMonth()
            }
            loadRecentEntries()
            loadSelectedDateEntries()
            errorMessage = nil
        } catch {
            loadRecentEntries()
            loadSelectedDateEntries()
            errorMessage = "作業記録を同期できませんでした。"
        }
    }

    private func retryUnsyncedEntries() async throws {
        let unsyncedEntries = try entryStore.fetchUnsynced()
        for localEntry in unsyncedEntries {
            do {
                let remoteEntry = try await entryClient.createEntry(
                    title: localEntry.title,
                    notes: localEntry.notes,
                    projectId: localEntry.projectID,
                    startedAt: localEntry.startedAt,
                    endedAt: localEntry.endedAt,
                    isBreak: false,
                    tagIds: localEntry.tagIDList
                )
                try entryStore.markSynced(localEntry, remoteEntryID: remoteEntry.id)
            } catch {
                try? entryStore.markSyncFailed(localEntry)
            }
        }
    }

    private func importRemoteEntriesForCurrentMonth() async throws {
        let interval = currentMonthInterval()
        try await importRemoteEntries(from: interval.start, to: interval.end)
    }

    private func importRemoteEntries(for date: Date) async throws {
        let interval = dayInterval(for: date)
        try await importRemoteEntries(from: interval.start, to: interval.end)
    }

    private func importRemoteEntries(from start: Date, to end: Date) async throws {
        let remoteEntries = try await entryClient.listEntries(from: start, to: end)
        for remoteEntry in remoteEntries {
            let project = projects.first { $0.id == remoteEntry.projectId }
            try entryStore.upsertRemoteEntry(remoteEntry, project: project, tags: remoteEntry.tags)
        }
    }

    private func updateEntry(_ draft: EntryEditDraft) async {
        let project = projects.first { $0.id == draft.projectID }
        let selectedTags = tags.filter { draft.selectedTagIDs.contains($0.id) }
        let title = project?.name ?? draft.entry.title
        let notes = draft.notes.trimmingCharacters(in: .whitespacesAndNewlines)

        do {
            if skipsAuthentication {
                try entryStore.updateEntry(
                    draft.entry,
                    title: title,
                    notes: notes,
                    project: project,
                    tags: selectedTags,
                    syncStatus: "synced"
                )
                entryEditDraft = nil
                loadRecentEntries()
                loadSelectedDateEntries()
                errorMessage = nil
                return
            }

            if let remoteEntryID = draft.entry.remoteEntryID {
                do {
                    let remoteEntry = try await entryClient.updateEntry(
                        id: remoteEntryID,
                        title: title,
                        notes: notes,
                        projectId: project?.id,
                        tagIds: selectedTags.map(\.id)
                    )
                    try entryStore.upsertRemoteEntry(remoteEntry, project: project, tags: remoteEntry.tags.isEmpty ? selectedTags : remoteEntry.tags)
                } catch {
                    try entryStore.updateEntry(
                        draft.entry,
                        title: title,
                        notes: notes,
                        project: project,
                        tags: selectedTags,
                        syncStatus: "failed"
                    )
                    errorMessage = "作業記録をローカル更新しましたが、同期に失敗しました。"
                }
            } else {
                try entryStore.updateEntry(
                    draft.entry,
                    title: title,
                    notes: notes,
                    project: project,
                    tags: selectedTags,
                    syncStatus: "pending"
                )
            }

            entryEditDraft = nil
            loadRecentEntries()
            loadSelectedDateEntries()
        } catch {
            errorMessage = "作業記録を更新できませんでした。"
        }
    }

    private func saveProject(_ draft: ProjectEditDraft) async {
        let name = draft.name.trimmingCharacters(in: .whitespacesAndNewlines)
        let description = draft.description.trimmingCharacters(in: .whitespacesAndNewlines)
        let color = normalizedHexColor(draft.color, fallback: "#3B82F6")

        guard !name.isEmpty else {
            errorMessage = "プロジェクト名を入力してください。"
            return
        }

        do {
            if skipsAuthentication {
                if let project = draft.project {
                    let updated = Project(
                        id: project.id,
                        userId: project.userId,
                        name: name,
                        description: description,
                        color: color,
                        isArchived: draft.isArchived,
                        createdAt: project.createdAt,
                        updatedAt: Date()
                    )
                    try workspaceStore?.saveProject(updated)
                    if let index = projects.firstIndex(where: { $0.id == updated.id }) {
                        projects[index] = updated
                    }
                } else {
                    let created = Project(
                        id: UUID().uuidString,
                        userId: "local-development-user",
                        name: name,
                        description: description,
                        color: color,
                        isArchived: false,
                        createdAt: Date(),
                        updatedAt: Date()
                    )
                    try workspaceStore?.saveProject(created)
                    projects.insert(created, at: 0)
                }
                projectEditDraft = nil
                errorMessage = nil
                return
            }

            if let project = draft.project {
                let updated = try await projectClient.updateProject(
                    id: project.id,
                    name: name,
                    description: description,
                    color: color,
                    isArchived: draft.isArchived
                )
                if let index = projects.firstIndex(where: { $0.id == updated.id }) {
                    projects[index] = updated
                }
            } else {
                let created = try await projectClient.createProject(
                    name: name,
                    description: description,
                    color: color
                )
                projects.insert(created, at: 0)
            }
            projectEditDraft = nil
            errorMessage = nil
        } catch {
            errorMessage = "プロジェクトを保存できませんでした。"
        }
    }

    private func saveTag(_ draft: TagEditDraft) async {
        let name = draft.name.trimmingCharacters(in: .whitespacesAndNewlines)
        let color = normalizedHexColor(draft.color, fallback: "#F97316")

        guard !name.isEmpty else {
            errorMessage = "タグ名を入力してください。"
            return
        }

        do {
            if skipsAuthentication {
                if let tag = draft.tag {
                    let updated = Tag(
                        id: tag.id,
                        userId: tag.userId,
                        name: name,
                        color: color,
                        createdAt: tag.createdAt,
                        updatedAt: Date()
                    )
                    try workspaceStore?.saveTag(updated)
                    if let index = tags.firstIndex(where: { $0.id == updated.id }) {
                        tags[index] = updated
                    }
                } else {
                    let created = Tag(
                        id: UUID().uuidString,
                        userId: "local-development-user",
                        name: name,
                        color: color,
                        createdAt: Date(),
                        updatedAt: Date()
                    )
                    try workspaceStore?.saveTag(created)
                    tags.insert(created, at: 0)
                }
                tagEditDraft = nil
                errorMessage = nil
                return
            }

            if let tag = draft.tag {
                let updated = try await tagClient.updateTag(
                    id: tag.id,
                    name: name,
                    color: color
                )
                if let index = tags.firstIndex(where: { $0.id == updated.id }) {
                    tags[index] = updated
                }
            } else {
                let created = try await tagClient.createTag(name: name, color: color)
                tags.insert(created, at: 0)
            }
            tagEditDraft = nil
            errorMessage = nil
        } catch {
            errorMessage = "タグを保存できませんでした。"
        }
    }

    private func normalizedHexColor(_ value: String, fallback: String) -> String {
        var color = value.trimmingCharacters(in: .whitespacesAndNewlines)
        if !color.hasPrefix("#") {
            color = "#\(color)"
        }
        let hexDigits = CharacterSet(charactersIn: "0123456789ABCDEFabcdef")
        let digits = String(color.dropFirst())
        guard color.count == 7,
              color.first == "#",
              digits.unicodeScalars.allSatisfy({ hexDigits.contains($0) })
        else {
            return fallback
        }
        return color.uppercased()
    }

    private func deleteEntryFromStoreAndRemote(_ entry: TimeEntryRecord) async {
        do {
            if !skipsAuthentication, let remoteEntryID = entry.remoteEntryID {
                try await entryClient.deleteEntry(id: remoteEntryID)
            }
            try entryStore.deleteEntry(entry)
            loadRecentEntries()
            loadSelectedDateEntries()
        } catch {
            errorMessage = "作業記録を削除できませんでした。"
        }
    }

    private func currentMonthInterval() -> (start: Date, end: Date) {
        let calendar = Calendar.current
        let now = Date()
        let start = calendar.date(from: calendar.dateComponents([.year, .month], from: now)) ?? now
        let end = calendar.date(byAdding: DateComponents(month: 1), to: start) ?? now
        return (start, end)
    }

    private func dayInterval(for date: Date) -> (start: Date, end: Date) {
        let calendar = Calendar.current
        let start = calendar.startOfDay(for: date)
        let end = calendar.date(byAdding: DateComponents(day: 1), to: start) ?? date
        return (start, end)
    }

    deinit {
        sessionTimer?.invalidate()
    }

    private static func localDevelopmentUser() -> AuthUser {
        AuthUser(
            id: "local-development-user",
            email: "local@chronome.dev",
            displayName: "ローカル確認",
            timeZone: TimeZone.current.identifier,
            createdAt: nil,
            updatedAt: nil
        )
    }
}

private extension Array {
    subscript(safe index: Int) -> Element? {
        indices.contains(index) ? self[index] : nil
    }
}
