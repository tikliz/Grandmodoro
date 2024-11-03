export namespace main {
	
	export class ProgramInfo {
	    PID: number;
	    Name: string;
	
	    static createFrom(source: any = {}) {
	        return new ProgramInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PID = source["PID"];
	        this.Name = source["Name"];
	    }
	}

}

