import Foundation
import SwiftData

@MainActor
protocol TimeEntryStoring {
    func fetchRecent(limit: Int) throws -> [TimeEntryRecord]
    func fetchEntries(from: Date, to: Date) throws -> [TimeEntryRecord]
    func fetchUnsynced() throws -> [TimeEntryRecord]
    @discardableResult
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
    ) throws -> TimeEntryRecord
    @discardableResult
    func upsertRemoteEntry(_ entry: Entry, project: Project?, tags: [Tag]) throws -> TimeEntryRecord
    func updateEntry(_ entry: TimeEntryRecord, title: String, notes: String, project: Project?, tags: [Tag], syncStatus: String) throws
    func deleteEntry(_ entry: TimeEntryRecord) throws
    func markSynced(_ entry: TimeEntryRecord, remoteEntryID: String) throws
    func markSyncFailed(_ entry: TimeEntryRecord) throws
}

@MainActor
final class SwiftDataTimeEntryStore: TimeEntryStoring {
    private let modelContext: ModelContext

    init(modelContext: ModelContext) {
        self.modelContext = modelContext
    }

    func fetchRecent(limit: Int = 20) throws -> [TimeEntryRecord] {
        var descriptor = FetchDescriptor<TimeEntryRecord>(
            sortBy: [SortDescriptor(\.startedAt, order: .reverse)]
        )
        descriptor.fetchLimit = limit
        return try modelContext.fetch(descriptor)
    }

    func fetchEntries(from: Date, to: Date) throws -> [TimeEntryRecord] {
        let descriptor = FetchDescriptor<TimeEntryRecord>(
            predicate: #Predicate { entry in
                entry.startedAt >= from && entry.startedAt < to
            },
            sortBy: [SortDescriptor(\.startedAt, order: .reverse)]
        )
        return try modelContext.fetch(descriptor)
    }

    func fetchUnsynced() throws -> [TimeEntryRecord] {
        let descriptor = FetchDescriptor<TimeEntryRecord>(
            predicate: #Predicate { entry in
                entry.syncStatus != "synced"
            },
            sortBy: [SortDescriptor(\.startedAt, order: .forward)]
        )
        return try modelContext.fetch(descriptor)
    }

    @discardableResult
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
        let entry = TimeEntryRecord(
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
        modelContext.insert(entry)
        try modelContext.save()
        return entry
    }

    @discardableResult
    func upsertRemoteEntry(_ entry: Entry, project: Project?, tags: [Tag]) throws -> TimeEntryRecord {
        let remoteEntryID = entry.id
        var descriptor = FetchDescriptor<TimeEntryRecord>(
            predicate: #Predicate { record in
                record.remoteEntryID == remoteEntryID
            }
        )
        descriptor.fetchLimit = 1

        let record = try modelContext.fetch(descriptor).first ?? TimeEntryRecord(
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

        record.remoteEntryID = entry.id
        record.title = entry.title
        record.notes = entry.notes ?? ""
        record.projectID = entry.projectId
        record.projectName = project?.name
        record.tagIDs = tags.map(\.id).joined(separator: ",")
        record.tagNames = tags.map(\.name).joined(separator: ",")
        record.startedAt = entry.startedAt
        record.endedAt = entry.endedAt ?? entry.startedAt
        record.durationSeconds = entry.durationSec
        record.syncStatus = "synced"

        if record.modelContext == nil {
            modelContext.insert(record)
        }
        try modelContext.save()
        return record
    }

    func updateEntry(_ entry: TimeEntryRecord, title: String, notes: String, project: Project?, tags: [Tag], syncStatus: String) throws {
        entry.title = title
        entry.notes = notes
        entry.projectID = project?.id
        entry.projectName = project?.name
        entry.tagIDs = tags.map(\.id).joined(separator: ",")
        entry.tagNames = tags.map(\.name).joined(separator: ",")
        entry.syncStatus = syncStatus
        try modelContext.save()
    }

    func deleteEntry(_ entry: TimeEntryRecord) throws {
        modelContext.delete(entry)
        try modelContext.save()
    }

    func markSynced(_ entry: TimeEntryRecord, remoteEntryID: String) throws {
        entry.remoteEntryID = remoteEntryID
        entry.syncStatus = "synced"
        try modelContext.save()
    }

    func markSyncFailed(_ entry: TimeEntryRecord) throws {
        entry.syncStatus = "failed"
        try modelContext.save()
    }
}
