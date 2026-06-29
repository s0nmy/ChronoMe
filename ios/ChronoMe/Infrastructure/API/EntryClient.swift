import Foundation

struct EntryCreateRequest: Encodable, Equatable {
    let title: String
    let notes: String
    let projectId: String?
    let startedAt: String
    let endedAt: String
    let isBreak: Bool
    let tagIds: [String]
}

struct EntryUpdateRequest: Encodable, Equatable {
    let title: String
    let notes: String
    let projectId: String?
    let tagIds: [String]
}

protocol EntryClientProtocol {
    func listEntries(from: Date?, to: Date?) async throws -> [Entry]
    func createEntry(
        title: String,
        notes: String,
        projectId: String?,
        startedAt: Date,
        endedAt: Date,
        isBreak: Bool,
        tagIds: [String]
    ) async throws -> Entry
    func updateEntry(
        id: String,
        title: String,
        notes: String,
        projectId: String?,
        tagIds: [String]
    ) async throws -> Entry
    func deleteEntry(id: String) async throws
}

final class EntryClient: EntryClientProtocol {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func listEntries(from: Date? = nil, to: Date? = nil) async throws -> [Entry] {
        let response: EntriesResponse = try await apiClient.request(Self.listPath(from: from, to: to))
        return response.entries
    }

    func createEntry(
        title: String,
        notes: String,
        projectId: String?,
        startedAt: Date,
        endedAt: Date,
        isBreak: Bool = false,
        tagIds: [String]
    ) async throws -> Entry {
        try await apiClient.request(
            "/api/entries/",
            method: .post,
            body: EntryCreateRequest(
                title: title,
                notes: notes,
                projectId: projectId,
                startedAt: Self.formatDate(startedAt),
                endedAt: Self.formatDate(endedAt),
                isBreak: isBreak,
                tagIds: tagIds
            )
        )
    }

    func updateEntry(
        id: String,
        title: String,
        notes: String,
        projectId: String?,
        tagIds: [String]
    ) async throws -> Entry {
        try await apiClient.request(
            "/api/entries/\(id)",
            method: .patch,
            body: EntryUpdateRequest(
                title: title,
                notes: notes,
                projectId: projectId,
                tagIds: tagIds
            )
        )
    }

    func deleteEntry(id: String) async throws {
        let _: EmptyResponse = try await apiClient.request(
            "/api/entries/\(id)",
            method: .delete
        )
    }

    private static func formatDate(_ date: Date) -> String {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return formatter.string(from: date)
    }

    private static func listPath(from: Date?, to: Date?) -> String {
        var components = URLComponents()
        components.path = "/api/entries/"
        components.queryItems = [
            from.map { URLQueryItem(name: "from", value: formatDate($0)) },
            to.map { URLQueryItem(name: "to", value: formatDate($0)) }
        ].compactMap { $0 }
        return components.string ?? "/api/entries/"
    }
}

private struct EntriesResponse: Decodable {
    let entries: [Entry]
}
