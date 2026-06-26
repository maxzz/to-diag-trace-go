export namespace backend {
	
	export class TraceStartOptions {
	    accumulateTraces: boolean;
	    enableOtsTrace: boolean;
	    verbosity: number;
	
	    static createFrom(source: any = {}) {
	        return new TraceStartOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.accumulateTraces = source["accumulateTraces"];
	        this.enableOtsTrace = source["enableOtsTrace"];
	        this.verbosity = source["verbosity"];
	    }
	}

}

export namespace diag {
	
	export class TraceSettings {
	    isTracing: boolean;
	    accumulateTraces: boolean;
	    enableOtsTrace: boolean;
	    verbosity: number;
	    tracePath: string;
	    passwordManagerFound: boolean;
	    is64BitWindows: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TraceSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isTracing = source["isTracing"];
	        this.accumulateTraces = source["accumulateTraces"];
	        this.enableOtsTrace = source["enableOtsTrace"];
	        this.verbosity = source["verbosity"];
	        this.tracePath = source["tracePath"];
	        this.passwordManagerFound = source["passwordManagerFound"];
	        this.is64BitWindows = source["is64BitWindows"];
	    }
	}

}

