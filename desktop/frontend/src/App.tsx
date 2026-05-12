import {useEffect, useState} from 'react';
import './App.css';
import {
    AppVersion,
    AvailableProviders,
    CleanCache,
    ConvertAudio,
    InspectPath,
    ListCache,
    PreviewMetadata,
    ResolveMetadata,
    SearchMetadata,
    SelectAudioDirectory,
    SelectAudioFile,
    SelectCacheDirectory,
    SelectCoverFile,
    SelectMetadataFile,
    SelectOutputFile,
} from '../wailsjs/go/main/App';

type Screen = 'import' | 'metadata' | 'convert' | 'cache';

type ProviderInfo = {
    Name: string;
    Enabled: boolean;
};

type InspectFileView = {
    Path: string;
    Name: string;
    Duration: string;
    Codec: string;
    Bitrate: string;
    Channels: number;
    Chapters: number;
    Tags: {
        Title?: string;
        Artist?: string;
        Album?: string;
        Genre?: string;
        Date?: string;
    };
};

type InspectView = {
    Path: string;
    Files: InspectFileView[];
    TotalTime: string;
};

type MetadataCandidateView = {
    Provider: string;
    ID: string;
    Title: string;
    Authors: string[];
    Narrators: string[];
    Series: string;
    SeriesIndex: string;
    Year: number;
    Duration: string;
    CoverURL: string;
    Confidence: string;
};

type BookMetadataView = {
    Title: string;
    Subtitle: string;
    Authors: string[];
    Author: string;
    Narrators: string[];
    Narrator: string;
    Series: string;
    SeriesIndex: string;
    Language: string;
    Genre: string;
    Description: string;
    Publisher: string;
    PublishedYear: number;
    Cover: string;
};

type MetadataPreviewView = {
    Candidate: MetadataCandidateView;
    Book: BookMetadataView;
};

type ConvertView = {
    InputPath: string;
    Files: InspectFileView[];
    TotalTime: string;
    MetadataPath: string;
    Title: string;
    CoverPath: string;
    ChapterEvery: string;
    OutputPath: string;
    DryRun: boolean;
    Command: string[];
    Status: string;
};

type CacheEntryView = {
    Path: string;
    Name: string;
    Kind: string;
    Size: string;
};

type CacheListView = {
    Path: string;
    Entries: CacheEntryView[];
    Size: string;
};

const screens: Array<{ id: Screen; label: string }> = [
    {id: 'import', label: 'Import'},
    {id: 'metadata', label: 'Metadata'},
    {id: 'convert', label: 'Convert'},
    {id: 'cache', label: 'Cache'},
];

function App() {
    const [activeScreen, setActiveScreen] = useState<Screen>('import');
    const [version, setVersion] = useState('');
    const [providers, setProviders] = useState<ProviderInfo[]>([]);
    const [inputPath, setInputPath] = useState('');
    const [metadataPath, setMetadataPath] = useState('');
    const [coverPath, setCoverPath] = useState('');
    const [inspectResult, setInspectResult] = useState<InspectView | null>(null);
    const [inspectError, setInspectError] = useState('');
    const [isInspecting, setIsInspecting] = useState(false);
    const [metadataTitle, setMetadataTitle] = useState('');
    const [metadataAuthor, setMetadataAuthor] = useState('');
    const [selectedProviders, setSelectedProviders] = useState<string[]>([]);
    const [metadataCandidates, setMetadataCandidates] = useState<MetadataCandidateView[]>([]);
    const [selectedCandidate, setSelectedCandidate] = useState<MetadataCandidateView | null>(null);
    const [metadataPreview, setMetadataPreview] = useState<MetadataPreviewView | null>(null);
    const [metadataOutputPath, setMetadataOutputPath] = useState('bookbind.yaml');
    const [overwriteMetadata, setOverwriteMetadata] = useState(false);
    const [metadataStatus, setMetadataStatus] = useState('');
    const [metadataError, setMetadataError] = useState('');
    const [isSearchingMetadata, setIsSearchingMetadata] = useState(false);
    const [isPreviewingMetadata, setIsPreviewingMetadata] = useState(false);
    const [isSavingMetadata, setIsSavingMetadata] = useState(false);
    const [outputPath, setOutputPath] = useState('book.m4b');
    const [chapterEvery, setChapterEvery] = useState('');
    const [overwriteOutput, setOverwriteOutput] = useState(false);
    const [convertResult, setConvertResult] = useState<ConvertView | null>(null);
    const [convertError, setConvertError] = useState('');
    const [isConverting, setIsConverting] = useState(false);
    const [cachePath, setCachePath] = useState('');
    const [cacheResult, setCacheResult] = useState<CacheListView | null>(null);
    const [cacheStatus, setCacheStatus] = useState('');
    const [cacheError, setCacheError] = useState('');
    const [isCacheBusy, setIsCacheBusy] = useState(false);

    useEffect(() => {
        AppVersion().then(setVersion).catch(() => setVersion('unknown'));
        AvailableProviders()
            .then((availableProviders) => {
                setProviders(availableProviders);
                setSelectedProviders(availableProviders.filter((provider) => provider.Enabled).map((provider) => provider.Name));
            })
            .catch(() => {
                setProviders([]);
                setSelectedProviders([]);
            });
    }, []);

    function inspectInput() {
        const path = inputPath.trim();
        if (!path) {
            setInspectError('Input path is required.');
            return;
        }

        setIsInspecting(true);
        setInspectError('');
        InspectPath(path)
            .then((result) => setInspectResult(result as InspectView))
            .catch((error) => {
                setInspectResult(null);
                setInspectError(String(error));
            })
            .finally(() => setIsInspecting(false));
    }

    function selectPath(action: () => Promise<string>, update: (path: string) => void) {
        setInspectError('');
        setConvertError('');
        setCacheError('');
        action()
            .then((path) => {
                if (path) {
                    update(path);
                }
            })
            .catch((error) => setInspectError(String(error)));
    }

    function toggleProvider(providerName: string) {
        setSelectedProviders((current) => {
            if (current.includes(providerName)) {
                return current.filter((name) => name !== providerName);
            }
            return [...current, providerName];
        });
    }

    function searchMetadata() {
        if (!metadataTitle.trim() && !metadataAuthor.trim()) {
            setMetadataError('Title or author is required.');
            return;
        }
        if (providers.length > 0 && selectedProviders.length === 0) {
            setMetadataError('Select at least one provider.');
            return;
        }

        setIsSearchingMetadata(true);
        setMetadataError('');
        setMetadataStatus('');
        setSelectedCandidate(null);
        setMetadataPreview(null);
        SearchMetadata(metadataTitle, metadataAuthor, selectedProviders)
            .then((result) => setMetadataCandidates((result.Candidates || []) as MetadataCandidateView[]))
            .catch((error) => {
                setMetadataCandidates([]);
                setMetadataError(String(error));
            })
            .finally(() => setIsSearchingMetadata(false));
    }

    function selectMetadataCandidate(candidate: MetadataCandidateView) {
        setSelectedCandidate(candidate);
        setMetadataError('');
        setMetadataStatus('');
        setMetadataPreview(null);
        setIsPreviewingMetadata(true);
        PreviewMetadata(candidate.Provider, candidate.ID)
            .then((result) => setMetadataPreview(result as MetadataPreviewView))
            .catch((error) => setMetadataError(String(error)))
            .finally(() => setIsPreviewingMetadata(false));
    }

    function saveMetadata() {
        if (!selectedCandidate) {
            setMetadataError('Select a candidate first.');
            return;
        }

        setIsSavingMetadata(true);
        setMetadataError('');
        setMetadataStatus('');
        ResolveMetadata(selectedCandidate.Provider, selectedCandidate.ID, metadataOutputPath, overwriteMetadata)
            .then((result) => setMetadataStatus(`Saved ${result.OutputPath}`))
            .catch((error) => setMetadataError(String(error)))
            .finally(() => setIsSavingMetadata(false));
    }

    function convertAudio(dryRun: boolean) {
        if (!inputPath.trim()) {
            setConvertError('Input path is required.');
            return;
        }

        setIsConverting(true);
        setConvertError('');
        setConvertResult(null);
        ConvertAudio(inputPath, outputPath, metadataPath, coverPath, chapterEvery, dryRun, overwriteOutput)
            .then((result) => setConvertResult(result as ConvertView))
            .catch((error) => setConvertError(String(error)))
            .finally(() => setIsConverting(false));
    }

    function refreshCache() {
        setIsCacheBusy(true);
        setCacheError('');
        setCacheStatus('');
        ListCache(cachePath)
            .then((result) => setCacheResult(result as CacheListView))
            .catch((error) => {
                setCacheResult(null);
                setCacheError(String(error));
            })
            .finally(() => setIsCacheBusy(false));
    }

    function cleanCache() {
        setIsCacheBusy(true);
        setCacheError('');
        setCacheStatus('');
        CleanCache(cachePath)
            .then((result) => {
                setCacheStatus(`Removed ${result.Removed} entries · ${result.RemovedSize}`);
                return ListCache(cachePath);
            })
            .then((result) => setCacheResult(result as CacheListView))
            .catch((error) => setCacheError(String(error)))
            .finally(() => setIsCacheBusy(false));
    }

    const convertLog = convertResult
        ? [
            `Input: ${convertResult.InputPath}`,
            `Files: ${convertResult.Files?.length || 0}${convertResult.TotalTime ? ` · ${convertResult.TotalTime}` : ''}`,
            convertResult.MetadataPath ? `Metadata: ${convertResult.MetadataPath}` : '',
            convertResult.Title ? `Title: ${convertResult.Title}` : '',
            convertResult.CoverPath ? `Cover: ${convertResult.CoverPath}` : '',
            convertResult.ChapterEvery ? `Chapter every: ${convertResult.ChapterEvery}` : '',
            `Output: ${convertResult.OutputPath}`,
            convertResult.DryRun ? 'Mode: dry-run' : 'Mode: convert',
            convertResult.Command?.length ? `Command: ${convertResult.Command.join(' ')}` : '',
            `Status: ${convertResult.Status}`,
        ].filter(Boolean).join('\n')
        : 'Conversion plan and progress events will appear here.';

    return (
        <main className="app-shell">
            <aside className="sidebar">
                <div className="brand">
                    <span className="brand-mark">B</span>
                    <div>
                        <h1>bookbind</h1>
                        <p>{version ? `Version ${version}` : 'Desktop preview'}</p>
                    </div>
                </div>

                <nav className="nav-list" aria-label="Primary">
                    {screens.map((screen) => (
                        <button
                            className={screen.id === activeScreen ? 'nav-item active' : 'nav-item'}
                            key={screen.id}
                            onClick={() => setActiveScreen(screen.id)}
                            type="button"
                        >
                            {screen.label}
                        </button>
                    ))}
                </nav>
            </aside>

            <section className="workspace">
                <header className="workspace-header">
                    <div>
                        <p className="eyebrow">Desktop shell</p>
                        <h2>{screens.find((screen) => screen.id === activeScreen)?.label}</h2>
                    </div>
                    <span className="status-pill">Go bridge connected</span>
                </header>

                {activeScreen === 'import' && (
                    <section className="panel">
                        <h3>Import source</h3>
                        <div className="inspect-row">
                            <label className="path-field">
                                Input path
                                <input
                                    onChange={(event) => setInputPath(event.target.value)}
                                    placeholder="/path/to/book.mp3 or /path/to/book-folder"
                                    value={inputPath}
                                />
                            </label>
                            <div className="button-group">
                                <button className="secondary-button" onClick={() => selectPath(SelectAudioFile, setInputPath)} type="button">
                                    File
                                </button>
                                <button className="secondary-button" onClick={() => selectPath(SelectAudioDirectory, setInputPath)} type="button">
                                    Folder
                                </button>
                            </div>
                            <button className="primary-button" disabled={isInspecting} onClick={inspectInput} type="button">
                                {isInspecting ? 'Inspecting' : 'Inspect'}
                            </button>
                        </div>

                        <div className="form-grid">
                            <label>
                                Metadata
                                <div className="field-with-button">
                                    <input
                                        onChange={(event) => setMetadataPath(event.target.value)}
                                        placeholder="Optional bookbind.yaml"
                                        value={metadataPath}
                                    />
                                    <button className="secondary-button" onClick={() => selectPath(SelectMetadataFile, setMetadataPath)} type="button">
                                        Browse
                                    </button>
                                </div>
                            </label>
                            <label>
                                Cover
                                <div className="field-with-button">
                                    <input
                                        onChange={(event) => setCoverPath(event.target.value)}
                                        placeholder="Optional JPG or PNG cover"
                                        value={coverPath}
                                    />
                                    <button className="secondary-button" onClick={() => selectPath(SelectCoverFile, setCoverPath)} type="button">
                                        Browse
                                    </button>
                                </div>
                            </label>
                        </div>

                        {inspectError && <div className="error-box">{inspectError}</div>}

                        {inspectResult && (
                            <div className="inspect-result">
                                <div className="summary-row">
                                    <span>{inspectResult.Path}</span>
                                    <strong>{inspectResult.Files.length} files · {inspectResult.TotalTime || 'duration unknown'}</strong>
                                </div>
                                <div className="file-list">
                                    {inspectResult.Files.map((file) => (
                                        <div className="file-row" key={file.Path}>
                                            <div>
                                                <strong>{file.Name || file.Path}</strong>
                                                <p>
                                                    {[file.Duration, file.Codec, file.Bitrate, file.Channels ? `${file.Channels} ch` : '']
                                                        .filter(Boolean)
                                                        .join(' · ')}
                                                </p>
                                            </div>
                                            <div className="tag-stack">
                                                {file.Tags?.Title && <span>{file.Tags.Title}</span>}
                                                {file.Tags?.Artist && <span>{file.Tags.Artist}</span>}
                                                {file.Chapters > 0 && <span>{file.Chapters} chapters</span>}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}
                    </section>
                )}

                {activeScreen === 'metadata' && (
                    <section className="panel">
                        <h3>Metadata search</h3>
                        <div className="metadata-search">
                            <label>
                                Title
                                <input
                                    onChange={(event) => setMetadataTitle(event.target.value)}
                                    placeholder="Book title"
                                    value={metadataTitle}
                                />
                            </label>
                            <label>
                                Author
                                <input
                                    onChange={(event) => setMetadataAuthor(event.target.value)}
                                    placeholder="Optional author"
                                    value={metadataAuthor}
                                />
                            </label>
                            <button className="primary-button" disabled={isSearchingMetadata} onClick={searchMetadata} type="button">
                                {isSearchingMetadata ? 'Searching' : 'Search'}
                            </button>
                        </div>

                        <div className="provider-strip" aria-label="Metadata providers">
                            {providers.length === 0 && <span>No providers available</span>}
                            {providers.map((provider) => (
                                <label className={provider.Enabled ? 'provider-toggle' : 'provider-toggle disabled'} key={provider.Name}>
                                    <input
                                        checked={selectedProviders.includes(provider.Name)}
                                        disabled={!provider.Enabled}
                                        onChange={() => toggleProvider(provider.Name)}
                                        type="checkbox"
                                    />
                                    {provider.Name}
                                </label>
                            ))}
                        </div>

                        {metadataError && <div className="error-box">{metadataError}</div>}
                        {metadataStatus && <div className="success-box">{metadataStatus}</div>}

                        <div className="metadata-layout">
                            <div className="table-shell">
                                <div className="metadata-table-header">
                                    <span>Provider</span>
                                    <span>Title</span>
                                    <span>Authors</span>
                                    <span>Year</span>
                                    <span>Confidence</span>
                                </div>
                                {metadataCandidates.length === 0 && (
                                    <div className="table-empty">
                                        {isSearchingMetadata ? 'Searching providers...' : 'No candidates yet.'}
                                    </div>
                                )}
                                {metadataCandidates.map((candidate) => {
                                    const selected = selectedCandidate?.Provider === candidate.Provider && selectedCandidate?.ID === candidate.ID;
                                    return (
                                        <button
                                            className={selected ? 'candidate-row selected' : 'candidate-row'}
                                            key={`${candidate.Provider}:${candidate.ID}`}
                                            onClick={() => selectMetadataCandidate(candidate)}
                                            type="button"
                                        >
                                            <span>{candidate.Provider}</span>
                                            <strong>{candidate.Title || 'Untitled'}</strong>
                                            <span>{candidate.Authors?.join(', ') || 'Unknown'}</span>
                                            <span>{candidate.Year || ''}</span>
                                            <span>{candidate.Confidence || ''}</span>
                                        </button>
                                    );
                                })}
                            </div>

                            <aside className="preview-pane">
                                <div className="preview-heading">
                                    <h4>Preview</h4>
                                    {isPreviewingMetadata && <span>Loading</span>}
                                </div>
                                {metadataPreview ? (
                                    <div className="preview-content">
                                        <div className="cover-preview">
                                            {metadataPreview.Book.Cover ? (
                                                <img alt="" src={metadataPreview.Book.Cover} />
                                            ) : (
                                                <span>No cover</span>
                                            )}
                                        </div>
                                        <dl>
                                            <dt>Title</dt>
                                            <dd>{metadataPreview.Book.Title || '-'}</dd>
                                            <dt>Authors</dt>
                                            <dd>{metadataPreview.Book.Authors?.join(', ') || metadataPreview.Book.Author || '-'}</dd>
                                            <dt>Narrators</dt>
                                            <dd>{metadataPreview.Book.Narrators?.join(', ') || metadataPreview.Book.Narrator || '-'}</dd>
                                            <dt>Series</dt>
                                            <dd>
                                                {[metadataPreview.Book.Series, metadataPreview.Book.SeriesIndex].filter(Boolean).join(' #') || '-'}
                                            </dd>
                                            <dt>Year</dt>
                                            <dd>{metadataPreview.Book.PublishedYear || '-'}</dd>
                                        </dl>
                                        <label>
                                            Output
                                            <input
                                                onChange={(event) => setMetadataOutputPath(event.target.value)}
                                                placeholder="bookbind.yaml"
                                                value={metadataOutputPath}
                                            />
                                        </label>
                                        <label className="checkbox-line">
                                            <input
                                                checked={overwriteMetadata}
                                                onChange={(event) => setOverwriteMetadata(event.target.checked)}
                                                type="checkbox"
                                            />
                                            Overwrite existing file
                                        </label>
                                        <button className="primary-button" disabled={isSavingMetadata} onClick={saveMetadata} type="button">
                                            {isSavingMetadata ? 'Saving' : 'Save metadata'}
                                        </button>
                                    </div>
                                ) : (
                                    <div className="table-empty">Select a candidate to preview metadata.</div>
                                )}
                            </aside>
                        </div>
                    </section>
                )}

                {activeScreen === 'convert' && (
                    <section className="panel">
                        <h3>Convert</h3>
                        <div className="convert-grid">
                            <label>
                                Input
                                <input
                                    onChange={(event) => setInputPath(event.target.value)}
                                    placeholder="/path/to/book.mp3 or /path/to/book-folder"
                                    value={inputPath}
                                />
                            </label>
                            <label>
                                Output
                                <div className="field-with-button">
                                    <input
                                        onChange={(event) => setOutputPath(event.target.value)}
                                        placeholder="book.m4b"
                                        value={outputPath}
                                    />
                                    <button className="secondary-button" onClick={() => selectPath(SelectOutputFile, setOutputPath)} type="button">
                                        Browse
                                    </button>
                                </div>
                            </label>
                            <label>
                                Chapter interval
                                <input
                                    onChange={(event) => setChapterEvery(event.target.value)}
                                    placeholder="Optional, for example 10m"
                                    value={chapterEvery}
                                />
                            </label>
                        </div>
                        <div className="convert-grid secondary">
                            <label>
                                Metadata
                                <input
                                    onChange={(event) => setMetadataPath(event.target.value)}
                                    placeholder="Optional bookbind.yaml"
                                    value={metadataPath}
                                />
                            </label>
                            <label>
                                Cover
                                <input
                                    onChange={(event) => setCoverPath(event.target.value)}
                                    placeholder="Optional JPG or PNG cover"
                                    value={coverPath}
                                />
                            </label>
                            <label className="checkbox-line convert-checkbox">
                                <input
                                    checked={overwriteOutput}
                                    onChange={(event) => setOverwriteOutput(event.target.checked)}
                                    type="checkbox"
                                />
                                Overwrite output
                            </label>
                        </div>
                        <div className="action-row">
                            <button className="secondary-button" disabled={isConverting} onClick={() => convertAudio(true)} type="button">
                                {isConverting ? 'Working' : 'Dry run'}
                            </button>
                            <button className="primary-button" disabled={isConverting} onClick={() => convertAudio(false)} type="button">
                                {isConverting ? 'Working' : 'Convert'}
                            </button>
                        </div>
                        {convertError && <div className="error-box">{convertError}</div>}
                        <pre className="log-box">{convertLog}</pre>
                    </section>
                )}

                {activeScreen === 'cache' && (
                    <section className="panel">
                        <h3>Cache</h3>
                        <div className="cache-toolbar">
                            <label className="path-field">
                                Cache path
                                <div className="field-with-button">
                                    <input
                                        onChange={(event) => setCachePath(event.target.value)}
                                        placeholder="Default bookbind cache"
                                        value={cachePath}
                                    />
                                    <button className="secondary-button" onClick={() => selectPath(SelectCacheDirectory, setCachePath)} type="button">
                                        Browse
                                    </button>
                                </div>
                            </label>
                            <button className="secondary-button" disabled={isCacheBusy} onClick={refreshCache} type="button">
                                {isCacheBusy ? 'Working' : 'Refresh'}
                            </button>
                            <button
                                className="primary-button danger"
                                disabled={isCacheBusy || !cacheResult || cacheResult.Entries.length === 0}
                                onClick={cleanCache}
                                type="button"
                            >
                                Clean
                            </button>
                        </div>
                        {cacheError && <div className="error-box">{cacheError}</div>}
                        {cacheStatus && <div className="success-box">{cacheStatus}</div>}
                        {cacheResult && (
                            <div className="summary-row">
                                <span>{cacheResult.Path}</span>
                                <strong>{cacheResult.Entries.length} entries · {cacheResult.Size}</strong>
                            </div>
                        )}
                        <div className="table-shell">
                            <div className="cache-table-header">
                                <span>Type</span>
                                <span>Name</span>
                                <span>Size</span>
                            </div>
                            {!cacheResult && <div className="table-empty">Refresh cache to see local provider responses.</div>}
                            {cacheResult && cacheResult.Entries.length === 0 && <div className="table-empty">Cache is empty.</div>}
                            {cacheResult?.Entries.map((entry) => (
                                <div className="cache-row" key={entry.Path}>
                                    <span>{entry.Kind}</span>
                                    <strong>{entry.Name}</strong>
                                    <span>{entry.Size}</span>
                                </div>
                            ))}
                        </div>
                    </section>
                )}
            </section>
        </main>
    );
}

export default App;
