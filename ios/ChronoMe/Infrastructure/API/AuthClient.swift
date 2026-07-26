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
    private let supabase: SupabaseAuthClient

    init(apiClient: APIClient, supabase: SupabaseAuthClient = .configured()) {
        self.apiClient = apiClient
        self.supabase = supabase
    }

    func login(email: String, password: String) async throws -> AuthUser {
        try await supabase.signIn(email: email, password: password)
        guard let user = try await currentUser() else { throw SupabaseAuthError.sessionUnavailable }
        return user
    }

    func signup(email: String, password: String, displayName: String?, timeZone: String?) async throws -> AuthUser {
        _ = displayName
        _ = timeZone
        try await supabase.signUp(email: email, password: password)
        guard let user = try await currentUser() else { throw SupabaseAuthError.emailConfirmationRequired }
        return user
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
        try await supabase.signOut()
    }

    func accessToken() async -> String? { await supabase.accessToken() }
}

private struct AuthResponse: Decodable {
    let user: AuthUser
}

enum SupabaseAuthError: LocalizedError {
    case configurationMissing
    case requestFailed(String)
    case sessionUnavailable
    case emailConfirmationRequired

    var errorDescription: String? {
        switch self {
        case .configurationMissing:
            return "Supabase の iOS 設定が不足しています。"
        case let .requestFailed(message):
            return message
        case .sessionUnavailable:
            return "認証セッションを取得できませんでした。"
        case .emailConfirmationRequired:
            return "確認メールを開いてからログインしてください。"
        }
    }
}

final class SupabaseAuthClient {
    private let url: URL?
    private let anonKey: String?
    private var token: String?

    init(url: URL?, anonKey: String?) {
        self.url = url
        self.anonKey = anonKey
    }

    static func configured() -> SupabaseAuthClient {
        let rawURL = Bundle.main.object(forInfoDictionaryKey: "SUPABASE_URL") as? String
        let key = Bundle.main.object(forInfoDictionaryKey: "SUPABASE_ANON_KEY") as? String
        return SupabaseAuthClient(url: rawURL.flatMap(URL.init(string:)), anonKey: key)
    }

    func accessToken() async -> String? { token }

    func signIn(email: String, password: String) async throws {
        token = try await authenticate(path: "/auth/v1/token?grant_type=password", body: LoginRequest(email: email, password: password))
    }

    func signUp(email: String, password: String) async throws {
        token = try await authenticate(path: "/auth/v1/signup", body: LoginRequest(email: email, password: password))
    }

    func signOut() async throws {
        guard let token else { return }
        _ = try await request(path: "/auth/v1/logout", method: "POST", body: SupabaseLogoutPayload(), token: token)
        self.token = nil
    }

    private func authenticate(path: String, body: Encodable) async throws -> String? {
        let data = try await request(path: path, method: "POST", body: body, token: nil)
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        return try? decoder.decode(SupabaseSession.self, from: data).accessToken
    }

    private func request(path: String, method: String, body: Encodable, token: String?) async throws -> Data {
        guard let url, let anonKey, !anonKey.isEmpty else { throw SupabaseAuthError.configurationMissing }
        guard let endpoint = URL(string: path, relativeTo: url) else { throw SupabaseAuthError.configurationMissing }
        var request = URLRequest(url: endpoint)
        request.httpMethod = method
        request.setValue(anonKey, forHTTPHeaderField: "apikey")
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue("Bearer \(token ?? anonKey)", forHTTPHeaderField: "Authorization")
        request.httpBody = try JSONEncoder().encode(SupabaseAnyEncodable(body))
        let (data, response) = try await URLSession.shared.data(for: request)
        guard let http = response as? HTTPURLResponse, (200..<300).contains(http.statusCode) else {
            let message = (try? JSONDecoder().decode(SupabaseError.self, from: data).message) ?? "Supabase 認証に失敗しました。"
            throw SupabaseAuthError.requestFailed(message)
        }
        return data
    }
}

private struct SupabaseSession: Decodable { let accessToken: String? }
private struct SupabaseError: Decodable { let message: String }
private struct SupabaseLogoutPayload: Encodable {}
private struct SupabaseAnyEncodable: Encodable {
    private let encodeValue: (Encoder) throws -> Void

    init(_ value: Encodable) { encodeValue = value.encode(to:) }

    func encode(to encoder: Encoder) throws { try encodeValue(encoder) }
}
