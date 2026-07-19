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
        ScrollView {
            VStack(spacing: 24) {
                activeTimersSection
                    .padding(.horizontal)

                if feature.timerSessions.count < 5 {
                    quickStartSection
                        .padding(.horizontal)
                }
            }
            .padding(.vertical, 20)
        }
    }

    private var entriesList: some View {
        ScrollView {
            VStack(spacing: 24) {
                datePickerSection
                    .padding(.horizontal)

                entriesGridSection
                    .padding(.horizontal)
            }
            .padding(.vertical, 20)
        }
    }

    private var reportsList: some View {
        DailyGanttChartView(
            entries: feature.selectedDateEntries,
            projects: feature.projects,
            selectedDate: Binding(
                get: { feature.selectedEntryDate },
                set: { feature.selectedEntryDateChanged($0) }
            )
        )
    }

    private var managementList: some View {
        ScrollView {
            VStack(spacing: 24) {
                projectsManagementSection
                    .padding(.horizontal)

                tagsManagementSection
                    .padding(.horizontal)
            }
            .padding(.vertical, 20)
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

    // MARK: - 作業履歴セクション

    private var datePickerSection: some View {
        VStack(spacing: 16) {
            DatePicker(
                "",
                selection: Binding(
                    get: { feature.selectedEntryDate },
                    set: { feature.selectedEntryDateChanged($0) }
                ),
                displayedComponents: .date
            )
            .datePickerStyle(.graphical)
            .padding(20)
            .background(
                RoundedRectangle(cornerRadius: 16)
                    .fill(.thinMaterial)
            )

            HStack(spacing: 24) {
                VStack(alignment: .leading, spacing: 4) {
                    Text("合計時間")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                    Text(durationText(feature.selectedDateTotalSeconds))
                        .font(.system(.title2, design: .rounded, weight: .semibold))
                        .monospacedDigit()
                }

                Spacer()

                VStack(alignment: .trailing, spacing: 4) {
                    Text("記録数")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                    Text("\(feature.selectedDateEntries.count)")
                        .font(.system(.title2, design: .rounded, weight: .semibold))
                }
            }
            .padding(20)
            .background(
                RoundedRectangle(cornerRadius: 16)
                    .fill(.thinMaterial)
            )
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

    // MARK: - 複数タイマー管理セクション

    private var activeTimersSection: some View {
        VStack(spacing: 20) {
            if !feature.timerSessions.isEmpty {
                // サマリー表示
                HStack(spacing: 24) {
                    VStack(alignment: .leading, spacing: 4) {
                        Text("実行中")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                        Text("\(feature.timerSessions.count)")
                            .font(.system(.title, design: .rounded, weight: .semibold))
                    }

                    Spacer()

                    VStack(alignment: .trailing, spacing: 4) {
                        Text("合計時間")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                        let totalSeconds = feature.timerSessions.reduce(0) { $0 + $1.currentElapsedSeconds() }
                        Text(durationText(totalSeconds))
                            .font(.system(.title2, design: .rounded, weight: .semibold))
                            .monospacedDigit()
                    }
                }
                .padding(20)
                .background(.thinMaterial, in: RoundedRectangle(cornerRadius: 16))

                // 各タイマーカード
                ForEach(feature.timerSessions) { session in
                    TimerSessionCard(
                        feature: feature,
                        session: session,
                        durationText: durationText
                    )
                }
            }
        }
    }

    // MARK: - クイックスタートセクション

    private var quickStartSection: some View {
        VStack(alignment: .leading, spacing: 16) {
            HStack {
                Image(systemName: "play.circle.fill")
                    .font(.title3)
                    .foregroundStyle(.tint)
                Text(feature.timerSessions.isEmpty ? "プロジェクトを選んで開始" : "別のタイマーを追加")
                    .font(.headline)
            }

            let activeProjects = feature.projects.filter { !$0.isArchived }

            if activeProjects.isEmpty {
                VStack(spacing: 12) {
                    Image(systemName: "folder.badge.plus")
                        .font(.system(size: 40))
                        .foregroundStyle(.secondary)
                    Text("プロジェクトがありません")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                    Text("プロジェクトタブからプロジェクトを作成してください")
                        .font(.caption)
                        .foregroundStyle(.tertiary)
                        .multilineTextAlignment(.center)
                }
                .frame(maxWidth: .infinity)
                .padding(.vertical, 32)
            } else {
                LazyVGrid(columns: [GridItem(.adaptive(minimum: 150), spacing: 12)], spacing: 12) {
                    ForEach(activeProjects) { project in
                        Button {
                            feature.startTimerSession(
                                projectID: project.id,
                                notes: "",
                                tagIDs: []
                            )
                        } label: {
                            VStack(alignment: .leading, spacing: 8) {
                                HStack {
                                    Circle()
                                        .fill(Color(hex: project.color))
                                        .frame(width: 12, height: 12)
                                    Spacer()
                                    Image(systemName: "play.fill")
                                        .font(.caption)
                                        .foregroundStyle(.secondary)
                                }

                                Text(project.name)
                                    .font(.subheadline.weight(.medium))
                                    .foregroundStyle(.primary)
                                    .lineLimit(2)
                                    .frame(maxWidth: .infinity, alignment: .leading)
                            }
                            .padding(16)
                            .background(
                                RoundedRectangle(cornerRadius: 12)
                                    .fill(Color(hex: project.color).opacity(0.1))
                            )
                        }
                        .buttonStyle(.plain)
                    }
                }
            }
        }
    }

    private var entriesGridSection: some View {
        VStack(alignment: .leading, spacing: 16) {
            HStack {
                Image(systemName: "list.bullet.rectangle")
                    .font(.title3)
                    .foregroundStyle(.tint)
                Text("作業記録")
                    .font(.headline)
            }

            if feature.selectedDateEntries.isEmpty {
                VStack(spacing: 12) {
                    Image(systemName: "clock.badge.questionmark")
                        .font(.system(size: 40))
                        .foregroundStyle(.secondary)
                    Text("この日の作業記録はありません")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }
                .frame(maxWidth: .infinity)
                .padding(.vertical, 40)
            } else {
                VStack(spacing: 12) {
                    ForEach(feature.selectedDateEntries) { entry in
                        Button {
                            feature.entryTapped(entry)
                        } label: {
                            EntryCard(entry: entry, durationText: durationText)
                        }
                        .buttonStyle(.plain)
                    }
                }
            }
        }
    }

    // MARK: - プロジェクト・タグ管理セクション

    private var projectsManagementSection: some View {
        VStack(alignment: .leading, spacing: 16) {
            HStack {
                Image(systemName: "folder.fill")
                    .font(.title3)
                    .foregroundStyle(.tint)
                Text("プロジェクト")
                    .font(.headline)
                Spacer()
                Button {
                    feature.addProjectButtonTapped()
                } label: {
                    Image(systemName: "plus.circle.fill")
                        .font(.title3)
                }
            }

            if feature.projects.isEmpty {
                VStack(spacing: 12) {
                    Image(systemName: "folder.badge.plus")
                        .font(.system(size: 40))
                        .foregroundStyle(.secondary)
                    Text("プロジェクトがありません")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                    Text("プロジェクトを作成して作業を整理しましょう")
                        .font(.caption)
                        .foregroundStyle(.tertiary)
                        .multilineTextAlignment(.center)
                }
                .frame(maxWidth: .infinity)
                .padding(.vertical, 40)
            } else {
                VStack(spacing: 12) {
                    ForEach(feature.projects) { project in
                        Button {
                            feature.projectTapped(project)
                        } label: {
                            ProjectCard(project: project)
                        }
                        .buttonStyle(.plain)
                    }
                }
            }
        }
    }

    private var tagsManagementSection: some View {
        VStack(alignment: .leading, spacing: 16) {
            HStack {
                Image(systemName: "tag.fill")
                    .font(.title3)
                    .foregroundStyle(.tint)
                Text("タグ")
                    .font(.headline)
                Spacer()
                Button {
                    feature.addTagButtonTapped()
                } label: {
                    Image(systemName: "plus.circle.fill")
                        .font(.title3)
                }
            }

            if feature.tags.isEmpty {
                VStack(spacing: 12) {
                    Image(systemName: "tag.slash")
                        .font(.system(size: 40))
                        .foregroundStyle(.secondary)
                    Text("タグがありません")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                    Text("タグを作成して作業を分類しましょう")
                        .font(.caption)
                        .foregroundStyle(.tertiary)
                        .multilineTextAlignment(.center)
                }
                .frame(maxWidth: .infinity)
                .padding(.vertical, 40)
            } else {
                FlowLayout(spacing: 12, lineSpacing: 12) {
                    ForEach(feature.tags) { tag in
                        Button {
                            feature.tagTapped(tag)
                        } label: {
                            HStack(spacing: 8) {
                                Circle()
                                    .fill(Color(hex: tag.color))
                                    .frame(width: 10, height: 10)
                                Text(tag.name)
                                    .font(.subheadline)
                            }
                            .padding(.horizontal, 16)
                            .padding(.vertical, 12)
                            .background(
                                RoundedRectangle(cornerRadius: 20)
                                    .fill(Color(hex: tag.color).opacity(0.1))
                            )
                        }
                        .buttonStyle(.plain)
                    }
                }
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

private struct ProjectCard: View {
    let project: Project

    var body: some View {
        HStack(spacing: 16) {
            Circle()
                .fill(Color(hex: project.color))
                .frame(width: 16, height: 16)

            VStack(alignment: .leading, spacing: 4) {
                Text(project.name)
                    .font(.subheadline.weight(.medium))
                    .foregroundStyle(.primary)

                if let description = project.description, !description.isEmpty {
                    Text(description)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                }

                if project.isArchived {
                    HStack(spacing: 4) {
                        Image(systemName: "archivebox.fill")
                            .font(.caption2)
                        Text("アーカイブ済み")
                            .font(.caption2.weight(.medium))
                    }
                    .foregroundStyle(.orange)
                }
            }

            Spacer()

            Image(systemName: "chevron.right")
                .font(.caption)
                .foregroundStyle(.tertiary)
        }
        .padding(16)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(Color(hex: project.color).opacity(0.08))
        )
    }
}

private struct EntryCard: View {
    let entry: TimeEntryRecord
    let durationText: (Int) -> String

    var body: some View {
        HStack(spacing: 16) {
            // プロジェクトカラーバー
            RoundedRectangle(cornerRadius: 3)
                .fill(Color.gray)
                .frame(width: 4)

            VStack(alignment: .leading, spacing: 8) {
                // タイトルと時間
                HStack {
                    Text(entry.title)
                        .font(.subheadline.weight(.medium))
                    Spacer()
                    Text(durationText(entry.durationSeconds))
                        .font(.subheadline.weight(.semibold))
                        .monospacedDigit()
                        .foregroundStyle(.secondary)
                }

                // 時刻表示
                HStack(spacing: 4) {
                    Image(systemName: "clock")
                        .font(.caption2)
                    Text("\(entry.startedAt.formatted(date: .omitted, time: .shortened)) - \(entry.endedAt.formatted(date: .omitted, time: .shortened))")
                        .font(.caption)
                }
                .foregroundStyle(.secondary)

                // プロジェクト名とタグ
                if let projectName = entry.projectName {
                    HStack(spacing: 4) {
                        Image(systemName: "folder")
                            .font(.caption2)
                        Text(projectName)
                            .font(.caption)
                    }
                    .foregroundStyle(.secondary)
                }

                // 同期ステータス
                if entry.syncStatus != "synced" {
                    HStack(spacing: 4) {
                        Image(systemName: "exclamationmark.triangle.fill")
                            .font(.caption2)
                        Text("未同期")
                            .font(.caption2.weight(.medium))
                    }
                    .foregroundStyle(.orange)
                }
            }
        }
        .padding(16)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(Color.gray.opacity(0.08))
        )
    }
}

private struct TimerSessionCard: View {
    @ObservedObject var feature: AppFeature
    let session: TimerSession
    let durationText: (Int) -> String

    @State private var editedNotes: String
    @State private var showingDetails = false

    init(feature: AppFeature, session: TimerSession, durationText: @escaping (Int) -> String) {
        self.feature = feature
        self.session = session
        self.durationText = durationText
        self._editedNotes = State(initialValue: session.notes)
    }

    var body: some View {
        VStack(spacing: 0) {
            // メインカード
            VStack(spacing: 16) {
                // 経過時間（大きく表示）
                VStack(spacing: 8) {
                    Text(durationText(session.currentElapsedSeconds()))
                        .font(.system(size: 48, weight: .bold, design: .rounded))
                        .monospacedDigit()
                        .foregroundStyle(
                            session.isPaused ? .secondary : Color(hex: projectColor)
                        )

                    if session.isPaused {
                        HStack(spacing: 4) {
                            Image(systemName: "pause.circle.fill")
                                .font(.caption)
                            Text("一時停止中")
                                .font(.caption.weight(.medium))
                        }
                        .foregroundStyle(.orange)
                    }
                }

                // プロジェクト名
                HStack(spacing: 8) {
                    Circle()
                        .fill(Color(hex: projectColor))
                        .frame(width: 10, height: 10)
                    Text(projectName)
                        .font(.subheadline.weight(.medium))
                        .foregroundStyle(.secondary)
                }

                // コントロールボタン
                HStack(spacing: 12) {
                    // 一時停止/再開
                    Button {
                        if session.isPaused {
                            feature.resumeTimerSession(session.id)
                        } else {
                            feature.pauseTimerSession(session.id)
                        }
                    } label: {
                        Image(systemName: session.isPaused ? "play.fill" : "pause.fill")
                            .font(.title3)
                            .frame(width: 44, height: 44)
                    }
                    .buttonStyle(.bordered)
                    .tint(.primary)

                    // 休憩トグル
                    Button {
                        feature.updateTimerSession(session.id, notes: nil, tagIDs: nil, isBreak: !session.isBreak)
                    } label: {
                        Image(systemName: session.isBreak ? "cup.and.saucer.fill" : "briefcase.fill")
                            .font(.title3)
                            .frame(width: 44, height: 44)
                    }
                    .buttonStyle(.bordered)
                    .tint(session.isBreak ? .orange : .blue)

                    Spacer()

                    // 詳細トグル
                    Button {
                        withAnimation(.spring(response: 0.3)) {
                            showingDetails.toggle()
                        }
                    } label: {
                        Image(systemName: showingDetails ? "chevron.up" : "chevron.down")
                            .font(.title3)
                            .frame(width: 44, height: 44)
                    }
                    .buttonStyle(.bordered)

                    // 終了ボタン
                    Button {
                        feature.stopTimerSession(session.id)
                    } label: {
                        Image(systemName: "stop.fill")
                            .font(.title3)
                            .frame(width: 44, height: 44)
                    }
                    .buttonStyle(.borderedProminent)
                    .tint(.red)
                }
            }
            .padding(24)
            .background(
                RoundedRectangle(cornerRadius: 20)
                    .fill(Color(hex: projectColor).opacity(0.08))
            )

            // 詳細セクション（展開可能）
            if showingDetails {
                VStack(alignment: .leading, spacing: 16) {
                    Divider()
                        .padding(.vertical, 8)

                    // メモ
                    VStack(alignment: .leading, spacing: 8) {
                        Text("メモ")
                            .font(.caption.weight(.semibold))
                            .foregroundStyle(.secondary)
                        TextField("メモを入力", text: Binding(
                            get: { editedNotes },
                            set: { newValue in
                                editedNotes = newValue
                                feature.updateTimerSession(session.id, notes: newValue, tagIDs: nil, isBreak: nil)
                            }
                        ), axis: .vertical)
                        .textFieldStyle(.roundedBorder)
                        .lineLimit(2...4)
                    }

                    // 開始時刻
                    HStack {
                        Text("開始時刻")
                            .font(.caption.weight(.semibold))
                            .foregroundStyle(.secondary)
                        Spacer()
                        Text(session.startedAt.formatted(date: .omitted, time: .shortened))
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
                .padding(.horizontal, 24)
                .padding(.bottom, 20)
                .transition(.opacity.combined(with: .move(edge: .top)))
            }
        }
    }

    private var projectName: String {
        if let projectID = session.projectID,
           let project = feature.projects.first(where: { $0.id == projectID }) {
            return project.name
        }
        return "未分類"
    }

    private var projectColor: String {
        if let projectID = session.projectID,
           let project = feature.projects.first(where: { $0.id == projectID }) {
            return project.color
        }
        return "#94A3B8"
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
        ProjectRecord.self,
        TagRecord.self,
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
