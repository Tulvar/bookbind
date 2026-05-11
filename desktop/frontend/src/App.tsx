import {useEffect, useMemo, useState} from 'react';
import './App.css';
import {AppVersion, AvailableProviders, InspectPath} from '../wailsjs/go/main/App';

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
    const [inspectResult, setInspectResult] = useState<InspectView | null>(null);
    const [inspectError, setInspectError] = useState('');
    const [isInspecting, setIsInspecting] = useState(false);

    useEffect(() => {
        AppVersion().then(setVersion).catch(() => setVersion('unknown'));
        AvailableProviders().then(setProviders).catch(() => setProviders([]));
    }, []);

    const providerNames = useMemo(
        () => providers.filter((provider) => provider.Enabled).map((provider) => provider.Name).join(', '),
        [providers],
    );

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
                            <button className="primary-button" disabled={isInspecting} onClick={inspectInput} type="button">
                                {isInspecting ? 'Inspecting' : 'Inspect'}
                            </button>
                        </div>

                        <div className="form-grid">
                            <label>
                                Metadata
                                <input readOnly value="Optional bookbind.yaml" />
                            </label>
                            <label>
                                Cover
                                <input readOnly value="Optional JPG or PNG cover" />
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
                        <div className="summary-row">
                            <span>Providers</span>
                            <strong>{providerNames || 'No providers available'}</strong>
                        </div>
                        <div className="table-shell">
                            <div className="table-header">
                                <span>Provider</span>
                                <span>Title</span>
                                <span>Authors</span>
                                <span>Confidence</span>
                            </div>
                            <div className="table-empty">Search and candidate selection land here next.</div>
                        </div>
                    </section>
                )}

                {activeScreen === 'convert' && (
                    <section className="panel">
                        <h3>Convert</h3>
                        <div className="form-grid">
                            <label>
                                Output
                                <input readOnly value="book.m4b" />
                            </label>
                            <label>
                                Mode
                                <input readOnly value="Dry-run preview first" />
                            </label>
                        </div>
                        <div className="log-box">Conversion plan and progress events will appear here.</div>
                    </section>
                )}

                {activeScreen === 'cache' && (
                    <section className="panel">
                        <h3>Cache</h3>
                        <div className="summary-row">
                            <span>Provider response cache</span>
                            <strong>Ready</strong>
                        </div>
                        <div className="log-box">Cache list and clean actions will be wired to the Go bridge next.</div>
                    </section>
                )}
            </section>
        </main>
    );
}

export default App;
