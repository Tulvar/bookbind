import {useEffect, useMemo, useState} from 'react';
import './App.css';
import {AppVersion, AvailableProviders} from '../wailsjs/go/main/App';

type Screen = 'import' | 'metadata' | 'convert' | 'cache';

type ProviderInfo = {
    Name: string;
    Enabled: boolean;
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

    useEffect(() => {
        AppVersion().then(setVersion).catch(() => setVersion('unknown'));
        AvailableProviders().then(setProviders).catch(() => setProviders([]));
    }, []);

    const providerNames = useMemo(
        () => providers.filter((provider) => provider.Enabled).map((provider) => provider.Name).join(', '),
        [providers],
    );

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
                        <div className="form-grid">
                            <label>
                                Input
                                <input readOnly value="MP3 file or folder picker will be connected next" />
                            </label>
                            <label>
                                Metadata
                                <input readOnly value="Optional bookbind.yaml" />
                            </label>
                            <label>
                                Cover
                                <input readOnly value="Optional JPG or PNG cover" />
                            </label>
                        </div>
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
