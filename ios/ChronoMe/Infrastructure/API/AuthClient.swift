import Foundation
import Security

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
        try await supabase.signUp(
            email: email,
            password: password,
            displayName: displayName,
            timeZone: timeZone
        )
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
    case sessionPersistenceFailed(Int32)

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
        case .sessionPersistenceFailed:
            return "認証セッションを安全に保存できませんでした。"
        }
    }
}

protocol AuthSessionStore: Sendable {
    func load() -> Data?
    func save(_ data: Data) throws
    func remove()
}

struct KeychainAuthSessionStore: AuthSessionStore {
    private let service: String
    private let account = "supabase-session"

    init(service: String = Bundle.main.bundleIdentifier ?? "com.chronome.app") {
        self.service = service
    }

    func load() -> Data? {
        var result: CFTypeRef?
        let status = SecItemCopyMatching(query(returnData: true) as CFDictionary, &result)
        guard status == errSecSuccess else { return nil }
        return result as? Data
    }

    func save(_ data: Data) throws {
        let updateStatus = SecItemUpdate(
            query() as CFDictionary,
            [kSecValueData as String: data] as CFDictionary
        )
        if updateStatus == errSecSuccess {
            return
        }

        guard updateStatus == errSecItemNotFound else {
            throw SupabaseAuthError.sessionPersistenceFailed(updateStatus)
        }

        var item = query()
        item[kSecValueData as String] = data
        let addStatus = SecItemAdd(item as CFDictionary, nil)
        guard addStatus == errSecSuccess else {
            throw SupabaseAuthError.sessionPersistenceFailed(addStatus)
        }
    }

    func remove() {
        _ = SecItemDelete(query() as CFDictionary)
    }

    private func query(returnData: Bool = false) -> [String: Any] {
        var query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account,
        ]
        if returnData {
            query[kSecReturnData as String] = true
            query[kSecMatchLimit as String] = kSecMatchLimitOne
        }
        return query
    }
}

actor SupabaseAuthClient {
    private let url: URL?
    private let anonKey: String?
    private let sessionStore: any AuthSessionStore
    private var session: SupabaseSession?

    init(
        url: URL?,
        anonKey: String?,
        sessionStore: any AuthSessionStore = KeychainAuthSessionStore()
    ) {
        self.url = url
        self.anonKey = anonKey
        self.sessionStore = sessionStore
        self.session = sessionStore.load().flatMap { data in
            try? JSONDecoder().decode(SupabaseSession.self, from: data)
        }
    }

    static func configured() -> SupabaseAuthClient {
        let rawURL = Bundle.main.object(forInfoDictionaryKey: "SUPABASE_URL") as? String
        let key = Bundle.main.object(forInfoDictionaryKey: "SUPABASE_ANON_KEY") as? String
        return SupabaseAuthClient(url: rawURL.flatMap(URL.init(string:)), anonKey: key)
    }

    func accessToken() async -> String? {
        if let token = currentAccessToken() {
            return token
        }

        guard let refreshToken = session?.refreshToken else { return nil }
        do {
            let refreshed = try await refresh(refreshToken: refreshToken)
            try persist(refreshed)
            session = refreshed
            return refreshed.accessToken
        } catch {
            // 一時的な通信失敗でセッションを消さず、次回のAPI呼び出しで再試行する。
            return nil
        }
    }

    func signIn(email: String, password: String) async throws {
        let response = try await authenticate(
            path: "/auth/v1/token?grant_type=password",
            body: LoginRequest(email: email, password: password)
        )
        guard let accessToken = response.accessToken, !accessToken.isEmpty else {
            throw SupabaseAuthError.sessionUnavailable
        }
        try persist(response)
        session = response
    }

    func signUp(email: String, password: String, displayName: String?, timeZone: String?) async throws {
        clearSession()
        var metadata: [String: String] = [:]
        if let displayName = displayName?.trimmingCharacters(in: .whitespacesAndNewlines), !displayName.isEmpty {
            metadata["full_name"] = displayName
        }
        if let timeZone = timeZone?.trimmingCharacters(in: .whitespacesAndNewlines), !timeZone.isEmpty {
            metadata["time_zone"] = timeZone
        }
        let response = try await authenticate(
            path: "/auth/v1/signup",
            body: SupabaseSignupRequest(
                email: email,
                password: password,
                options: SupabaseSignupOptions(data: metadata)
            )
        )
        guard let accessToken = response.accessToken, !accessToken.isEmpty else {
            return
        }
        try persist(response)
        session = response
    }

    func signOut() async throws {
        guard let token = await accessToken() else {
            clearSession()
            return
        }
        _ = try await request(path: "/auth/v1/logout", method: "POST", body: SupabaseLogoutPayload(), token: token)
        clearSession()
    }

    private func currentAccessToken() -> String? {
        guard let token = session?.accessToken, !token.isEmpty else { return nil }
        guard let expiresAt = session?.expiresAt else { return token }
        guard expiresAt > Int64(Date().timeIntervalSince1970) + 60 else { return nil }
        return token
    }

    private func refresh(refreshToken: String) async throws -> SupabaseSession {
        let response = try await authenticate(
            path: "/auth/v1/token?grant_type=refresh_token",
            body: SupabaseRefreshRequest(refreshToken: refreshToken)
        )
        guard let accessToken = response.accessToken, !accessToken.isEmpty else {
            throw SupabaseAuthError.sessionUnavailable
        }
        return SupabaseSession(
            accessToken: accessToken,
            refreshToken: response.refreshToken ?? refreshToken,
            expiresAt: response.expiresAt,
            tokenType: response.tokenType
        )
    }

    private func persist(_ session: SupabaseSession) throws {
        let data = try JSONEncoder().encode(session)
        try sessionStore.save(data)
    }

    private func clearSession() {
        session = nil
        sessionStore.remove()
    }

    private func authenticate(path: String, body: Encodable) async throws -> SupabaseSession {
        let data = try await request(path: path, method: "POST", body: body, token: nil)
        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        return try decoder.decode(SupabaseSession.self, from: data)
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

private struct SupabaseSignupRequest: Encodable {
    let email: String
    let password: String
    let options: SupabaseSignupOptions
}

private struct SupabaseSignupOptions: Encodable {
    let data: [String: String]
}

private struct SupabaseRefreshRequest: Encodable {
    let refreshToken: String

    enum CodingKeys: String, CodingKey {
        case refreshToken = "refresh_token"
    }
}

private struct SupabaseSession: Codable, Sendable {
    let accessToken: String?
    let refreshToken: String?
    let expiresAt: Int64?
    let tokenType: String?

    init(accessToken: String?, refreshToken: String?, expiresAt: Int64?, tokenType: String?) {
        self.accessToken = accessToken
        self.refreshToken = refreshToken
        self.expiresAt = expiresAt
        self.tokenType = tokenType
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        accessToken = try container.decodeIfPresent(String.self, forKey: .accessToken)
        refreshToken = try container.decodeIfPresent(String.self, forKey: .refreshToken)
        tokenType = try container.decodeIfPresent(String.self, forKey: .tokenType)

        let explicitExpiry = try container.decodeIfPresent(Int64.self, forKey: .expiresAt)
        let expiresIn = try container.decodeIfPresent(Int64.self, forKey: .expiresIn)
        expiresAt = explicitExpiry ?? expiresIn.map { Int64(Date().timeIntervalSince1970) + $0 }
    }

    func encode(to encoder: Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)
        try container.encodeIfPresent(accessToken, forKey: .accessToken)
        try container.encodeIfPresent(refreshToken, forKey: .refreshToken)
        try container.encodeIfPresent(expiresAt, forKey: .expiresAt)
        try container.encodeIfPresent(tokenType, forKey: .tokenType)
    }

    private enum CodingKeys: String, CodingKey {
        case accessToken
        case refreshToken
        case expiresAt
        case expiresIn
        case tokenType
    }
}

private struct SupabaseError: Decodable { let message: String }
private struct SupabaseLogoutPayload: Encodable {}
private struct SupabaseAnyEncodable: Encodable {
    private let encodeValue: (Encoder) throws -> Void

    init(_ value: Encodable) { encodeValue = value.encode(to:) }

    func encode(to encoder: Encoder) throws { try encodeValue(encoder) }
}
