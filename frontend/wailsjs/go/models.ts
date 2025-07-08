export namespace config {
	
	export class ConfigDataMySQL {
	    version: string;
	    gpg?: string;
	    link: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigDataMySQL(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.gpg = source["gpg"];
	        this.link = source["link"];
	    }
	}
	export class ConfigArchInfoMySQL {
	    os: string;
	    data: ConfigDataMySQL[];
	
	    static createFrom(source: any = {}) {
	        return new ConfigArchInfoMySQL(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.os = source["os"];
	        this.data = this.convertValues(source["data"], ConfigDataMySQL);
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
	
	export class ConfigVersionMySQL {
	    mysql: ConfigArchInfoMySQL[];
	
	    static createFrom(source: any = {}) {
	        return new ConfigVersionMySQL(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mysql = this.convertValues(source["mysql"], ConfigArchInfoMySQL);
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

