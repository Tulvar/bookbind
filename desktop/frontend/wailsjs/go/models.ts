export namespace app {

	export class ProviderInfo {
	    Name: string;
	    Enabled: boolean;

	    static createFrom(source: any = {}) {
	        return new ProviderInfo(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Enabled = source["Enabled"];
	    }
	}

}

export namespace audio {

	export class EmbeddedTags {
	    Title: string;
	    Artist: string;
	    Album: string;
	    AlbumArtist: string;
	    Composer: string;
	    Genre: string;
	    Date: string;
	    Comment: string;
	    Language: string;

	    static createFrom(source: any = {}) {
	        return new EmbeddedTags(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Title = source["Title"];
	        this.Artist = source["Artist"];
	        this.Album = source["Album"];
	        this.AlbumArtist = source["AlbumArtist"];
	        this.Composer = source["Composer"];
	        this.Genre = source["Genre"];
	        this.Date = source["Date"];
	        this.Comment = source["Comment"];
	        this.Language = source["Language"];
	    }
	}

}

export namespace main {

	export class BookMetadataView {
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

	    static createFrom(source: any = {}) {
	        return new BookMetadataView(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Title = source["Title"];
	        this.Subtitle = source["Subtitle"];
	        this.Authors = source["Authors"];
	        this.Author = source["Author"];
	        this.Narrators = source["Narrators"];
	        this.Narrator = source["Narrator"];
	        this.Series = source["Series"];
	        this.SeriesIndex = source["SeriesIndex"];
	        this.Language = source["Language"];
	        this.Genre = source["Genre"];
	        this.Description = source["Description"];
	        this.Publisher = source["Publisher"];
	        this.PublishedYear = source["PublishedYear"];
	        this.Cover = source["Cover"];
	    }
	}
	export class CacheCleanView {
	    Path: string;
	    Removed: number;
	    RemovedSize: string;

	    static createFrom(source: any = {}) {
	        return new CacheCleanView(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Removed = source["Removed"];
	        this.RemovedSize = source["RemovedSize"];
	    }
	}
	export class CacheEntryView {
	    Path: string;
	    Name: string;
	    Kind: string;
	    Size: string;

	    static createFrom(source: any = {}) {
	        return new CacheEntryView(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Name = source["Name"];
	        this.Kind = source["Kind"];
	        this.Size = source["Size"];
	    }
	}
	export class CacheListView {
	    Path: string;
	    Entries: CacheEntryView[];
	    Size: string;

	    static createFrom(source: any = {}) {
	        return new CacheListView(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Entries = this.convertValues(source["Entries"], CacheEntryView);
	        this.Size = source["Size"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ConversionPreparationView {
	    Book: BookMetadataView;
	    Missing: string[];
	    Files: number;

	    static createFrom(source: any = {}) {
	        return new ConversionPreparationView(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Book = this.convertValues(source["Book"], BookMetadataView);
	        this.Missing = source["Missing"];
	        this.Files = source["Files"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class InspectFileView {
	    Path: string;
	    Name: string;
	    Duration: string;
	    Codec: string;
	    Bitrate: string;
	    Channels: number;
	    Tags: audio.EmbeddedTags;
	    Chapters: number;

	    static createFrom(source: any = {}) {
	        return new InspectFileView(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Name = source["Name"];
	        this.Duration = source["Duration"];
	        this.Codec = source["Codec"];
	        this.Bitrate = source["Bitrate"];
	        this.Channels = source["Channels"];
	        this.Tags = this.convertValues(source["Tags"], audio.EmbeddedTags);
	        this.Chapters = source["Chapters"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ConvertView {
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

	    static createFrom(source: any = {}) {
	        return new ConvertView(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.InputPath = source["InputPath"];
	        this.Files = this.convertValues(source["Files"], InspectFileView);
	        this.TotalTime = source["TotalTime"];
	        this.MetadataPath = source["MetadataPath"];
	        this.Title = source["Title"];
	        this.CoverPath = source["CoverPath"];
	        this.ChapterEvery = source["ChapterEvery"];
	        this.OutputPath = source["OutputPath"];
	        this.DryRun = source["DryRun"];
	        this.Command = source["Command"];
	        this.Status = source["Status"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

	export class InspectView {
	    Path: string;
	    Files: InspectFileView[];
	    TotalTime: string;

	    static createFrom(source: any = {}) {
	        return new InspectView(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Files = this.convertValues(source["Files"], InspectFileView);
	        this.TotalTime = source["TotalTime"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MetadataCandidateView {
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

	    static createFrom(source: any = {}) {
	        return new MetadataCandidateView(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Provider = source["Provider"];
	        this.ID = source["ID"];
	        this.Title = source["Title"];
	        this.Authors = source["Authors"];
	        this.Narrators = source["Narrators"];
	        this.Series = source["Series"];
	        this.SeriesIndex = source["SeriesIndex"];
	        this.Year = source["Year"];
	        this.Duration = source["Duration"];
	        this.CoverURL = source["CoverURL"];
	        this.Confidence = source["Confidence"];
	    }
	}
	export class MetadataPreviewView {
	    Candidate: MetadataCandidateView;
	    Book: BookMetadataView;

	    static createFrom(source: any = {}) {
	        return new MetadataPreviewView(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Candidate = this.convertValues(source["Candidate"], MetadataCandidateView);
	        this.Book = this.convertValues(source["Book"], BookMetadataView);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MetadataResolveView {
	    OutputPath: string;
	    Candidate: MetadataCandidateView;
	    Book: BookMetadataView;

	    static createFrom(source: any = {}) {
	        return new MetadataResolveView(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.OutputPath = source["OutputPath"];
	        this.Candidate = this.convertValues(source["Candidate"], MetadataCandidateView);
	        this.Book = this.convertValues(source["Book"], BookMetadataView);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MetadataSearchView {
	    Candidates: MetadataCandidateView[];

	    static createFrom(source: any = {}) {
	        return new MetadataSearchView(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Candidates = this.convertValues(source["Candidates"], MetadataCandidateView);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

