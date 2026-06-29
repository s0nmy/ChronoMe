import Foundation

protocol URLSessionProtocol {
    func data(for request: URLRequest) async throws -> (Data, URLResponse)
}

extension URLSession: URLSessionProtocol {}

enum HTTPMethod: String {
    case get = "GET"
    case post = "POST"
    case patch = "PATCH"
    case delete = "DELETE"
}

enum APIClientError: LocalizedError, Equatable {
    case invalidURL(String)
    case missingCSRFToken
    case invalidResponse
    case httpStatus(Int, String)
    case decodingFailed

    var errorDescription: String? {
        switch self {
        case let .invalidURL(path):
            return "Invalid API path: \(path)"
        case .missingCSRFToken:
            return "CSRF token is missing. Please log in again."
        case .invalidResponse:
            return "Invalid server response."
        case let .httpStatus(status, message):
            return "\(message) (\(status))"
        case .decodingFailed:
            return "Failed to decode server response."
        }
    }
}

struct EmptyResponse: Decodable, Equatable {}

final class APIClient {
    private let baseURL: URL
    private let session: URLSessionProtocol
    private let cookieStorage: HTTPCookieStorage
    private let encoder: JSONEncoder
    private let decoder: JSONDecoder

    init(
        baseURL: URL = URL(string: "http://localhost:8080")!,
        session: URLSessionProtocol? = nil,
        cookieStorage: HTTPCookieStorage = .shared
    ) {
        self.baseURL = baseURL
        self.cookieStorage = cookieStorage

        if let session {
            self.session = session
        } else {
            let configuration = URLSessionConfiguration.default
            configuration.httpCookieStorage = cookieStorage
            configuration.httpCookieAcceptPolicy = .always
            configuration.httpShouldSetCookies = true
            self.session = URLSession(configuration: configuration)
        }

        let encoder = JSONEncoder()
        encoder.keyEncodingStrategy = .convertToSnakeCase
        self.encoder = encoder

        let decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
        decoder.dateDecodingStrategy = .custom { decoder in
            try Self.decodeISO8601Date(from: decoder)
        }
        self.decoder = decoder
    }

    func request<Response: Decodable>(
        _ path: String,
        method: HTTPMethod = .get,
        body: Encodable? = nil,
        requiresCSRF: Bool? = nil
    ) async throws -> Response {
        let url = try makeURL(path: path)
        var request = URLRequest(url: url)
        request.httpMethod = method.rawValue
        request.setValue("application/json", forHTTPHeaderField: "Accept")

        if let body {
            request.httpBody = try encoder.encode(AnyEncodable(body))
            request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        }

        if requiresCSRF ?? method.requiresCSRF {
            guard let token = csrfToken(for: url) else {
                throw APIClientError.missingCSRFToken
            }
            request.setValue(token, forHTTPHeaderField: "X-CSRF-Token")
        }

        let (data, response) = try await session.data(for: request)
        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIClientError.invalidResponse
        }

        guard (200..<300).contains(httpResponse.statusCode) else {
            throw APIClientError.httpStatus(
                httpResponse.statusCode,
                parseErrorMessage(from: data) ?? "Request failed"
            )
        }

        if Response.self == EmptyResponse.self || httpResponse.statusCode == 204 || data.isEmpty {
            return EmptyResponse() as! Response
        }

        do {
            return try decoder.decode(Response.self, from: data)
        } catch {
            throw APIClientError.decodingFailed
        }
    }

    private func makeURL(path: String) throws -> URL {
        guard let url = URL(string: path, relativeTo: baseURL)?.absoluteURL else {
            throw APIClientError.invalidURL(path)
        }
        return url
    }

    private func csrfToken(for url: URL) -> String? {
        cookieStorage.cookies(for: url)?
            .first { $0.name == "chronome_csrf" }?
            .value
    }

    private func parseErrorMessage(from data: Data) -> String? {
        guard !data.isEmpty,
              let payload = try? JSONSerialization.jsonObject(with: data) as? [String: Any]
        else {
            return nil
        }

        if let message = payload["error"] as? String {
            return message
        }

        if let error = payload["error"] as? [String: Any],
           let message = error["message"] as? String {
            return message
        }

        return nil
    }

    private static func decodeISO8601Date(from decoder: Decoder) throws -> Date {
        let container = try decoder.singleValueContainer()
        let value = try container.decode(String.self)

        let fractionalFormatter = ISO8601DateFormatter()
        fractionalFormatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        if let date = fractionalFormatter.date(from: value) {
            return date
        }

        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime]
        if let date = formatter.date(from: value) {
            return date
        }

        throw DecodingError.dataCorruptedError(
            in: container,
            debugDescription: "Invalid ISO8601 date: \(value)"
        )
    }
}

private extension HTTPMethod {
    var requiresCSRF: Bool {
        switch self {
        case .get:
            return false
        case .post, .patch, .delete:
            return true
        }
    }
}

private struct AnyEncodable: Encodable {
    private let encode: (Encoder) throws -> Void

    init(_ value: Encodable) {
        encode = value.encode(to:)
    }

    func encode(to encoder: Encoder) throws {
        try encode(encoder)
    }
}
