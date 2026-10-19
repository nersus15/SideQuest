export namespace main {
	
	export class CompletionPayload {
	    quest: questbook.Quest;
	    xp: number;
	    level: number;
	    badges: string[];
	    progress: questbook.PlayerProgress;
	
	    static createFrom(source: any = {}) {
	        return new CompletionPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.quest = this.convertValues(source["quest"], questbook.Quest);
	        this.xp = source["xp"];
	        this.level = source["level"];
	        this.badges = source["badges"];
	        this.progress = this.convertValues(source["progress"], questbook.PlayerProgress);
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

export namespace questbook {
	
	export class PlayerProgress {
	    xp: number;
	    level: number;
	    completed: Record<string, time.Time>;
	    badges: Record<string, boolean>;
	    // Go type: time
	    lastUpdated: any;
	
	    static createFrom(source: any = {}) {
	        return new PlayerProgress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.xp = source["xp"];
	        this.level = source["level"];
	        this.completed = this.convertValues(source["completed"], time.Time, true);
	        this.badges = source["badges"];
	        this.lastUpdated = this.convertValues(source["lastUpdated"], null);
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
	export class Quest {
	    id: string;
	    title: string;
	    description: string;
	    category: string;
	    durationMinutes: number;
	    difficulty: number;
	    context?: string;
	    xp: number;
	    repeatable: boolean;
	    cooldownMinutes?: number;
	    source: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    lastCompleted?: any;
	
	    static createFrom(source: any = {}) {
	        return new Quest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.durationMinutes = source["durationMinutes"];
	        this.difficulty = source["difficulty"];
	        this.context = source["context"];
	        this.xp = source["xp"];
	        this.repeatable = source["repeatable"];
	        this.cooldownMinutes = source["cooldownMinutes"];
	        this.source = source["source"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.lastCompleted = this.convertValues(source["lastCompleted"], null);
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

