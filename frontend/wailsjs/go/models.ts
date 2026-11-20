export namespace main {
	
	export class State {
	    error: string;
	    self: node.SelfInfo;
	    peers: node.PeerView[];
	    transfers: node.Transfer[];
	
	    static createFrom(source: any = {}) {
	        return new State(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.error = source["error"];
	        this.self = this.convertValues(source["self"], node.SelfInfo);
	        this.peers = this.convertValues(source["peers"], node.PeerView);
	        this.transfers = this.convertValues(source["transfers"], node.Transfer);
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

export namespace node {
	
	export class PeerView {
	    id: string;
	    name: string;
	    os: string;
	    ip: string;
	    port: number;
	    fp: string;
	    link: string;
	    source: string;
	    online: boolean;
	    favorite: boolean;
	    autoAccept: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PeerView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.os = source["os"];
	        this.ip = source["ip"];
	        this.port = source["port"];
	        this.fp = source["fp"];
	        this.link = source["link"];
	        this.source = source["source"];
	        this.online = source["online"];
	        this.favorite = source["favorite"];
	        this.autoAccept = source["autoAccept"];
	    }
	}
	export class SelfInfo {
	    id: string;
	    name: string;
	    os: string;
	    port: number;
	    fp: string;
	    ips: string[];
	    downloadDir: string;
	
	    static createFrom(source: any = {}) {
	        return new SelfInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.os = source["os"];
	        this.port = source["port"];
	        this.fp = source["fp"];
	        this.ips = source["ips"];
	        this.downloadDir = source["downloadDir"];
	    }
	}
	export class Transfer {
	    id: string;
	    direction: string;
	    peerName: string;
	    label: string;
	    files: number;
	    total: number;
	    done: number;
	    speed: number;
	    status: string;
	    error?: string;
	    saved?: string[];
	    startedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new Transfer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.direction = source["direction"];
	        this.peerName = source["peerName"];
	        this.label = source["label"];
	        this.files = source["files"];
	        this.total = source["total"];
	        this.done = source["done"];
	        this.speed = source["speed"];
	        this.status = source["status"];
	        this.error = source["error"];
	        this.saved = source["saved"];
	        this.startedAt = source["startedAt"];
	    }
	}

}

export namespace protocol {
	
	export class FileMeta {
	    id: string;
	    name: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new FileMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.size = source["size"];
	    }
	}
	export class Info {
	    v: number;
	    id: string;
	    name: string;
	    os: string;
	    port: number;
	    fp: string;
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.v = source["v"];
	        this.id = source["id"];
	        this.name = source["name"];
	        this.os = source["os"];
	        this.port = source["port"];
	        this.fp = source["fp"];
	    }
	}

}

export namespace server {
	
	export class IncomingOffer {
	    id: string;
	    sender: protocol.Info;
	    senderIp: string;
	    files: protocol.FileMeta[];
	    totalBytes: number;
	
	    static createFrom(source: any = {}) {
	        return new IncomingOffer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sender = this.convertValues(source["sender"], protocol.Info);
	        this.senderIp = source["senderIp"];
	        this.files = this.convertValues(source["files"], protocol.FileMeta);
	        this.totalBytes = source["totalBytes"];
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

