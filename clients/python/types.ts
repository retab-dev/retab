export interface MIMEData {
    id: string;
    name: string;
    size: number;
    mime_type: string;
    content: string;
}

export interface BaseMIMEData {
    id: string;
    name: string;
    size: number;
    mime_type: string;
}


export interface RegexInstruction {
    name: string;
    pattern: string;
    description: string;
}

export interface DocumentSection {
    title: string;
    contentSummary?: string;
}

export interface DomainTerm {
    term: string;
    definition: string;
}

export interface UserInfo {
    first_name?: string;
    last_name?: string;
    email?: string;
    organization?: string;
    phone?: string;
    date?: Date;
}

export interface TextOperations {
    userInfo?: UserInfo;
    regexInstructions?: RegexInstruction[];
    documentStructure?: DocumentSection[];
    domainGlossary?: DomainTerm[];
    otherContext?: Record<string, any>;
}

export interface RegexInstructionResult {
    instruction: RegexInstruction;
    hits: string[];
}

export interface BaseDocumentAPIRequest {
    json_schema: Record<string, any>;
    model: string;
    seed?: number;
    temperature?: number;
    TextOperations?: TextOperations[];
}

export interface DocumentAPIRequest extends BaseDocumentAPIRequest {
    documents: MIMEData[];
}

export interface DocumentAPIResponse extends DocumentAPIRequest {
    response: Record<string, any>;
    prompt: string;
    document_results: BaseMIMEData[];
    regexInstructionResults?: RegexInstructionResult[];
}

export interface DocumentAPIError extends DocumentAPIRequest {
    error: string;
}
