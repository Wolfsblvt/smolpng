// SPDX-License-Identifier: MIT
// Independent native pasteboard consumer/producer, invoked only on disposable CI.
ObjC.import('AppKit');
ObjC.import('Foundation');

function run(args) {
    const [operation, kind, path] = args;
    const board = $.NSPasteboard.generalPasteboard;
    if (operation === 'read' && kind === 'types') {
        const types = [];
        for (let index = 0; index < board.types.count; index++) types.push(board.types.objectAtIndex(index).js);
        if (!$(JSON.stringify(types)).writeToFileAtomicallyEncodingError(path, true, $.NSUTF8StringEncoding, null)) {
            throw new Error('Cannot retain native types');
        }
        return 'Native types read';
    }
    if (operation === 'read' && kind === 'files') {
        const urls = board.readObjectsForClassesOptions($.NSArray.arrayWithObject($.NSURL), $({}));
        if (!urls) throw new Error('No native file URL items');
        const paths = [];
        for (let index = 0; index < urls.count; index++) paths.push(urls.objectAtIndex(index).path.js);
        if (!$(JSON.stringify(paths)).writeToFileAtomicallyEncodingError(path, true, $.NSUTF8StringEncoding, null)) {
            throw new Error('Cannot retain native file URL read');
        }
        return 'Native file URLs read';
    }
    if (operation === 'write' && kind === 'files') {
        const input = $.NSString.stringWithContentsOfFileEncodingError(path, $.NSUTF8StringEncoding, null).js;
        const paths = JSON.parse(input);
        const items = $.NSMutableArray.alloc.init;
        for (const file of paths) {
            const item = $.NSPasteboardItem.alloc.init;
            const url = $.NSURL.fileURLWithPath($(file));
            if (!item.setStringForType(url.absoluteString, $('public.file-url'))) {
                throw new Error('Cannot populate native file URL item');
            }
            items.addObject(item);
        }
        board.clearContents;
        if (!board.writeObjects(items)) throw new Error('Native file URL write failed');
        const observed = board.readObjectsForClassesOptions($.NSArray.arrayWithObject($.NSURL), $({}));
        if (!observed || observed.count !== paths.length) throw new Error('Native file URL write did not populate its items');
        return 'Native file URLs written';
    }
    const type = kind === 'image' ? 'public.png' : 'public.utf8-plain-text';
    if (operation === 'read') {
        const data = board.dataForType(type);
        if (!data || !data.writeToFileAtomically(path, true)) throw new Error(`Cannot read ${type}`);
        return `Read ${type}`;
    }
    const data = $.NSData.dataWithContentsOfFile(path);
    if (!data) throw new Error('Missing fixture bytes');
    board.clearContents;
    if (!board.setDataForType(data, type)) throw new Error(`Cannot write ${type}`);
    return `Wrote ${type}`;
}
