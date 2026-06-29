import Foundation

struct Entry: Decodable, Equatable, Identifiable {
    let id: String
    let userId: String?
    let projectId: String?
    let title: String
    let notes: String?
    let startedAt: Date
    let endedAt: Date?
    let durationSec: Int
    let ratio: Double
    let isBreak: Bool
    let tags: [Tag]
    let createdAt: Date?
    let updatedAt: Date?
}
