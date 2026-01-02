import { CompositionClient } from "../../client.js";
import APIEditTemplates from "./templates/client.js";
import APIEditAgent from "./agent/client.js";

/**
 * Edit API client for document editing functionality.
 * 
 * Sub-clients:
 * - agent: Agent-based document editing (fill any document with AI)
 * - templates: Template-based PDF form filling (for batch processing)
 */
export default class APIEdit extends CompositionClient {
    public agent: APIEditAgent;
    public templates: APIEditTemplates;

    constructor(client: CompositionClient) {
        super(client);
        this.agent = new APIEditAgent(this);
        this.templates = new APIEditTemplates(this);
    }
}

