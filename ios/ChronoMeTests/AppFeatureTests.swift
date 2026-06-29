import SwiftData
import XCTest
@testable import ChronoMe

@MainActor
final class AppFeatureTests: XCTestCase {
    func testTimerStartsTicksAndStops() async throws {
        let configuration = ModelConfiguration(isStoredInMemoryOnly: true)
        let container = try ModelContainer(
            for: TimeEntryRecord.self,
            configurations: configuration
        )
        let store = SwiftDataTimeEntryStore(modelContext: container.mainContext)
        let feature = AppFeature(
            entryStore: store,
            authClient: MockAuthClient(),
            projectClient: MockProjectClient(),
            tagClient: MockTagClient(),
            entryClient: MockEntryClient()
        )

        feature.timerButtonTapped()
        XCTAssertTrue(feature.isTimerRunning)

        feature.timerTicked()
        XCTAssertEqual(feature.elapsedSeconds, 1)

        feature.timerButtonTapped()
        try await waitForRecentEntries(in: feature)
        XCTAssertFalse(feature.isTimerRunning)
        XCTAssertEqual(feature.recentEntries.count, 1)
        XCTAssertEqual(feature.recentEntries.first?.durationSeconds, 1)
        XCTAssertEqual(feature.recentEntries.first?.syncStatus, "synced")
    }

    func testRecentEntriesAreSortedNewestFirst() throws {
        let configuration = ModelConfiguration(isStoredInMemoryOnly: true)
        let container = try ModelContainer(
            for: TimeEntryRecord.self,
            configurations: configuration
        )
        let store = SwiftDataTimeEntryStore(modelContext: container.mainContext)
        let now = Date()

        try store.save(title: "Older", notes: "", project: nil, tags: [], startedAt: now.addingTimeInterval(-120), endedAt: now, durationSeconds: 120, remoteEntryID: nil, syncStatus: "pending")
        try store.save(title: "Newer", notes: "", project: nil, tags: [], startedAt: now, endedAt: now.addingTimeInterval(60), durationSeconds: 60, remoteEntryID: nil, syncStatus: "pending")

        let entries = try store.fetchRecent(limit: 20)
        XCTAssertEqual(entries.map(\.durationSeconds), [60, 120])
    }

    func testUnsyncedEntriesExcludeSyncedEntries() throws {
        let configuration = ModelConfiguration(isStoredInMemoryOnly: true)
        let container = try ModelContainer(
            for: TimeEntryRecord.self,
            configurations: configuration
        )
        let store = SwiftDataTimeEntryStore(modelContext: container.mainContext)
        let now = Date()

        try store.save(title: "Pending", notes: "", project: nil, tags: [], startedAt: now, endedAt: now.addingTimeInterval(60), durationSeconds: 60, remoteEntryID: nil, syncStatus: "pending")
        try store.save(title: "Synced", notes: "", project: nil, tags: [], startedAt: now, endedAt: now.addingTimeInterval(120), durationSeconds: 120, remoteEntryID: "remote-1", syncStatus: "synced")

        let unsynced = try store.fetchUnsynced()

        XCTAssertEqual(unsynced.map(\.title), ["Pending"])
    }

    func testEntriesCanBeFetchedByDateRange() throws {
        let configuration = ModelConfiguration(isStoredInMemoryOnly: true)
        let container = try ModelContainer(
            for: TimeEntryRecord.self,
            configurations: configuration
        )
        let store = SwiftDataTimeEntryStore(modelContext: container.mainContext)
        let calendar = Calendar(identifier: .gregorian)
        let targetDay = calendar.date(from: DateComponents(year: 2026, month: 6, day: 26))!
        let nextDay = calendar.date(byAdding: .day, value: 1, to: targetDay)!

        try store.save(title: "Target", notes: "", project: nil, tags: [], startedAt: targetDay.addingTimeInterval(3_600), endedAt: targetDay.addingTimeInterval(7_200), durationSeconds: 3_600, remoteEntryID: nil, syncStatus: "pending")
        try store.save(title: "Other", notes: "", project: nil, tags: [], startedAt: nextDay.addingTimeInterval(3_600), endedAt: nextDay.addingTimeInterval(7_200), durationSeconds: 3_600, remoteEntryID: nil, syncStatus: "pending")

        let entries = try store.fetchEntries(from: targetDay, to: nextDay)

        XCTAssertEqual(entries.map(\.title), ["Target"])
    }

    func testUpsertRemoteEntryDoesNotDuplicateExistingRemoteEntry() throws {
        let configuration = ModelConfiguration(isStoredInMemoryOnly: true)
        let container = try ModelContainer(
            for: TimeEntryRecord.self,
            configurations: configuration
        )
        let store = SwiftDataTimeEntryStore(modelContext: container.mainContext)
        let now = Date()
        let remote = Entry(
            id: "remote-1",
            userId: "user-1",
            projectId: nil,
            title: "Remote",
            notes: "",
            startedAt: now,
            endedAt: now.addingTimeInterval(60),
            durationSec: 60,
            ratio: 1,
            isBreak: false,
            tags: [],
            createdAt: nil,
            updatedAt: nil
        )

        try store.upsertRemoteEntry(remote, project: nil, tags: [])
        try store.upsertRemoteEntry(remote, project: nil, tags: [])

        XCTAssertEqual(try store.fetchRecent(limit: 20).count, 1)
    }

    func testLocalEntryCanBeUpdatedAndDeleted() throws {
        let configuration = ModelConfiguration(isStoredInMemoryOnly: true)
        let container = try ModelContainer(
            for: TimeEntryRecord.self,
            configurations: configuration
        )
        let store = SwiftDataTimeEntryStore(modelContext: container.mainContext)
        let now = Date()
        let project = Project(id: "project-1", userId: "user-1", name: "Client A", description: nil, color: "#3B82F6", isArchived: false, createdAt: nil, updatedAt: nil)
        let tag = Tag(id: "tag-1", userId: "user-1", name: "Deep Work", color: "#F97316", createdAt: nil, updatedAt: nil)
        let entry = try store.save(title: "作業", notes: "", project: nil, tags: [], startedAt: now, endedAt: now.addingTimeInterval(60), durationSeconds: 60, remoteEntryID: nil, syncStatus: "pending")

        try store.updateEntry(entry, title: "Client A", notes: "updated", project: project, tags: [tag], syncStatus: "pending")
        let updated = try XCTUnwrap(store.fetchRecent(limit: 20).first)

        XCTAssertEqual(updated.title, "Client A")
        XCTAssertEqual(updated.notes, "updated")
        XCTAssertEqual(updated.projectID, "project-1")
        XCTAssertEqual(updated.tagIDList, ["tag-1"])

        try store.deleteEntry(updated)
        XCTAssertTrue(try store.fetchRecent(limit: 20).isEmpty)
    }

    func testSelectedDateSummaryAggregatesProjectsAndTags() throws {
        let project = Project(id: "project-1", userId: "user-1", name: "Client A", description: nil, color: "#3B82F6", isArchived: false, createdAt: nil, updatedAt: nil)
        let tag = Tag(id: "tag-1", userId: "user-1", name: "Deep Work", color: "#F97316", createdAt: nil, updatedAt: nil)
        let now = Date()
        let entries = [
            TimeEntryRecord(
                remoteEntryID: nil,
                title: "Client A",
                notes: "",
                projectID: project.id,
                projectName: project.name,
                tagIDs: [tag.id],
                tagNames: [tag.name],
                startedAt: now,
                endedAt: now.addingTimeInterval(3_600),
                durationSeconds: 3_600,
                syncStatus: "synced"
            ),
            TimeEntryRecord(
                remoteEntryID: nil,
                title: "作業",
                notes: "",
                projectID: nil,
                projectName: nil,
                tagIDs: [],
                tagNames: [],
                startedAt: now,
                endedAt: now.addingTimeInterval(1_800),
                durationSeconds: 1_800,
                syncStatus: "synced"
            )
        ]
        let feature = AppFeature(
            entryStore: MockTimeEntryStore(entries: entries),
            authClient: MockAuthClient(),
            projectClient: MockProjectClient(projects: [project]),
            tagClient: MockTagClient(tags: [tag]),
            entryClient: MockEntryClient()
        )

        feature.loadSelectedDateEntries()

        XCTAssertEqual(feature.selectedDateTotalSeconds, 5_400)
        XCTAssertEqual(feature.selectedDateEntryCount, 2)
        XCTAssertEqual(feature.selectedDateProjectSummaries.map(\.name), ["Client A", "未分類"])
        XCTAssertEqual(feature.selectedDateTagSummaries.map(\.name), ["Deep Work", "タグなし"])
    }

    func testRestoreSessionSignsInWhenUserExists() async throws {
        let feature = AppFeature(
            entryStore: MockTimeEntryStore(),
            authClient: MockAuthClient(currentUserResult: AuthUser(
                id: "user-1",
                email: "miyu@example.com",
                displayName: "Miyu",
                timeZone: "Asia/Tokyo",
                createdAt: nil,
                updatedAt: nil
            )),
            projectClient: MockProjectClient(projects: [
                Project(id: "project-1", userId: "user-1", name: "Client A", description: nil, color: "#3B82F6", isArchived: false, createdAt: nil, updatedAt: nil)
            ]),
            tagClient: MockTagClient(tags: [
                Tag(id: "tag-1", userId: "user-1", name: "Deep Work", color: "#F97316", createdAt: nil, updatedAt: nil)
            ]),
            entryClient: MockEntryClient()
        )

        feature.restoreSession()
        try await waitForAuthRequestToFinish(in: feature)

        XCTAssertEqual(
            feature.authState,
            .signedIn(AuthUser(
                id: "user-1",
                email: "miyu@example.com",
                displayName: "Miyu",
                timeZone: "Asia/Tokyo",
                createdAt: nil,
                updatedAt: nil
            ))
        )
        XCTAssertEqual(feature.projects.map(\.name), ["Client A"])
        XCTAssertEqual(feature.tags.map(\.name), ["Deep Work"])
    }

    func testLoginSignsIn() async throws {
        let feature = AppFeature(
            entryStore: MockTimeEntryStore(),
            authClient: MockAuthClient(loginResult: AuthUser(
                id: "user-1",
                email: "miyu@example.com",
                displayName: nil,
                timeZone: "Asia/Tokyo",
                createdAt: nil,
                updatedAt: nil
            )),
            projectClient: MockProjectClient(projects: [
                Project(id: "project-1", userId: "user-1", name: "Client A", description: nil, color: "#3B82F6", isArchived: false, createdAt: nil, updatedAt: nil)
            ]),
            tagClient: MockTagClient(tags: [
                Tag(id: "tag-1", userId: "user-1", name: "Deep Work", color: "#F97316", createdAt: nil, updatedAt: nil)
            ]),
            entryClient: MockEntryClient()
        )

        let expectedUser = AuthUser(
            id: "user-1",
            email: "miyu@example.com",
            displayName: nil,
            timeZone: "Asia/Tokyo",
            createdAt: nil,
            updatedAt: nil
        )

        feature.loginButtonTapped(email: "miyu@example.com", password: "Password1")
        try await waitForAuthState(.signedIn(expectedUser), in: feature)

        XCTAssertEqual(feature.authState, .signedIn(expectedUser))
        XCTAssertEqual(feature.projects.count, 1)
        XCTAssertEqual(feature.tags.count, 1)
    }

    func testLogoutClearsWorkspaceData() async throws {
        let feature = AppFeature(
            entryStore: MockTimeEntryStore(),
            authClient: MockAuthClient(),
            projectClient: MockProjectClient(projects: [
                Project(id: "project-1", userId: "user-1", name: "Client A", description: nil, color: "#3B82F6", isArchived: false, createdAt: nil, updatedAt: nil)
            ]),
            tagClient: MockTagClient(tags: [
                Tag(id: "tag-1", userId: "user-1", name: "Deep Work", color: "#F97316", createdAt: nil, updatedAt: nil)
            ]),
            entryClient: MockEntryClient()
        )

        feature.loginButtonTapped(email: "miyu@example.com", password: "Password1")
        try await waitForAuthRequestToFinish(in: feature)
        XCTAssertEqual(feature.projects.count, 1)

        feature.logoutButtonTapped()
        try await waitForAuthState(.signedOut, in: feature)

        XCTAssertEqual(feature.authState, .signedOut)
        XCTAssertTrue(feature.projects.isEmpty)
        XCTAssertTrue(feature.tags.isEmpty)
    }

    private func waitForAuthRequestToFinish(
        in feature: AppFeature,
        timeout: TimeInterval = 1,
        file: StaticString = #filePath,
        line: UInt = #line
    ) async throws {
        let deadline = Date().addingTimeInterval(timeout)
        while (feature.isAuthRequestInFlight || feature.authState == .checking), Date() < deadline {
            await Task.yield()
        }

        XCTAssertFalse(feature.isAuthRequestInFlight, file: file, line: line)
        XCTAssertNotEqual(feature.authState, .checking, file: file, line: line)
    }

    private func waitForAuthState(
        _ expectedState: AppFeature.AuthState,
        in feature: AppFeature,
        timeout: TimeInterval = 1,
        file: StaticString = #filePath,
        line: UInt = #line
    ) async throws {
        let deadline = Date().addingTimeInterval(timeout)
        while (feature.isAuthRequestInFlight || feature.authState != expectedState), Date() < deadline {
            await Task.yield()
        }

        XCTAssertFalse(feature.isAuthRequestInFlight, file: file, line: line)
        XCTAssertEqual(feature.authState, expectedState, file: file, line: line)
    }

    private func waitForRecentEntries(
        in feature: AppFeature,
        timeout: TimeInterval = 1,
        file: StaticString = #filePath,
        line: UInt = #line
    ) async throws {
        let deadline = Date().addingTimeInterval(timeout)
        while feature.recentEntries.isEmpty && Date() < deadline {
            await Task.yield()
        }

        XCTAssertFalse(feature.recentEntries.isEmpty, file: file, line: line)
    }
}

private final class MockTimeEntryStore: TimeEntryStoring {
    let entries: [TimeEntryRecord]

    init(entries: [TimeEntryRecord] = []) {
        self.entries = entries
    }

    func fetchRecent(limit: Int) throws -> [TimeEntryRecord] { Array(entries.prefix(limit)) }
    func fetchEntries(from: Date, to: Date) throws -> [TimeEntryRecord] { entries }
    func fetchUnsynced() throws -> [TimeEntryRecord] { [] }

    func save(
        title: String,
        notes: String,
        project: Project?,
        tags: [Tag],
        startedAt: Date,
        endedAt: Date,
        durationSeconds: Int,
        remoteEntryID: String?,
        syncStatus: String
    ) throws -> TimeEntryRecord {
        TimeEntryRecord(
            remoteEntryID: remoteEntryID,
            title: title,
            notes: notes,
            projectID: project?.id,
            projectName: project?.name,
            tagIDs: tags.map(\.id),
            tagNames: tags.map(\.name),
            startedAt: startedAt,
            endedAt: endedAt,
            durationSeconds: durationSeconds,
            syncStatus: syncStatus
        )
    }

    func markSynced(_ entry: TimeEntryRecord, remoteEntryID: String) throws {
        entry.remoteEntryID = remoteEntryID
        entry.syncStatus = "synced"
    }

    func upsertRemoteEntry(_ entry: Entry, project: Project?, tags: [Tag]) throws -> TimeEntryRecord {
        TimeEntryRecord(
            remoteEntryID: entry.id,
            title: entry.title,
            notes: entry.notes ?? "",
            projectID: entry.projectId,
            projectName: project?.name,
            tagIDs: tags.map(\.id),
            tagNames: tags.map(\.name),
            startedAt: entry.startedAt,
            endedAt: entry.endedAt ?? entry.startedAt,
            durationSeconds: entry.durationSec,
            syncStatus: "synced"
        )
    }

    func markSyncFailed(_ entry: TimeEntryRecord) throws {
        entry.syncStatus = "failed"
    }

    func updateEntry(_ entry: TimeEntryRecord, title: String, notes: String, project: Project?, tags: [Tag], syncStatus: String) throws {
        entry.title = title
        entry.notes = notes
        entry.projectID = project?.id
        entry.projectName = project?.name
        entry.tagIDs = tags.map(\.id).joined(separator: ",")
        entry.tagNames = tags.map(\.name).joined(separator: ",")
        entry.syncStatus = syncStatus
    }

    func deleteEntry(_ entry: TimeEntryRecord) throws {}
}

private final class MockAuthClient: AuthClientProtocol {
    var currentUserResult: AuthUser?
    var loginResult: AuthUser

    init(
        currentUserResult: AuthUser? = nil,
        loginResult: AuthUser = AuthUser(
            id: "default",
            email: "default@example.com",
            displayName: nil,
            timeZone: "Asia/Tokyo",
            createdAt: nil,
            updatedAt: nil
        )
    ) {
        self.currentUserResult = currentUserResult
        self.loginResult = loginResult
    }

    func login(email: String, password: String) async throws -> AuthUser {
        loginResult
    }

    func signup(email: String, password: String, displayName: String?, timeZone: String?) async throws -> AuthUser {
        AuthUser(
            id: "signup",
            email: email,
            displayName: displayName,
            timeZone: timeZone,
            createdAt: nil,
            updatedAt: nil
        )
    }

    func currentUser() async throws -> AuthUser? {
        currentUserResult
    }

    func logout() async throws {}
}

private final class MockProjectClient: ProjectClientProtocol {
    var projects: [Project]

    init(projects: [Project] = []) {
        self.projects = projects
    }

    func listProjects() async throws -> [Project] {
        projects
    }

    func createProject(name: String, description: String, color: String) async throws -> Project {
        let project = Project(id: "project-new", userId: "user-1", name: name, description: description, color: color, isArchived: false, createdAt: nil, updatedAt: nil)
        projects.insert(project, at: 0)
        return project
    }

    func updateProject(id: String, name: String, description: String, color: String, isArchived: Bool) async throws -> Project {
        let project = Project(id: id, userId: "user-1", name: name, description: description, color: color, isArchived: isArchived, createdAt: nil, updatedAt: nil)
        if let index = projects.firstIndex(where: { $0.id == id }) {
            projects[index] = project
        }
        return project
    }
}

private final class MockTagClient: TagClientProtocol {
    var tags: [Tag]

    init(tags: [Tag] = []) {
        self.tags = tags
    }

    func listTags() async throws -> [Tag] {
        tags
    }

    func createTag(name: String, color: String) async throws -> Tag {
        let tag = Tag(id: "tag-new", userId: "user-1", name: name, color: color, createdAt: nil, updatedAt: nil)
        tags.insert(tag, at: 0)
        return tag
    }

    func updateTag(id: String, name: String, color: String) async throws -> Tag {
        let tag = Tag(id: id, userId: "user-1", name: name, color: color, createdAt: nil, updatedAt: nil)
        if let index = tags.firstIndex(where: { $0.id == id }) {
            tags[index] = tag
        }
        return tag
    }
}

private final class MockEntryClient: EntryClientProtocol {
    var listEntriesResult: [Entry]

    init(listEntriesResult: [Entry] = []) {
        self.listEntriesResult = listEntriesResult
    }

    func listEntries(from: Date?, to: Date?) async throws -> [Entry] {
        listEntriesResult
    }

    func createEntry(
        title: String,
        notes: String,
        projectId: String?,
        startedAt: Date,
        endedAt: Date,
        isBreak: Bool,
        tagIds: [String]
    ) async throws -> Entry {
        Entry(
            id: "remote-entry-1",
            userId: "user-1",
            projectId: projectId,
            title: title,
            notes: notes,
            startedAt: startedAt,
            endedAt: endedAt,
            durationSec: Int(endedAt.timeIntervalSince(startedAt)),
            ratio: 1,
            isBreak: isBreak,
            tags: [],
            createdAt: nil,
            updatedAt: nil
        )
    }

    func updateEntry(id: String, title: String, notes: String, projectId: String?, tagIds: [String]) async throws -> Entry {
        Entry(
            id: id,
            userId: "user-1",
            projectId: projectId,
            title: title,
            notes: notes,
            startedAt: Date(),
            endedAt: Date(),
            durationSec: 0,
            ratio: 1,
            isBreak: false,
            tags: [],
            createdAt: nil,
            updatedAt: nil
        )
    }

    func deleteEntry(id: String) async throws {}
}
