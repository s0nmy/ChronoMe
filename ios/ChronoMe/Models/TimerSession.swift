import Foundation

/// 複数タイマー同時実行用のタイマーセッション
/// Web版のActiveEntryに相当する構造体
struct TimerSession: Identifiable {
    let id: UUID
    var projectID: String?
    var notes: String
    var tagIDs: Set<String>
    var isBreak: Bool
    let startedAt: Date
    var pausedDurationSeconds: Int
    var lastPausedAt: Date?

    init(
        id: UUID = UUID(),
        projectID: String? = nil,
        notes: String = "",
        tagIDs: Set<String> = [],
        isBreak: Bool = false,
        startedAt: Date = Date(),
        pausedDurationSeconds: Int = 0,
        lastPausedAt: Date? = nil
    ) {
        self.id = id
        self.projectID = projectID
        self.notes = notes
        self.tagIDs = tagIDs
        self.isBreak = isBreak
        self.startedAt = startedAt
        self.pausedDurationSeconds = pausedDurationSeconds
        self.lastPausedAt = lastPausedAt
    }

    /// 一時停止を考慮した現在の経過秒数を計算
    func currentElapsedSeconds() -> Int {
        let totalElapsed = Int(Date().timeIntervalSince(startedAt))

        // 一時停止中の場合、最後の一時停止時点からの時間を除外
        if let pausedAt = lastPausedAt {
            let currentPauseDuration = Int(Date().timeIntervalSince(pausedAt))
            return totalElapsed - pausedDurationSeconds - currentPauseDuration
        }

        return totalElapsed - pausedDurationSeconds
    }

    /// 一時停止中かどうか
    var isPaused: Bool {
        lastPausedAt != nil
    }

    /// 実行中かどうか
    var isRunning: Bool {
        !isPaused
    }
}
