import Foundation

/// ガントチャート表示用のデータ構造
/// エントリの時間軸上の位置とレーン番号を保持
struct GanttBlock: Identifiable {
    let id: UUID
    let entryID: UUID
    let projectID: String?
    let projectName: String
    let projectColor: String
    let notes: String
    let startedAt: Date
    let endedAt: Date
    let durationSeconds: Int
    var laneIndex: Int
    var left: Double  // 0.0-1.0 (24時間を1.0とした位置)
    var width: Double // 0.0-1.0 (24時間を1.0とした幅)

    init(
        id: UUID = UUID(),
        entryID: UUID,
        projectID: String?,
        projectName: String,
        projectColor: String,
        notes: String,
        startedAt: Date,
        endedAt: Date,
        durationSeconds: Int,
        laneIndex: Int = 0,
        left: Double = 0.0,
        width: Double = 0.0
    ) {
        self.id = id
        self.entryID = entryID
        self.projectID = projectID
        self.projectName = projectName
        self.projectColor = projectColor
        self.notes = notes
        self.startedAt = startedAt
        self.endedAt = endedAt
        self.durationSeconds = durationSeconds
        self.laneIndex = laneIndex
        self.left = left
        self.width = width
    }
}
