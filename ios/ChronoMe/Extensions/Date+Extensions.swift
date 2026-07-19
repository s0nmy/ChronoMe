import Foundation

extension Date {
    /// 時刻の「時」部分を取得
    var hour: Int {
        Calendar.current.component(.hour, from: self)
    }

    /// 時刻の「分」部分を取得
    var minute: Int {
        Calendar.current.component(.minute, from: self)
    }

    /// 時刻の「秒」部分を取得
    var second: Int {
        Calendar.current.component(.second, from: self)
    }

    /// 0時からの経過分数を計算（ガントチャート用）
    var minutesFromMidnight: Double {
        Double(hour * 60 + minute) + Double(second) / 60.0
    }
}
