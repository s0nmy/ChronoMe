import SwiftUI

struct LoginView: View {
    @ObservedObject var feature: AppFeature

    @State private var email = ""
    @State private var password = ""
    @State private var displayName = ""
    @State private var isSignupMode = false

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("メールアドレス", text: $email)
                        .textContentType(.emailAddress)
                        .keyboardType(.emailAddress)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled()

                    SecureField("パスワード", text: $password)
                        .textContentType(isSignupMode ? .newPassword : .password)

                    if isSignupMode {
                        TextField("表示名（任意）", text: $displayName)
                            .textContentType(.name)
                    }
                }

                Section {
                    Button(isSignupMode ? "アカウント作成" : "ログイン") {
                        submit()
                    }
                    .disabled(!canSubmit || feature.isAuthRequestInFlight)

                    Button(isSignupMode ? "ログインに戻る" : "アカウントを作成") {
                        isSignupMode.toggle()
                    }
                    .disabled(feature.isAuthRequestInFlight)
                }

                if feature.isAuthRequestInFlight {
                    Section {
                        ProgressView()
                    }
                }
            }
            .navigationTitle(isSignupMode ? "アカウント作成" : "ログイン")
            .alert(
                "エラー",
                isPresented: Binding(
                    get: { feature.errorMessage != nil },
                    set: { isPresented in
                        if !isPresented {
                            feature.dismissError()
                        }
                    }
                )
            ) {
                Button("OK") {}
            } message: {
                Text(feature.errorMessage ?? "")
            }
        }
        .tint(Color("AccentColor"))
    }

    private var canSubmit: Bool {
        !email.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty &&
            !password.isEmpty
    }

    private func submit() {
        if isSignupMode {
            feature.signupButtonTapped(
                email: email,
                password: password,
                displayName: displayName.isEmpty ? nil : displayName
            )
        } else {
            feature.loginButtonTapped(email: email, password: password)
        }
    }
}

#Preview {
    LoginView(
        feature: AppFeature(
            entryStore: PreviewTimeEntryStore(),
            authClient: PreviewAuthClient(),
            projectClient: PreviewAuthProjectClient(),
            tagClient: PreviewAuthTagClient(),
            entryClient: PreviewAuthEntryClient()
        )
    )
}

private final class PreviewTimeEntryStore: TimeEntryStoring {
    func fetchRecent(limit: Int) throws -> [TimeEntryRecord] { [] }
    func fetchEntries(from: Date, to: Date) throws -> [TimeEntryRecord] { [] }
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

private final class PreviewAuthClient: AuthClientProtocol {
    func login(email: String, password: String) async throws -> AuthUser {
        AuthUser(id: "preview", email: email, displayName: "Preview", timeZone: "Asia/Tokyo", createdAt: nil, updatedAt: nil)
    }

    func signup(email: String, password: String, displayName: String?, timeZone: String?) async throws -> AuthUser {
        AuthUser(id: "preview", email: email, displayName: displayName, timeZone: timeZone, createdAt: nil, updatedAt: nil)
    }

    func currentUser() async throws -> AuthUser? { nil }
    func logout() async throws {}
}

private final class PreviewAuthProjectClient: ProjectClientProtocol {
    func listProjects() async throws -> [Project] { [] }
    func createProject(name: String, description: String, color: String) async throws -> Project {
        Project(id: "preview", userId: "preview", name: name, description: description, color: color, isArchived: false, createdAt: nil, updatedAt: nil)
    }
    func updateProject(id: String, name: String, description: String, color: String, isArchived: Bool) async throws -> Project {
        Project(id: id, userId: "preview", name: name, description: description, color: color, isArchived: isArchived, createdAt: nil, updatedAt: nil)
    }
}

private final class PreviewAuthTagClient: TagClientProtocol {
    func listTags() async throws -> [Tag] { [] }
    func createTag(name: String, color: String) async throws -> Tag {
        Tag(id: "preview", userId: "preview", name: name, color: color, createdAt: nil, updatedAt: nil)
    }
    func updateTag(id: String, name: String, color: String) async throws -> Tag {
        Tag(id: id, userId: "preview", name: name, color: color, createdAt: nil, updatedAt: nil)
    }
}

private final class PreviewAuthEntryClient: EntryClientProtocol {
    func listEntries(from: Date?, to: Date?) async throws -> [Entry] { [] }

    func createEntry(
        title: String,
        notes: String,
        projectId: String?,
        startedAt: Date,
        endedAt: Date,
        isBreak: Bool,
        tagIds: [String]
    ) async throws -> Entry {
        Entry(id: "preview", userId: "preview", projectId: projectId, title: title, notes: notes, startedAt: startedAt, endedAt: endedAt, durationSec: 0, ratio: 1, isBreak: isBreak, tags: [], createdAt: nil, updatedAt: nil)
    }

    func updateEntry(id: String, title: String, notes: String, projectId: String?, tagIds: [String]) async throws -> Entry {
        Entry(id: id, userId: "preview", projectId: projectId, title: title, notes: notes, startedAt: Date(), endedAt: Date(), durationSec: 0, ratio: 1, isBreak: false, tags: [], createdAt: nil, updatedAt: nil)
    }

    func deleteEntry(id: String) async throws {}
}
