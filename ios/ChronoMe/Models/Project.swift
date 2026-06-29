import Foundation

struct Project: Decodable, Equatable, Identifiable {
    let id: String
    let userId: String?
    let name: String
    let description: String?
    let color: String
    let isArchived: Bool
    let createdAt: Date?
    let updatedAt: Date?
}
