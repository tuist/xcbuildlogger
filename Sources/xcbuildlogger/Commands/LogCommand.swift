import ArgumentParser

struct LogCommand: AsyncParsableCommand {
    func run() async throws {
        print("foo")
    }
}
