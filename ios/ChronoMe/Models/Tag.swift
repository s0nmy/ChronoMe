import Foundation

struct Tag: Decodable, Equatable, Identifiable {
    let id: String
    let userId: String?
    let name: String
    let color: String
    let createdAt: Date?
    let updatedAt: Date?
}
