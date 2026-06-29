import Foundation
import SwiftData

@Model
final class TimeEntryRecord {
    @Attribute(.unique) var id: UUID
    var remoteEntryID: String?
    var title: String = "作業"
    var notes: String = ""
    var projectID: String?
    var projectName: String?
    var tagIDs: String = ""
    var tagNames: String = ""
    var startedAt: Date
    var endedAt: Date
    var durationSeconds: Int
    var syncStatus: String = "pending"

    init(
        id: UUID = UUID(),
        remoteEntryID: String? = nil,
        title: String,
        notes: String = "",
        projectID: String? = nil,
        projectName: String? = nil,
        tagIDs: [String] = [],
        tagNames: [String] = [],
        startedAt: Date,
        endedAt: Date,
        durationSeconds: Int,
        syncStatus: String = "pending"
    ) {
        self.id = id
        self.remoteEntryID = remoteEntryID
        self.title = title
        self.notes = notes
        self.projectID = projectID
        self.projectName = projectName
        self.tagIDs = tagIDs.joined(separator: ",")
        self.tagNames = tagNames.joined(separator: ",")
        self.startedAt = startedAt
        self.endedAt = endedAt
        self.durationSeconds = durationSeconds
        self.syncStatus = syncStatus
    }

    var tagNameList: [String] {
        tagNames.split(separator: ",").map(String.init)
    }

    var tagIDList: [String] {
        tagIDs.split(separator: ",").map(String.init)
    }
}
