// SPDX-License-Identifier: MIT
// Native test consumer/producer only; never included in application packages.
import AppKit
import Foundation
import Darwin

enum ClipboardFailure: Error { case failed(String) }

func filePaths(_ board: NSPasteboard) throws -> [String] {
    let objects = board.readObjects(forClasses: [NSURL.self], options: nil) as? [NSURL] ?? []
    return try objects.map { url in
        guard url.isFileURL, let path = url.path else {
            throw ClipboardFailure.failed("Pasteboard item is not a local file URL")
        }
        return path
    }
}

do {
    guard CommandLine.arguments.count == 4 else {
        throw ClipboardFailure.failed("Expected operation, representation and fixture path")
    }
    let operation = CommandLine.arguments[1]
    let kind = CommandLine.arguments[2]
    let file = URL(fileURLWithPath: CommandLine.arguments[3])
    let board = NSPasteboard.general
    if operation == "read" {
        let data: Data
        if kind == "files" {
            data = try JSONEncoder().encode(filePaths(board))
        } else if kind == "types" {
            data = try JSONEncoder().encode(board.types?.map { $0.rawValue } ?? [])
        } else {
            let type: NSPasteboard.PasteboardType = kind == "image" ? .png : .string
            guard let observed = board.data(forType: type) else {
                throw ClipboardFailure.failed("No data for \(type.rawValue)")
            }
            data = observed
        }
        try data.write(to: file, options: .atomic)
    } else if operation == "write" {
        let data = try Data(contentsOf: file)
        board.clearContents()
        if kind == "files" {
            let paths = try JSONDecoder().decode([String].self, from: data)
            let objects: [NSPasteboardWriting] = paths.map { NSURL(fileURLWithPath: $0) }
            guard board.writeObjects(objects), try filePaths(board).count == paths.count else {
                throw ClipboardFailure.failed("Native file URL write/readback failed; item count \(board.pasteboardItems?.count ?? 0)")
            }
        } else {
            let type: NSPasteboard.PasteboardType = kind == "image" ? .png : .string
            guard board.setData(data, forType: type) else {
                throw ClipboardFailure.failed("Cannot write \(type.rawValue)")
            }
        }
    } else {
        throw ClipboardFailure.failed("Unknown operation")
    }
} catch {
    FileHandle.standardError.write(Data("\(error)\n".utf8))
    exit(1)
}
