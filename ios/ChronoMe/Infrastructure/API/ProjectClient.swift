import Foundation

struct ProjectMutationRequest: Encodable, Equatable {
    let name: String
    let description: String
    let color: String
    let isArchived: Bool?
}

protocol ProjectClientProtocol {
    func listProjects() async throws -> [Project]
    func createProject(name: String, description: String, color: String) async throws -> Project
    func updateProject(id: String, name: String, description: String, color: String, isArchived: Bool) async throws -> Project
}

final class ProjectClient: ProjectClientProtocol {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func listProjects() async throws -> [Project] {
        let response: ProjectsResponse = try await apiClient.request("/api/projects/")
        return response.projects
    }

    func createProject(name: String, description: String, color: String) async throws -> Project {
        try await apiClient.request(
            "/api/projects/",
            method: .post,
            body: ProjectMutationRequest(
                name: name,
                description: description,
                color: color,
                isArchived: nil
            )
        )
    }

    func updateProject(id: String, name: String, description: String, color: String, isArchived: Bool) async throws -> Project {
        try await apiClient.request(
            "/api/projects/\(id)",
            method: .patch,
            body: ProjectMutationRequest(
                name: name,
                description: description,
                color: color,
                isArchived: isArchived
            )
        )
    }
}

private struct ProjectsResponse: Decodable {
    let projects: [Project]
}
