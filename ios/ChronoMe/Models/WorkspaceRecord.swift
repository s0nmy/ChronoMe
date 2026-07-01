import Foundation
import SwiftData

@Model
final class ProjectRecord {
    @Attribute(.unique) var id: String
    var userId: String?
    var name: String
    var projectDescription: String?
    var color: String
    var isArchived: Bool
    var createdAt: Date?
    var updatedAt: Date?

    init(project: Project) {
        self.id = project.id
        self.userId = project.userId
        self.name = project.name
        self.projectDescription = project.description
        self.color = project.color
        self.isArchived = project.isArchived
        self.createdAt = project.createdAt
        self.updatedAt = project.updatedAt
    }

    func update(from project: Project) {
        userId = project.userId
        name = project.name
        projectDescription = project.description
        color = project.color
        isArchived = project.isArchived
        createdAt = project.createdAt
        updatedAt = project.updatedAt
    }

    var project: Project {
        Project(
            id: id,
            userId: userId,
            name: name,
            description: projectDescription,
            color: color,
            isArchived: isArchived,
            createdAt: createdAt,
            updatedAt: updatedAt
        )
    }
}

@Model
final class TagRecord {
    @Attribute(.unique) var id: String
    var userId: String?
    var name: String
    var color: String
    var createdAt: Date?
    var updatedAt: Date?

    init(tag: Tag) {
        self.id = tag.id
        self.userId = tag.userId
        self.name = tag.name
        self.color = tag.color
        self.createdAt = tag.createdAt
        self.updatedAt = tag.updatedAt
    }

    func update(from tag: Tag) {
        userId = tag.userId
        name = tag.name
        color = tag.color
        createdAt = tag.createdAt
        updatedAt = tag.updatedAt
    }

    var tag: Tag {
        Tag(
            id: id,
            userId: userId,
            name: name,
            color: color,
            createdAt: createdAt,
            updatedAt: updatedAt
        )
    }
}
