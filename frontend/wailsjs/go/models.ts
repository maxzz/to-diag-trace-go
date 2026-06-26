export namespace backend {
	
	export class AppLaunchSettings {
	    requireElevationAtLaunch: boolean;
	    isElevated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppLaunchSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requireElevationAtLaunch = source["requireElevationAtLaunch"];
	        this.isElevated = source["isElevated"];
	    }
	}
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
	export class PendingAction {
	    type: string;
	    startOpts?: TraceStartOptions;
	
	    static createFrom(source: any = {}) {
	        return new PendingAction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.startOpts = this.convertValues(source["startOpts"], TraceStartOptions);
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

