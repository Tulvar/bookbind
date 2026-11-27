import {deepStrictEqual} from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {test} from 'node:test';
import ts from 'typescript';

const source = await readFile(new URL('../src/latestTask.ts', import.meta.url), 'utf8');
const {outputText} = ts.transpileModule(source, {
    compilerOptions: {module: ts.ModuleKind.ESNext, target: ts.ScriptTarget.ES2020},
});
const {createLatestTask} = await import('data:text/javascript;base64,' + Buffer.from(outputText).toString('base64'));

function deferred() {
    let resolve;
    let reject;
    const promise = new Promise((accept, fail) => {
        resolve = accept;
        reject = fail;
    });
    return {promise, resolve, reject};
}

function record(events, name) {
    return {
        success: (value) => events.push(['success', name, value]),
        failure: (error) => events.push(['failure', name, error.message]),
        finish: () => events.push(['finish', name]),
    };
}

test('a slow first book cannot replace the second book', async () => {
    const preview = createLatestTask();
    const events = [];
    const first = deferred();
    const second = deferred();

    const firstRequest = preview.run(() => first.promise, record(events, 'first'));
    const secondRequest = preview.run(() => second.promise, record(events, 'second'));

    second.resolve('Second book');
    await secondRequest;
    first.resolve('First book');
    await firstRequest;
    deepStrictEqual(events, [['success', 'second', 'Second book'], ['finish', 'second']]);
});

test('an outdated error cannot stop the current preview', async () => {
    const preview = createLatestTask();
    const events = [];
    const first = deferred();
    const second = deferred();

    const firstRequest = preview.run(() => first.promise, record(events, 'first'));
    const secondRequest = preview.run(() => second.promise, record(events, 'second'));

    first.reject(new Error('First failed'));
    await firstRequest;
    deepStrictEqual(events, []);
    second.reject(new Error('Second failed'));
    await secondRequest;
    deepStrictEqual(events, [['failure', 'second', 'Second failed'], ['finish', 'second']]);
});

test('starting a new search discards an unfinished preview', async () => {
    const preview = createLatestTask();
    const events = [];
    const pending = deferred();
    const request = preview.run(() => pending.promise, record(events, 'pending'));

    preview.invalidate();
    pending.resolve('Old book');
    await request;
    deepStrictEqual(events, []);
});
