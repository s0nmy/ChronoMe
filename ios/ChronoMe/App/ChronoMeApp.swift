import SwiftData
import SwiftUI

@main
struct ChronoMeApp: App {
    private let modelContainer: ModelContainer
    @StateObject private var feature: AppFeature

    init() {
        do {
            let container = try ModelContainer(for: TimeEntryRecord.self)
            let apiClient = APIClient()
            let authClient = AuthClient(apiClient: apiClient)
            let projectClient = ProjectClient(apiClient: apiClient)
            let tagClient = TagClient(apiClient: apiClient)
            let entryClient = EntryClient(apiClient: apiClient)
            modelContainer = container
            _feature = StateObject(
                wrappedValue: AppFeature(
                    entryStore: SwiftDataTimeEntryStore(modelContext: container.mainContext),
                    authClient: authClient,
                    projectClient: projectClient,
                    tagClient: tagClient,
                    entryClient: entryClient,
                    skipsAuthentication: true
                )
            )
        } catch {
            fatalError("Failed to initialize SwiftData: \(error)")
        }
    }

    var body: some Scene {
        WindowGroup {
            ContentView(feature: feature)
                .modelContainer(modelContainer)
        }
    }
}
