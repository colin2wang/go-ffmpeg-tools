export namespace ffmpeg {
	
	export class SubStream {
	    index: number;
	    relIndex: number;
	    codecName: string;
	    language: string;
	    title: string;
	
	    static createFrom(source: any = {}) {
	        return new SubStream(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.relIndex = source["relIndex"];
	        this.codecName = source["codecName"];
	        this.language = source["language"];
	        this.title = source["title"];
	    }
	}
	export class VersionInfo {
	    version: string;
	    firstLine: string;
	    builtWith: string;
	    config: string;
	    path: string;
	    raw: string;
	
	    static createFrom(source: any = {}) {
	        return new VersionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.firstLine = source["firstLine"];
	        this.builtWith = source["builtWith"];
	        this.config = source["config"];
	        this.path = source["path"];
	        this.raw = source["raw"];
	    }
	}

}

export namespace frontend {
	
	export class FileFilter {
	    DisplayName: string;
	    Pattern: string;
	
	    static createFrom(source: any = {}) {
	        return new FileFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.DisplayName = source["DisplayName"];
	        this.Pattern = source["Pattern"];
	    }
	}

}

export namespace main {
	
	export class SubmitJobReq {
	    inputVideo: string;
	    outputVideo: string;
	    outputMode: string;
	    outputSuffix: string;
	    subConfig: task.SubConfig;
	
	    static createFrom(source: any = {}) {
	        return new SubmitJobReq(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inputVideo = source["inputVideo"];
	        this.outputVideo = source["outputVideo"];
	        this.outputMode = source["outputMode"];
	        this.outputSuffix = source["outputSuffix"];
	        this.subConfig = this.convertValues(source["subConfig"], task.SubConfig);
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

export namespace task {
	
	export class SubTrack {
	    file: string;
	    lang: string;
	    title: string;
	
	    static createFrom(source: any = {}) {
	        return new SubTrack(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file = source["file"];
	        this.lang = source["lang"];
	        this.title = source["title"];
	    }
	}
	export class SubConfig {
	    mode: string;
	    tracks: SubTrack[];
	    fontName?: string;
	    fontSize?: number;
	    fontFile?: string;
	
	    static createFrom(source: any = {}) {
	        return new SubConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.tracks = this.convertValues(source["tracks"], SubTrack);
	        this.fontName = source["fontName"];
	        this.fontSize = source["fontSize"];
	        this.fontFile = source["fontFile"];
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
	export class Job {
	    jobId: string;
	    inputVideo: string;
	    outputVideo: string;
	    subConfig: SubConfig;
	    totalSeconds: number;
	    currentSeconds: number;
	    progress: number;
	    status: string;
	    log: string[];
	
	    static createFrom(source: any = {}) {
	        return new Job(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.jobId = source["jobId"];
	        this.inputVideo = source["inputVideo"];
	        this.outputVideo = source["outputVideo"];
	        this.subConfig = this.convertValues(source["subConfig"], SubConfig);
	        this.totalSeconds = source["totalSeconds"];
	        this.currentSeconds = source["currentSeconds"];
	        this.progress = source["progress"];
	        this.status = source["status"];
	        this.log = source["log"];
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

