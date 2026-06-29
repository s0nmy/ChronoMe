import Foundation

struct TagMutationRequest: Encodable, Equatable {
    let name: String
    let color: String
}

protocol TagClientProtocol {
    func listTags() async throws -> [Tag]
    func createTag(name: String, color: String) async throws -> Tag
    func updateTag(id: String, name: String, color: String) async throws -> Tag
}

final class TagClient: TagClientProtocol {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func listTags() async throws -> [Tag] {
        let response: TagsResponse = try await apiClient.request("/api/tags/")
        return response.tags
    }

    func createTag(name: String, color: String) async throws -> Tag {
        try await apiClient.request(
            "/api/tags/",
            method: .post,
            body: TagMutationRequest(name: name, color: color)
        )
    }

    func updateTag(id: String, name: String, color: String) async throws -> Tag {
        try await apiClient.request(
            "/api/tags/\(id)",
            method: .patch,
            body: TagMutationRequest(name: name, color: color)
        )
    }
}

private struct TagsResponse: Decodable {
    let tags: [Tag]
}
