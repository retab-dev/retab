import { CompositionClient } from "../../client.js";
import APIEvalsClassify from "./classify/client.js";
import APIEvalsExtract from "./extract/client.js";
import APIEvalsSplit from "./split/client.js";

export default class APIEvals extends CompositionClient {
    public extract: APIEvalsExtract;
    public split: APIEvalsSplit;
    public classify: APIEvalsClassify;

    constructor(client: CompositionClient) {
        super(client);
        this.extract = new APIEvalsExtract(this);
        this.split = new APIEvalsSplit(this);
        this.classify = new APIEvalsClassify(this);
    }
}
