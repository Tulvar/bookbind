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

