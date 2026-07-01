import Foundation
import SwiftData

@MainActor
protocol WorkspaceStoring {
    func fetchProjects() throws -> [Project]
    func fetchTags() throws -> [Tag]
    func saveProject(_ project: Project) throws
    func saveTag(_ tag: Tag) throws
}

@MainActor
final class SwiftDataWorkspaceStore: WorkspaceStoring {
    private let modelContext: ModelContext

    init(modelContext: ModelContext) {
        self.modelContext = modelContext
    }

    func fetchProjects() throws -> [Project] {
        let records = try modelContext.fetch(FetchDescriptor<ProjectRecord>())
        return records
            .sorted { ($0.createdAt ?? .distantPast) > ($1.createdAt ?? .distantPast) }
            .map(\.project)
    }

    func fetchTags() throws -> [Tag] {
        let records = try modelContext.fetch(FetchDescriptor<TagRecord>())
        return records
            .sorted { ($0.createdAt ?? .distantPast) > ($1.createdAt ?? .distantPast) }
            .map(\.tag)
    }

    func saveProject(_ project: Project) throws {
        let id = project.id
        var descriptor = FetchDescriptor<ProjectRecord>(
            predicate: #Predicate { record in
                record.id == id
            }
        )
        descriptor.fetchLimit = 1

        if let record = try modelContext.fetch(descriptor).first {
            record.update(from: project)
        } else {
            modelContext.insert(ProjectRecord(project: project))
        }

        try modelContext.save()
    }

    func saveTag(_ tag: Tag) throws {
        let id = tag.id
        var descriptor = FetchDescriptor<TagRecord>(
            predicate: #Predicate { record in
                record.id == id
            }
        )
        descriptor.fetchLimit = 1

        if let record = try modelContext.fetch(descriptor).first {
            record.update(from: tag)
        } else {
            modelContext.insert(TagRecord(tag: tag))
        }

        try modelContext.save()
    }
}
