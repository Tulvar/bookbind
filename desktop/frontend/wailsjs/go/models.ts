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

}

