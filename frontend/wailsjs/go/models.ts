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

export namespace desktop {
	
	export class Anchor {
	    right: boolean;
	    bottom: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Anchor(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.right = source["right"];
	        this.bottom = source["bottom"];
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
	export class Look {
	    hat: string;
	    extra: string;
	
	    static createFrom(source: any = {}) {
	        return new Look(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hat = source["hat"];
	        this.extra = source["extra"];
	    }
	}
	export class LookView {
	    id: string;
	    kind: string;
	    level: number;
	    unlocked: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LookView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.kind = source["kind"];
	        this.level = source["level"];
	        this.unlocked = source["unlocked"];
	    }
	}
	export class Nudges {
	    breakMinutes: number;
	    stayWhenBusy: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Nudges(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.breakMinutes = source["breakMinutes"];
	        this.stayWhenBusy = source["stayWhenBusy"];
	    }
	}
	export class Sound {
	    volume: number;
	    muted: boolean;
	    nudge: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Sound(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.volume = source["volume"];
	        this.muted = source["muted"];
	        this.nudge = source["nudge"];
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
	    lang: string;
	    onboarded: boolean;
	    look: Look;
	    looks: LookView[];
	    nudges: Nudges;
	    schedule: schedule.Schedule;
	    sound: Sound;
	    maxDifficulty: number;
	    nextDifficultyAt: number;
	
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
	        this.lang = source["lang"];
	        this.onboarded = source["onboarded"];
	        this.look = this.convertValues(source["look"], Look);
	        this.looks = this.convertValues(source["looks"], LookView);
	        this.nudges = this.convertValues(source["nudges"], Nudges);
	        this.schedule = this.convertValues(source["schedule"], schedule.Schedule);
	        this.sound = this.convertValues(source["sound"], Sound);
	        this.maxDifficulty = source["maxDifficulty"];
	        this.nextDifficultyAt = source["nextDifficultyAt"];
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

export namespace journal {
	
	export class Entry {
	    id: string;
	    questId: string;
	    titles: Record<string, string>;
	    category: string;
	    kind: string;
	    xp: number;
	    // Go type: time
	    at: any;
	    answers?: string[];
	    count?: number;
	    doodle?: string;
	    mood?: string;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.questId = source["questId"];
	        this.titles = source["titles"];
	        this.category = source["category"];
	        this.kind = source["kind"];
	        this.xp = source["xp"];
	        this.at = this.convertValues(source["at"], null);
	        this.answers = source["answers"];
	        this.count = source["count"];
	        this.doodle = source["doodle"];
	        this.mood = source["mood"];
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
	export class Input {
	    answers: string[];
	    count: number;
	    doodle: string;
	
	    static createFrom(source: any = {}) {
	        return new Input(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.answers = source["answers"];
	        this.count = source["count"];
	        this.doodle = source["doodle"];
	    }
	}
	export class Insight {
	    category: string;
	    moods: number;
	
	    static createFrom(source: any = {}) {
	        return new Insight(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.moods = source["moods"];
	    }
	}

}

export namespace main {
	
	export class CompleteResult {
	    view: game.View;
	    done: quest.Quest;
	    rating: string;
	    gained: number;
	    leveledUp: boolean;
	    newBadges: progress.Badge[];
	    newLooks: progress.Look[];
	    newDifficulty: number;
	    entry: string;
	
	    static createFrom(source: any = {}) {
	        return new CompleteResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.view = this.convertValues(source["view"], game.View);
	        this.done = this.convertValues(source["done"], quest.Quest);
	        this.rating = source["rating"];
	        this.gained = source["gained"];
	        this.leveledUp = source["leveledUp"];
	        this.newBadges = this.convertValues(source["newBadges"], progress.Badge);
	        this.newLooks = this.convertValues(source["newLooks"], progress.Look);
	        this.newDifficulty = source["newDifficulty"];
	        this.entry = source["entry"];
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
	export class JournalView {
	    entries: journal.Entry[];
	    insight: journal.Insight;
	
	    static createFrom(source: any = {}) {
	        return new JournalView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.entries = this.convertValues(source["entries"], journal.Entry);
	        this.insight = this.convertValues(source["insight"], journal.Insight);
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
	
	export class Interaction {
	    kind?: string;
	    prompts?: string[];
	    seconds?: number;
	    guide?: string;
	    target?: number;
	
	    static createFrom(source: any = {}) {
	        return new Interaction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.prompts = source["prompts"];
	        this.seconds = source["seconds"];
	        this.guide = source["guide"];
	        this.target = source["target"];
	    }
	}
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
	export class Translation {
	    title?: string;
	    description?: string;
	    prompts?: string[];
	    twists?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Translation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.description = source["description"];
	        this.prompts = source["prompts"];
	        this.twists = source["twists"];
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
	    interaction: Interaction;
	    i18n?: Record<string, Translation>;
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
	        this.interaction = this.convertValues(source["interaction"], Interaction);
	        this.i18n = this.convertValues(source["i18n"], Translation, true);
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

export namespace schedule {
	
	export class Schedule {
	    times: string[];
	    days: string;
	    notify: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Schedule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.times = source["times"];
	        this.days = source["days"];
	        this.notify = source["notify"];
	    }
	}

}

