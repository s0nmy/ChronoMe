import Foundation
import XCTest
@testable import ChronoMe

final class APIClientTests: XCTestCase {
    func testMutatingRequestAddsCSRFHeaderFromCookieStorage() async throws {
        let baseURL = URL(string: "https://example.com")!
        let cookieStorage = HTTPCookieStorage()
        cookieStorage.setCookie(HTTPCookie(
            properties: [
                .domain: "example.com",
                .path: "/",
                .name: "chronome_csrf",
                .value: "csrf-token",
                .secure: "TRUE"
            ]
        )!)

        let session = MockURLSession(
            data: #"{"user":{"id":"1","email":"miyu@example.com"}}"#.data(using: .utf8)!,
            statusCode: 200,
            url: baseURL
        )
        let apiClient = APIClient(baseURL: baseURL, session: session, cookieStorage: cookieStorage)
        let authClient = AuthClient(apiClient: apiClient)

        _ = try await authClient.logout()

        XCTAssertEqual(session.lastRequest?.value(forHTTPHeaderField: "X-CSRF-Token"), "csrf-token")
    }

    func testCurrentUserReturnsNilForUnauthorizedResponse() async throws {
        let baseURL = URL(string: "https://example.com")!
        let session = MockURLSession(
            data: #"{"error":"unauthorized"}"#.data(using: .utf8)!,
            statusCode: 401,
            url: baseURL
        )
        let apiClient = APIClient(baseURL: baseURL, session: session, cookieStorage: HTTPCookieStorage())
        let authClient = AuthClient(apiClient: apiClient)

        let user = try await authClient.currentUser()

        XCTAssertNil(user)
    }

    func testMutatingRequestWithoutCSRFTokenFailsBeforeNetworkRequest() async throws {
        let baseURL = URL(string: "https://example.com")!
        let session = MockURLSession(data: Data(), statusCode: 204, url: baseURL)
        let apiClient = APIClient(baseURL: baseURL, session: session, cookieStorage: HTTPCookieStorage())
        let authClient = AuthClient(apiClient: apiClient)

        do {
            try await authClient.logout()
            XCTFail("logout should fail without CSRF token")
        } catch let error as APIClientError {
            XCTAssertEqual(error, .missingCSRFToken)
            XCTAssertNil(session.lastRequest)
        }
    }

    func testProjectClientDecodesProjects() async throws {
        let baseURL = URL(string: "https://example.com")!
        let session = MockURLSession(
            data: """
            {
              "projects": [
                {
                  "id": "project-1",
                  "user_id": "user-1",
                  "name": "Client A",
                  "description": "Important work",
                  "color": "#3B82F6",
                  "is_archived": false,
                  "created_at": "2026-06-26T01:00:00.123Z",
                  "updated_at": "2026-06-26T02:00:00Z"
                }
              ]
            }
            """.data(using: .utf8)!,
            statusCode: 200,
            url: baseURL
        )
        let apiClient = APIClient(baseURL: baseURL, session: session, cookieStorage: HTTPCookieStorage())
        let projectClient = ProjectClient(apiClient: apiClient)

        let projects = try await projectClient.listProjects()

        XCTAssertEqual(projects.count, 1)
        XCTAssertEqual(projects.first?.id, "project-1")
        XCTAssertEqual(projects.first?.name, "Client A")
        XCTAssertEqual(projects.first?.color, "#3B82F6")
        XCTAssertEqual(session.lastRequest?.url?.path, "/api/projects/")
    }

    func testTagClientDecodesTags() async throws {
        let baseURL = URL(string: "https://example.com")!
        let session = MockURLSession(
            data: """
            {
              "tags": [
                {
                  "id": "tag-1",
                  "user_id": "user-1",
                  "name": "Deep Work",
                  "color": "#F97316",
                  "created_at": "2026-06-26T01:00:00Z",
                  "updated_at": "2026-06-26T02:00:00.456Z"
                }
              ]
            }
            """.data(using: .utf8)!,
            statusCode: 200,
            url: baseURL
        )
        let apiClient = APIClient(baseURL: baseURL, session: session, cookieStorage: HTTPCookieStorage())
        let tagClient = TagClient(apiClient: apiClient)

        let tags = try await tagClient.listTags()

        XCTAssertEqual(tags.count, 1)
        XCTAssertEqual(tags.first?.id, "tag-1")
        XCTAssertEqual(tags.first?.name, "Deep Work")
        XCTAssertEqual(tags.first?.color, "#F97316")
        XCTAssertEqual(session.lastRequest?.url?.path, "/api/tags/")
    }

    func testProjectClientCreatesProjectWithCSRFHeader() async throws {
        let baseURL = URL(string: "https://example.com")!
        let cookieStorage = csrfCookieStorage(for: baseURL)
        let session = MockURLSession(
            data: """
            {
              "id": "project-1",
              "user_id": "user-1",
              "name": "Client A",
              "description": "Work",
              "color": "#3B82F6",
              "is_archived": false,
              "created_at": "2026-06-26T01:00:00Z",
              "updated_at": "2026-06-26T01:00:00Z"
            }
            """.data(using: .utf8)!,
            statusCode: 201,
            url: baseURL
        )
        let apiClient = APIClient(baseURL: baseURL, session: session, cookieStorage: cookieStorage)
        let projectClient = ProjectClient(apiClient: apiClient)

        let project = try await projectClient.createProject(name: "Client A", description: "Work", color: "#3B82F6")
        let payload = try JSONSerialization.jsonObject(with: try XCTUnwrap(session.lastRequest?.httpBody)) as? [String: Any]

        XCTAssertEqual(project.id, "project-1")
        XCTAssertEqual(session.lastRequest?.httpMethod, "POST")
        XCTAssertEqual(session.lastRequest?.url?.path, "/api/projects/")
        XCTAssertEqual(session.lastRequest?.value(forHTTPHeaderField: "X-CSRF-Token"), "csrf-token")
        XCTAssertEqual(payload?["name"] as? String, "Client A")
        XCTAssertEqual(payload?["color"] as? String, "#3B82F6")
    }

    func testTagClientUpdatesTagWithCSRFHeader() async throws {
        let baseURL = URL(string: "https://example.com")!
        let cookieStorage = csrfCookieStorage(for: baseURL)
        let session = MockURLSession(
            data: """
            {
              "id": "tag-1",
              "user_id": "user-1",
              "name": "Focus",
              "color": "#22C55E",
              "created_at": "2026-06-26T01:00:00Z",
              "updated_at": "2026-06-26T02:00:00Z"
            }
            """.data(using: .utf8)!,
            statusCode: 200,
            url: baseURL
        )
        let apiClient = APIClient(baseURL: baseURL, session: session, cookieStorage: cookieStorage)
        let tagClient = TagClient(apiClient: apiClient)

        let tag = try await tagClient.updateTag(id: "tag-1", name: "Focus", color: "#22C55E")
        let payload = try JSONSerialization.jsonObject(with: try XCTUnwrap(session.lastRequest?.httpBody)) as? [String: Any]

        XCTAssertEqual(tag.name, "Focus")
        XCTAssertEqual(session.lastRequest?.httpMethod, "PATCH")
        XCTAssertEqual(session.lastRequest?.url?.path, "/api/tags/tag-1")
        XCTAssertEqual(session.lastRequest?.value(forHTTPHeaderField: "X-CSRF-Token"), "csrf-token")
        XCTAssertEqual(payload?["name"] as? String, "Focus")
        XCTAssertEqual(payload?["color"] as? String, "#22C55E")
    }

    func testEntryClientCreatesEntryWithSnakeCasePayloadAndCSRFHeader() async throws {
        let baseURL = URL(string: "https://example.com")!
        let cookieStorage = HTTPCookieStorage()
        cookieStorage.setCookie(HTTPCookie(
            properties: [
                .domain: "example.com",
                .path: "/",
                .name: "chronome_csrf",
                .value: "csrf-token",
                .secure: "TRUE"
            ]
        )!)

        let session = MockURLSession(
            data: """
            {
              "id": "entry-1",
              "user_id": "user-1",
              "project_id": "project-1",
              "title": "Client A",
              "notes": "Initial drafting",
              "started_at": "2026-06-26T01:00:00Z",
              "ended_at": "2026-06-26T02:00:00Z",
              "duration_sec": 3600,
              "ratio": 1,
              "is_break": false,
              "tags": [],
              "created_at": "2026-06-26T02:00:01Z",
              "updated_at": "2026-06-26T02:00:01Z"
            }
            """.data(using: .utf8)!,
            statusCode: 201,
            url: baseURL
        )
        let apiClient = APIClient(baseURL: baseURL, session: session, cookieStorage: cookieStorage)
        let entryClient = EntryClient(apiClient: apiClient)

        let startedAt = Date(timeIntervalSince1970: 1_782_435_600)
        let endedAt = Date(timeIntervalSince1970: 1_782_439_200)
        let entry = try await entryClient.createEntry(
            title: "Client A",
            notes: "Initial drafting",
            projectId: "project-1",
            startedAt: startedAt,
            endedAt: endedAt,
            isBreak: false,
            tagIds: ["tag-1"]
        )

        let body = try XCTUnwrap(session.lastRequest?.httpBody)
        let payload = try JSONSerialization.jsonObject(with: body) as? [String: Any]

        XCTAssertEqual(entry.id, "entry-1")
        XCTAssertEqual(session.lastRequest?.url?.path, "/api/entries/")
        XCTAssertEqual(session.lastRequest?.value(forHTTPHeaderField: "X-CSRF-Token"), "csrf-token")
        XCTAssertEqual(payload?["project_id"] as? String, "project-1")
        XCTAssertEqual(payload?["tag_ids"] as? [String], ["tag-1"])
        XCTAssertEqual(payload?["is_break"] as? Bool, false)
        XCTAssertNotNil(payload?["started_at"] as? String)
        XCTAssertNotNil(payload?["ended_at"] as? String)
    }

    func testEntryClientListsEntriesWithDateRangeQuery() async throws {
        let baseURL = URL(string: "https://example.com")!
        let session = MockURLSession(
            data: """
            {
              "entries": [
                {
                  "id": "entry-1",
                  "user_id": "user-1",
                  "project_id": "project-1",
                  "title": "Client A",
                  "notes": "Initial drafting",
                  "started_at": "2026-06-26T01:00:00Z",
                  "ended_at": "2026-06-26T02:00:00Z",
                  "duration_sec": 3600,
                  "ratio": 1,
                  "is_break": false,
                  "tags": [],
                  "created_at": "2026-06-26T02:00:01Z",
                  "updated_at": "2026-06-26T02:00:01Z"
                }
              ]
            }
            """.data(using: .utf8)!,
            statusCode: 200,
            url: baseURL
        )
        let apiClient = APIClient(baseURL: baseURL, session: session, cookieStorage: HTTPCookieStorage())
        let entryClient = EntryClient(apiClient: apiClient)

        let entries = try await entryClient.listEntries(
            from: Date(timeIntervalSince1970: 1_782_432_000),
            to: Date(timeIntervalSince1970: 1_785_024_000)
        )

        XCTAssertEqual(entries.map(\.id), ["entry-1"])
        XCTAssertEqual(session.lastRequest?.url?.path, "/api/entries/")
        XCTAssertNotNil(URLComponents(url: try XCTUnwrap(session.lastRequest?.url), resolvingAgainstBaseURL: false)?.queryItems?.first { $0.name == "from" })
        XCTAssertNotNil(URLComponents(url: try XCTUnwrap(session.lastRequest?.url), resolvingAgainstBaseURL: false)?.queryItems?.first { $0.name == "to" })
    }

    func testEntryClientUpdatesEntryWithSnakeCasePayload() async throws {
        let baseURL = URL(string: "https://example.com")!
        let cookieStorage = HTTPCookieStorage()
        cookieStorage.setCookie(HTTPCookie(
            properties: [
                .domain: "example.com",
                .path: "/",
                .name: "chronome_csrf",
                .value: "csrf-token",
                .secure: "TRUE"
            ]
        )!)
        let session = MockURLSession(
            data: """
            {
              "id": "entry-1",
              "user_id": "user-1",
              "project_id": "project-1",
              "title": "Client A",
              "notes": "updated",
              "started_at": "2026-06-26T01:00:00Z",
              "ended_at": "2026-06-26T02:00:00Z",
              "duration_sec": 3600,
              "ratio": 1,
              "is_break": false,
              "tags": [],
              "created_at": "2026-06-26T02:00:01Z",
              "updated_at": "2026-06-26T03:00:01Z"
            }
            """.data(using: .utf8)!,
            statusCode: 200,
            url: baseURL
        )
        let apiClient = APIClient(baseURL: baseURL, session: session, cookieStorage: cookieStorage)
        let entryClient = EntryClient(apiClient: apiClient)

        let entry = try await entryClient.updateEntry(id: "entry-1", title: "Client A", notes: "updated", projectId: "project-1", tagIds: ["tag-1"])
        let body = try XCTUnwrap(session.lastRequest?.httpBody)
        let payload = try JSONSerialization.jsonObject(with: body) as? [String: Any]

        XCTAssertEqual(entry.notes, "updated")
        XCTAssertEqual(session.lastRequest?.httpMethod, "PATCH")
        XCTAssertEqual(session.lastRequest?.url?.path, "/api/entries/entry-1")
        XCTAssertEqual(session.lastRequest?.value(forHTTPHeaderField: "X-CSRF-Token"), "csrf-token")
        XCTAssertEqual(payload?["project_id"] as? String, "project-1")
        XCTAssertEqual(payload?["tag_ids"] as? [String], ["tag-1"])
    }

    func testEntryClientDeletesEntryWithCSRFHeader() async throws {
        let baseURL = URL(string: "https://example.com")!
        let cookieStorage = HTTPCookieStorage()
        cookieStorage.setCookie(HTTPCookie(
            properties: [
                .domain: "example.com",
                .path: "/",
                .name: "chronome_csrf",
                .value: "csrf-token",
                .secure: "TRUE"
            ]
        )!)
        let session = MockURLSession(data: Data(), statusCode: 204, url: baseURL)
        let apiClient = APIClient(baseURL: baseURL, session: session, cookieStorage: cookieStorage)
        let entryClient = EntryClient(apiClient: apiClient)

        try await entryClient.deleteEntry(id: "entry-1")

        XCTAssertEqual(session.lastRequest?.httpMethod, "DELETE")
        XCTAssertEqual(session.lastRequest?.url?.path, "/api/entries/entry-1")
        XCTAssertEqual(session.lastRequest?.value(forHTTPHeaderField: "X-CSRF-Token"), "csrf-token")
    }
}

private final class MockURLSession: URLSessionProtocol {
    private let data: Data
    private let statusCode: Int
    private let url: URL

    private(set) var lastRequest: URLRequest?

    init(data: Data, statusCode: Int, url: URL) {
        self.data = data
        self.statusCode = statusCode
        self.url = url
    }

    func data(for request: URLRequest) async throws -> (Data, URLResponse) {
        lastRequest = request
        return (
            data,
            HTTPURLResponse(
                url: url,
                statusCode: statusCode,
                httpVersion: nil,
                headerFields: nil
            )!
        )
    }
}

private func csrfCookieStorage(for baseURL: URL) -> HTTPCookieStorage {
    let storage = HTTPCookieStorage()
    storage.setCookie(HTTPCookie(
        properties: [
            .domain: baseURL.host ?? "example.com",
            .path: "/",
            .name: "chronome_csrf",
            .value: "csrf-token",
            .secure: "TRUE"
        ]
    )!)
    return storage
}
