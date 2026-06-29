import Foundation

struct AuthUser: Decodable, Equatable {
    let id: String
    let email: String
    let displayName: String?
    let timeZone: String?
    let createdAt: Date?
    let updatedAt: Date?
}

struct LoginRequest: Encodable, Equatable {
    let email: String
    let password: String
}

struct SignupRequest: Encodable, Equatable {
    let email: String
    let password: String
    let displayName: String?
    let timeZone: String?
}

protocol AuthClientProtocol {
    func login(email: String, password: String) async throws -> AuthUser
    func signup(email: String, password: String, displayName: String?, timeZone: String?) async throws -> AuthUser
    func currentUser() async throws -> AuthUser?
    func logout() async throws
}

final class AuthClient: AuthClientProtocol {
    private let apiClient: APIClient

    init(apiClient: APIClient) {
        self.apiClient = apiClient
    }

    func login(email: String, password: String) async throws -> AuthUser {
        let response: AuthResponse = try await apiClient.request(
            "/api/auth/login",
            method: .post,
            body: LoginRequest(email: email, password: password),
            requiresCSRF: false
        )
        return response.user
    }

    func signup(email: String, password: String, displayName: String?, timeZone: String?) async throws -> AuthUser {
        let response: AuthResponse = try await apiClient.request(
            "/api/auth/signup",
            method: .post,
            body: SignupRequest(
                email: email,
                password: password,
                displayName: displayName,
                timeZone: timeZone
            ),
            requiresCSRF: false
        )
        return response.user
    }

    func currentUser() async throws -> AuthUser? {
        do {
            let response: AuthResponse = try await apiClient.request("/api/auth/me")
            return response.user
        } catch let error as APIClientError {
            if case .httpStatus(401, _) = error {
                return nil
            }
            throw error
        }
    }

    func logout() async throws {
        let _: EmptyResponse = try await apiClient.request(
            "/api/auth/logout",
            method: .post
        )
    }
}

private struct AuthResponse: Decodable {
    let user: AuthUser
}
