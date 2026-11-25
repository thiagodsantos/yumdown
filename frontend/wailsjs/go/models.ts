export namespace downloader {
	
	export class Capabilities {
	    toolsAvailable: boolean;
	    ytDlpVersion?: string;
	    ffmpegVersion?: string;
	    ffprobeFound: boolean;
	    toolsDirectory?: string;
	    formats: string[];
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new Capabilities(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.toolsAvailable = source["toolsAvailable"];
	        this.ytDlpVersion = source["ytDlpVersion"];
	        this.ffmpegVersion = source["ffmpegVersion"];
	        this.ffprobeFound = source["ffprobeFound"];
	        this.toolsDirectory = source["toolsDirectory"];
	        this.formats = source["formats"];
	        this.message = source["message"];
	    }
	}
	export class VideoRequest {
	    url: string;
	    filename: string;
	
	    static createFrom(source: any = {}) {
	        return new VideoRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.filename = source["filename"];
	    }
	}
	export class DownloadRequest {
	    videos: VideoRequest[];
	    outputDir: string;
	    format: string;
	    audioQuality: string;
	
	    static createFrom(source: any = {}) {
	        return new DownloadRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.videos = this.convertValues(source["videos"], VideoRequest);
	        this.outputDir = source["outputDir"];
	        this.format = source["format"];
	        this.audioQuality = source["audioQuality"];
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
	export class HistoryEntry {
	    id: string;
	    url: string;
	    outputDir: string;
	    outputPath?: string;
	    format: string;
	    status: string;
	    outputSize?: number;
	    error?: string;
	    finishedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new HistoryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.url = source["url"];
	        this.outputDir = source["outputDir"];
	        this.outputPath = source["outputPath"];
	        this.format = source["format"];
	        this.status = source["status"];
	        this.outputSize = source["outputSize"];
	        this.error = source["error"];
	        this.finishedAt = source["finishedAt"];
	    }
	}
	export class Progress {
	    percent?: number;
	    downloadedBytes?: number;
	    totalBytes?: number;
	    speedBytesPerSecond?: number;
	    etaSeconds?: number;
	
	    static createFrom(source: any = {}) {
	        return new Progress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.percent = source["percent"];
	        this.downloadedBytes = source["downloadedBytes"];
	        this.totalBytes = source["totalBytes"];
	        this.speedBytesPerSecond = source["speedBytesPerSecond"];
	        this.etaSeconds = source["etaSeconds"];
	    }
	}
	export class Job {
	    id: string;
	    url: string;
	    outputDir: string;
	    outputName: string;
	    format: string;
	    audioQuality: string;
	    status: string;
	    progress: Progress;
	    outputPath?: string;
	    outputSize?: number;
	    error?: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Job(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.url = source["url"];
	        this.outputDir = source["outputDir"];
	        this.outputName = source["outputName"];
	        this.format = source["format"];
	        this.audioQuality = source["audioQuality"];
	        this.status = source["status"];
	        this.progress = this.convertValues(source["progress"], Progress);
	        this.outputPath = source["outputPath"];
	        this.outputSize = source["outputSize"];
	        this.error = source["error"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
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
	
	export class Settings {
	    outputDir: string;
	    format: string;
	    concurrency: number;
	    audioQuality: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.outputDir = source["outputDir"];
	        this.format = source["format"];
	        this.concurrency = source["concurrency"];
	        this.audioQuality = source["audioQuality"];
	    }
	}

}

