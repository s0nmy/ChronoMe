import SwiftUI

/// 1日の作業時間を24時間の時間軸で視覚化するガントチャート
struct DailyGanttChartView: View {
    let entries: [TimeEntryRecord]
    let projects: [Project]
    @Binding var selectedDate: Date

    private let laneHeight: CGFloat = 32
    private let laneGap: CGFloat = 4

    var body: some View {
        ScrollView {
            VStack(spacing: 24) {
                dateNavigationHeader
                    .padding(.horizontal)

                ganttTimeline
                    .padding(.horizontal)

                entryDetailsList
                    .padding(.horizontal)
            }
            .padding(.vertical, 20)
        }
    }

    // MARK: - 日付ナビゲーションヘッダー

    private var dateNavigationHeader: some View {
        VStack(spacing: 16) {
            // 日付ナビゲーション
            HStack(spacing: 16) {
                Button {
                    selectedDate = Calendar.current.date(byAdding: .day, value: -1, to: selectedDate) ?? selectedDate
                } label: {
                    Image(systemName: "chevron.left")
                        .font(.title3)
                        .frame(width: 44, height: 44)
                }
                .buttonStyle(.bordered)

                VStack(spacing: 4) {
                    Text(selectedDate, format: .dateTime.year().month().day())
                        .font(.headline)

                    if !Calendar.current.isDateInToday(selectedDate) {
                        Button {
                            selectedDate = Date()
                        } label: {
                            Text("今日に戻る")
                                .font(.caption)
                        }
                    }
                }
                .frame(maxWidth: .infinity)

                Button {
                    selectedDate = Calendar.current.date(byAdding: .day, value: 1, to: selectedDate) ?? selectedDate
                } label: {
                    Image(systemName: "chevron.right")
                        .font(.title3)
                        .frame(width: 44, height: 44)
                }
                .buttonStyle(.bordered)
            }

            // サマリーカード
            HStack(spacing: 24) {
                VStack(alignment: .leading, spacing: 4) {
                    Text("合計時間")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                    Text(durationText(totalSeconds))
                        .font(.system(.title2, design: .rounded, weight: .semibold))
                        .monospacedDigit()
                }

                Spacer()

                VStack(alignment: .trailing, spacing: 4) {
                    Text("記録数")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                    Text("\(entries.count)")
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

    // MARK: - ガントタイムライン

    private var ganttTimeline: some View {
        VStack(alignment: .leading, spacing: 16) {
            HStack {
                Image(systemName: "chart.bar.xaxis")
                    .font(.title3)
                    .foregroundStyle(.tint)
                Text("タイムライン")
                    .font(.headline)
            }

            if entries.isEmpty {
                VStack(spacing: 12) {
                    Image(systemName: "calendar.badge.clock")
                        .font(.system(size: 40))
                        .foregroundStyle(.secondary)
                    Text("この日の作業記録はありません")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }
                .frame(maxWidth: .infinity)
                .padding(.vertical, 40)
            } else {
                ganttChart
                    .padding(20)
                    .background(
                        RoundedRectangle(cornerRadius: 16)
                            .fill(.thinMaterial)
                    )
            }
        }
    }

    private var ganttChart: some View {
        let ganttBlocks = buildGanttBlocks(from: entries)
        let maxLaneIndex = ganttBlocks.map(\.laneIndex).max() ?? 0
        let chartHeight = CGFloat(maxLaneIndex + 1) * (laneHeight + laneGap) - laneGap

        return VStack(spacing: 8) {
            // 時間軸目盛り
            timeAxisLabels

            // ガントブロック
            GeometryReader { geometry in
                ZStack(alignment: .topLeading) {
                    // 背景グリッド
                    ForEach(0..<13, id: \.self) { index in
                        let x = CGFloat(index) / 12.0 * geometry.size.width
                        Rectangle()
                            .fill(Color.secondary.opacity(0.1))
                            .frame(width: 1)
                            .offset(x: x)
                    }

                    // 現在時刻インジケーター（当日のみ）
                    if Calendar.current.isDateInToday(selectedDate) {
                        currentTimeIndicator(width: geometry.size.width)
                    }

                    // エントリブロック
                    ForEach(ganttBlocks) { block in
                        ganttBlock(block, totalWidth: geometry.size.width)
                    }
                }
                .frame(height: chartHeight)
            }
            .frame(height: chartHeight)
        }
    }

    private var timeAxisLabels: some View {
        HStack(spacing: 0) {
            ForEach(stride(from: 0, through: 24, by: 2).map { $0 }, id: \.self) { hour in
                Text("\(hour)")
                    .font(.caption2)
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity, alignment: hour == 0 ? .leading : (hour == 24 ? .trailing : .center))
            }
        }
    }

    private func currentTimeIndicator(width: CGFloat) -> some View {
        let now = Date()
        let left = now.minutesFromMidnight / (24 * 60)
        let xOffset = left * width

        return Rectangle()
            .fill(Color.red)
            .frame(width: 2)
            .offset(x: xOffset)
    }

    private func ganttBlock(_ block: GanttBlock, totalWidth: CGFloat) -> some View {
        let xOffset = block.left * totalWidth
        let blockWidth = max(block.width * totalWidth, 4)
        let yOffset = CGFloat(block.laneIndex) * (laneHeight + laneGap)

        return RoundedRectangle(cornerRadius: 6)
            .fill(Color(hex: block.projectColor))
            .frame(width: blockWidth, height: laneHeight)
            .overlay(
                Text(block.projectName)
                    .font(.caption2)
                    .foregroundStyle(.white)
                    .lineLimit(1)
                    .padding(.horizontal, 4)
                    .frame(maxWidth: .infinity, alignment: .leading)
            )
            .offset(x: xOffset, y: yOffset)
    }

    // MARK: - エントリ詳細リスト

    private var entryDetailsList: some View {
        VStack(alignment: .leading, spacing: 16) {
            HStack {
                Image(systemName: "list.bullet")
                    .font(.title3)
                    .foregroundStyle(.tint)
                Text("詳細")
                    .font(.headline)
            }

            if entries.isEmpty {
                Text("表示する記録はありません")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 40)
            } else {
                VStack(spacing: 12) {
                    ForEach(entries) { entry in
                        entryRow(entry)
                    }
                }
            }
        }
    }

    private func entryRow(_ entry: TimeEntryRecord) -> some View {
        let project = projects.first { $0.id == entry.projectID }
        let projectColor = project?.color ?? "#94A3B8"

        return HStack(spacing: 16) {
            RoundedRectangle(cornerRadius: 3)
                .fill(Color(hex: projectColor))
                .frame(width: 4)

            VStack(alignment: .leading, spacing: 8) {
                HStack {
                    Text(entry.projectName ?? "未分類")
                        .font(.subheadline.weight(.medium))
                    Spacer()
                    Text(durationText(entry.durationSeconds))
                        .font(.subheadline.weight(.semibold))
                        .monospacedDigit()
                        .foregroundStyle(.secondary)
                }

                if !entry.notes.isEmpty {
                    Text(entry.notes)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                        .lineLimit(2)
                }

                HStack(spacing: 4) {
                    Image(systemName: "clock")
                        .font(.caption2)
                    Text("\(entry.startedAt.formatted(date: .omitted, time: .shortened)) - \(entry.endedAt.formatted(date: .omitted, time: .shortened))")
                        .font(.caption)
                }
                .foregroundStyle(.secondary)
            }
        }
        .padding(16)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(Color(hex: projectColor).opacity(0.08))
        )
    }

    // MARK: - ヘルパー

    private var totalSeconds: Int {
        entries.reduce(0) { $0 + $1.durationSeconds }
    }

    private func durationText(_ seconds: Int) -> String {
        let hours = seconds / 3_600
        let minutes = (seconds % 3_600) / 60
        let secs = seconds % 60
        return String(format: "%02d:%02d:%02d", hours, minutes, secs)
    }

    private func buildGanttBlocks(from entries: [TimeEntryRecord]) -> [GanttBlock] {
        // 時系列順にソート
        let sortedEntries = entries.sorted { $0.startedAt < $1.startedAt }

        var blocks: [GanttBlock] = []
        var laneEndTimes: [Date] = []

        for entry in sortedEntries {
            // 空いているレーンを探す
            let laneIndex = laneEndTimes.firstIndex { $0 <= entry.startedAt } ?? laneEndTimes.count

            // レーンが存在しない場合は新しいレーンを追加
            if laneIndex == laneEndTimes.count {
                laneEndTimes.append(entry.endedAt)
            } else {
                laneEndTimes[laneIndex] = entry.endedAt
            }

            // 時間軸上の位置を計算
            let startMinutes = entry.startedAt.minutesFromMidnight
            let durationMinutes = Double(entry.durationSeconds) / 60.0
            let left = startMinutes / (24 * 60)
            let width = durationMinutes / (24 * 60)

            let project = projects.first { $0.id == entry.projectID }
            let block = GanttBlock(
                entryID: entry.id,
                projectID: entry.projectID,
                projectName: entry.projectName ?? "未分類",
                projectColor: project?.color ?? "#94A3B8",
                notes: entry.notes,
                startedAt: entry.startedAt,
                endedAt: entry.endedAt,
                durationSeconds: entry.durationSeconds,
                laneIndex: laneIndex,
                left: left,
                width: max(width, 0.005) // 最小幅を確保
            )

            blocks.append(block)
        }

        return blocks
    }
}

// MARK: - Color拡張（既存のものと重複する場合は削除）

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
