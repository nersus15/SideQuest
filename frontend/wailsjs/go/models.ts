export namespace custom {
	
	export class Draft {
	    id: string;
	    title: string;
	    description: string;
	    category: string;
	    minutes: number;
	    difficulty: number;
	
	    static createFrom(source: any = {}) {
	        return new Draft(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.minutes = source["minutes"];
	        this.difficulty = source["difficulty"];
	    }
	}

}

export namespace game {
	
	export class BadgeView {
	    id: string;
	    name: string;
	    emoji: string;
	    hint: string;
	    earned: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BadgeView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.emoji = source["emoji"];
	        this.hint = source["hint"];
	        this.earned = source["earned"];
	    }
	}
	export class LogEntry {
	    questId: string;
	    title: string;
	    category: string;
	    xp: number;
	    // Go type: time
	    at: any;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.questId = source["questId"];
	        this.title = source["title"];
	        this.category = source["category"];
	        this.xp = source["xp"];
	        this.at = this.convertValues(source["at"], null);
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
	export class View {
	    quest?: quest.Quest;
	    twist?: quest.Twist;
	    twistOn: boolean;
	    accepted: boolean;
	    reward: number;
	    xp: number;
	    level: progress.Level;
	    completed: number;
	    badges: BadgeView[];
	    prefs: quest.Prefs;
	    recent: LogEntry[];
	
	    static createFrom(source: any = {}) {
	        return new View(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.quest = this.convertValues(source["quest"], quest.Quest);
	        this.twist = this.convertValues(source["twist"], quest.Twist);
	        this.twistOn = source["twistOn"];
	        this.accepted = source["accepted"];
	        this.reward = source["reward"];
	        this.xp = source["xp"];
	        this.level = this.convertValues(source["level"], progress.Level);
	        this.completed = source["completed"];
	        this.badges = this.convertValues(source["badges"], BadgeView);
	        this.prefs = this.convertValues(source["prefs"], quest.Prefs);
	        this.recent = this.convertValues(source["recent"], LogEntry);
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
	export class Completion {
	    view: View;
	    gained: number;
	    leveledUp: boolean;
	    newBadges: progress.Badge[];
	
	    static createFrom(source: any = {}) {
	        return new Completion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.view = this.convertValues(source["view"], View);
	        this.gained = source["gained"];
	        this.leveledUp = source["leveledUp"];
	        this.newBadges = this.convertValues(source["newBadges"], progress.Badge);
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

export namespace progress {
	
	export class Badge {
	    id: string;
	    name: string;
	    emoji: string;
	    hint: string;
	
	    static createFrom(source: any = {}) {
	        return new Badge(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.emoji = source["emoji"];
	        this.hint = source["hint"];
	    }
	}
	export class Level {
	    number: number;
	    title: string;
	    xpInLevel: number;
	    xpToNext: number;
	
	    static createFrom(source: any = {}) {
	        return new Level(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.number = source["number"];
	        this.title = source["title"];
	        this.xpInLevel = source["xpInLevel"];
	        this.xpToNext = source["xpToNext"];
	    }
	}

}

export namespace quest {
	
	export class Prefs {
	    muted: string[];
	    avoid: string[];
	
	    static createFrom(source: any = {}) {
	        return new Prefs(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.muted = source["muted"];
	        this.avoid = source["avoid"];
	    }
	}
	export class Twist {
	    id: string;
	    text: string;
	    bonus?: number;
	
	    static createFrom(source: any = {}) {
	        return new Twist(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.text = source["text"];
	        this.bonus = source["bonus"];
	    }
	}
	export class Quest {
	    id: string;
	    title: string;
	    description: string;
	    category: string;
	    minutes: number;
	    difficulty: number;
	    needs?: string[];
	    xp: number;
	    repeatable: boolean;
	    cooldownHours?: number;
	    twists?: Twist[];
	    source: string;
	    pack: string;
	
	    static createFrom(source: any = {}) {
	        return new Quest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.minutes = source["minutes"];
	        this.difficulty = source["difficulty"];
	        this.needs = source["needs"];
	        this.xp = source["xp"];
	        this.repeatable = source["repeatable"];
	        this.cooldownHours = source["cooldownHours"];
	        this.twists = this.convertValues(source["twists"], Twist);
	        this.source = source["source"];
	        this.pack = source["pack"];
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

