import {useEffect, useState} from 'react';
import './App.css';
import {
    AppVersion,
    AvailableProviders,
    CancelConvert,
    CleanCache,
    ConvertAudio,
    ConvertAudioWithMetadata,
    InspectPath,
    ListCache,
    PreviewMetadata,
    PrepareConversion,
    ResolveMetadata,
    SearchMetadata,
    SelectAudioDirectory,
    SelectAudioFile,
    SelectCacheDirectory,
    SelectCoverFile,
    SelectMetadataFile,
    SelectOutputFile,
} from '../wailsjs/go/main/App';
import {EventsOn} from '../wailsjs/runtime/runtime';

type Screen = 'import' | 'metadata' | 'convert' | 'cache';
type Locale = 'en' | 'ru';

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
    Translators: string[];
    Translator: string;
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

type ConversionPreparationView = {
    Book: BookMetadataView;
    Missing: string[];
    Files: number;
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

type ConvertProgressEvent = {
    Phase: string;
    Line: string;
    Percent: number;
    Elapsed: string;
    Total: string;
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

const workflowScreens: Screen[] = ['import', 'metadata', 'convert'];

const translations = {
    en: {
        language: 'Language',
        languageEnglish: 'English',
        languageRussian: 'Russian',
        version: 'Version',
        desktopPreview: 'Desktop preview',
        desktopShell: 'Desktop shell',
        goBridgeConnected: 'Go bridge connected',
        nav: {
            import: 'Book',
            metadata: 'Enrich',
            convert: 'Convert',
            cache: 'Cache',
        },
        workflow: 'Workflow',
        step: 'Step',
        continueToMetadata: 'Continue to metadata',
        continueToConvert: 'Continue to conversion',
        chooseBookFirst: 'Choose an audiobook first.',
        metadataStepIntro: 'Search external sources, preview a candidate, then use it for conversion.',
        convertStepIntro: 'Review the selected book and run the final M4B conversion.',
        importSource: 'Import source',
        inputPath: 'Input path',
        inputPlaceholder: '/path/to/book.mp3 or /path/to/book-folder',
        file: 'File',
        folder: 'Folder',
        inspecting: 'Inspecting',
        inspect: 'Inspect',
        metadata: 'Metadata',
        metadataPlaceholder: 'Optional bookbind.yaml',
        cover: 'Cover',
        coverPlaceholder: 'Optional JPG or PNG cover',
        durationUnknown: 'duration unknown',
        files: 'files',
        channel: 'ch',
        chapters: 'chapters',
        metadataSearch: 'Metadata search',
        title: 'Title',
        bookTitle: 'Book title',
        author: 'Author',
        optionalAuthor: 'Optional author',
        searching: 'Searching',
        search: 'Search',
        noProviders: 'No providers available',
        provider: 'Provider',
        authors: 'Authors',
        year: 'Year',
        confidence: 'Confidence',
        searchingProviders: 'Searching providers...',
        noCandidates: 'No candidates yet.',
        untitled: 'Untitled',
        unknown: 'Unknown',
        preview: 'Preview',
        loading: 'Loading',
        noCover: 'No cover',
        narrators: 'Narrators',
        translators: 'Translators',
        series: 'Series',
        genre: 'Genre',
        output: 'Output',
        overwriteExistingFile: 'Overwrite existing file',
        saving: 'Saving',
        saveMetadata: 'Save metadata',
        saveYamlInstead: 'Save YAML instead',
        close: 'Close',
        clearMetadata: 'Clear metadata',
        useForConvert: 'Use for conversion',
        metadataReadyTitle: 'Metadata ready',
        metadataReadyText: 'Use this metadata for conversion without writing bookbind.yaml, or save it as YAML.',
        usingInlineMetadata: 'Using selected metadata for conversion',
        selectedMetadata: 'Selected metadata',
        preparingMetadata: 'Collecting metadata',
        reviewBeforeConvert: 'Review metadata before conversion',
        missingMetadataHint: 'Bookbind filled what it could. Add or adjust missing fields before conversion starts.',
        startConversion: 'Start conversion',
        authorsInput: 'Authors',
        narrator: 'Narrator',
        translator: 'Translator',
        selectCandidatePreview: 'Select a candidate to preview metadata.',
        input: 'Input',
        outputPlaceholder: 'Optional, defaults next to input',
        chapterInterval: 'Chapter interval',
        chapterEvery: 'Chapter every',
        chapterIntervalPlaceholder: 'Optional, for example 10m',
        overwriteOutput: 'Overwrite output',
        working: 'Working',
        dryRun: 'Dry run',
        converting: 'Converting',
        convert: 'Convert',
        done: 'Done',
        progress: 'Progress',
        elapsedAudio: 'Elapsed audio',
        ffmpegRunning: 'ffmpeg is running',
        conversionPlanPlaceholder: 'Conversion plan and progress events will appear here.',
        preparingConversion: 'Preparing conversion...',
        conversionFinished: 'Conversion finished.',
        cancellationRequested: 'Cancellation requested...',
        cancel: 'Cancel',
        confirmCancel: 'Cancel current conversion?',
        mode: 'Mode',
        command: 'Command',
        status: 'Status',
        cachePath: 'Cache path',
        defaultCache: 'Default bookbind cache',
        browse: 'Browse',
        refresh: 'Refresh',
        clean: 'Clean',
        entries: 'entries',
        type: 'Type',
        name: 'Name',
        size: 'Size',
        refreshCacheEmpty: 'Refresh cache to see local provider responses.',
        cacheEmpty: 'Cache is empty.',
        inputRequired: 'Input path is required.',
        titleOrAuthorRequired: 'Title or author is required.',
        selectProvider: 'Select at least one provider.',
        selectCandidateFirst: 'Select a candidate first.',
        googleBooksAPIKey: 'Google Books API key',
        googleBooksAPIKeyPlaceholder: 'Optional API key',
        googleBooksAPIKeyHelp: 'Optional. A personal key can make Google Books search more reliable and reduce anonymous rate limits. Create it in Google Cloud Console, enable Books API, then paste the key here. It is saved only on this computer.',
        saved: 'Saved',
        removed: 'Removed',
    },
    ru: {
        language: 'Язык',
        languageEnglish: 'Английский',
        languageRussian: 'Русский',
        version: 'Версия',
        desktopPreview: 'Desktop preview',
        desktopShell: 'Desktop',
        goBridgeConnected: 'Go bridge подключен',
        nav: {
            import: 'Книга',
            metadata: 'Обогащение',
            convert: 'Конвертация',
            cache: 'Кэш',
        },
        workflow: 'Процесс',
        step: 'Шаг',
        continueToMetadata: 'Перейти к метаданным',
        continueToConvert: 'Перейти к конвертации',
        chooseBookFirst: 'Сначала выбери аудиокнигу.',
        metadataStepIntro: 'Найди книгу во внешних источниках, проверь вариант и используй его для конвертации.',
        convertStepIntro: 'Проверь выбранную книгу и запусти финальную конвертацию M4B.',
        importSource: 'Источник',
        inputPath: 'Путь',
        inputPlaceholder: '/путь/к/book.mp3 или /путь/к/папке',
        file: 'Файл',
        folder: 'Папка',
        inspecting: 'Проверка',
        inspect: 'Проверить',
        metadata: 'Метаданные',
        metadataPlaceholder: 'Необязательный bookbind.yaml',
        cover: 'Обложка',
        coverPlaceholder: 'Необязательная JPG или PNG обложка',
        durationUnknown: 'длительность неизвестна',
        files: 'файлов',
        channel: 'кан.',
        chapters: 'глав',
        metadataSearch: 'Поиск метаданных',
        title: 'Название',
        bookTitle: 'Название книги',
        author: 'Автор',
        optionalAuthor: 'Автор, необязательно',
        searching: 'Ищем',
        search: 'Найти',
        noProviders: 'Нет доступных источников',
        provider: 'Источник',
        authors: 'Авторы',
        year: 'Год',
        confidence: 'Совпадение',
        searchingProviders: 'Ищем по источникам...',
        noCandidates: 'Пока нет вариантов.',
        untitled: 'Без названия',
        unknown: 'Неизвестно',
        preview: 'Превью',
        loading: 'Загрузка',
        noCover: 'Нет обложки',
        narrators: 'Чтецы',
        translators: 'Переводчики',
        series: 'Серия',
        genre: 'Жанр',
        output: 'Выходной файл',
        overwriteExistingFile: 'Перезаписать существующий файл',
        saving: 'Сохраняем',
        saveMetadata: 'Сохранить метаданные',
        saveYamlInstead: 'Сохранить YAML',
        close: 'Закрыть',
        clearMetadata: 'Сбросить метаданные',
        useForConvert: 'Использовать в конвертации',
        metadataReadyTitle: 'Метаданные готовы',
        metadataReadyText: 'Можно использовать эти метаданные в конвертации без сохранения bookbind.yaml или сохранить YAML-файл.',
        usingInlineMetadata: 'Используются выбранные метаданные',
        selectedMetadata: 'Выбранные метаданные',
        preparingMetadata: 'Собираем метаданные',
        reviewBeforeConvert: 'Проверь метаданные перед конвертацией',
        missingMetadataHint: 'Bookbind заполнил всё, что смог. Дополни или поправь поля перед стартом конвертации.',
        startConversion: 'Начать конвертацию',
        authorsInput: 'Авторы',
        narrator: 'Чтец',
        translator: 'Переводчик',
        selectCandidatePreview: 'Выбери вариант, чтобы посмотреть метаданные.',
        input: 'Источник',
        outputPlaceholder: 'Необязательно, по умолчанию рядом с источником',
        chapterInterval: 'Интервал глав',
        chapterEvery: 'Интервал глав',
        chapterIntervalPlaceholder: 'Необязательно, например 10m',
        overwriteOutput: 'Перезаписать результат',
        working: 'Работаем',
        dryRun: 'План',
        converting: 'Конвертация',
        convert: 'Конвертировать',
        done: 'Готово',
        progress: 'Прогресс',
        elapsedAudio: 'Обработано аудио',
        ffmpegRunning: 'ffmpeg работает',
        conversionPlanPlaceholder: 'План конвертации и прогресс появятся здесь.',
        preparingConversion: 'Готовим конвертацию...',
        conversionFinished: 'Конвертация завершена.',
        cancellationRequested: 'Запрошена отмена...',
        cancel: 'Отмена',
        confirmCancel: 'Отменить текущую конвертацию?',
        mode: 'Режим',
        command: 'Команда',
        status: 'Статус',
        cachePath: 'Путь к кэшу',
        defaultCache: 'Кэш bookbind по умолчанию',
        browse: 'Выбрать',
        refresh: 'Обновить',
        clean: 'Очистить',
        entries: 'записей',
        type: 'Тип',
        name: 'Имя',
        size: 'Размер',
        refreshCacheEmpty: 'Обнови кэш, чтобы увидеть локальные ответы источников.',
        cacheEmpty: 'Кэш пуст.',
        inputRequired: 'Нужно выбрать путь к источнику.',
        titleOrAuthorRequired: 'Нужно указать название или автора.',
        selectProvider: 'Выбери хотя бы один источник.',
        selectCandidateFirst: 'Сначала выбери вариант.',
        googleBooksAPIKey: 'Google Books API key',
        googleBooksAPIKeyPlaceholder: 'Необязательный ключ',
        googleBooksAPIKeyHelp: 'Необязательно. Личный ключ делает поиск Google Books стабильнее и снижает шанс anonymous rate limit. Создай ключ в Google Cloud Console, включи Books API и вставь его сюда. Ключ хранится только на этом компьютере.',
        saved: 'Сохранено',
        removed: 'Удалено',
    },
} satisfies Record<Locale, Record<string, any>>;

function initialLocale(): Locale {
    return localStorage.getItem('bookbind-locale') === 'ru' ? 'ru' : 'en';
}

function formatEntryCount(count: number, locale: Locale, entryLabel: string) {
    if (locale === 'ru') {
        return `${count} ${entryLabel}`;
    }
    return `${count} ${entryLabel}`;
}

const localeOptions: Array<{ id: Locale; labelKey: 'languageEnglish' | 'languageRussian' }> = [
    {id: 'en', labelKey: 'languageEnglish'},
    {id: 'ru', labelKey: 'languageRussian'},
];

function App() {
    const [activeScreen, setActiveScreen] = useState<Screen>('import');
    const [locale, setLocale] = useState<Locale>(initialLocale);
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
    const [googleBooksAPIKey, setGoogleBooksAPIKey] = useState(() => localStorage.getItem('bookbind-google-books-api-key') || '');
    const [selectedProviders, setSelectedProviders] = useState<string[]>([]);
    const [metadataCandidates, setMetadataCandidates] = useState<MetadataCandidateView[]>([]);
    const [selectedCandidate, setSelectedCandidate] = useState<MetadataCandidateView | null>(null);
    const [metadataPreview, setMetadataPreview] = useState<MetadataPreviewView | null>(null);
    const [conversionMetadata, setConversionMetadata] = useState<BookMetadataView | null>(null);
    const [showMetadataPrompt, setShowMetadataPrompt] = useState(false);
    const [metadataOutputPath, setMetadataOutputPath] = useState('bookbind.yaml');
    const [overwriteMetadata, setOverwriteMetadata] = useState(false);
    const [metadataStatus, setMetadataStatus] = useState('');
    const [metadataError, setMetadataError] = useState('');
    const [isSearchingMetadata, setIsSearchingMetadata] = useState(false);
    const [isPreviewingMetadata, setIsPreviewingMetadata] = useState(false);
    const [isSavingMetadata, setIsSavingMetadata] = useState(false);
    const [outputPath, setOutputPath] = useState('');
    const [chapterEvery, setChapterEvery] = useState('');
    const [overwriteOutput, setOverwriteOutput] = useState(false);
    const [convertResult, setConvertResult] = useState<ConvertView | null>(null);
    const [convertError, setConvertError] = useState('');
    const [isConverting, setIsConverting] = useState(false);
    const [isPreparingConversion, setIsPreparingConversion] = useState(false);
    const [showConversionPreparation, setShowConversionPreparation] = useState(false);
    const [preparedMetadata, setPreparedMetadata] = useState<BookMetadataView | null>(null);
    const [preparedMissing, setPreparedMissing] = useState<string[]>([]);
    const [currentConversionDryRun, setCurrentConversionDryRun] = useState(false);
    const [convertProgress, setConvertProgress] = useState<ConvertProgressEvent | null>(null);
    const [convertProgressLog, setConvertProgressLog] = useState<string[]>([]);
    const [cachePath, setCachePath] = useState('');
    const [cacheResult, setCacheResult] = useState<CacheListView | null>(null);
    const [cacheStatus, setCacheStatus] = useState('');
    const [cacheError, setCacheError] = useState('');
    const [isCacheBusy, setIsCacheBusy] = useState(false);
    const copy = translations[locale];

    function changeLocale(nextLocale: Locale) {
        setLocale(nextLocale);
        localStorage.setItem('bookbind-locale', nextLocale);
    }

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

    useEffect(() => {
        const unsubscribe = EventsOn('convert:progress', (event: ConvertProgressEvent) => {
            setConvertProgress(event);
            if (event.Line) {
                setConvertProgressLog((current) => [...current.slice(-120), event.Line]);
            }
        });
        return unsubscribe;
    }, []);

    function inspectInput() {
        const path = inputPath.trim();
        if (!path) {
            setInspectError(copy.inputRequired);
            return;
        }

        setIsInspecting(true);
        setInspectError('');
        InspectPath(path)
            .then((result) => {
                const inspect = result as InspectView;
                setInspectResult(inspect);
                hydrateMetadataSearch(inspect);
            })
            .catch((error) => {
                setInspectResult(null);
                setInspectError(String(error));
            })
            .finally(() => setIsInspecting(false));
    }

    function hydrateMetadataSearch(inspect: InspectView) {
        const firstFile = inspect.Files[0];
        if (!firstFile) {
            return;
        }
        if (!metadataTitle.trim()) {
            setMetadataTitle(firstFile.Tags?.Album || firstFile.Tags?.Title || '');
        }
        if (!metadataAuthor.trim()) {
            setMetadataAuthor(firstFile.Tags?.Artist || '');
        }
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

    function updateGoogleBooksAPIKey(value: string) {
        setGoogleBooksAPIKey(value);
        if (value.trim()) {
            localStorage.setItem('bookbind-google-books-api-key', value.trim());
        } else {
            localStorage.removeItem('bookbind-google-books-api-key');
        }
    }

    function searchMetadata() {
        if (!metadataTitle.trim() && !metadataAuthor.trim()) {
            setMetadataError(copy.titleOrAuthorRequired);
            return;
        }
        if (providers.length > 0 && selectedProviders.length === 0) {
            setMetadataError(copy.selectProvider);
            return;
        }

        setIsSearchingMetadata(true);
        setMetadataError('');
        setMetadataStatus('');
        setSelectedCandidate(null);
        setMetadataPreview(null);
        SearchMetadata(metadataTitle, metadataAuthor, selectedProviders, googleBooksAPIKey.trim())
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
        PreviewMetadata(candidate.Provider, candidate.ID, googleBooksAPIKey.trim())
            .then((result) => {
                setMetadataPreview(result as MetadataPreviewView);
                setShowMetadataPrompt(true);
            })
            .catch((error) => setMetadataError(String(error)))
            .finally(() => setIsPreviewingMetadata(false));
    }

    function useMetadataForConvert(book: BookMetadataView) {
        setConversionMetadata(book);
        setMetadataPath('');
        setShowMetadataPrompt(false);
        setMetadataStatus(copy.usingInlineMetadata);
        setActiveScreen('convert');
    }

    function saveMetadata() {
        if (!selectedCandidate) {
            setMetadataError(copy.selectCandidateFirst);
            return;
        }

        setIsSavingMetadata(true);
        setMetadataError('');
        setMetadataStatus('');
        ResolveMetadata(selectedCandidate.Provider, selectedCandidate.ID, metadataOutputPath, overwriteMetadata, googleBooksAPIKey.trim())
            .then((result) => {
                setMetadataStatus(`${copy.saved} ${result.OutputPath}`);
                setMetadataPath(result.OutputPath);
                setShowMetadataPrompt(false);
            })
            .catch((error) => setMetadataError(String(error)))
            .finally(() => setIsSavingMetadata(false));
    }

    function convertAudio(dryRun: boolean) {
        if (!inputPath.trim()) {
            setConvertError(copy.inputRequired);
            return;
        }

        if (!dryRun) {
            setIsPreparingConversion(true);
            setConvertError('');
            const inlineBook = conversionMetadata || emptyBookMetadata();
            PrepareConversion(inputPath, metadataPath, inlineBook)
                .then((result) => {
                    const preparation = result as ConversionPreparationView;
                    setPreparedMetadata(preparation.Book);
                    setPreparedMissing(preparation.Missing || []);
                    setShowConversionPreparation(true);
                })
                .catch((error) => setConvertError(String(error)))
                .finally(() => setIsPreparingConversion(false));
            return;
        }

        executeConversion(dryRun, conversionMetadata);
    }

    function executeConversion(dryRun: boolean, metadataOverride: BookMetadataView | null) {
        setIsConverting(true);
        setCurrentConversionDryRun(dryRun);
        setConvertError('');
        setConvertResult(null);
        setConvertProgress(dryRun ? null : {Phase: 'preparing', Line: copy.preparingConversion, Percent: 0, Elapsed: '', Total: ''});
        setConvertProgressLog(dryRun ? [] : [copy.preparingConversion]);
        const convertAction = metadataOverride
            ? ConvertAudioWithMetadata(inputPath, outputPath, metadataOverride, coverPath, chapterEvery, dryRun, overwriteOutput)
            : ConvertAudio(inputPath, outputPath, metadataPath, coverPath, chapterEvery, dryRun, overwriteOutput);
        convertAction
            .then((result) => {
                setConvertResult(result as ConvertView);
                const summary = convertSummary(result as ConvertView);
                if (!dryRun) {
                    setConvertProgress({Phase: 'done', Line: copy.conversionFinished, Percent: 100, Elapsed: '', Total: ''});
                    setConvertProgressLog((current) => [...current.slice(-120), copy.conversionFinished, '', ...summary]);
                } else {
                    setConvertProgressLog(summary);
                }
            })
            .catch((error) => {
                const message = String(error);
                setConvertError(message);
                setConvertProgressLog((current) => [...current.slice(-120), message]);
            })
            .finally(() => setIsConverting(false));
    }

    function confirmPreparedConversion() {
        if (!preparedMetadata) {
            return;
        }
        setConversionMetadata(preparedMetadata);
        setShowConversionPreparation(false);
        executeConversion(false, preparedMetadata);
    }

    function updatePreparedMetadata(patch: Partial<BookMetadataView>) {
        setPreparedMetadata((current) => current ? {...current, ...patch} : current);
    }

    function cancelConvert() {
        if (!window.confirm(copy.confirmCancel)) {
            return;
        }
        CancelConvert()
            .then((cancelled) => {
                if (cancelled) {
                    setConvertProgressLog((current) => [...current.slice(-120), copy.cancellationRequested]);
                }
            })
            .catch((error) => setConvertError(String(error)));
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
                setCacheStatus(`${copy.removed} ${result.Removed} ${copy.entries} · ${result.RemovedSize}`);
                return ListCache(cachePath);
            })
            .then((result) => setCacheResult(result as CacheListView))
            .catch((error) => setCacheError(String(error)))
            .finally(() => setIsCacheBusy(false));
    }

    function convertSummary(result: ConvertView) {
        return [
            `${copy.input}: ${result.InputPath}`,
            `${copy.files}: ${result.Files?.length || 0}${result.TotalTime ? ` · ${result.TotalTime}` : ''}`,
            result.MetadataPath ? `${copy.metadata}: ${result.MetadataPath}` : '',
            conversionMetadata && !result.MetadataPath ? `${copy.metadata}: ${copy.usingInlineMetadata}` : '',
            result.Title ? `${copy.title}: ${result.Title}` : '',
            result.CoverPath ? `${copy.cover}: ${result.CoverPath}` : '',
            result.ChapterEvery ? `${copy.chapterEvery}: ${result.ChapterEvery}` : '',
            `${copy.output}: ${result.OutputPath}`,
            result.DryRun ? `${copy.mode}: ${copy.dryRun}` : `${copy.mode}: ${copy.convert}`,
            result.Command?.length ? `${copy.command}: ${result.Command.join(' ')}` : '',
            `${copy.status}: ${result.Status}`,
        ].filter(Boolean);
    }

    function canOpenScreen(screen: Screen) {
        if (screen === 'cache' || screen === 'import') {
            return true;
        }
        return !!inputPath.trim();
    }

    function openScreen(screen: Screen) {
        if (canOpenScreen(screen)) {
            setActiveScreen(screen);
        }
    }

    return (
        <main className="app-shell">
            {showMetadataPrompt && metadataPreview && (
                <div className="modal-backdrop" role="presentation">
                    <section aria-modal="true" className="modal" role="dialog">
                        <h3>{copy.metadataReadyTitle}</h3>
                        <p>{copy.metadataReadyText}</p>
                        <div className="modal-summary">
                            <strong>{metadataPreview.Book.Title || copy.untitled}</strong>
                            <span>{metadataPreview.Book.Authors?.join(', ') || metadataPreview.Book.Author || copy.unknown}</span>
                        </div>
                        <div className="modal-actions">
                            <button className="primary-button" onClick={() => useMetadataForConvert(metadataPreview.Book)} type="button">
                                {copy.useForConvert}
                            </button>
                            <button className="secondary-button" onClick={() => setShowMetadataPrompt(false)} type="button">
                                {copy.saveYamlInstead}
                            </button>
                            <button className="secondary-button" onClick={() => setShowMetadataPrompt(false)} type="button">
                                {copy.close}
                            </button>
                        </div>
                    </section>
                </div>
            )}
            {showConversionPreparation && preparedMetadata && (
                <div className="modal-backdrop" role="presentation">
                    <section aria-modal="true" className="modal wide-modal" role="dialog">
                        <h3>{copy.reviewBeforeConvert}</h3>
                        <p>{copy.missingMetadataHint}</p>
                        {preparedMissing.length > 0 && (
                            <div className="missing-strip">
                                {preparedMissing.map((field) => <span key={field}>{field}</span>)}
                            </div>
                        )}
                        <div className="metadata-form">
                            <label>
                                {copy.title}
                                <input
                                    onChange={(event) => updatePreparedMetadata({Title: event.target.value})}
                                    value={preparedMetadata.Title}
                                />
                            </label>
                            <label>
                                {copy.authorsInput}
                                <input
                                    onChange={(event) => updatePreparedMetadata({
                                        Author: event.target.value,
                                        Authors: [],
                                    })}
                                    value={preparedMetadata.Authors?.join(', ') || preparedMetadata.Author}
                                />
                            </label>
                            <label>
                                {copy.narrator}
                                <input
                                    onChange={(event) => updatePreparedMetadata({
                                        Narrator: event.target.value,
                                        Narrators: [],
                                    })}
                                    value={preparedMetadata.Narrators?.join(', ') || preparedMetadata.Narrator}
                                />
                            </label>
                            <label>
                                {copy.translator}
                                <input
                                    onChange={(event) => updatePreparedMetadata({
                                        Translator: event.target.value,
                                        Translators: [],
                                    })}
                                    value={preparedMetadata.Translators?.join(', ') || preparedMetadata.Translator}
                                />
                            </label>
                            <label>
                                {copy.language}
                                <input
                                    onChange={(event) => updatePreparedMetadata({Language: event.target.value})}
                                    value={preparedMetadata.Language}
                                />
                            </label>
                            <label>
                                {copy.genre}
                                <input
                                    onChange={(event) => updatePreparedMetadata({Genre: event.target.value})}
                                    value={preparedMetadata.Genre}
                                />
                            </label>
                            <label>
                                {copy.cover}
                                <input
                                    onChange={(event) => updatePreparedMetadata({Cover: event.target.value})}
                                    value={preparedMetadata.Cover}
                                />
                            </label>
                        </div>
                        <div className="modal-actions">
                            <button className="secondary-button" onClick={() => setShowConversionPreparation(false)} type="button">
                                {copy.cancel}
                            </button>
                            <button className="primary-button" onClick={confirmPreparedConversion} type="button">
                                {copy.startConversion}
                            </button>
                        </div>
                    </section>
                </div>
            )}
            <aside className="sidebar">
                <div className="brand">
                    <span className="brand-mark">B</span>
                    <div>
                        <h1>bookbind</h1>
                        <p>{version ? `${copy.version} ${version}` : copy.desktopPreview}</p>
                    </div>
                </div>

                <nav className="nav-list workflow-list" aria-label="Primary">
                    {workflowScreens.map((screen, index) => {
                        const locked = !canOpenScreen(screen);
                        const completed = screen === 'import' ? !!inspectResult : screen === 'metadata' ? !!conversionMetadata || !!metadataPath.trim() : !!convertResult;
                        return (
                        <button
                            className={[
                                'nav-item',
                                'workflow-item',
                                screen === activeScreen ? 'active' : '',
                                completed ? 'completed' : '',
                                locked ? 'locked' : '',
                            ].filter(Boolean).join(' ')}
                            disabled={locked}
                            key={screen}
                            onClick={() => openScreen(screen)}
                            type="button"
                        >
                            <span className="step-number">{index + 1}</span>
                            <span>
                                <strong>{copy.nav[screen]}</strong>
                                <small>{locked ? copy.chooseBookFirst : `${copy.step} ${index + 1}`}</small>
                            </span>
                        </button>
                        );
                    })}
                    <button
                        className={activeScreen === 'cache' ? 'nav-item active utility-item' : 'nav-item utility-item'}
                        onClick={() => openScreen('cache')}
                        type="button"
                    >
                        {copy.nav.cache}
                    </button>
                </nav>

                <label className="language-picker">
                    {copy.language}
                    <select onChange={(event) => changeLocale(event.target.value as Locale)} value={locale}>
                        {localeOptions.map((option) => (
                            <option key={option.id} value={option.id}>
                                {copy[option.labelKey]}
                            </option>
                        ))}
                    </select>
                </label>
            </aside>

            <section className="workspace">
                <header className="workspace-header">
                    <div>
                        <p className="eyebrow">{activeScreen === 'cache' ? copy.desktopShell : copy.workflow}</p>
                        <h2>{copy.nav[activeScreen]}</h2>
                    </div>
                    <span className="status-pill">{copy.goBridgeConnected}</span>
                </header>

                {activeScreen === 'import' && (
                    <section className="panel">
                        <h3>{copy.importSource}</h3>
                        <div className="inspect-row">
                            <label className="path-field">
                                {copy.inputPath}
                                <input
                                    onChange={(event) => setInputPath(event.target.value)}
                                    placeholder={copy.inputPlaceholder}
                                    value={inputPath}
                                />
                            </label>
                            <div className="button-group">
                                <button className="secondary-button" onClick={() => selectPath(SelectAudioFile, setInputPath)} type="button">
                                    {copy.file}
                                </button>
                                <button className="secondary-button" onClick={() => selectPath(SelectAudioDirectory, setInputPath)} type="button">
                                    {copy.folder}
                                </button>
                            </div>
                            <button className="primary-button" disabled={isInspecting} onClick={inspectInput} type="button">
                                {isInspecting ? copy.inspecting : copy.inspect}
                            </button>
                        </div>

                        <div className="form-grid">
                            <label>
                                {copy.metadata}
                                <div className="field-with-button">
                                    <input
                                        onChange={(event) => setMetadataPath(event.target.value)}
                                        placeholder={copy.metadataPlaceholder}
                                        value={metadataPath}
                                    />
                                    <button className="secondary-button" onClick={() => selectPath(SelectMetadataFile, setMetadataPath)} type="button">
                                        {copy.browse}
                                    </button>
                                </div>
                            </label>
                            <label>
                                {copy.cover}
                                <div className="field-with-button">
                                    <input
                                        onChange={(event) => setCoverPath(event.target.value)}
                                        placeholder={copy.coverPlaceholder}
                                        value={coverPath}
                                    />
                                    <button className="secondary-button" onClick={() => selectPath(SelectCoverFile, setCoverPath)} type="button">
                                        {copy.browse}
                                    </button>
                                </div>
                            </label>
                        </div>

                        {inspectError && <div className="error-box">{inspectError}</div>}

                        {inspectResult && (
                            <div className="inspect-result">
                                <div className="summary-row">
                                    <span>{inspectResult.Path}</span>
                                    <strong>{inspectResult.Files.length} {copy.files} · {inspectResult.TotalTime || copy.durationUnknown}</strong>
                                </div>
                                <div className="file-list">
                                    {inspectResult.Files.map((file) => (
                                        <div className="file-row" key={file.Path}>
                                            <div>
                                                <strong>{file.Name || file.Path}</strong>
                                                <p>
                                                    {[file.Duration, file.Codec, file.Bitrate, file.Channels ? `${file.Channels} ${copy.channel}` : '']
                                                        .filter(Boolean)
                                                        .join(' · ')}
                                                </p>
                                            </div>
                                            <div className="tag-stack">
                                                {file.Tags?.Title && <span>{file.Tags.Title}</span>}
                                                {file.Tags?.Artist && <span>{file.Tags.Artist}</span>}
                                                {file.Chapters > 0 && <span>{file.Chapters} {copy.chapters}</span>}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}
                        {inspectResult && (
                            <div className="continue-row">
                                <button className="primary-button" onClick={() => openScreen('metadata')} type="button">
                                    {copy.continueToMetadata}
                                </button>
                            </div>
                        )}
                    </section>
                )}

                {activeScreen === 'metadata' && (
                    <section className="panel">
                        <h3>{copy.metadataSearch}</h3>
                        <p className="step-intro">{copy.metadataStepIntro}</p>
                        <div className="metadata-search">
                            <label>
                                {copy.title}
                                <input
                                    onChange={(event) => setMetadataTitle(event.target.value)}
                                    placeholder={copy.bookTitle}
                                    value={metadataTitle}
                                />
                            </label>
                            <label>
                                {copy.author}
                                <input
                                    onChange={(event) => setMetadataAuthor(event.target.value)}
                                    placeholder={copy.optionalAuthor}
                                    value={metadataAuthor}
                                />
                            </label>
                            <button className="primary-button" disabled={isSearchingMetadata} onClick={searchMetadata} type="button">
                                {isSearchingMetadata ? copy.searching : copy.search}
                            </button>
                        </div>

                        <div className="provider-strip" aria-label={copy.provider}>
                            {providers.length === 0 && <span>{copy.noProviders}</span>}
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
                        <label className="api-key-field">
                            {copy.googleBooksAPIKey}
                            <input
                                onChange={(event) => updateGoogleBooksAPIKey(event.target.value)}
                                placeholder={copy.googleBooksAPIKeyPlaceholder}
                                type="password"
                                value={googleBooksAPIKey}
                            />
                            <span>{copy.googleBooksAPIKeyHelp}</span>
                        </label>

                        {metadataError && <div className="error-box">{metadataError}</div>}
                        {metadataStatus && <div className="success-box">{metadataStatus}</div>}

                        <div className="metadata-layout">
                            <div className="table-shell">
                                <div className="metadata-table-header">
                                    <span>{copy.provider}</span>
                                    <span>{copy.title}</span>
                                    <span>{copy.authors}</span>
                                    <span>{copy.year}</span>
                                    <span>{copy.confidence}</span>
                                </div>
                                {metadataCandidates.length === 0 && (
                                    <div className="table-empty">
                                        {isSearchingMetadata ? copy.searchingProviders : copy.noCandidates}
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
                                            <strong>{candidate.Title || copy.untitled}</strong>
                                            <span>{candidate.Authors?.join(', ') || copy.unknown}</span>
                                            <span>{candidate.Year || ''}</span>
                                            <span>{candidate.Confidence || ''}</span>
                                        </button>
                                    );
                                })}
                            </div>

                            <aside className="preview-pane">
                                <div className="preview-heading">
                                    <h4>{copy.preview}</h4>
                                    {isPreviewingMetadata && <span>{copy.loading}</span>}
                                </div>
                                {metadataPreview ? (
                                    <div className="preview-content">
                                        <div className="cover-preview">
                                            {metadataPreview.Book.Cover ? (
                                                <img alt="" src={metadataPreview.Book.Cover} />
                                            ) : (
                                                <span>{copy.noCover}</span>
                                            )}
                                        </div>
                                        <dl>
                                            <dt>{copy.title}</dt>
                                            <dd>{metadataPreview.Book.Title || '-'}</dd>
                                            <dt>{copy.authors}</dt>
                                            <dd>{metadataPreview.Book.Authors?.join(', ') || metadataPreview.Book.Author || '-'}</dd>
                                            <dt>{copy.narrators}</dt>
                                            <dd>{metadataPreview.Book.Narrators?.join(', ') || metadataPreview.Book.Narrator || '-'}</dd>
                                            <dt>{copy.translators}</dt>
                                            <dd>{metadataPreview.Book.Translators?.join(', ') || metadataPreview.Book.Translator || '-'}</dd>
                                            <dt>{copy.series}</dt>
                                            <dd>
                                                {[metadataPreview.Book.Series, metadataPreview.Book.SeriesIndex].filter(Boolean).join(' #') || '-'}
                                            </dd>
                                            <dt>{copy.year}</dt>
                                            <dd>{metadataPreview.Book.PublishedYear || '-'}</dd>
                                        </dl>
                                        <label>
                                            {copy.output}
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
                                            {copy.overwriteExistingFile}
                                        </label>
                                        <button className="primary-button" disabled={isSavingMetadata} onClick={saveMetadata} type="button">
                                            {isSavingMetadata ? copy.saving : copy.saveMetadata}
                                        </button>
                                        <button className="secondary-button" onClick={() => useMetadataForConvert(metadataPreview.Book)} type="button">
                                            {copy.useForConvert}
                                        </button>
                                    </div>
                                ) : (
                                    <div className="table-empty">{copy.selectCandidatePreview}</div>
                                )}
                            </aside>
                        </div>
                        {(conversionMetadata || metadataPath.trim()) && (
                            <div className="continue-row">
                                <button className="primary-button" onClick={() => openScreen('convert')} type="button">
                                    {copy.continueToConvert}
                                </button>
                            </div>
                        )}
                    </section>
                )}

                {activeScreen === 'convert' && (
                    <section className="panel">
                        <h3>{copy.convert}</h3>
                        <p className="step-intro">{copy.convertStepIntro}</p>
                        <div className="convert-grid">
                            <label>
                                {copy.input}
                                <input
                                    onChange={(event) => setInputPath(event.target.value)}
                                    placeholder={copy.inputPlaceholder}
                                    value={inputPath}
                                />
                            </label>
                            <label>
                                {copy.output}
                                <div className="field-with-button">
                                    <input
                                        onChange={(event) => setOutputPath(event.target.value)}
                                        placeholder={copy.outputPlaceholder}
                                        value={outputPath}
                                    />
                                    <button className="secondary-button" onClick={() => selectPath(SelectOutputFile, setOutputPath)} type="button">
                                        {copy.browse}
                                    </button>
                                </div>
                            </label>
                            <label>
                                {copy.chapterInterval}
                                <input
                                    onChange={(event) => setChapterEvery(event.target.value)}
                                    placeholder={copy.chapterIntervalPlaceholder}
                                    value={chapterEvery}
                                />
                            </label>
                        </div>
                        <div className="convert-grid secondary">
                            <label>
                                {copy.metadata}
                                <input
                                    disabled={!!conversionMetadata}
                                    onChange={(event) => {
                                        setMetadataPath(event.target.value);
                                        if (event.target.value.trim()) {
                                            setConversionMetadata(null);
                                        }
                                    }}
                                    placeholder={copy.metadataPlaceholder}
                                    value={metadataPath}
                                />
                            </label>
                            <label>
                                {copy.cover}
                                <input
                                    onChange={(event) => setCoverPath(event.target.value)}
                                    placeholder={copy.coverPlaceholder}
                                    value={coverPath}
                                />
                            </label>
                            <label className="checkbox-line convert-checkbox">
                                <input
                                    checked={overwriteOutput}
                                    onChange={(event) => setOverwriteOutput(event.target.checked)}
                                    type="checkbox"
                                />
                                {copy.overwriteOutput}
                            </label>
                        </div>
                        {conversionMetadata && (
                            <div className="selected-metadata">
                                <div>
                                    <strong>{copy.selectedMetadata}</strong>
                                    <span>{conversionMetadata.Title || copy.untitled}</span>
                                </div>
                                <button className="secondary-button" onClick={() => setConversionMetadata(null)} type="button">
                                    {copy.clearMetadata}
                                </button>
                            </div>
                        )}
                        <div className="action-row">
                            <button className="secondary-button" disabled={isConverting || isPreparingConversion} onClick={() => convertAudio(true)} type="button">
                                {isConverting ? copy.working : copy.dryRun}
                            </button>
                            {isConverting && !currentConversionDryRun && (
                                <button className="secondary-button danger-secondary" onClick={cancelConvert} type="button">
                                    {copy.cancel}
                                </button>
                            )}
                            <button className="primary-button" disabled={isConverting || isPreparingConversion} onClick={() => convertAudio(false)} type="button">
                                {isPreparingConversion ? copy.preparingMetadata : isConverting ? copy.working : copy.convert}
                            </button>
                        </div>
                        {convertError && <div className="error-box">{convertError}</div>}
                        {(isConverting || convertProgress || convertProgressLog.length > 0 || convertResult) && (
                            <div className="progress-panel">
                                <div className="progress-header">
                                    <strong>{convertProgress?.Phase === 'done' ? copy.done : isConverting ? copy.converting : copy.progress}</strong>
                                    <span>
                                        {convertProgress?.Elapsed
                                            ? `${copy.elapsedAudio}: ${convertProgress.Elapsed}${convertProgress.Total ? ` / ${convertProgress.Total}` : ''}`
                                            : isConverting ? copy.ffmpegRunning : ''}
                                    </span>
                                </div>
                                <div className="progress-track">
                                    <span
                                        style={{width: `${convertProgress?.Phase === 'done' ? 100 : Math.max(convertProgress?.Percent || 0, 8)}%`}}
                                    />
                                </div>
                                <pre className="progress-log">
                                    {convertProgressLog.length > 0 ? convertProgressLog.join('\n') : copy.conversionPlanPlaceholder}
                                </pre>
                            </div>
                        )}
                    </section>
                )}

                {activeScreen === 'cache' && (
                    <section className="panel">
                        <h3>{copy.nav.cache}</h3>
                        <div className="cache-toolbar">
                            <label className="path-field">
                                {copy.cachePath}
                                <div className="field-with-button">
                                    <input
                                        onChange={(event) => setCachePath(event.target.value)}
                                        placeholder={copy.defaultCache}
                                        value={cachePath}
                                    />
                                    <button className="secondary-button" onClick={() => selectPath(SelectCacheDirectory, setCachePath)} type="button">
                                        {copy.browse}
                                    </button>
                                </div>
                            </label>
                            <button className="secondary-button" disabled={isCacheBusy} onClick={refreshCache} type="button">
                                {isCacheBusy ? copy.working : copy.refresh}
                            </button>
                            <button
                                className="primary-button danger"
                                disabled={isCacheBusy || !cacheResult || cacheResult.Entries.length === 0}
                                onClick={cleanCache}
                                type="button"
                            >
                                {copy.clean}
                            </button>
                        </div>
                        {cacheError && <div className="error-box">{cacheError}</div>}
                        {cacheStatus && <div className="success-box">{cacheStatus}</div>}
                        {cacheResult && (
                            <div className="summary-row">
                                <span>{cacheResult.Path}</span>
                                <strong>{formatEntryCount(cacheResult.Entries.length, locale, copy.entries)} · {cacheResult.Size}</strong>
                            </div>
                        )}
                        <div className="table-shell">
                            <div className="cache-table-header">
                                <span>{copy.type}</span>
                                <span>{copy.name}</span>
                                <span>{copy.size}</span>
                            </div>
                            {!cacheResult && <div className="table-empty">{copy.refreshCacheEmpty}</div>}
                            {cacheResult && cacheResult.Entries.length === 0 && <div className="table-empty">{copy.cacheEmpty}</div>}
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

function emptyBookMetadata(): BookMetadataView {
    return {
        Title: '',
        Subtitle: '',
        Authors: [],
        Author: '',
        Narrators: [],
        Narrator: '',
        Translators: [],
        Translator: '',
        Series: '',
        SeriesIndex: '',
        Language: '',
        Genre: '',
        Description: '',
        Publisher: '',
        PublishedYear: 0,
        Cover: '',
    };
}

export default App;
