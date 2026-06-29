import SwiftData
import SwiftUI

struct ContentView: View {
    @ObservedObject var feature: AppFeature

    var body: some View {
        Group {
            switch feature.authState {
            case .checking:
                ProgressView("ログイン状態を確認中")
                    .tint(Color("AccentColor"))
            case .signedOut:
                LoginView(feature: feature)
            case let .signedIn(user):
                TimerHomeView(feature: feature, user: user)
            }
        }
        .task {
            if feature.authState == .checking {
                feature.restoreSession()
            }
        }
    }
}

private struct TimerHomeView: View {
    @ObservedObject var feature: AppFeature
    let user: AuthUser
    @State private var selectedTab: MainTab = .timer

    private enum MainTab: Hashable {
        case timer
        case entries
        case reports
        case management
    }

    var body: some View {
        TabView(selection: $selectedTab) {
            NavigationStack {
                timerList
                    .navigationTitle("時間記録")
                    .toolbar { workspaceToolbar }
            }
            .tabItem {
                Label("時間記録", systemImage: "clock")
            }
            .tag(MainTab.timer)

            NavigationStack {
                entriesList
                    .navigationTitle("作業履歴")
                    .toolbar { workspaceToolbar }
            }
            .tabItem {
                Label("作業履歴", systemImage: "doc.text")
            }
            .tag(MainTab.entries)

            NavigationStack {
                reportsList
                    .navigationTitle("集計")
                    .toolbar { workspaceToolbar }
            }
            .tabItem {
                Label("集計", systemImage: "chart.bar")
            }
            .tag(MainTab.reports)

            NavigationStack {
                managementList
                    .navigationTitle("プロジェクト")
                    .toolbar { workspaceToolbar }
            }
            .tabItem {
                Label("プロジェクト", systemImage: "folder")
            }
            .tag(MainTab.management)
        }
        .tint(Color("AccentColor"))
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
        .sheet(item: $feature.entryEditDraft) { _ in
            EntryEditView(feature: feature)
        }
        .sheet(item: $feature.projectEditDraft) { _ in
            ProjectEditView(feature: feature)
        }
        .sheet(item: $feature.tagEditDraft) { _ in
            TagEditView(feature: feature)
        }
    }

    private var timerList: some View {
        List {
            userSection
            workspaceStatusSection
            recordingDetailsSection
            timerControlSection
        }
    }

    private var entriesList: some View {
        List {
            displayDateSection
            entriesSection
        }
        .overlay {
            if feature.selectedDateEntries.isEmpty {
                ContentUnavailableView(
                    "選択日の作業記録はありません",
                    systemImage: "clock",
                    description: Text("時間記録タブで作業を終了すると記録が保存されます。")
                )
                .allowsHitTesting(false)
            }
        }
    }

    private var reportsList: some View {
        List {
            displayDateSection
            dailySummarySection
        }
    }

    private var managementList: some View {
        List {
            userSection
            workspaceStatusSection
            projectsSection
            tagsSection
        }
    }

    private var userSection: some View {
        Section {
            VStack(alignment: .leading, spacing: 6) {
                Text(user.displayName ?? user.email)
                    .font(.headline)
                Text(user.email)
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
    }

    private var workspaceStatusSection: some View {
        Section {
            if feature.isLoadingWorkspaceData {
                HStack {
                    ProgressView()
                    Text("プロジェクトとタグを読み込み中")
                        .foregroundStyle(.secondary)
                }
            } else if feature.isSyncingEntries {
                HStack {
                    ProgressView()
                    Text("作業記録を同期中")
                        .foregroundStyle(.secondary)
                }
            } else {
                HStack {
                    Label("\(feature.projects.count)", systemImage: "folder")
                    Text("プロジェクト")
                    Spacer()
                    Label("\(feature.tags.count)", systemImage: "tag")
                    Text("タグ")
                }
                .foregroundStyle(.secondary)
            }
        }
    }

    private var displayDateSection: some View {
        Section("表示日") {
            DatePicker(
                "日付",
                selection: Binding(
                    get: { feature.selectedEntryDate },
                    set: { feature.selectedEntryDateChanged($0) }
                ),
                displayedComponents: .date
            )

            HStack {
                Text("選択日の合計")
                    .foregroundStyle(.secondary)
                Spacer()
                Text(durationText(feature.selectedDateTotalSeconds))
                    .monospacedDigit()
            }
        }
    }

    private var dailySummarySection: some View {
        Section("日次サマリー") {
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    Text("合計作業時間")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                    Text(durationText(feature.selectedDateTotalSeconds))
                        .font(.title3.weight(.semibold))
                        .monospacedDigit()
                }
                Spacer()
                VStack(alignment: .trailing, spacing: 4) {
                    Text("記録数")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                    Text("\(feature.selectedDateEntryCount)件")
                        .font(.title3.weight(.semibold))
                }
            }

            if feature.selectedDateProjectSummaries.isEmpty {
                Text("集計できるプロジェクトはありません")
                    .foregroundStyle(.secondary)
            } else {
                VStack(alignment: .leading, spacing: 10) {
                    Text("プロジェクト別")
                        .font(.subheadline.weight(.semibold))
                    ForEach(feature.selectedDateProjectSummaries) { item in
                        SummaryBarView(item: item, durationText: durationText(item.durationSeconds))
                    }
                }
                .padding(.vertical, 4)
            }

            if !feature.selectedDateTagSummaries.isEmpty {
                VStack(alignment: .leading, spacing: 10) {
                    Text("タグ別")
                        .font(.subheadline.weight(.semibold))
                    ForEach(feature.selectedDateTagSummaries) { item in
                        SummaryBarView(item: item, durationText: durationText(item.durationSeconds))
                    }
                }
                .padding(.vertical, 4)
            }
        }
    }

    private var recordingDetailsSection: some View {
        Section("記録内容") {
            Picker(
                "プロジェクト",
                selection: Binding(
                    get: { feature.selectedProjectID },
                    set: { feature.projectSelectionChanged($0) }
                )
            ) {
                Text("未選択").tag(String?.none)
                ForEach(feature.projects.filter { !$0.isArchived }) { project in
                    Text(project.name).tag(String?.some(project.id))
                }
            }

            if feature.tags.isEmpty {
                Text("選択できるタグはありません")
                    .foregroundStyle(.secondary)
            } else {
                ForEach(feature.tags) { tag in
                    Button {
                        feature.tagSelectionToggled(tag.id)
                    } label: {
                        HStack {
                            Circle()
                                .fill(Color(hex: tag.color))
                                .frame(width: 10, height: 10)
                            Text(tag.name)
                            Spacer()
                            if feature.selectedTagIDs.contains(tag.id) {
                                Image(systemName: "checkmark")
                                    .foregroundStyle(.tint)
                            }
                        }
                    }
                    .buttonStyle(.plain)
                }
            }

            TextField(
                "メモ",
                text: Binding(
                    get: { feature.draftNotes },
                    set: { feature.draftNotesChanged($0) }
                ),
                axis: .vertical
            )
            .lineLimit(2...4)
        }
    }

    private var timerControlSection: some View {
        Section {
            VStack(spacing: 28) {
                Image(systemName: feature.isTimerRunning ? "clock.badge.checkmark.fill" : "clock.fill")
                    .font(.system(size: 72))
                    .foregroundStyle(.tint)
                    .accessibilityHidden(true)

                VStack(spacing: 8) {
                    Text("今日の作業時間")
                        .font(.headline)
                        .foregroundStyle(.secondary)
                    Text(durationText(feature.elapsedSeconds))
                        .font(.system(.largeTitle, design: .rounded, weight: .semibold))
                        .monospacedDigit()
                        .accessibilityLabel("今日の作業時間、\(durationAccessibilityText(feature.elapsedSeconds))")
                }

                Button(feature.isTimerRunning ? "作業を終了" : "作業を開始") {
                    feature.timerButtonTapped()
                }
                .buttonStyle(.borderedProminent)
                .controlSize(.large)
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, 20)
        }
        .listRowBackground(Color.clear)
    }

    private var entriesSection: some View {
        Section("選択日の作業記録") {
            if feature.selectedDateEntries.isEmpty {
                Text("作業記録はありません")
                    .foregroundStyle(.secondary)
            } else {
                ForEach(feature.selectedDateEntries) { entry in
                    Button {
                        feature.entryTapped(entry)
                    } label: {
                        entryRow(entry)
                    }
                    .buttonStyle(.plain)
                }
                .onDelete { offsets in
                    for index in offsets {
                        feature.deleteEntry(feature.selectedDateEntries[index])
                    }
                }
            }
        }
    }

    private var projectsSection: some View {
        Section {
            if feature.projects.isEmpty {
                Text("プロジェクトはありません")
                    .foregroundStyle(.secondary)
            } else {
                ForEach(feature.projects) { project in
                    Button {
                        feature.projectTapped(project)
                    } label: {
                        HStack(spacing: 12) {
                            Circle()
                                .fill(Color(hex: project.color))
                                .frame(width: 12, height: 12)
                            VStack(alignment: .leading, spacing: 2) {
                                Text(project.name)
                                if project.isArchived {
                                    Text("アーカイブ済み")
                                        .font(.caption)
                                        .foregroundStyle(.secondary)
                                }
                            }
                        }
                    }
                    .buttonStyle(.plain)
                }
            }
        } header: {
            HStack {
                Text("プロジェクト")
                Spacer()
                Button("追加") {
                    feature.addProjectButtonTapped()
                }
                .font(.caption)
            }
        }
    }

    private var tagsSection: some View {
        Section {
            if feature.tags.isEmpty {
                Text("タグはありません")
                    .foregroundStyle(.secondary)
            } else {
                ForEach(feature.tags) { tag in
                    Button {
                        feature.tagTapped(tag)
                    } label: {
                        HStack(spacing: 10) {
                            Circle()
                                .fill(Color(hex: tag.color))
                                .frame(width: 10, height: 10)
                            Text(tag.name)
                        }
                    }
                    .buttonStyle(.plain)
                }
            }
        } header: {
            HStack {
                Text("タグ")
                Spacer()
                Button("追加") {
                    feature.addTagButtonTapped()
                }
                .font(.caption)
            }
        }
    }

    @ToolbarContentBuilder
    private var workspaceToolbar: some ToolbarContent {
        ToolbarItem(placement: .topBarLeading) {
            Button {
                feature.refreshWorkspaceData()
            } label: {
                Image(systemName: "arrow.clockwise")
            }
            .disabled(feature.isLoadingWorkspaceData || feature.isSyncingEntries)
        }

        ToolbarItem(placement: .topBarLeading) {
            Button {
                feature.syncButtonTapped()
            } label: {
                Image(systemName: "icloud.and.arrow.up")
            }
            .disabled(feature.isSyncingEntries)
        }

        ToolbarItem(placement: .topBarTrailing) {
            Button("ログアウト") {
                feature.logoutButtonTapped()
            }
            .disabled(feature.isAuthRequestInFlight)
        }
    }

    private func entryRow(_ entry: TimeEntryRecord) -> some View {
        HStack {
            VStack(alignment: .leading, spacing: 4) {
                Text(entry.startedAt, format: .dateTime.month().day().hour().minute())
                Text(entry.title)
                    .font(.subheadline)
                if let projectName = entry.projectName {
                    Text(projectName)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
                if !entry.tagNameList.isEmpty {
                    Text(entry.tagNameList.joined(separator: ", "))
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
                Text("終了 \(entry.endedAt.formatted(date: .omitted, time: .shortened))")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
            Spacer()
            VStack(alignment: .trailing, spacing: 4) {
                Text(durationText(entry.durationSeconds))
                    .monospacedDigit()
                if entry.syncStatus != "synced" {
                    Text("未同期")
                        .font(.caption2)
                        .foregroundStyle(.orange)
                }
            }
        }
    }

    private func durationText(_ elapsedSeconds: Int) -> String {
        let hours = elapsedSeconds / 3_600
        let minutes = (elapsedSeconds % 3_600) / 60
        let seconds = elapsedSeconds % 60
        return String(format: "%02d:%02d:%02d", hours, minutes, seconds)
    }

    private func durationAccessibilityText(_ elapsedSeconds: Int) -> String {
        let hours = elapsedSeconds / 3_600
        let minutes = (elapsedSeconds % 3_600) / 60
        let seconds = elapsedSeconds % 60
        return "\(hours)時間\(minutes)分\(seconds)秒"
    }
}

private struct FlowTagList: View {
    let tags: [Tag]

    var body: some View {
        LazyVGrid(columns: [GridItem(.adaptive(minimum: 96), spacing: 8)], alignment: .leading, spacing: 8) {
            ForEach(tags) { tag in
                HStack(spacing: 6) {
                    Circle()
                        .fill(Color(hex: tag.color))
                        .frame(width: 8, height: 8)
                    Text(tag.name)
                        .lineLimit(1)
                }
                .font(.caption)
                .padding(.horizontal, 10)
                .padding(.vertical, 6)
                .background(.thinMaterial, in: Capsule())
            }
        }
    }
}

private struct SummaryBarView: View {
    let item: AppFeature.DailySummaryItem
    let durationText: String

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack {
                HStack(spacing: 8) {
                    Circle()
                        .fill(Color(hex: item.color))
                        .frame(width: 10, height: 10)
                    Text(item.name)
                        .lineLimit(1)
                }
                Spacer()
                Text(durationText)
                    .font(.caption)
                    .monospacedDigit()
                    .foregroundStyle(.secondary)
            }

            GeometryReader { proxy in
                ZStack(alignment: .leading) {
                    Capsule()
                        .fill(.quaternary)
                    Capsule()
                        .fill(Color(hex: item.color))
                        .frame(width: max(4, proxy.size.width * item.ratio))
                }
            }
            .frame(height: 8)
        }
    }
}

private struct EntryEditView: View {
    @ObservedObject var feature: AppFeature
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            Form {
                Section("プロジェクト") {
                    Picker(
                        "プロジェクト",
                        selection: Binding(
                            get: { feature.entryEditDraft?.projectID },
                            set: { feature.editProjectSelectionChanged($0) }
                        )
                    ) {
                        Text("未選択").tag(String?.none)
                        ForEach(feature.projects.filter { !$0.isArchived }) { project in
                            Text(project.name).tag(String?.some(project.id))
                        }
                    }
                }

                Section("タグ") {
                    if feature.tags.isEmpty {
                        Text("タグはありません")
                            .foregroundStyle(.secondary)
                    } else {
                        ForEach(feature.tags) { tag in
                            Button {
                                feature.editTagSelectionToggled(tag.id)
                            } label: {
                                HStack {
                                    Circle()
                                        .fill(Color(hex: tag.color))
                                        .frame(width: 10, height: 10)
                                    Text(tag.name)
                                    Spacer()
                                    if feature.entryEditDraft?.selectedTagIDs.contains(tag.id) == true {
                                        Image(systemName: "checkmark")
                                            .foregroundStyle(.tint)
                                    }
                                }
                            }
                            .buttonStyle(.plain)
                        }
                    }
                }

                Section("メモ") {
                    TextField(
                        "メモ",
                        text: Binding(
                            get: { feature.entryEditDraft?.notes ?? "" },
                            set: { feature.editNotesChanged($0) }
                        ),
                        axis: .vertical
                    )
                    .lineLimit(3...6)
                }
            }
            .navigationTitle("作業記録を編集")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("キャンセル") {
                        feature.cancelEntryEdit()
                        dismiss()
                    }
                }

                ToolbarItem(placement: .confirmationAction) {
                    Button("保存") {
                        feature.saveEntryEdit()
                        dismiss()
                    }
                }
            }
        }
    }
}

private extension Color {
    init(hex: String) {
        var value = hex.trimmingCharacters(in: .whitespacesAndNewlines)
        if value.hasPrefix("#") {
            value.removeFirst()
        }

        guard value.count == 6, let integer = UInt64(value, radix: 16) else {
            self = .accentColor
            return
        }

        let red = Double((integer >> 16) & 0xFF) / 255
        let green = Double((integer >> 8) & 0xFF) / 255
        let blue = Double(integer & 0xFF) / 255
        self.init(red: red, green: green, blue: blue)
    }
}

private struct ProjectEditView: View {
    @ObservedObject var feature: AppFeature
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            Form {
                Section("基本情報") {
                    TextField(
                        "プロジェクト名",
                        text: Binding(
                            get: { feature.projectEditDraft?.name ?? "" },
                            set: { feature.projectDraftChanged(name: $0) }
                        )
                    )

                    TextField(
                        "説明",
                        text: Binding(
                            get: { feature.projectEditDraft?.description ?? "" },
                            set: { feature.projectDraftChanged(description: $0) }
                        ),
                        axis: .vertical
                    )
                    .lineLimit(2...4)

                    TextField(
                        "色（例: #3B82F6）",
                        text: Binding(
                            get: { feature.projectEditDraft?.color ?? "#3B82F6" },
                            set: { feature.projectDraftChanged(color: $0) }
                        )
                    )
                    .textInputAutocapitalization(.characters)
                    .autocorrectionDisabled()
                }

                if feature.projectEditDraft?.project != nil {
                    Section {
                        Toggle(
                            "アーカイブ",
                            isOn: Binding(
                                get: { feature.projectEditDraft?.isArchived ?? false },
                                set: { feature.projectDraftChanged(isArchived: $0) }
                            )
                        )
                    }
                }
            }
            .navigationTitle(feature.projectEditDraft?.project == nil ? "プロジェクト追加" : "プロジェクト編集")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("キャンセル") {
                        feature.cancelProjectEdit()
                        dismiss()
                    }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("保存") {
                        feature.saveProjectEdit()
                        dismiss()
                    }
                }
            }
        }
    }
}

private struct TagEditView: View {
    @ObservedObject var feature: AppFeature
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        NavigationStack {
            Form {
                Section("基本情報") {
                    TextField(
                        "タグ名",
                        text: Binding(
                            get: { feature.tagEditDraft?.name ?? "" },
                            set: { feature.tagDraftChanged(name: $0) }
                        )
                    )

                    TextField(
                        "色（例: #F97316）",
                        text: Binding(
                            get: { feature.tagEditDraft?.color ?? "#F97316" },
                            set: { feature.tagDraftChanged(color: $0) }
                        )
                    )
                    .textInputAutocapitalization(.characters)
                    .autocorrectionDisabled()
                }
            }
            .navigationTitle(feature.tagEditDraft?.tag == nil ? "タグ追加" : "タグ編集")
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("キャンセル") {
                        feature.cancelTagEdit()
                        dismiss()
                    }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("保存") {
                        feature.saveTagEdit()
                        dismiss()
                    }
                }
            }
        }
    }
}

#Preview {
    let configuration = ModelConfiguration(isStoredInMemoryOnly: true)
    let container = try! ModelContainer(
        for: TimeEntryRecord.self,
        configurations: configuration
    )
        ContentView(
            feature: AppFeature(
                entryStore: SwiftDataTimeEntryStore(modelContext: container.mainContext),
                authClient: PreviewContentAuthClient(),
                projectClient: PreviewProjectClient(),
                tagClient: PreviewTagClient(),
                entryClient: PreviewEntryClient()
        )
    )
    .modelContainer(container)
}

private final class PreviewContentAuthClient: AuthClientProtocol {
    func login(email: String, password: String) async throws -> AuthUser {
        AuthUser(id: "preview", email: email, displayName: "Preview", timeZone: "Asia/Tokyo", createdAt: nil, updatedAt: nil)
    }

    func signup(email: String, password: String, displayName: String?, timeZone: String?) async throws -> AuthUser {
        AuthUser(id: "preview", email: email, displayName: displayName, timeZone: timeZone, createdAt: nil, updatedAt: nil)
    }

    func currentUser() async throws -> AuthUser? {
        AuthUser(id: "preview", email: "preview@example.com", displayName: "Preview", timeZone: "Asia/Tokyo", createdAt: nil, updatedAt: nil)
    }

    func logout() async throws {}
}

private final class PreviewProjectClient: ProjectClientProtocol {
    func listProjects() async throws -> [Project] {
        [
            Project(id: "project-1", userId: "preview", name: "Client A", description: nil, color: "#3B82F6", isArchived: false, createdAt: nil, updatedAt: nil),
            Project(id: "project-2", userId: "preview", name: "Research", description: nil, color: "#22C55E", isArchived: false, createdAt: nil, updatedAt: nil)
        ]
    }

    func createProject(name: String, description: String, color: String) async throws -> Project {
        Project(id: "project-new", userId: "preview", name: name, description: description, color: color, isArchived: false, createdAt: nil, updatedAt: nil)
    }

    func updateProject(id: String, name: String, description: String, color: String, isArchived: Bool) async throws -> Project {
        Project(id: id, userId: "preview", name: name, description: description, color: color, isArchived: isArchived, createdAt: nil, updatedAt: nil)
    }
}

private final class PreviewTagClient: TagClientProtocol {
    func listTags() async throws -> [Tag] {
        [
            Tag(id: "tag-1", userId: "preview", name: "Deep Work", color: "#F97316", createdAt: nil, updatedAt: nil),
            Tag(id: "tag-2", userId: "preview", name: "Meeting", color: "#A855F7", createdAt: nil, updatedAt: nil)
        ]
    }

    func createTag(name: String, color: String) async throws -> Tag {
        Tag(id: "tag-new", userId: "preview", name: name, color: color, createdAt: nil, updatedAt: nil)
    }

    func updateTag(id: String, name: String, color: String) async throws -> Tag {
        Tag(id: id, userId: "preview", name: name, color: color, createdAt: nil, updatedAt: nil)
    }
}

private final class PreviewEntryClient: EntryClientProtocol {
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
        Entry(
            id: "entry-preview",
            userId: "preview",
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
            userId: "preview",
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
