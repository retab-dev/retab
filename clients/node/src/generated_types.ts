import * as z from 'zod';

export const ZDistancesResult = z.lazy(() => z.object({
    distances: z.record(z.string(), z.any()),
    mean_distance: z.number(),
    metric_type: z.union([z.literal("levenshtein"), z.literal("jaccard"), z.literal("hamming")]),
}));
export type DistancesResult = z.infer<typeof ZDistancesResult>;

export const ZItemMetric = z.lazy(() => z.object({
    id: z.string(),
    name: z.string(),
    similarity: z.number(),
    similarities: z.record(z.string(), z.any()),
    flat_similarities: z.record(z.string(), z.number().optional()),
    aligned_similarity: z.number(),
    aligned_similarities: z.record(z.string(), z.any()),
    aligned_flat_similarities: z.record(z.string(), z.number().optional()),
}));
export type ItemMetric = z.infer<typeof ZItemMetric>;

export const ZMetricResult = z.lazy(() => z.object({
    item_metrics: z.array(ZItemMetric),
    mean_similarity: z.number(),
    aligned_mean_similarity: z.number(),
    metric_type: z.union([z.literal("levenshtein"), z.literal("jaccard"), z.literal("hamming")]),
}));
export type MetricResult = z.infer<typeof ZMetricResult>;

export const ZMetricType = z.lazy(() => z.union([z.literal("levenshtein"), z.literal("jaccard"), z.literal("hamming")]));
export type MetricType = z.infer<typeof ZMetricType>;

export const ZAttachmentMIMEData = z.lazy(() => z.object({
    filename: z.string(),
    url: z.string(),
    metadata: ZAttachmentMetadata,
}).merge(ZMIMEData.schema));
export type AttachmentMIMEData = z.infer<typeof ZAttachmentMIMEData>;

export const ZAttachmentMetadata = z.lazy(() => z.object({
    is_inline: z.boolean(),
    inline_cid: z.string().optional(),
    source: z.string().optional(),
}));
export type AttachmentMetadata = z.infer<typeof ZAttachmentMetadata>;

export const ZBaseAttachmentMIMEData = z.lazy(() => z.object({
    filename: z.string(),
    url: z.string(),
    metadata: ZAttachmentMetadata,
}).merge(ZBaseMIMEData.schema));
export type BaseAttachmentMIMEData = z.infer<typeof ZBaseAttachmentMIMEData>;

export const ZBaseEmailData = z.lazy(() => z.object({
    id: z.string(),
    tree_id: z.string(),
    subject: z.string().optional(),
    body_plain: z.string().optional(),
    body_html: z.string().optional(),
    sender: ZEmailAddressData,
    recipients_to: z.array(ZEmailAddressData),
    recipients_cc: z.array(ZEmailAddressData),
    recipients_bcc: z.array(ZEmailAddressData),
    sent_at: z.string(),
    received_at: z.string().optional(),
    in_reply_to: z.string().optional(),
    references: z.array(z.string()),
    headers: z.record(z.string(), z.string()),
    url: z.string().optional(),
    attachments: z.array(ZBaseAttachmentMIMEData),
}));
export type BaseEmailData = z.infer<typeof ZBaseEmailData>;

export const ZBaseMIMEData = z.lazy(() => z.object({
    filename: z.string(),
    url: z.string(),
}).merge(ZMIMEData.schema));
export type BaseMIMEData = z.infer<typeof ZBaseMIMEData>;

export const ZEmailAddressData = z.lazy(() => z.object({
    email: z.string(),
    display_name: z.string().optional(),
}));
export type EmailAddressData = z.infer<typeof ZEmailAddressData>;

export const ZEmailData = z.lazy(() => z.object({
    id: z.string(),
    tree_id: z.string(),
    subject: z.string().optional(),
    body_plain: z.string().optional(),
    body_html: z.string().optional(),
    sender: ZEmailAddressData,
    recipients_to: z.array(ZEmailAddressData),
    recipients_cc: z.array(ZEmailAddressData),
    recipients_bcc: z.array(ZEmailAddressData),
    sent_at: z.string(),
    received_at: z.string().optional(),
    in_reply_to: z.string().optional(),
    references: z.array(z.string()),
    headers: z.record(z.string(), z.string()),
    url: z.string().optional(),
    attachments: z.array(ZAttachmentMIMEData),
}).merge(ZBaseEmailData.schema));
export type EmailData = z.infer<typeof ZEmailData>;

export const ZMIMEData = z.lazy(() => z.object({
    filename: z.string(),
    url: z.string(),
}));
export type MIMEData = z.infer<typeof ZMIMEData>;

export const ZMatrix = z.lazy(() => z.object({
    rows: z.number(),
    cols: z.number(),
    type_: z.number(),
    data: z.string(),
}));
export type Matrix = z.infer<typeof ZMatrix>;

export const ZOCR = z.lazy(() => z.object({
    pages: z.array(ZPage),
}));
export type OCR = z.infer<typeof ZOCR>;

export const ZPage = z.lazy(() => z.object({
    page_number: z.number(),
    width: z.number(),
    height: z.number(),
    unit: z.string(),
    blocks: z.array(ZTextBox),
    lines: z.array(ZTextBox),
    tokens: z.array(ZTextBox),
    transforms: z.array(ZMatrix),
}));
export type Page = z.infer<typeof ZPage>;

export const ZPoint = z.lazy(() => z.object({
    x: z.number(),
    y: z.number(),
}));
export type Point = z.infer<typeof ZPoint>;

export const ZTextBox = z.lazy(() => z.object({
    width: z.number(),
    height: z.number(),
    center: ZPoint,
    vertices: z.tuple([ZPoint, ZPoint, ZPoint, ZPoint]),
    text: z.string(),
}));
export type TextBox = z.infer<typeof ZTextBox>;

export const ZAmount = z.lazy(() => z.object({
    value: z.number(),
    currency: z.string(),
}));
export type Amount = z.infer<typeof ZAmount>;

export const ZPredictionData = z.lazy(() => z.object({
    prediction: z.record(z.string(), z.any()),
    metadata: ZPredictionMetadata.optional(),
    updated_at: z.string().optional(),
}));
export type PredictionData = z.infer<typeof ZPredictionData>;

export const ZPredictionMetadata = z.lazy(() => z.object({
    extraction_id: z.string().optional(),
    likelihoods: z.record(z.string(), z.any()).optional(),
    field_locations: z.record(z.string(), z.any()).optional(),
    agentic_field_locations: z.record(z.string(), z.any()).optional(),
    consensus_details: z.array(z.record(z.string(), z.any())).optional(),
    api_cost: ZAmount.optional(),
}));
export type PredictionMetadata = z.infer<typeof ZPredictionMetadata>;

export const ZBrowserCanvas = z.lazy(() => z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]));
export type BrowserCanvas = z.infer<typeof ZBrowserCanvas>;

export const ZChatCompletion = z.lazy(() => z.object({
    id: z.string(),
    choices: z.array(ZChoice),
    created: z.number(),
    model: z.string(),
    object: z.literal("chat.completion"),
    service_tier: z.union([z.literal("auto"), z.literal("default"), z.literal("flex"), z.literal("scale"), z.literal("priority")]).optional(),
    system_fingerprint: z.string().optional(),
    usage: ZCompletionUsage.optional(),
}));
export type ChatCompletion = z.infer<typeof ZChatCompletion>;

export const ZChatCompletionReasoningEffort = z.lazy(() => z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).optional());
export type ChatCompletionReasoningEffort = z.infer<typeof ZChatCompletionReasoningEffort>;

export const ZChatCompletionRetabMessage = z.lazy(() => z.object({
    role: z.union([z.literal("user"), z.literal("system"), z.literal("assistant"), z.literal("developer")]),
    content: z.union([z.string(), z.array(z.union([ZChatCompletionContentPartTextParam, ZChatCompletionContentPartImageParam, ZChatCompletionContentPartInputAudioParam, ZFile]))]),
}));
export type ChatCompletionRetabMessage = z.infer<typeof ZChatCompletionRetabMessage>;

export const ZCostBreakdown = z.lazy(() => z.object({
    total: ZAmount,
    text_prompt_cost: ZAmount,
    text_cached_cost: ZAmount,
    text_completion_cost: ZAmount,
    text_total_cost: ZAmount,
    audio_prompt_cost: ZAmount.optional(),
    audio_completion_cost: ZAmount.optional(),
    audio_total_cost: ZAmount.optional(),
    token_counts: ZTokenCounts,
    model: z.string(),
    is_fine_tuned: z.boolean(),
}));
export type CostBreakdown = z.infer<typeof ZCostBreakdown>;

export const ZExtraction = z.lazy(() => z.object({
    id: z.string(),
    messages: z.array(ZChatCompletionRetabMessage),
    messages_gcs: z.string(),
    file_gcs_paths: z.array(z.string()),
    file_ids: z.array(z.string()),
    file_gcs: z.string(),
    file_id: z.string(),
    status: z.union([z.literal("success"), z.literal("failed")]),
    completion: z.union([ZRetabParsedChatCompletion, ZChatCompletion]),
    json_schema: z.any(),
    model: z.string(),
    temperature: z.number(),
    source: ZExtractionSource,
    image_resolution_dpi: z.number(),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).optional(),
    n_consensus: z.number(),
    timings: z.array(ZExtractionTimingStep),
    schema_id: z.string(),
    schema_data_id: z.string(),
    created_at: z.string(),
    request_at: z.string().optional(),
    organization_id: z.string(),
    validation_state: z.union([z.literal("pending"), z.literal("validated"), z.literal("invalid")]).optional(),
    billed: z.boolean(),
}));
export type Extraction = z.infer<typeof ZExtraction>;

export const ZExtractionSource = z.lazy(() => z.object({
    type: z.union([z.literal("api"), z.literal("annotation"), z.literal("processor"), z.literal("automation"), z.literal("automation.link"), z.literal("automation.mailbox"), z.literal("automation.cron"), z.literal("automation.outlook"), z.literal("automation.endpoint"), z.literal("schema.extract")]),
    id: z.string().optional(),
}));
export type ExtractionSource = z.infer<typeof ZExtractionSource>;

export const ZExtractionSteps = z.lazy(() => z.union([z.string(), z.union([z.literal("initialization"), z.literal("prepare_messages"), z.literal("yield_first_token"), z.literal("completion")])]));
export type ExtractionSteps = z.infer<typeof ZExtractionSteps>;

export const ZExtractionTimingStep = z.lazy(() => z.object({
    name: z.union([z.string(), z.union([z.literal("initialization"), z.literal("prepare_messages"), z.literal("yield_first_token"), z.literal("completion")])]),
    duration: z.number(),
    notes: z.string().optional(),
}));
export type ExtractionTimingStep = z.infer<typeof ZExtractionTimingStep>;

export const ZModality = z.lazy(() => z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]));
export type Modality = z.infer<typeof ZModality>;

export const ZRetabParsedChatCompletion = z.lazy(() => z.object({
    id: z.string(),
    choices: z.array(ZRetabParsedChoice),
    created: z.number(),
    model: z.string(),
    object: z.literal("chat.completion"),
    service_tier: z.union([z.literal("auto"), z.literal("default"), z.literal("flex"), z.literal("scale"), z.literal("priority")]).optional(),
    system_fingerprint: z.string().optional(),
    usage: ZCompletionUsage.optional(),
    extraction_id: z.string().optional(),
    likelihoods: z.record(z.string(), z.any()).optional(),
    schema_validation_error: ZErrorDetail.optional(),
    request_at: z.string().optional(),
    first_token_at: z.string().optional(),
    last_token_at: z.string().optional(),
}).merge(ZParsedChatCompletion.schema));
export type RetabParsedChatCompletion = z.infer<typeof ZRetabParsedChatCompletion>;

export const ZValidationsState = z.lazy(() => z.union([z.literal("pending"), z.literal("validated"), z.literal("invalid")]));
export type ValidationsState = z.infer<typeof ZValidationsState>;

export const ZEvent = z.lazy(() => z.object({
    object: z.literal("event"),
    id: z.string(),
    event: z.string(),
    created_at: z.string(),
    data: z.record(z.string(), z.any()),
    metadata: z.record(z.union([z.literal("automation"), z.literal("cron"), z.literal("data_structure"), z.literal("dataset"), z.literal("dataset_membership"), z.literal("endpoint"), z.literal("evaluation"), z.literal("extraction"), z.literal("file"), z.literal("files"), z.literal("link"), z.literal("mailbox"), z.literal("organization"), z.literal("outlook"), z.literal("preprocessing"), z.literal("reconciliation"), z.literal("schema"), z.literal("schema_data"), z.literal("template"), z.literal("user"), z.literal("webhook")]), z.string()).optional(),
}));
export type Event = z.infer<typeof ZEvent>;

export const ZStoredEvent = z.lazy(() => z.object({
    object: z.literal("event"),
    id: z.string(),
    event: z.string(),
    created_at: z.string(),
    data: z.record(z.string(), z.any()),
    metadata: z.record(z.union([z.literal("automation"), z.literal("cron"), z.literal("data_structure"), z.literal("dataset"), z.literal("dataset_membership"), z.literal("endpoint"), z.literal("evaluation"), z.literal("extraction"), z.literal("file"), z.literal("files"), z.literal("link"), z.literal("mailbox"), z.literal("organization"), z.literal("outlook"), z.literal("preprocessing"), z.literal("reconciliation"), z.literal("schema"), z.literal("schema_data"), z.literal("template"), z.literal("user"), z.literal("webhook")]), z.string()).optional(),
    organization_id: z.string(),
}).merge(ZEvent.schema));
export type StoredEvent = z.infer<typeof ZStoredEvent>;

export const ZAUDIO_TYPES = z.lazy(() => z.union([z.literal(".mp3"), z.literal(".mp4"), z.literal(".mpeg"), z.literal(".mpga"), z.literal(".m4a"), z.literal(".wav"), z.literal(".webm")]));
export type AUDIO_TYPES = z.infer<typeof ZAUDIO_TYPES>;

export const ZBaseModality = z.lazy(() => z.union([z.literal("text"), z.literal("image")]));
export type BaseModality = z.infer<typeof ZBaseModality>;

export const ZEMAIL_TYPES = z.lazy(() => z.union([z.literal(".eml"), z.literal(".msg")]));
export type EMAIL_TYPES = z.infer<typeof ZEMAIL_TYPES>;

export const ZEXCEL_TYPES = z.lazy(() => z.union([z.literal(".xls"), z.literal(".xlsx"), z.literal(".ods")]));
export type EXCEL_TYPES = z.infer<typeof ZEXCEL_TYPES>;

export const ZHTML_TYPES = z.lazy(() => z.union([z.literal(".html"), z.literal(".htm")]));
export type HTML_TYPES = z.infer<typeof ZHTML_TYPES>;

export const ZIMAGE_TYPES = z.lazy(() => z.union([z.literal(".jpg"), z.literal(".jpeg"), z.literal(".png"), z.literal(".gif"), z.literal(".bmp"), z.literal(".tiff"), z.literal(".webp")]));
export type IMAGE_TYPES = z.infer<typeof ZIMAGE_TYPES>;

export const ZPDF_TYPES = z.lazy(() => z.literal(".pdf"));
export type PDF_TYPES = z.infer<typeof ZPDF_TYPES>;

export const ZPPT_TYPES = z.lazy(() => z.union([z.literal(".ppt"), z.literal(".pptx"), z.literal(".odp")]));
export type PPT_TYPES = z.infer<typeof ZPPT_TYPES>;

export const ZSUPPORTED_TYPES = z.lazy(() => z.union([z.literal(".xls"), z.literal(".xlsx"), z.literal(".ods"), z.literal(".doc"), z.literal(".docx"), z.literal(".odt"), z.literal(".ppt"), z.literal(".pptx"), z.literal(".odp"), z.literal(".pdf"), z.literal(".jpg"), z.literal(".jpeg"), z.literal(".png"), z.literal(".gif"), z.literal(".bmp"), z.literal(".tiff"), z.literal(".webp"), z.literal(".txt"), z.literal(".csv"), z.literal(".tsv"), z.literal(".md"), z.literal(".log"), z.literal(".xml"), z.literal(".json"), z.literal(".yaml"), z.literal(".yml"), z.literal(".rtf"), z.literal(".ini"), z.literal(".conf"), z.literal(".cfg"), z.literal(".nfo"), z.literal(".srt"), z.literal(".sql"), z.literal(".sh"), z.literal(".bat"), z.literal(".ps1"), z.literal(".js"), z.literal(".jsx"), z.literal(".ts"), z.literal(".tsx"), z.literal(".py"), z.literal(".java"), z.literal(".c"), z.literal(".cpp"), z.literal(".cs"), z.literal(".rb"), z.literal(".php"), z.literal(".swift"), z.literal(".kt"), z.literal(".go"), z.literal(".rs"), z.literal(".pl"), z.literal(".r"), z.literal(".m"), z.literal(".scala"), z.literal(".html"), z.literal(".htm"), z.literal(".mhtml"), z.literal(".eml"), z.literal(".msg"), z.literal(".mp3"), z.literal(".mp4"), z.literal(".mpeg"), z.literal(".mpga"), z.literal(".m4a"), z.literal(".wav"), z.literal(".webm")]));
export type SUPPORTED_TYPES = z.infer<typeof ZSUPPORTED_TYPES>;

export const ZTEXT_TYPES = z.lazy(() => z.union([z.literal(".txt"), z.literal(".csv"), z.literal(".tsv"), z.literal(".md"), z.literal(".log"), z.literal(".xml"), z.literal(".json"), z.literal(".yaml"), z.literal(".yml"), z.literal(".rtf"), z.literal(".ini"), z.literal(".conf"), z.literal(".cfg"), z.literal(".nfo"), z.literal(".srt"), z.literal(".sql"), z.literal(".sh"), z.literal(".bat"), z.literal(".ps1"), z.literal(".js"), z.literal(".jsx"), z.literal(".ts"), z.literal(".tsx"), z.literal(".py"), z.literal(".java"), z.literal(".c"), z.literal(".cpp"), z.literal(".cs"), z.literal(".rb"), z.literal(".php"), z.literal(".swift"), z.literal(".kt"), z.literal(".go"), z.literal(".rs"), z.literal(".pl"), z.literal(".r"), z.literal(".m"), z.literal(".scala")]));
export type TEXT_TYPES = z.infer<typeof ZTEXT_TYPES>;

export const ZTYPE_FAMILIES = z.lazy(() => z.union([z.literal("excel"), z.literal("word"), z.literal("powerpoint"), z.literal("pdf"), z.literal("image"), z.literal("text"), z.literal("email"), z.literal("audio"), z.literal("html"), z.literal("web")]));
export type TYPE_FAMILIES = z.infer<typeof ZTYPE_FAMILIES>;

export const ZWEB_TYPES = z.lazy(() => z.literal(".mhtml"));
export type WEB_TYPES = z.infer<typeof ZWEB_TYPES>;

export const ZWORD_TYPES = z.lazy(() => z.union([z.literal(".doc"), z.literal(".docx"), z.literal(".odt")]));
export type WORD_TYPES = z.infer<typeof ZWORD_TYPES>;

export const ZDeleteResponse = z.lazy(() => z.object({
    success: z.boolean(),
    id: z.string(),
}));
export type DeleteResponse = z.infer<typeof ZDeleteResponse>;

export const ZDocumentPreprocessResponseContent = z.lazy(() => z.object({
    messages: z.array(z.record(z.string(), z.any())),
    json_schema: z.record(z.string(), z.any()),
}));
export type DocumentPreprocessResponseContent = z.infer<typeof ZDocumentPreprocessResponseContent>;

export const ZErrorDetail = z.lazy(() => z.object({
    code: z.string(),
    message: z.string(),
    details: z.record(z.any()).optional(),
}));
export type ErrorDetail = z.infer<typeof ZErrorDetail>;

export const ZExportResponse = z.lazy(() => z.object({
    success: z.boolean(),
    path: z.string(),
}));
export type ExportResponse = z.infer<typeof ZExportResponse>;

export const ZPreparedRequest = z.lazy(() => z.object({
    method: z.union([z.literal("POST"), z.literal("GET"), z.literal("PUT"), z.literal("PATCH"), z.literal("DELETE"), z.literal("HEAD"), z.literal("OPTIONS"), z.literal("CONNECT"), z.literal("TRACE")]),
    url: z.string(),
    data: z.record(z.any()).optional(),
    params: z.record(z.any()).optional(),
    form_data: z.record(z.any()).optional(),
    files: z.union([z.record(z.any()), z.array(z.tuple([z.string(), z.tuple([z.string(), z.instanceof(Uint8Array), z.string()])]))]).optional(),
    idempotency_key: z.string().optional(),
    raise_for_status: z.boolean(),
}));
export type PreparedRequest = z.infer<typeof ZPreparedRequest>;

export const ZStandardErrorResponse = z.lazy(() => z.object({
    detail: ZErrorDetail,
}));
export type StandardErrorResponse = z.infer<typeof ZStandardErrorResponse>;

export const ZStreamingBaseModel = z.lazy(() => z.object({
    streaming_error: ZErrorDetail.optional(),
}));
export type StreamingBaseModel = z.infer<typeof ZStreamingBaseModel>;

export const ZT = z.lazy(() => z.any());
export type T = z.infer<typeof ZT>;

export const ZTuple = z.lazy(() => z.tuple([]));
export type Tuple = z.infer<typeof ZTuple>;

export const ZAIProvider = z.lazy(() => z.union([z.literal("OpenAI"), z.literal("Anthropic"), z.literal("Gemini"), z.literal("xAI"), z.literal("Retab")]));
export type AIProvider = z.infer<typeof ZAIProvider>;

export const ZReasoning = z.lazy(() => z.object({
    effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).optional(),
    generate_summary: z.union([z.literal("auto"), z.literal("concise"), z.literal("detailed")]).optional(),
    summary: z.union([z.literal("auto"), z.literal("concise"), z.literal("detailed")]).optional(),
}));
export type Reasoning = z.infer<typeof ZReasoning>;

export const ZResponseFormatJSONSchema = z.lazy(() => z.object({
    json_schema: ZJSONSchema,
    type: z.literal("json_schema"),
}));
export type ResponseFormatJSONSchema = z.infer<typeof ZResponseFormatJSONSchema>;

export const ZResponseInputParam = z.lazy(() => z.array(z.union([ZEasyInputMessageParam, ZMessage, ZResponseOutputMessageParam, ZResponseFileSearchToolCallParam, ZResponseComputerToolCallParam, ZComputerCallOutput, ZResponseFunctionWebSearchParam, ZResponseFunctionToolCallParam, ZFunctionCallOutput, ZResponseReasoningItemParam, ZImageGenerationCall, ZResponseCodeInterpreterToolCallParam, ZLocalShellCall, ZLocalShellCallOutput, ZMcpListTools, ZMcpApprovalRequest, ZMcpApprovalResponse, ZMcpCall, ZItemReference])));
export type ResponseInputParam = z.infer<typeof ZResponseInputParam>;

export const ZResponseTextConfigParam = z.lazy(() => z.object({
    format: z.union([ZResponseFormatText, ZResponseFormatTextJSONSchemaConfigParam, ZResponseFormatJSONObject]),
}));
export type ResponseTextConfigParam = z.infer<typeof ZResponseTextConfigParam>;

export const ZRetabChatCompletionsParseRequest = z.lazy(() => z.object({
    model: z.string(),
    messages: z.array(ZChatCompletionRetabMessage),
    json_schema: z.record(z.string(), z.any()),
    temperature: z.number(),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).optional(),
    stream: z.boolean(),
    seed: z.number().optional(),
    n_consensus: z.number(),
}));
export type RetabChatCompletionsParseRequest = z.infer<typeof ZRetabChatCompletionsParseRequest>;

export const ZRetabChatCompletionsRequest = z.lazy(() => z.object({
    model: z.string(),
    messages: z.array(ZChatCompletionRetabMessage),
    response_format: ZResponseFormatJSONSchema,
    temperature: z.number(),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).optional(),
    stream: z.boolean(),
    seed: z.number().optional(),
    n_consensus: z.number(),
}));
export type RetabChatCompletionsRequest = z.infer<typeof ZRetabChatCompletionsRequest>;

export const ZRetabChatResponseCreateRequest = z.lazy(() => z.object({
    input: z.union([z.string(), z.array(z.union([ZEasyInputMessageParam, ZMessage, ZResponseOutputMessageParam, ZResponseFileSearchToolCallParam, ZResponseComputerToolCallParam, ZComputerCallOutput, ZResponseFunctionWebSearchParam, ZResponseFunctionToolCallParam, ZFunctionCallOutput, ZResponseReasoningItemParam, ZImageGenerationCall, ZResponseCodeInterpreterToolCallParam, ZLocalShellCall, ZLocalShellCallOutput, ZMcpListTools, ZMcpApprovalRequest, ZMcpApprovalResponse, ZMcpCall, ZItemReference]))]),
    instructions: z.string().optional(),
    model: z.string(),
    temperature: z.number().optional(),
    reasoning: ZReasoning.optional(),
    stream: z.boolean().optional(),
    seed: z.number().optional(),
    text: ZResponseTextConfigParam,
    n_consensus: z.number(),
}));
export type RetabChatResponseCreateRequest = z.infer<typeof ZRetabChatResponseCreateRequest>;

export const ZChatCompletionContentPartParam = z.lazy(() => z.union([ZChatCompletionContentPartTextParam, ZChatCompletionContentPartImageParam, ZChatCompletionContentPartInputAudioParam, ZFile]));
export type ChatCompletionContentPartParam = z.infer<typeof ZChatCompletionContentPartParam>;

export const ZReconciliationRequest = z.lazy(() => z.object({
    list_dicts: z.array(z.record(z.any())),
    reference_schema: z.record(z.string(), z.any()).optional(),
    mode: z.union([z.literal("direct"), z.literal("aligned")]),
}));
export type ReconciliationRequest = z.infer<typeof ZReconciliationRequest>;

export const ZReconciliationResponse = z.lazy(() => z.object({
    consensus_dict: z.record(z.any()),
    likelihoods: z.record(z.any()),
}));
export type ReconciliationResponse = z.infer<typeof ZReconciliationResponse>;

export const ZAutomationConfig = z.lazy(() => z.object({
    id: z.string(),
    name: z.string(),
    processor_id: z.string(),
    updated_at: z.string(),
    default_language: z.string(),
    webhook_url: z.string(),
    webhook_headers: z.record(z.string(), z.string()),
    need_validation: z.boolean(),
}));
export type AutomationConfig = z.infer<typeof ZAutomationConfig>;

export const ZAutomationLog = z.lazy(() => z.object({
    object: z.literal("automation_log"),
    id: z.string(),
    user_email: z.string().email().optional(),
    organization_id: z.string(),
    created_at: z.string(),
    automation_snapshot: ZAutomationConfig,
    completion: z.union([ZRetabParsedChatCompletion, ZChatCompletion]),
    file_metadata: ZBaseMIMEData.optional(),
    external_request_log: ZExternalRequestLog.optional(),
    extraction_id: z.string().optional(),
}));
export type AutomationLog = z.infer<typeof ZAutomationLog>;

export const ZDict = z.lazy(() => z.record(z.any()));
export type Dict = z.infer<typeof ZDict>;

export const ZEmailStr = z.lazy(() => z.string().email());
export type EmailStr = z.infer<typeof ZEmailStr>;

export const ZExternalRequestLog = z.lazy(() => z.object({
    webhook_url: z.string().optional(),
    request_body: z.record(z.string(), z.any()),
    request_headers: z.record(z.string(), z.string()),
    request_at: z.string(),
    response_body: z.record(z.string(), z.any()),
    response_headers: z.record(z.string(), z.string()),
    response_at: z.string(),
    status_code: z.number(),
    error: z.string().optional(),
    duration_ms: z.number(),
}));
export type ExternalRequestLog = z.infer<typeof ZExternalRequestLog>;

export const ZListLogs = z.lazy(() => z.object({
    data: z.array(ZAutomationLog),
    list_metadata: ZListMetadata,
}));
export type ListLogs = z.infer<typeof ZListLogs>;

export const ZListMetadata = z.lazy(() => z.object({
    before: z.string().optional(),
    after: z.string().optional(),
}));
export type ListMetadata = z.infer<typeof ZListMetadata>;

export const ZLogCompletionRequest = z.lazy(() => z.object({
    json_schema: z.record(z.string(), z.any()),
    completion: ZChatCompletion,
}));
export type LogCompletionRequest = z.infer<typeof ZLogCompletionRequest>;

export const ZOpenAIRequestConfig = z.lazy(() => z.object({
    object: z.literal("openai_request"),
    id: z.string(),
    model: z.string(),
    json_schema: z.record(z.string(), z.any()),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).optional(),
}));
export type OpenAIRequestConfig = z.infer<typeof ZOpenAIRequestConfig>;

export const ZProcessorConfig = z.lazy(() => z.object({
    object: z.string(),
    id: z.string(),
    updated_at: z.string(),
    name: z.string(),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]),
    image_resolution_dpi: z.number(),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]),
    model: z.string(),
    json_schema: z.record(z.string(), z.any()),
    temperature: z.number(),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).optional(),
    n_consensus: z.number(),
}));
export type ProcessorConfig = z.infer<typeof ZProcessorConfig>;

export const ZUpdateAutomationRequest = z.lazy(() => z.object({
    name: z.string().optional(),
    default_language: z.string().optional(),
    webhook_url: z.string().optional(),
    webhook_headers: z.record(z.string(), z.string()).optional(),
    need_validation: z.boolean().optional(),
}));
export type UpdateAutomationRequest = z.infer<typeof ZUpdateAutomationRequest>;

export const ZUpdateProcessorRequest = z.lazy(() => z.object({
    name: z.string().optional(),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]).optional(),
    image_resolution_dpi: z.number().optional(),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]).optional(),
    model: z.string().optional(),
    json_schema: z.record(z.any()).optional(),
    temperature: z.number().optional(),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).optional(),
    n_consensus: z.number().optional(),
}));
export type UpdateProcessorRequest = z.infer<typeof ZUpdateProcessorRequest>;

export const ZAnthropicModel = z.lazy(() => z.union([z.literal("claude-3-5-sonnet-latest"), z.literal("claude-3-5-sonnet-20241022"), z.literal("claude-3-opus-20240229"), z.literal("claude-3-sonnet-20240229"), z.literal("claude-3-haiku-20240307")]));
export type AnthropicModel = z.infer<typeof ZAnthropicModel>;

export const ZEndpointType = z.lazy(() => z.union([z.literal("chat_completions"), z.literal("responses"), z.literal("assistants"), z.literal("batch"), z.literal("fine_tuning"), z.literal("embeddings"), z.literal("speech_generation"), z.literal("translation"), z.literal("completions_legacy"), z.literal("image_generation"), z.literal("transcription"), z.literal("moderation"), z.literal("realtime")]));
export type EndpointType = z.infer<typeof ZEndpointType>;

export const ZFeatureType = z.lazy(() => z.union([z.literal("streaming"), z.literal("function_calling"), z.literal("structured_outputs"), z.literal("distillation"), z.literal("fine_tuning"), z.literal("predicted_outputs"), z.literal("schema_generation")]));
export type FeatureType = z.infer<typeof ZFeatureType>;

export const ZFinetunedModel = z.lazy(() => z.object({
    object: z.literal("finetuned_model"),
    organization_id: z.string(),
    model: z.string(),
    schema_id: z.string(),
    schema_data_id: z.string(),
    finetuning_props: ZInferenceSettings,
    evaluation_id: z.string().optional(),
    created_at: z.string(),
}));
export type FinetunedModel = z.infer<typeof ZFinetunedModel>;

export const ZGeminiModel = z.lazy(() => z.union([z.literal("gemini-2.5-pro"), z.literal("gemini-2.5-flash"), z.literal("gemini-2.5-pro-preview-06-05"), z.literal("gemini-2.5-pro-preview-05-06"), z.literal("gemini-2.5-pro-preview-03-25"), z.literal("gemini-2.5-flash-preview-05-20"), z.literal("gemini-2.5-flash-preview-04-17"), z.literal("gemini-2.5-flash-lite-preview-06-17"), z.literal("gemini-2.5-pro-exp-03-25"), z.literal("gemini-2.0-flash-lite"), z.literal("gemini-2.0-flash")]));
export type GeminiModel = z.infer<typeof ZGeminiModel>;

export const ZInferenceSettings = z.lazy(() => z.object({
    model: z.string(),
    temperature: z.number(),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).optional(),
    image_resolution_dpi: z.number(),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]),
    n_consensus: z.number(),
}));
export type InferenceSettings = z.infer<typeof ZInferenceSettings>;

export const ZLLMModel = z.lazy(() => z.union([z.literal("gpt-4o"), z.literal("gpt-4o-mini"), z.literal("chatgpt-4o-latest"), z.literal("gpt-4.1"), z.literal("gpt-4.1-mini"), z.literal("gpt-4.1-mini-2025-04-14"), z.literal("gpt-4.1-2025-04-14"), z.literal("gpt-4.1-nano"), z.literal("gpt-4.1-nano-2025-04-14"), z.literal("gpt-4o-2024-11-20"), z.literal("gpt-4o-2024-08-06"), z.literal("gpt-4o-2024-05-13"), z.literal("gpt-4o-mini-2024-07-18"), z.literal("o1"), z.literal("o1-2024-12-17"), z.literal("o3"), z.literal("o3-2025-04-16"), z.literal("o4-mini"), z.literal("o4-mini-2025-04-16"), z.literal("gpt-4o-audio-preview-2024-12-17"), z.literal("gpt-4o-audio-preview-2024-10-01"), z.literal("gpt-4o-realtime-preview-2024-12-17"), z.literal("gpt-4o-realtime-preview-2024-10-01"), z.literal("gpt-4o-mini-audio-preview-2024-12-17"), z.literal("gpt-4o-mini-realtime-preview-2024-12-17"), z.literal("claude-3-5-sonnet-latest"), z.literal("claude-3-5-sonnet-20241022"), z.literal("claude-3-opus-20240229"), z.literal("claude-3-sonnet-20240229"), z.literal("claude-3-haiku-20240307"), z.literal("grok-3"), z.literal("grok-3-mini"), z.literal("gemini-2.5-pro"), z.literal("gemini-2.5-flash"), z.literal("gemini-2.5-pro-preview-06-05"), z.literal("gemini-2.5-pro-preview-05-06"), z.literal("gemini-2.5-pro-preview-03-25"), z.literal("gemini-2.5-flash-preview-05-20"), z.literal("gemini-2.5-flash-preview-04-17"), z.literal("gemini-2.5-flash-lite-preview-06-17"), z.literal("gemini-2.5-pro-exp-03-25"), z.literal("gemini-2.0-flash-lite"), z.literal("gemini-2.0-flash"), z.literal("auto-large"), z.literal("auto-small"), z.literal("auto-micro"), z.literal("human")]));
export type LLMModel = z.infer<typeof ZLLMModel>;

export const ZModel = z.lazy(() => z.object({
    id: z.string(),
    created: z.number(),
    object: z.literal("model"),
    owned_by: z.string(),
}));
export type Model = z.infer<typeof ZModel>;

export const ZModelCapabilities = z.lazy(() => z.object({
    modalities: z.array(z.union([z.literal("text"), z.literal("audio"), z.literal("image")])),
    endpoints: z.array(z.union([z.literal("chat_completions"), z.literal("responses"), z.literal("assistants"), z.literal("batch"), z.literal("fine_tuning"), z.literal("embeddings"), z.literal("speech_generation"), z.literal("translation"), z.literal("completions_legacy"), z.literal("image_generation"), z.literal("transcription"), z.literal("moderation"), z.literal("realtime")])),
    features: z.array(z.union([z.literal("streaming"), z.literal("function_calling"), z.literal("structured_outputs"), z.literal("distillation"), z.literal("fine_tuning"), z.literal("predicted_outputs"), z.literal("schema_generation")])),
}));
export type ModelCapabilities = z.infer<typeof ZModelCapabilities>;

export const ZModelCard = z.lazy(() => z.object({
    model: z.union([z.union([z.literal("gpt-4o"), z.literal("gpt-4o-mini"), z.literal("chatgpt-4o-latest"), z.literal("gpt-4.1"), z.literal("gpt-4.1-mini"), z.literal("gpt-4.1-mini-2025-04-14"), z.literal("gpt-4.1-2025-04-14"), z.literal("gpt-4.1-nano"), z.literal("gpt-4.1-nano-2025-04-14"), z.literal("gpt-4o-2024-11-20"), z.literal("gpt-4o-2024-08-06"), z.literal("gpt-4o-2024-05-13"), z.literal("gpt-4o-mini-2024-07-18"), z.literal("o1"), z.literal("o1-2024-12-17"), z.literal("o3"), z.literal("o3-2025-04-16"), z.literal("o4-mini"), z.literal("o4-mini-2025-04-16"), z.literal("gpt-4o-audio-preview-2024-12-17"), z.literal("gpt-4o-audio-preview-2024-10-01"), z.literal("gpt-4o-realtime-preview-2024-12-17"), z.literal("gpt-4o-realtime-preview-2024-10-01"), z.literal("gpt-4o-mini-audio-preview-2024-12-17"), z.literal("gpt-4o-mini-realtime-preview-2024-12-17"), z.literal("claude-3-5-sonnet-latest"), z.literal("claude-3-5-sonnet-20241022"), z.literal("claude-3-opus-20240229"), z.literal("claude-3-sonnet-20240229"), z.literal("claude-3-haiku-20240307"), z.literal("grok-3"), z.literal("grok-3-mini"), z.literal("gemini-2.5-pro"), z.literal("gemini-2.5-flash"), z.literal("gemini-2.5-pro-preview-06-05"), z.literal("gemini-2.5-pro-preview-05-06"), z.literal("gemini-2.5-pro-preview-03-25"), z.literal("gemini-2.5-flash-preview-05-20"), z.literal("gemini-2.5-flash-preview-04-17"), z.literal("gemini-2.5-flash-lite-preview-06-17"), z.literal("gemini-2.5-pro-exp-03-25"), z.literal("gemini-2.0-flash-lite"), z.literal("gemini-2.0-flash"), z.literal("auto-large"), z.literal("auto-small"), z.literal("auto-micro"), z.literal("human")]), z.string()]),
    pricing: ZPricing,
    capabilities: ZModelCapabilities,
    temperature_support: z.boolean(),
    reasoning_effort_support: z.boolean(),
    permissions: ZModelCardPermissions,
}));
export type ModelCard = z.infer<typeof ZModelCard>;

export const ZModelCardPermissions = z.lazy(() => z.object({
    show_in_free_picker: z.boolean(),
    show_in_paid_picker: z.boolean(),
}));
export type ModelCardPermissions = z.infer<typeof ZModelCardPermissions>;

export const ZModelModality = z.lazy(() => z.union([z.literal("text"), z.literal("audio"), z.literal("image")]));
export type ModelModality = z.infer<typeof ZModelModality>;

export const ZMonthlyUsageResponse = z.lazy(() => z.object({
    credits_count: z.number(),
}));
export type MonthlyUsageResponse = z.infer<typeof ZMonthlyUsageResponse>;

export const ZMonthlyUsageResponseContent = z.lazy(() => z.object({
    credits_count: z.number(),
}));
export type MonthlyUsageResponseContent = z.infer<typeof ZMonthlyUsageResponseContent>;

export const ZOpenAICompatibleProvider = z.lazy(() => z.union([z.literal("OpenAI"), z.literal("xAI")]));
export type OpenAICompatibleProvider = z.infer<typeof ZOpenAICompatibleProvider>;

export const ZOpenAIModel = z.lazy(() => z.union([z.literal("gpt-4o"), z.literal("gpt-4o-mini"), z.literal("chatgpt-4o-latest"), z.literal("gpt-4.1"), z.literal("gpt-4.1-mini"), z.literal("gpt-4.1-mini-2025-04-14"), z.literal("gpt-4.1-2025-04-14"), z.literal("gpt-4.1-nano"), z.literal("gpt-4.1-nano-2025-04-14"), z.literal("gpt-4o-2024-11-20"), z.literal("gpt-4o-2024-08-06"), z.literal("gpt-4o-2024-05-13"), z.literal("gpt-4o-mini-2024-07-18"), z.literal("o1"), z.literal("o1-2024-12-17"), z.literal("o3"), z.literal("o3-2025-04-16"), z.literal("o4-mini"), z.literal("o4-mini-2025-04-16"), z.literal("gpt-4o-audio-preview-2024-12-17"), z.literal("gpt-4o-audio-preview-2024-10-01"), z.literal("gpt-4o-realtime-preview-2024-12-17"), z.literal("gpt-4o-realtime-preview-2024-10-01"), z.literal("gpt-4o-mini-audio-preview-2024-12-17"), z.literal("gpt-4o-mini-realtime-preview-2024-12-17")]));
export type OpenAIModel = z.infer<typeof ZOpenAIModel>;

export const ZPricing = z.lazy(() => z.object({
    text: ZTokenPrice,
    audio: ZTokenPrice.optional(),
    ft_price_hike: z.number(),
}));
export type Pricing = z.infer<typeof ZPricing>;

export const ZPureLLMModel = z.lazy(() => z.union([z.literal("gpt-4o"), z.literal("gpt-4o-mini"), z.literal("chatgpt-4o-latest"), z.literal("gpt-4.1"), z.literal("gpt-4.1-mini"), z.literal("gpt-4.1-mini-2025-04-14"), z.literal("gpt-4.1-2025-04-14"), z.literal("gpt-4.1-nano"), z.literal("gpt-4.1-nano-2025-04-14"), z.literal("gpt-4o-2024-11-20"), z.literal("gpt-4o-2024-08-06"), z.literal("gpt-4o-2024-05-13"), z.literal("gpt-4o-mini-2024-07-18"), z.literal("o1"), z.literal("o1-2024-12-17"), z.literal("o3"), z.literal("o3-2025-04-16"), z.literal("o4-mini"), z.literal("o4-mini-2025-04-16"), z.literal("gpt-4o-audio-preview-2024-12-17"), z.literal("gpt-4o-audio-preview-2024-10-01"), z.literal("gpt-4o-realtime-preview-2024-12-17"), z.literal("gpt-4o-realtime-preview-2024-10-01"), z.literal("gpt-4o-mini-audio-preview-2024-12-17"), z.literal("gpt-4o-mini-realtime-preview-2024-12-17"), z.literal("claude-3-5-sonnet-latest"), z.literal("claude-3-5-sonnet-20241022"), z.literal("claude-3-opus-20240229"), z.literal("claude-3-sonnet-20240229"), z.literal("claude-3-haiku-20240307"), z.literal("grok-3"), z.literal("grok-3-mini"), z.literal("gemini-2.5-pro"), z.literal("gemini-2.5-flash"), z.literal("gemini-2.5-pro-preview-06-05"), z.literal("gemini-2.5-pro-preview-05-06"), z.literal("gemini-2.5-pro-preview-03-25"), z.literal("gemini-2.5-flash-preview-05-20"), z.literal("gemini-2.5-flash-preview-04-17"), z.literal("gemini-2.5-flash-lite-preview-06-17"), z.literal("gemini-2.5-pro-exp-03-25"), z.literal("gemini-2.0-flash-lite"), z.literal("gemini-2.0-flash"), z.literal("auto-large"), z.literal("auto-small"), z.literal("auto-micro")]));
export type PureLLMModel = z.infer<typeof ZPureLLMModel>;

export const ZRetabModel = z.lazy(() => z.union([z.literal("auto-large"), z.literal("auto-small"), z.literal("auto-micro")]));
export type RetabModel = z.infer<typeof ZRetabModel>;

export const ZTokenPrice = z.lazy(() => z.object({
    prompt: z.number(),
    completion: z.number(),
    cached_discount: z.number(),
}));
export type TokenPrice = z.infer<typeof ZTokenPrice>;

export const ZAddIterationFromJsonlRequest = z.lazy(() => z.object({
    jsonl_gcs_path: z.string(),
}));
export type AddIterationFromJsonlRequest = z.infer<typeof ZAddIterationFromJsonlRequest>;

export const ZAnnotatedDocument = z.lazy(() => z.object({
    mime_data: ZMIMEData,
    annotation: z.record(z.string(), z.any()),
}));
export type AnnotatedDocument = z.infer<typeof ZAnnotatedDocument>;

export const ZCreateIterationRequest = z.lazy(() => z.object({
    inference_settings: ZInferenceSettings,
    json_schema: z.record(z.string(), z.any()).optional(),
}));
export type CreateIterationRequest = z.infer<typeof ZCreateIterationRequest>;

export const ZDocumentItem = z.lazy(() => z.object({
    mime_data: ZMIMEData,
    annotation: z.record(z.string(), z.any()),
    annotation_metadata: ZPredictionMetadata.optional(),
}).merge(ZAnnotatedDocument.schema));
export type DocumentItem = z.infer<typeof ZDocumentItem>;

export const ZEvaluation = z.lazy(() => z.object({
    id: z.string(),
    updated_at: z.string(),
    name: z.string(),
    old_documents: z.array(ZEvaluationDocument).optional(),
    documents: z.array(ZEvaluationDocument),
    iterations: z.array(ZIteration),
    json_schema: z.record(z.string(), z.any()),
    project_id: z.string(),
    default_inference_settings: ZInferenceSettings.optional(),
}));
export type Evaluation = z.infer<typeof ZEvaluation>;

export const ZEvaluationDocument = z.lazy(() => z.object({
    mime_data: ZMIMEData,
    annotation: z.record(z.string(), z.any()),
    annotation_metadata: ZPredictionMetadata.optional(),
    id: z.string(),
}).merge(ZDocumentItem.schema));
export type EvaluationDocument = z.infer<typeof ZEvaluationDocument>;

export const ZIteration = z.lazy(() => z.object({
    id: z.string(),
    inference_settings: ZInferenceSettings,
    json_schema: z.record(z.string(), z.any()),
    predictions: z.array(ZPredictionData),
    metric_results: ZMetricResult.optional(),
}));
export type Iteration = z.infer<typeof ZIteration>;

export const ZUpdateEvaluationDocumentRequest = z.lazy(() => z.object({
    annotation: z.record(z.string(), z.any()).optional(),
    annotation_metadata: ZPredictionMetadata.optional(),
}));
export type UpdateEvaluationDocumentRequest = z.infer<typeof ZUpdateEvaluationDocumentRequest>;

export const ZUpdateEvaluationRequest = z.lazy(() => z.object({
    name: z.string().optional(),
    documents: z.array(ZEvaluationDocument).optional(),
    iterations: z.array(ZIteration).optional(),
    json_schema: z.record(z.string(), z.any()).optional(),
    project_id: z.string().optional(),
    default_inference_settings: ZInferenceSettings.optional(),
}));
export type UpdateEvaluationRequest = z.infer<typeof ZUpdateEvaluationRequest>;

export const ZBaseIteration = z.lazy(() => z.object({
    id: z.string(),
    inference_settings: ZInferenceSettings,
    json_schema: z.record(z.string(), z.any()),
    updated_at: z.string(),
}));
export type BaseIteration = z.infer<typeof ZBaseIteration>;

export const ZDocumentStatus = z.lazy(() => z.object({
    document_id: z.string(),
    filename: z.string(),
    needs_update: z.boolean(),
    has_prediction: z.boolean(),
    prediction_updated_at: z.string().optional(),
    iteration_updated_at: z.string(),
}));
export type DocumentStatus = z.infer<typeof ZDocumentStatus>;

export const ZIterationDocumentStatusResponse = z.lazy(() => z.object({
    iteration_id: z.string(),
    documents: z.array(ZDocumentStatus),
    total_documents: z.number(),
    documents_needing_update: z.number(),
    documents_up_to_date: z.number(),
}));
export type IterationDocumentStatusResponse = z.infer<typeof ZIterationDocumentStatusResponse>;

export const ZPatchIterationRequest = z.lazy(() => z.object({
    inference_settings: ZInferenceSettings.optional(),
    json_schema: z.record(z.string(), z.any()).optional(),
    version: z.number().optional(),
}));
export type PatchIterationRequest = z.infer<typeof ZPatchIterationRequest>;

export const ZProcessIterationRequest = z.lazy(() => z.object({
    document_ids: z.array(z.string()).optional(),
    only_outdated: z.boolean(),
}));
export type ProcessIterationRequest = z.infer<typeof ZProcessIterationRequest>;

export const ZCreateEvaluationDocumentRequest = z.lazy(() => z.object({
    mime_data: ZMIMEData,
    annotation: z.record(z.string(), z.any()),
    annotation_metadata: ZPredictionMetadata.optional(),
}).merge(ZDocumentItem.schema));
export type CreateEvaluationDocumentRequest = z.infer<typeof ZCreateEvaluationDocumentRequest>;

export const ZPatchEvaluationDocumentRequest = z.lazy(() => z.object({
    annotation: z.record(z.string(), z.any()).optional(),
    annotation_metadata: ZPredictionMetadata.optional(),
}));
export type PatchEvaluationDocumentRequest = z.infer<typeof ZPatchEvaluationDocumentRequest>;

export const ZBaseEvaluation = z.lazy(() => z.object({
    id: z.string(),
    name: z.string(),
    json_schema: z.record(z.string(), z.any()),
    project_id: z.string(),
    default_inference_settings: ZInferenceSettings,
    updated_at: z.string(),
}));
export type BaseEvaluation = z.infer<typeof ZBaseEvaluation>;

export const ZCreateEvaluationRequest = z.lazy(() => z.object({
    name: z.string(),
    project_id: z.string(),
    json_schema: z.record(z.string(), z.any()),
    default_inference_settings: ZInferenceSettings,
}));
export type CreateEvaluationRequest = z.infer<typeof ZCreateEvaluationRequest>;

export const ZListEvaluationParams = z.lazy(() => z.object({
    project_id: z.string().optional(),
    schema_id: z.string().optional(),
    schema_data_id: z.string().optional(),
}));
export type ListEvaluationParams = z.infer<typeof ZListEvaluationParams>;

export const ZPatchEvaluationRequest = z.lazy(() => z.object({
    name: z.string().optional(),
    json_schema: z.record(z.string(), z.any()).optional(),
    project_id: z.string().optional(),
    default_inference_settings: ZInferenceSettings.optional(),
}));
export type PatchEvaluationRequest = z.infer<typeof ZPatchEvaluationRequest>;

export const ZExternalAPIKey = z.lazy(() => z.object({
    provider: z.union([z.literal("OpenAI"), z.literal("Anthropic"), z.literal("Gemini"), z.literal("xAI"), z.literal("Retab")]),
    is_configured: z.boolean(),
    last_updated: z.string().optional(),
}));
export type ExternalAPIKey = z.infer<typeof ZExternalAPIKey>;

export const ZExternalAPIKeyRequest = z.lazy(() => z.object({
    provider: z.union([z.literal("OpenAI"), z.literal("Anthropic"), z.literal("Gemini"), z.literal("xAI"), z.literal("Retab")]),
    api_key: z.string(),
}));
export type ExternalAPIKeyRequest = z.infer<typeof ZExternalAPIKeyRequest>;

export const ZAnnotationInputData = z.lazy(() => z.object({
    data_file: z.string(),
    schema_id: z.string(),
    inference_settings: ZInferenceSettings,
}));
export type AnnotationInputData = z.infer<typeof ZAnnotationInputData>;

export const ZAnnotationModel = z.lazy(() => z.union([z.literal("human"), z.string()]));
export type AnnotationModel = z.infer<typeof ZAnnotationModel>;

export const ZEvaluationInputData = z.lazy(() => z.object({
    eval_data_file: z.string(),
    schema_id: z.string(),
    inference_settings_1: ZInferenceSettings.optional(),
    inference_settings_2: ZInferenceSettings,
}));
export type EvaluationInputData = z.infer<typeof ZEvaluationInputData>;

export const ZFinetuningWorkflowInputData = z.lazy(() => z.object({
    prepare_dataset_input_data: ZPrepareDatasetInputData,
    annotation_model: z.union([z.literal("human"), z.string()]),
    inference_settings: ZInferenceSettings.optional(),
    finetuning_props: ZInferenceSettings,
}));
export type FinetuningWorkflowInputData = z.infer<typeof ZFinetuningWorkflowInputData>;

export const ZPrepareDatasetInputData = z.lazy(() => z.object({
    dataset_id: z.string().optional(),
    schema_id: z.string().optional(),
    schema_data_id: z.string().optional(),
    selection_model: z.union([z.literal("all"), z.literal("manual")]),
}));
export type PrepareDatasetInputData = z.infer<typeof ZPrepareDatasetInputData>;

export const ZStandaloneAnnotationWorkflowInputData = z.lazy(() => z.object({
    data_file: z.string(),
    schema_id: z.string(),
    inference_settings: ZInferenceSettings,
}).merge(ZAnnotationInputData.schema));
export type StandaloneAnnotationWorkflowInputData = z.infer<typeof ZStandaloneAnnotationWorkflowInputData>;

export const ZStandaloneEvaluationWorkflowInputData = z.lazy(() => z.object({
    eval_data_file: z.string(),
    schema_id: z.string(),
    inference_settings_1: ZInferenceSettings.optional(),
    inference_settings_2: ZInferenceSettings,
}).merge(ZEvaluationInputData.schema));
export type StandaloneEvaluationWorkflowInputData = z.infer<typeof ZStandaloneEvaluationWorkflowInputData>;

export const ZWorkflows = z.lazy(() => z.union([z.literal("finetuning-workflow"), z.literal("annotation-workflow"), z.literal("evaluation-workflow")]));
export type Workflows = z.infer<typeof ZWorkflows>;

export const ZChatCompletionMessageParam = z.lazy(() => z.union([ZChatCompletionDeveloperMessageParam, ZChatCompletionSystemMessageParam, ZChatCompletionUserMessageParam, ZChatCompletionAssistantMessageParam, ZChatCompletionToolMessageParam, ZChatCompletionFunctionMessageParam]));
export type ChatCompletionMessageParam = z.infer<typeof ZChatCompletionMessageParam>;

export const ZContentUnionDict = z.lazy(() => z.union([ZContent, z.array(z.union([ZFile, ZPart, z.instanceof(Uint8Array), z.string()])), ZFile, ZPart, z.instanceof(Uint8Array), z.string(), ZContentDict]));
export type ContentUnionDict = z.infer<typeof ZContentUnionDict>;

export const ZMessageParam = z.lazy(() => z.object({
    content: z.union([z.string(), z.array(z.union([ZTextBlockParam, ZImageBlockParam, ZDocumentBlockParam, ZThinkingBlockParam, ZRedactedThinkingBlockParam, ZToolUseBlockParam, ZToolResultBlockParam, ZServerToolUseBlockParam, ZWebSearchToolResultBlockParam, z.union([ZTextBlock, ZThinkingBlock, ZRedactedThinkingBlock, ZToolUseBlock, ZServerToolUseBlock, ZWebSearchToolResultBlock])]))]),
    role: z.union([z.literal("user"), z.literal("assistant")]),
}));
export type MessageParam = z.infer<typeof ZMessageParam>;

export const ZPartialSchema = z.lazy(() => z.object({
    object: z.literal("schema"),
    created_at: z.string(),
    json_schema: z.record(z.string(), z.any()),
    strict: z.boolean(),
}));
export type PartialSchema = z.infer<typeof ZPartialSchema>;

export const ZPartialSchemaChunk = z.lazy(() => z.object({
    streaming_error: ZErrorDetail.optional(),
    object: z.literal("schema.chunk"),
    created_at: z.string(),
    delta_json_schema_flat: z.record(z.string(), z.any()),
}).merge(ZStreamingBaseModel.schema));
export type PartialSchemaChunk = z.infer<typeof ZPartialSchemaChunk>;

export const ZResponseInputItemParam = z.lazy(() => z.union([ZEasyInputMessageParam, ZMessage, ZResponseOutputMessageParam, ZResponseFileSearchToolCallParam, ZResponseComputerToolCallParam, ZComputerCallOutput, ZResponseFunctionWebSearchParam, ZResponseFunctionToolCallParam, ZFunctionCallOutput, ZResponseReasoningItemParam, ZImageGenerationCall, ZResponseCodeInterpreterToolCallParam, ZLocalShellCall, ZLocalShellCallOutput, ZMcpListTools, ZMcpApprovalRequest, ZMcpApprovalResponse, ZMcpCall, ZItemReference]));
export type ResponseInputItemParam = z.infer<typeof ZResponseInputItemParam>;

export const ZGenerateSchemaRequest = z.lazy(() => z.object({
    documents: z.array(ZMIMEData),
    model: z.string(),
    temperature: z.number(),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).optional(),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]),
    instructions: z.string().optional(),
    image_resolution_dpi: z.number(),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]),
    stream: z.boolean(),
}));
export type GenerateSchemaRequest = z.infer<typeof ZGenerateSchemaRequest>;

export const ZGenerateSystemPromptRequest = z.lazy(() => z.object({
    documents: z.array(ZMIMEData),
    model: z.string(),
    temperature: z.number(),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).optional(),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]),
    instructions: z.string().optional(),
    image_resolution_dpi: z.number(),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]),
    stream: z.boolean(),
    json_schema: z.record(z.string(), z.any()),
}).merge(ZGenerateSchemaRequest.schema));
export type GenerateSystemPromptRequest = z.infer<typeof ZGenerateSystemPromptRequest>;

export const ZColumn = z.lazy(() => z.object({
    type: z.literal("column"),
    size: z.number(),
    items: z.array(z.union([ZRow, ZFieldItem, ZRefObject, ZRowList])),
    name: z.string().optional(),
}));
export type Column = z.infer<typeof ZColumn>;

export const ZFieldItem = z.lazy(() => z.object({
    type: z.literal("field"),
    name: z.string(),
    size: z.number().optional(),
}));
export type FieldItem = z.infer<typeof ZFieldItem>;

export const ZLayout = z.lazy(() => z.object({
    defs: z.record(z.string(), ZColumn),
    type: z.literal("column"),
    size: z.number(),
    items: z.array(z.union([ZRow, ZRowList, ZFieldItem, ZRefObject])),
}));
export type Layout = z.infer<typeof ZLayout>;

export const ZRefObject = z.lazy(() => z.object({
    type: z.literal("object"),
    size: z.number().optional(),
    name: z.string().optional(),
    ref: z.string(),
}));
export type RefObject = z.infer<typeof ZRefObject>;

export const ZRow = z.lazy(() => z.object({
    type: z.literal("row"),
    name: z.string().optional(),
    items: z.array(z.union([ZColumn, ZFieldItem, ZRefObject])),
}));
export type Row = z.infer<typeof ZRow>;

export const ZRowList = z.lazy(() => z.object({
    type: z.literal("rowList"),
    name: z.string().optional(),
    items: z.array(z.union([ZColumn, ZFieldItem, ZRefObject])),
}));
export type RowList = z.infer<typeof ZRowList>;

export const ZEnhanceSchemaConfig = z.lazy(() => z.object({
    allow_reasoning_fields_added: z.boolean(),
    allow_field_description_update: z.boolean(),
    allow_system_prompt_update: z.boolean(),
    allow_field_simple_type_change: z.boolean(),
    allow_field_data_structure_breakdown: z.boolean(),
}));
export type EnhanceSchemaConfig = z.infer<typeof ZEnhanceSchemaConfig>;

export const ZEnhanceSchemaConfigDict = z.lazy(() => z.object({
    allow_reasoning_fields_added: z.boolean(),
    allow_field_description_update: z.boolean(),
    allow_system_prompt_update: z.boolean(),
    allow_field_simple_type_change: z.boolean(),
    allow_field_data_structure_breakdown: z.boolean(),
}));
export type EnhanceSchemaConfigDict = z.infer<typeof ZEnhanceSchemaConfigDict>;

export const ZEnhanceSchemaRequest = z.lazy(() => z.object({
    documents: z.array(ZMIMEData),
    ground_truths: z.array(z.record(z.string(), z.any())).optional(),
    model: z.string(),
    temperature: z.number(),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).optional(),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]),
    image_resolution_dpi: z.number(),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]),
    stream: z.boolean(),
    tools_config: ZEnhanceSchemaConfig,
    json_schema: z.record(z.string(), z.any()),
    instructions: z.string().optional(),
    flat_likelihoods: z.union([z.array(z.record(z.string(), z.number())), z.record(z.string(), z.number())]).optional(),
}));
export type EnhanceSchemaRequest = z.infer<typeof ZEnhanceSchemaRequest>;

export const ZUpdateTemplateRequest = z.lazy(() => z.object({
    id: z.string(),
    name: z.string().optional(),
    json_schema: z.record(z.string(), z.any()).optional(),
    python_code: z.string().optional(),
    sample_document: ZMIMEData.optional(),
}));
export type UpdateTemplateRequest = z.infer<typeof ZUpdateTemplateRequest>;

export const ZEvaluateSchemaRequest = z.lazy(() => z.object({
    documents: z.array(ZMIMEData),
    ground_truths: z.array(z.record(z.string(), z.any())).optional(),
    model: z.string(),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).optional(),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]),
    image_resolution_dpi: z.number(),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]),
    n_consensus: z.number(),
    json_schema: z.record(z.string(), z.any()),
}));
export type EvaluateSchemaRequest = z.infer<typeof ZEvaluateSchemaRequest>;

export const ZEvaluateSchemaResponse = z.lazy(() => z.object({
    item_metrics: z.array(ZItemMetric),
}));
export type EvaluateSchemaResponse = z.infer<typeof ZEvaluateSchemaResponse>;

export const ZBinaryIO = z.lazy(() => z.instanceof(Uint8Array));
export type BinaryIO = z.infer<typeof ZBinaryIO>;

export const ZDBFile = z.lazy(() => z.object({
    object: z.literal("file"),
    id: z.string(),
    filename: z.string(),
}));
export type DBFile = z.infer<typeof ZDBFile>;

export const ZFileData = z.lazy(() => z.tuple([z.string(), z.instanceof(Uint8Array), z.string()]));
export type FileData = z.infer<typeof ZFileData>;

export const ZFileLink = z.lazy(() => z.object({
    download_url: z.string(),
    expires_in: z.string(),
    filename: z.string(),
}));
export type FileLink = z.infer<typeof ZFileLink>;

export const ZFileTuple = z.lazy(() => z.tuple([z.string(), z.tuple([z.string(), z.instanceof(Uint8Array), z.string()])]));
export type FileTuple = z.infer<typeof ZFileTuple>;

export const ZAnnotation = z.lazy(() => z.object({
    file_id: z.string(),
    parameters: ZAnnotationParameters,
    data: z.record(z.string(), z.any()),
    schema_id: z.string(),
    organization_id: z.string(),
    updated_at: z.string(),
}));
export type Annotation = z.infer<typeof ZAnnotation>;

export const ZAnnotationParameters = z.lazy(() => z.object({
    model: z.string(),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]).optional(),
    image_resolution_dpi: z.number(),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]),
    temperature: z.number(),
}));
export type AnnotationParameters = z.infer<typeof ZAnnotationParameters>;

export const ZCronSchedule = z.lazy(() => z.object({
    second: z.number().optional(),
    minute: z.number(),
    hour: z.number(),
    day_of_month: z.number().optional(),
    month: z.number().optional(),
    day_of_week: z.number().optional(),
}));
export type CronSchedule = z.infer<typeof ZCronSchedule>;

export const ZScrappingConfig = z.lazy(() => z.object({
    id: z.string(),
    name: z.string(),
    processor_id: z.string(),
    updated_at: z.string(),
    default_language: z.string(),
    webhook_url: z.string(),
    webhook_headers: z.record(z.string(), z.string()),
    need_validation: z.boolean(),
    link: z.string(),
    schedule: ZCronSchedule,
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]),
    image_resolution_dpi: z.number(),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]),
    model: z.string(),
    json_schema: z.record(z.string(), z.any()),
    temperature: z.number(),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).optional(),
}).merge(ZAutomationConfig.schema));
export type ScrappingConfig = z.infer<typeof ZScrappingConfig>;

export const ZListMailboxes = z.lazy(() => z.object({
    data: z.array(ZMailbox),
    list_metadata: ZListMetadata,
}));
export type ListMailboxes = z.infer<typeof ZListMailboxes>;

export const ZMailbox = z.lazy(() => z.object({
    id: z.string(),
    name: z.string(),
    processor_id: z.string(),
    updated_at: z.string(),
    default_language: z.string(),
    webhook_url: z.string(),
    webhook_headers: z.record(z.string(), z.string()),
    need_validation: z.boolean(),
    email: z.string(),
    authorized_domains: z.array(z.string()),
    authorized_emails: z.array(z.string().email()),
}).merge(ZAutomationConfig.schema));
export type Mailbox = z.infer<typeof ZMailbox>;

export const ZUpdateMailboxRequest = z.lazy(() => z.object({
    name: z.string().optional(),
    default_language: z.string().optional(),
    webhook_url: z.string().optional(),
    webhook_headers: z.record(z.string(), z.string()).optional(),
    need_validation: z.boolean().optional(),
    authorized_domains: z.array(z.string()).optional(),
    authorized_emails: z.array(z.string().email()).optional(),
}).merge(ZUpdateAutomationRequest.schema));
export type UpdateMailboxRequest = z.infer<typeof ZUpdateMailboxRequest>;

export const ZLink = z.lazy(() => z.object({
    id: z.string(),
    name: z.string(),
    processor_id: z.string(),
    updated_at: z.string(),
    default_language: z.string(),
    webhook_url: z.string(),
    webhook_headers: z.record(z.string(), z.string()),
    need_validation: z.boolean(),
    password: z.string().optional(),
}).merge(ZAutomationConfig.schema));
export type Link = z.infer<typeof ZLink>;

export const ZListLinks = z.lazy(() => z.object({
    data: z.array(ZLink),
    list_metadata: ZListMetadata,
}));
export type ListLinks = z.infer<typeof ZListLinks>;

export const ZUpdateLinkRequest = z.lazy(() => z.object({
    name: z.string().optional(),
    default_language: z.string().optional(),
    webhook_url: z.string().optional(),
    webhook_headers: z.record(z.string(), z.string()).optional(),
    need_validation: z.boolean().optional(),
    password: z.string().optional(),
}).merge(ZUpdateAutomationRequest.schema));
export type UpdateLinkRequest = z.infer<typeof ZUpdateLinkRequest>;

export const ZAutomationLevel = z.lazy(() => z.object({
    distance_threshold: z.number(),
    score_threshold: z.number(),
}));
export type AutomationLevel = z.infer<typeof ZAutomationLevel>;

export const ZFetchParams = z.lazy(() => z.object({
    endpoint: z.string(),
    headers: z.record(z.string(), z.string()),
    name: z.string(),
}));
export type FetchParams = z.infer<typeof ZFetchParams>;

export const ZListOutlooks = z.lazy(() => z.object({
    data: z.array(ZOutlook),
    list_metadata: ZListMetadata,
}));
export type ListOutlooks = z.infer<typeof ZListOutlooks>;

export const ZMatchParams = z.lazy(() => z.object({
    endpoint: z.string(),
    headers: z.record(z.string(), z.string()),
    path: z.string(),
}));
export type MatchParams = z.infer<typeof ZMatchParams>;

export const ZOutlook = z.lazy(() => z.object({
    id: z.string(),
    name: z.string(),
    processor_id: z.string(),
    updated_at: z.string(),
    default_language: z.string(),
    webhook_url: z.string(),
    webhook_headers: z.record(z.string(), z.string()),
    need_validation: z.boolean(),
    authorized_domains: z.array(z.string()),
    authorized_emails: z.array(z.string().email()),
    layout_schema: z.record(z.string(), z.any()).optional(),
    match_params: z.array(ZMatchParams),
    fetch_params: z.array(ZFetchParams),
}).merge(ZAutomationConfig.schema));
export type Outlook = z.infer<typeof ZOutlook>;

export const ZUpdateOutlookRequest = z.lazy(() => z.object({
    name: z.string().optional(),
    default_language: z.string().optional(),
    webhook_url: z.string().optional(),
    webhook_headers: z.record(z.string(), z.string()).optional(),
    need_validation: z.boolean().optional(),
    authorized_domains: z.array(z.string()).optional(),
    authorized_emails: z.array(z.string().email()).optional(),
    match_params: z.array(ZMatchParams).optional(),
    fetch_params: z.array(ZFetchParams).optional(),
    layout_schema: z.record(z.string(), z.any()).optional(),
}).merge(ZUpdateAutomationRequest.schema));
export type UpdateOutlookRequest = z.infer<typeof ZUpdateOutlookRequest>;

export const ZBaseWebhookRequest = z.lazy(() => z.object({
    completion: ZRetabParsedChatCompletion,
    user: z.string().email().optional(),
    file_payload: ZBaseMIMEData,
    metadata: z.record(z.string(), z.any()).optional(),
}));
export type BaseWebhookRequest = z.infer<typeof ZBaseWebhookRequest>;

export const ZWebhookRequest = z.lazy(() => z.object({
    completion: ZRetabParsedChatCompletion,
    user: z.string().email().optional(),
    file_payload: ZMIMEData,
    metadata: z.record(z.string(), z.any()).optional(),
}));
export type WebhookRequest = z.infer<typeof ZWebhookRequest>;

export const ZEndpoint = z.lazy(() => z.object({
    id: z.string(),
    name: z.string(),
    processor_id: z.string(),
    updated_at: z.string(),
    default_language: z.string(),
    webhook_url: z.string(),
    webhook_headers: z.record(z.string(), z.string()),
    need_validation: z.boolean(),
}).merge(ZAutomationConfig.schema));
export type Endpoint = z.infer<typeof ZEndpoint>;

export const ZListEndpoints = z.lazy(() => z.object({
    data: z.array(ZEndpoint),
    list_metadata: ZListMetadata,
}));
export type ListEndpoints = z.infer<typeof ZListEndpoints>;

export const ZUpdateEndpointRequest = z.lazy(() => z.object({
    name: z.string().optional(),
    default_language: z.string().optional(),
    webhook_url: z.string().optional(),
    webhook_headers: z.record(z.string(), z.string()).optional(),
    need_validation: z.boolean().optional(),
}).merge(ZUpdateAutomationRequest.schema));
export type UpdateEndpointRequest = z.infer<typeof ZUpdateEndpointRequest>;

export const ZChatCompletionChunk = z.lazy(() => z.object({
    id: z.string(),
    choices: z.array(ZChoice),
    created: z.number(),
    model: z.string(),
    object: z.literal("chat.completion.chunk"),
    service_tier: z.union([z.literal("auto"), z.literal("default"), z.literal("flex"), z.literal("scale"), z.literal("priority")]).optional(),
    system_fingerprint: z.string().optional(),
    usage: ZCompletionUsage.optional(),
}));
export type ChatCompletionChunk = z.infer<typeof ZChatCompletionChunk>;

export const ZChoiceChunk = z.lazy(() => z.object({
    delta: ZChoiceDelta,
    finish_reason: z.union([z.literal("stop"), z.literal("length"), z.literal("tool_calls"), z.literal("content_filter"), z.literal("function_call")]).optional(),
    index: z.number(),
    logprobs: ZChoiceLogprobs.optional(),
}));
export type ChoiceChunk = z.infer<typeof ZChoiceChunk>;

export const ZChoiceDeltaChunk = z.lazy(() => z.object({
    content: z.string().optional(),
    function_call: ZChoiceDeltaFunctionCall.optional(),
    refusal: z.string().optional(),
    role: z.union([z.literal("developer"), z.literal("system"), z.literal("user"), z.literal("assistant"), z.literal("tool")]).optional(),
    tool_calls: z.array(ZChoiceDeltaToolCall).optional(),
}));
export type ChoiceDeltaChunk = z.infer<typeof ZChoiceDeltaChunk>;

export const ZConsensusModel = z.lazy(() => z.object({
    model: z.string(),
    temperature: z.number(),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).optional(),
}));
export type ConsensusModel = z.infer<typeof ZConsensusModel>;

export const ZDocumentExtractRequest = z.lazy(() => z.object({
    document: ZMIMEData,
    documents: z.array(ZMIMEData),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]),
    image_resolution_dpi: z.number(),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]),
    model: z.string(),
    json_schema: z.record(z.string(), z.any()),
    temperature: z.number(),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).optional(),
    n_consensus: z.number(),
    stream: z.boolean(),
    seed: z.number().optional(),
    store: z.boolean(),
    need_validation: z.boolean(),
}));
export type DocumentExtractRequest = z.infer<typeof ZDocumentExtractRequest>;

export const ZFieldLocation = z.lazy(() => z.object({
    label: z.string(),
    value: z.string(),
    quote: z.string(),
    file_id: z.string().optional(),
    page: z.number().optional(),
    bbox_normalized: z.tuple([z.number(), z.number(), z.number(), z.number()]).optional(),
    score: z.number().optional(),
    match_level: z.union([z.literal("token"), z.literal("line"), z.literal("block")]).optional(),
}));
export type FieldLocation = z.infer<typeof ZFieldLocation>;

export const ZLikelihoodsSource = z.lazy(() => z.union([z.literal("consensus"), z.literal("log_probs")]));
export type LikelihoodsSource = z.infer<typeof ZLikelihoodsSource>;

export const ZLogExtractionRequest = z.lazy(() => z.object({
    messages: z.array(ZChatCompletionRetabMessage).optional(),
    openai_messages: z.array(z.union([ZChatCompletionDeveloperMessageParam, ZChatCompletionSystemMessageParam, ZChatCompletionUserMessageParam, ZChatCompletionAssistantMessageParam, ZChatCompletionToolMessageParam, ZChatCompletionFunctionMessageParam])).optional(),
    openai_responses_input: z.array(z.union([ZEasyInputMessageParam, ZMessage, ZResponseOutputMessageParam, ZResponseFileSearchToolCallParam, ZResponseComputerToolCallParam, ZComputerCallOutput, ZResponseFunctionWebSearchParam, ZResponseFunctionToolCallParam, ZFunctionCallOutput, ZResponseReasoningItemParam, ZImageGenerationCall, ZResponseCodeInterpreterToolCallParam, ZLocalShellCall, ZLocalShellCallOutput, ZMcpListTools, ZMcpApprovalRequest, ZMcpApprovalResponse, ZMcpCall, ZItemReference])).optional(),
    anthropic_messages: z.array(ZMessageParam).optional(),
    anthropic_system_prompt: z.string().optional(),
    document: ZMIMEData,
    completion: z.union([z.record(z.any()), ZRetabParsedChatCompletion, ZMessage, ZParsedChatCompletion, ZChatCompletion]).optional(),
    openai_responses_output: ZResponse.optional(),
    json_schema: z.record(z.string(), z.any()),
    model: z.string(),
    temperature: z.number(),
}));
export type LogExtractionRequest = z.infer<typeof ZLogExtractionRequest>;

export const ZLogExtractionResponse = z.lazy(() => z.object({
    extraction_id: z.string().optional(),
    status: z.union([z.literal("success"), z.literal("error")]),
    error_message: z.string().optional(),
}));
export type LogExtractionResponse = z.infer<typeof ZLogExtractionResponse>;

export const ZMessage = z.lazy(() => z.object({
    id: z.string(),
    content: z.array(z.union([ZTextBlock, ZThinkingBlock, ZRedactedThinkingBlock, ZToolUseBlock, ZServerToolUseBlock, ZWebSearchToolResultBlock])),
    model: z.union([z.union([z.literal("claude-3-7-sonnet-latest"), z.literal("claude-3-7-sonnet-20250219"), z.literal("claude-3-5-haiku-latest"), z.literal("claude-3-5-haiku-20241022"), z.literal("claude-sonnet-4-20250514"), z.literal("claude-sonnet-4-0"), z.literal("claude-4-sonnet-20250514"), z.literal("claude-3-5-sonnet-latest"), z.literal("claude-3-5-sonnet-20241022"), z.literal("claude-3-5-sonnet-20240620"), z.literal("claude-opus-4-0"), z.literal("claude-opus-4-20250514"), z.literal("claude-4-opus-20250514"), z.literal("claude-3-opus-latest"), z.literal("claude-3-opus-20240229"), z.literal("claude-3-sonnet-20240229"), z.literal("claude-3-haiku-20240307"), z.literal("claude-2.1"), z.literal("claude-2.0")]), z.string()]),
    role: z.literal("assistant"),
    stop_reason: z.union([z.literal("end_turn"), z.literal("max_tokens"), z.literal("stop_sequence"), z.literal("tool_use"), z.literal("pause_turn"), z.literal("refusal")]).optional(),
    stop_sequence: z.string().optional(),
    type: z.literal("message"),
    usage: ZUsage,
}));
export type Message = z.infer<typeof ZMessage>;

export const ZParsedChatCompletion = z.lazy(() => z.object({
    id: z.string(),
    choices: z.array(ZParsedChoice),
    created: z.number(),
    model: z.string(),
    object: z.literal("chat.completion"),
    service_tier: z.union([z.literal("auto"), z.literal("default"), z.literal("flex"), z.literal("scale"), z.literal("priority")]).optional(),
    system_fingerprint: z.string().optional(),
    usage: ZCompletionUsage.optional(),
}).merge(ZChatCompletion.schema));
export type ParsedChatCompletion = z.infer<typeof ZParsedChatCompletion>;

export const ZParsedChoice = z.lazy(() => z.object({
    finish_reason: z.union([z.literal("stop"), z.literal("length"), z.literal("tool_calls"), z.literal("content_filter"), z.literal("function_call")]),
    index: z.number(),
    logprobs: ZChoiceLogprobs.optional(),
    message: ZParsedChatCompletionMessage,
}).merge(ZChoice.schema));
export type ParsedChoice = z.infer<typeof ZParsedChoice>;

export const ZResponse = z.lazy(() => z.object({
    id: z.string(),
    created_at: z.number(),
    error: ZResponseError.optional(),
    incomplete_details: ZIncompleteDetails.optional(),
    instructions: z.union([z.string(), z.array(z.union([ZEasyInputMessage, ZMessage, ZResponseOutputMessage, ZResponseFileSearchToolCall, ZResponseComputerToolCall, ZComputerCallOutput, ZResponseFunctionWebSearch, ZResponseFunctionToolCall, ZFunctionCallOutput, ZResponseReasoningItem, ZImageGenerationCall, ZResponseCodeInterpreterToolCall, ZLocalShellCall, ZLocalShellCallOutput, ZMcpListTools, ZMcpApprovalRequest, ZMcpApprovalResponse, ZMcpCall, ZItemReference]))]).optional(),
    metadata: z.record(z.string(), z.string()).optional(),
    model: z.union([z.string(), z.union([z.literal("gpt-4.1"), z.literal("gpt-4.1-mini"), z.literal("gpt-4.1-nano"), z.literal("gpt-4.1-2025-04-14"), z.literal("gpt-4.1-mini-2025-04-14"), z.literal("gpt-4.1-nano-2025-04-14"), z.literal("o4-mini"), z.literal("o4-mini-2025-04-16"), z.literal("o3"), z.literal("o3-2025-04-16"), z.literal("o3-mini"), z.literal("o3-mini-2025-01-31"), z.literal("o1"), z.literal("o1-2024-12-17"), z.literal("o1-preview"), z.literal("o1-preview-2024-09-12"), z.literal("o1-mini"), z.literal("o1-mini-2024-09-12"), z.literal("gpt-4o"), z.literal("gpt-4o-2024-11-20"), z.literal("gpt-4o-2024-08-06"), z.literal("gpt-4o-2024-05-13"), z.literal("gpt-4o-audio-preview"), z.literal("gpt-4o-audio-preview-2024-10-01"), z.literal("gpt-4o-audio-preview-2024-12-17"), z.literal("gpt-4o-audio-preview-2025-06-03"), z.literal("gpt-4o-mini-audio-preview"), z.literal("gpt-4o-mini-audio-preview-2024-12-17"), z.literal("gpt-4o-search-preview"), z.literal("gpt-4o-mini-search-preview"), z.literal("gpt-4o-search-preview-2025-03-11"), z.literal("gpt-4o-mini-search-preview-2025-03-11"), z.literal("chatgpt-4o-latest"), z.literal("codex-mini-latest"), z.literal("gpt-4o-mini"), z.literal("gpt-4o-mini-2024-07-18"), z.literal("gpt-4-turbo"), z.literal("gpt-4-turbo-2024-04-09"), z.literal("gpt-4-0125-preview"), z.literal("gpt-4-turbo-preview"), z.literal("gpt-4-1106-preview"), z.literal("gpt-4-vision-preview"), z.literal("gpt-4"), z.literal("gpt-4-0314"), z.literal("gpt-4-0613"), z.literal("gpt-4-32k"), z.literal("gpt-4-32k-0314"), z.literal("gpt-4-32k-0613"), z.literal("gpt-3.5-turbo"), z.literal("gpt-3.5-turbo-16k"), z.literal("gpt-3.5-turbo-0301"), z.literal("gpt-3.5-turbo-0613"), z.literal("gpt-3.5-turbo-1106"), z.literal("gpt-3.5-turbo-0125"), z.literal("gpt-3.5-turbo-16k-0613")]), z.union([z.literal("o1-pro"), z.literal("o1-pro-2025-03-19"), z.literal("o3-pro"), z.literal("o3-pro-2025-06-10"), z.literal("o3-deep-research"), z.literal("o3-deep-research-2025-06-26"), z.literal("o4-mini-deep-research"), z.literal("o4-mini-deep-research-2025-06-26"), z.literal("computer-use-preview"), z.literal("computer-use-preview-2025-03-11")])]),
    object: z.literal("response"),
    output: z.array(z.union([ZResponseOutputMessage, ZResponseFileSearchToolCall, ZResponseFunctionToolCall, ZResponseFunctionWebSearch, ZResponseComputerToolCall, ZResponseReasoningItem, ZImageGenerationCall, ZResponseCodeInterpreterToolCall, ZLocalShellCall, ZMcpCall, ZMcpListTools, ZMcpApprovalRequest])),
    parallel_tool_calls: z.boolean(),
    temperature: z.number().optional(),
    tool_choice: z.union([z.union([z.literal("none"), z.literal("auto"), z.literal("required")]), ZToolChoiceTypes, ZToolChoiceFunction, ZToolChoiceMcp]),
    tools: z.array(z.union([ZFunctionTool, ZFileSearchTool, ZWebSearchTool, ZComputerTool, ZMcp, ZCodeInterpreter, ZImageGeneration, ZLocalShell])),
    top_p: z.number().optional(),
    background: z.boolean().optional(),
    max_output_tokens: z.number().optional(),
    max_tool_calls: z.number().optional(),
    previous_response_id: z.string().optional(),
    prompt: ZResponsePrompt.optional(),
    reasoning: ZReasoning.optional(),
    service_tier: z.union([z.literal("auto"), z.literal("default"), z.literal("flex"), z.literal("scale"), z.literal("priority")]).optional(),
    status: z.union([z.literal("completed"), z.literal("failed"), z.literal("in_progress"), z.literal("cancelled"), z.literal("queued"), z.literal("incomplete")]).optional(),
    text: ZResponseTextConfig.optional(),
    top_logprobs: z.number().optional(),
    truncation: z.union([z.literal("auto"), z.literal("disabled")]).optional(),
    usage: ZResponseUsage.optional(),
    user: z.string().optional(),
}));
export type Response = z.infer<typeof ZResponse>;

export const ZRetabParsedChatCompletionChunk = z.lazy(() => z.object({
    id: z.string(),
    choices: z.array(ZRetabParsedChoiceChunk),
    created: z.number(),
    model: z.string(),
    object: z.literal("chat.completion.chunk"),
    service_tier: z.union([z.literal("auto"), z.literal("default"), z.literal("flex"), z.literal("scale"), z.literal("priority")]).optional(),
    system_fingerprint: z.string().optional(),
    usage: ZCompletionUsage.optional(),
    streaming_error: ZErrorDetail.optional(),
    extraction_id: z.string().optional(),
    schema_validation_error: ZErrorDetail.optional(),
    request_at: z.string().optional(),
    first_token_at: z.string().optional(),
    last_token_at: z.string().optional(),
}).merge(ZStreamingBaseModel.schema).merge(ZChatCompletionChunk.schema));
export type RetabParsedChatCompletionChunk = z.infer<typeof ZRetabParsedChatCompletionChunk>;

export const ZRetabParsedChoice = z.lazy(() => z.object({
    finish_reason: z.union([z.literal("stop"), z.literal("length"), z.literal("tool_calls"), z.literal("content_filter"), z.literal("function_call")]).optional(),
    index: z.number(),
    logprobs: ZChoiceLogprobs.optional(),
    message: ZParsedChatCompletionMessage,
    field_locations: z.record(z.string(), ZFieldLocation).optional(),
    key_mapping: z.record(z.string(), z.string().optional()).optional(),
}).merge(ZParsedChoice.schema));
export type RetabParsedChoice = z.infer<typeof ZRetabParsedChoice>;

export const ZRetabParsedChoiceChunk = z.lazy(() => z.object({
    delta: ZRetabParsedChoiceDeltaChunk,
    finish_reason: z.union([z.literal("stop"), z.literal("length"), z.literal("tool_calls"), z.literal("content_filter"), z.literal("function_call")]).optional(),
    index: z.number(),
    logprobs: ZChoiceLogprobs.optional(),
}).merge(ZChoice.schema));
export type RetabParsedChoiceChunk = z.infer<typeof ZRetabParsedChoiceChunk>;

export const ZRetabParsedChoiceDeltaChunk = z.lazy(() => z.object({
    content: z.string().optional(),
    function_call: ZChoiceDeltaFunctionCall.optional(),
    refusal: z.string().optional(),
    role: z.union([z.literal("developer"), z.literal("system"), z.literal("user"), z.literal("assistant"), z.literal("tool")]).optional(),
    tool_calls: z.array(ZChoiceDeltaToolCall).optional(),
    flat_likelihoods: z.record(z.string(), z.number()),
    flat_parsed: z.record(z.string(), z.any()),
    flat_deleted_keys: z.array(z.string()),
    field_locations: z.record(z.string(), z.array(ZFieldLocation)).optional(),
    is_valid_json: z.boolean(),
    key_mapping: z.record(z.string(), z.string().optional()).optional(),
}).merge(ZChoiceDelta.schema));
export type RetabParsedChoiceDeltaChunk = z.infer<typeof ZRetabParsedChoiceDeltaChunk>;

export const ZUiResponse = z.lazy(() => z.object({
    id: z.string(),
    created_at: z.number(),
    error: ZResponseError.optional(),
    incomplete_details: ZIncompleteDetails.optional(),
    instructions: z.union([z.string(), z.array(z.union([ZEasyInputMessage, ZMessage, ZResponseOutputMessage, ZResponseFileSearchToolCall, ZResponseComputerToolCall, ZComputerCallOutput, ZResponseFunctionWebSearch, ZResponseFunctionToolCall, ZFunctionCallOutput, ZResponseReasoningItem, ZImageGenerationCall, ZResponseCodeInterpreterToolCall, ZLocalShellCall, ZLocalShellCallOutput, ZMcpListTools, ZMcpApprovalRequest, ZMcpApprovalResponse, ZMcpCall, ZItemReference]))]).optional(),
    metadata: z.record(z.string(), z.string()).optional(),
    model: z.union([z.string(), z.union([z.literal("gpt-4.1"), z.literal("gpt-4.1-mini"), z.literal("gpt-4.1-nano"), z.literal("gpt-4.1-2025-04-14"), z.literal("gpt-4.1-mini-2025-04-14"), z.literal("gpt-4.1-nano-2025-04-14"), z.literal("o4-mini"), z.literal("o4-mini-2025-04-16"), z.literal("o3"), z.literal("o3-2025-04-16"), z.literal("o3-mini"), z.literal("o3-mini-2025-01-31"), z.literal("o1"), z.literal("o1-2024-12-17"), z.literal("o1-preview"), z.literal("o1-preview-2024-09-12"), z.literal("o1-mini"), z.literal("o1-mini-2024-09-12"), z.literal("gpt-4o"), z.literal("gpt-4o-2024-11-20"), z.literal("gpt-4o-2024-08-06"), z.literal("gpt-4o-2024-05-13"), z.literal("gpt-4o-audio-preview"), z.literal("gpt-4o-audio-preview-2024-10-01"), z.literal("gpt-4o-audio-preview-2024-12-17"), z.literal("gpt-4o-audio-preview-2025-06-03"), z.literal("gpt-4o-mini-audio-preview"), z.literal("gpt-4o-mini-audio-preview-2024-12-17"), z.literal("gpt-4o-search-preview"), z.literal("gpt-4o-mini-search-preview"), z.literal("gpt-4o-search-preview-2025-03-11"), z.literal("gpt-4o-mini-search-preview-2025-03-11"), z.literal("chatgpt-4o-latest"), z.literal("codex-mini-latest"), z.literal("gpt-4o-mini"), z.literal("gpt-4o-mini-2024-07-18"), z.literal("gpt-4-turbo"), z.literal("gpt-4-turbo-2024-04-09"), z.literal("gpt-4-0125-preview"), z.literal("gpt-4-turbo-preview"), z.literal("gpt-4-1106-preview"), z.literal("gpt-4-vision-preview"), z.literal("gpt-4"), z.literal("gpt-4-0314"), z.literal("gpt-4-0613"), z.literal("gpt-4-32k"), z.literal("gpt-4-32k-0314"), z.literal("gpt-4-32k-0613"), z.literal("gpt-3.5-turbo"), z.literal("gpt-3.5-turbo-16k"), z.literal("gpt-3.5-turbo-0301"), z.literal("gpt-3.5-turbo-0613"), z.literal("gpt-3.5-turbo-1106"), z.literal("gpt-3.5-turbo-0125"), z.literal("gpt-3.5-turbo-16k-0613")]), z.union([z.literal("o1-pro"), z.literal("o1-pro-2025-03-19"), z.literal("o3-pro"), z.literal("o3-pro-2025-06-10"), z.literal("o3-deep-research"), z.literal("o3-deep-research-2025-06-26"), z.literal("o4-mini-deep-research"), z.literal("o4-mini-deep-research-2025-06-26"), z.literal("computer-use-preview"), z.literal("computer-use-preview-2025-03-11")])]),
    object: z.literal("response"),
    output: z.array(z.union([ZResponseOutputMessage, ZResponseFileSearchToolCall, ZResponseFunctionToolCall, ZResponseFunctionWebSearch, ZResponseComputerToolCall, ZResponseReasoningItem, ZImageGenerationCall, ZResponseCodeInterpreterToolCall, ZLocalShellCall, ZMcpCall, ZMcpListTools, ZMcpApprovalRequest])),
    parallel_tool_calls: z.boolean(),
    temperature: z.number().optional(),
    tool_choice: z.union([z.union([z.literal("none"), z.literal("auto"), z.literal("required")]), ZToolChoiceTypes, ZToolChoiceFunction, ZToolChoiceMcp]),
    tools: z.array(z.union([ZFunctionTool, ZFileSearchTool, ZWebSearchTool, ZComputerTool, ZMcp, ZCodeInterpreter, ZImageGeneration, ZLocalShell])),
    top_p: z.number().optional(),
    background: z.boolean().optional(),
    max_output_tokens: z.number().optional(),
    max_tool_calls: z.number().optional(),
    previous_response_id: z.string().optional(),
    prompt: ZResponsePrompt.optional(),
    reasoning: ZReasoning.optional(),
    service_tier: z.union([z.literal("auto"), z.literal("default"), z.literal("flex"), z.literal("scale"), z.literal("priority")]).optional(),
    status: z.union([z.literal("completed"), z.literal("failed"), z.literal("in_progress"), z.literal("cancelled"), z.literal("queued"), z.literal("incomplete")]).optional(),
    text: ZResponseTextConfig.optional(),
    top_logprobs: z.number().optional(),
    truncation: z.union([z.literal("auto"), z.literal("disabled")]).optional(),
    usage: ZResponseUsage.optional(),
    user: z.string().optional(),
    extraction_id: z.string().optional(),
    likelihoods: z.record(z.string(), z.any()).optional(),
    schema_validation_error: ZErrorDetail.optional(),
    request_at: z.string().optional(),
    first_token_at: z.string().optional(),
    last_token_at: z.string().optional(),
}).merge(ZResponse.schema));
export type UiResponse = z.infer<typeof ZUiResponse>;

export const ZDocumentTransformRequest = z.lazy(() => z.object({
    document: ZMIMEData,
}));
export type DocumentTransformRequest = z.infer<typeof ZDocumentTransformRequest>;

export const ZDocumentTransformResponse = z.lazy(() => z.object({
    document: ZMIMEData,
}));
export type DocumentTransformResponse = z.infer<typeof ZDocumentTransformResponse>;

export const ZParseRequest = z.lazy(() => z.object({
    document: ZMIMEData,
    model: z.union([z.literal("gpt-4o"), z.literal("gpt-4o-mini"), z.literal("chatgpt-4o-latest"), z.literal("gpt-4.1"), z.literal("gpt-4.1-mini"), z.literal("gpt-4.1-mini-2025-04-14"), z.literal("gpt-4.1-2025-04-14"), z.literal("gpt-4.1-nano"), z.literal("gpt-4.1-nano-2025-04-14"), z.literal("gpt-4o-2024-11-20"), z.literal("gpt-4o-2024-08-06"), z.literal("gpt-4o-2024-05-13"), z.literal("gpt-4o-mini-2024-07-18"), z.literal("o1"), z.literal("o1-2024-12-17"), z.literal("o3"), z.literal("o3-2025-04-16"), z.literal("o4-mini"), z.literal("o4-mini-2025-04-16"), z.literal("gpt-4o-audio-preview-2024-12-17"), z.literal("gpt-4o-audio-preview-2024-10-01"), z.literal("gpt-4o-realtime-preview-2024-12-17"), z.literal("gpt-4o-realtime-preview-2024-10-01"), z.literal("gpt-4o-mini-audio-preview-2024-12-17"), z.literal("gpt-4o-mini-realtime-preview-2024-12-17"), z.literal("claude-3-5-sonnet-latest"), z.literal("claude-3-5-sonnet-20241022"), z.literal("claude-3-opus-20240229"), z.literal("claude-3-sonnet-20240229"), z.literal("claude-3-haiku-20240307"), z.literal("grok-3"), z.literal("grok-3-mini"), z.literal("gemini-2.5-pro"), z.literal("gemini-2.5-flash"), z.literal("gemini-2.5-pro-preview-06-05"), z.literal("gemini-2.5-pro-preview-05-06"), z.literal("gemini-2.5-pro-preview-03-25"), z.literal("gemini-2.5-flash-preview-05-20"), z.literal("gemini-2.5-flash-preview-04-17"), z.literal("gemini-2.5-flash-lite-preview-06-17"), z.literal("gemini-2.5-pro-exp-03-25"), z.literal("gemini-2.0-flash-lite"), z.literal("gemini-2.0-flash"), z.literal("auto-large"), z.literal("auto-small"), z.literal("auto-micro"), z.literal("human")]),
    table_parsing_format: z.union([z.literal("markdown"), z.literal("yaml"), z.literal("html"), z.literal("json")]),
    image_resolution_dpi: z.number(),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]),
}));
export type ParseRequest = z.infer<typeof ZParseRequest>;

export const ZParseResult = z.lazy(() => z.object({
    document: ZBaseMIMEData,
    usage: ZRetabUsage,
    pages: z.array(z.string()),
    text: z.string(),
}));
export type ParseResult = z.infer<typeof ZParseResult>;

export const ZRetabUsage = z.lazy(() => z.object({
    page_count: z.number(),
    credits: z.number(),
}));
export type RetabUsage = z.infer<typeof ZRetabUsage>;

export const ZTableParsingFormat = z.lazy(() => z.union([z.literal("markdown"), z.literal("yaml"), z.literal("html"), z.literal("json")]));
export type TableParsingFormat = z.infer<typeof ZTableParsingFormat>;

export const ZDocumentCreateInputRequest = z.lazy(() => z.object({
    document: ZMIMEData,
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]),
    image_resolution_dpi: z.number(),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]),
    json_schema: z.record(z.string(), z.any()),
}).merge(ZDocumentCreateMessageRequest.schema));
export type DocumentCreateInputRequest = z.infer<typeof ZDocumentCreateInputRequest>;

export const ZDocumentCreateMessageRequest = z.lazy(() => z.object({
    document: ZMIMEData,
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]),
    image_resolution_dpi: z.number(),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]),
}));
export type DocumentCreateMessageRequest = z.infer<typeof ZDocumentCreateMessageRequest>;

export const ZDocumentMessage = z.lazy(() => z.object({
    id: z.string(),
    object: z.literal("document_message"),
    messages: z.array(ZChatCompletionRetabMessage),
    created: z.number(),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]),
}));
export type DocumentMessage = z.infer<typeof ZDocumentMessage>;

export const ZMediaType = z.lazy(() => z.union([z.literal("image/jpeg"), z.literal("image/png"), z.literal("image/gif"), z.literal("image/webp")]));
export type MediaType = z.infer<typeof ZMediaType>;

export const ZTokenCount = z.lazy(() => z.object({
    total_tokens: z.number(),
    developer_tokens: z.number(),
    user_tokens: z.number(),
}));
export type TokenCount = z.infer<typeof ZTokenCount>;

export const ZDatasetSplitInputData = z.lazy(() => z.object({
    dataset_id: z.string(),
    train_size: z.union([z.number(), z.number()]).optional(),
    eval_size: z.union([z.number(), z.number()]).optional(),
}));
export type DatasetSplitInputData = z.infer<typeof ZDatasetSplitInputData>;

export const ZSelectionMode = z.lazy(() => z.union([z.literal("all"), z.literal("manual")]));
export type SelectionMode = z.infer<typeof ZSelectionMode>;

export const ZChoice = z.lazy(() => z.object({
    finish_reason: z.union([z.literal("stop"), z.literal("length"), z.literal("tool_calls"), z.literal("content_filter"), z.literal("function_call")]),
    index: z.number(),
    logprobs: ZChoiceLogprobs.optional(),
    message: ZChatCompletionMessage,
}));
export type Choice = z.infer<typeof ZChoice>;

export const ZCompletionUsage = z.lazy(() => z.object({
    completion_tokens: z.number(),
    prompt_tokens: z.number(),
    total_tokens: z.number(),
    completion_tokens_details: ZCompletionTokensDetails.optional(),
    prompt_tokens_details: ZPromptTokensDetails.optional(),
}));
export type CompletionUsage = z.infer<typeof ZCompletionUsage>;

export const ZChatCompletionContentPartTextParam = z.lazy(() => z.object({
    text: z.string(),
    type: z.literal("text"),
}));
export type ChatCompletionContentPartTextParam = z.infer<typeof ZChatCompletionContentPartTextParam>;

export const ZChatCompletionContentPartImageParam = z.lazy(() => z.object({
    image_url: ZImageURL,
    type: z.literal("image_url"),
}));
export type ChatCompletionContentPartImageParam = z.infer<typeof ZChatCompletionContentPartImageParam>;

export const ZChatCompletionContentPartInputAudioParam = z.lazy(() => z.object({
    input_audio: ZInputAudio,
    type: z.literal("input_audio"),
}));
export type ChatCompletionContentPartInputAudioParam = z.infer<typeof ZChatCompletionContentPartInputAudioParam>;

export const ZFile = z.lazy(() => z.object({
    file: ZFileFile,
    type: z.literal("file"),
}));
export type File = z.infer<typeof ZFile>;

export const ZTokenCounts = z.lazy(() => z.object({
    prompt_regular_text: z.number(),
    prompt_cached_text: z.number(),
    prompt_audio: z.number(),
    completion_regular_text: z.number(),
    completion_audio: z.number(),
    total_tokens: z.number(),
}));
export type TokenCounts = z.infer<typeof ZTokenCounts>;

export const ZJSONSchema = z.lazy(() => z.object({
    name: z.string(),
    description: z.string(),
    schema: z.record(z.string(), z.object({}).passthrough()),
    strict: z.boolean().optional(),
}));
export type JSONSchema = z.infer<typeof ZJSONSchema>;

export const ZEasyInputMessageParam = z.lazy(() => z.object({
    content: z.union([z.string(), z.array(z.union([ZResponseInputTextParam, ZResponseInputImageParam, ZResponseInputFileParam]))]),
    role: z.union([z.literal("user"), z.literal("assistant"), z.literal("system"), z.literal("developer")]),
    type: z.literal("message"),
}));
export type EasyInputMessageParam = z.infer<typeof ZEasyInputMessageParam>;

export const ZResponseOutputMessageParam = z.lazy(() => z.object({
    id: z.string(),
    content: z.array(z.union([ZResponseOutputTextParam, ZResponseOutputRefusalParam])),
    role: z.literal("assistant"),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
    type: z.literal("message"),
}));
export type ResponseOutputMessageParam = z.infer<typeof ZResponseOutputMessageParam>;

export const ZResponseFileSearchToolCallParam = z.lazy(() => z.object({
    id: z.string(),
    queries: z.array(z.string()),
    status: z.union([z.literal("in_progress"), z.literal("searching"), z.literal("completed"), z.literal("incomplete"), z.literal("failed")]),
    type: z.literal("file_search_call"),
    results: z.array(ZResult).optional(),
}));
export type ResponseFileSearchToolCallParam = z.infer<typeof ZResponseFileSearchToolCallParam>;

export const ZResponseComputerToolCallParam = z.lazy(() => z.object({
    id: z.string(),
    action: z.union([ZActionClick, ZActionDoubleClick, ZActionDrag, ZActionKeypress, ZActionMove, ZActionScreenshot, ZActionScroll, ZActionType, ZActionWait]),
    call_id: z.string(),
    pending_safety_checks: z.array(ZPendingSafetyCheck),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
    type: z.literal("computer_call"),
}));
export type ResponseComputerToolCallParam = z.infer<typeof ZResponseComputerToolCallParam>;

export const ZComputerCallOutput = z.lazy(() => z.object({
    call_id: z.string(),
    output: ZResponseComputerToolCallOutputScreenshotParam,
    type: z.literal("computer_call_output"),
    id: z.string().optional(),
    acknowledged_safety_checks: z.array(ZComputerCallOutputAcknowledgedSafetyCheck).optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).optional(),
}));
export type ComputerCallOutput = z.infer<typeof ZComputerCallOutput>;

export const ZResponseFunctionWebSearchParam = z.lazy(() => z.object({
    id: z.string(),
    action: z.union([ZActionSearch, ZActionOpenPage, ZActionFind]),
    status: z.union([z.literal("in_progress"), z.literal("searching"), z.literal("completed"), z.literal("failed")]),
    type: z.literal("web_search_call"),
}));
export type ResponseFunctionWebSearchParam = z.infer<typeof ZResponseFunctionWebSearchParam>;

export const ZResponseFunctionToolCallParam = z.lazy(() => z.object({
    arguments: z.string(),
    call_id: z.string(),
    name: z.string(),
    type: z.literal("function_call"),
    id: z.string(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
}));
export type ResponseFunctionToolCallParam = z.infer<typeof ZResponseFunctionToolCallParam>;

export const ZFunctionCallOutput = z.lazy(() => z.object({
    call_id: z.string(),
    output: z.string(),
    type: z.literal("function_call_output"),
    id: z.string().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).optional(),
}));
export type FunctionCallOutput = z.infer<typeof ZFunctionCallOutput>;

export const ZResponseReasoningItemParam = z.lazy(() => z.object({
    id: z.string(),
    summary: z.array(ZSummary),
    type: z.literal("reasoning"),
    encrypted_content: z.string().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
}));
export type ResponseReasoningItemParam = z.infer<typeof ZResponseReasoningItemParam>;

export const ZImageGenerationCall = z.lazy(() => z.object({
    id: z.string(),
    result: z.string().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("generating"), z.literal("failed")]),
    type: z.literal("image_generation_call"),
}));
export type ImageGenerationCall = z.infer<typeof ZImageGenerationCall>;

export const ZResponseCodeInterpreterToolCallParam = z.lazy(() => z.object({
    id: z.string(),
    code: z.string().optional(),
    container_id: z.string(),
    outputs: z.array(z.union([ZOutputLogs, ZOutputImage])).optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete"), z.literal("interpreting"), z.literal("failed")]),
    type: z.literal("code_interpreter_call"),
}));
export type ResponseCodeInterpreterToolCallParam = z.infer<typeof ZResponseCodeInterpreterToolCallParam>;

export const ZLocalShellCall = z.lazy(() => z.object({
    id: z.string(),
    action: ZLocalShellCallAction,
    call_id: z.string(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
    type: z.literal("local_shell_call"),
}));
export type LocalShellCall = z.infer<typeof ZLocalShellCall>;

export const ZLocalShellCallOutput = z.lazy(() => z.object({
    id: z.string(),
    output: z.string(),
    type: z.literal("local_shell_call_output"),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).optional(),
}));
export type LocalShellCallOutput = z.infer<typeof ZLocalShellCallOutput>;

export const ZMcpListTools = z.lazy(() => z.object({
    id: z.string(),
    server_label: z.string(),
    tools: z.array(ZMcpListToolsTool),
    type: z.literal("mcp_list_tools"),
    error: z.string().optional(),
}));
export type McpListTools = z.infer<typeof ZMcpListTools>;

export const ZMcpApprovalRequest = z.lazy(() => z.object({
    id: z.string(),
    arguments: z.string(),
    name: z.string(),
    server_label: z.string(),
    type: z.literal("mcp_approval_request"),
}));
export type McpApprovalRequest = z.infer<typeof ZMcpApprovalRequest>;

export const ZMcpApprovalResponse = z.lazy(() => z.object({
    approval_request_id: z.string(),
    approve: z.boolean(),
    type: z.literal("mcp_approval_response"),
    id: z.string().optional(),
    reason: z.string().optional(),
}));
export type McpApprovalResponse = z.infer<typeof ZMcpApprovalResponse>;

export const ZMcpCall = z.lazy(() => z.object({
    id: z.string(),
    arguments: z.string(),
    name: z.string(),
    server_label: z.string(),
    type: z.literal("mcp_call"),
    error: z.string().optional(),
    output: z.string().optional(),
}));
export type McpCall = z.infer<typeof ZMcpCall>;

export const ZItemReference = z.lazy(() => z.object({
    id: z.string(),
    type: z.literal("item_reference").optional(),
}));
export type ItemReference = z.infer<typeof ZItemReference>;

export const ZResponseFormatText = z.lazy(() => z.object({
    type: z.literal("text"),
}));
export type ResponseFormatText = z.infer<typeof ZResponseFormatText>;

export const ZResponseFormatTextJSONSchemaConfigParam = z.lazy(() => z.object({
    name: z.string(),
    schema: z.record(z.string(), z.object({}).passthrough()),
    type: z.literal("json_schema"),
    description: z.string(),
    strict: z.boolean().optional(),
}));
export type ResponseFormatTextJSONSchemaConfigParam = z.infer<typeof ZResponseFormatTextJSONSchemaConfigParam>;

export const ZResponseFormatJSONObject = z.lazy(() => z.object({
    type: z.literal("json_object"),
}));
export type ResponseFormatJSONObject = z.infer<typeof ZResponseFormatJSONObject>;

export const ZChatCompletionDeveloperMessageParam = z.lazy(() => z.object({
    content: z.union([z.string(), z.array(ZChatCompletionContentPartTextParam)]),
    role: z.literal("developer"),
    name: z.string(),
}));
export type ChatCompletionDeveloperMessageParam = z.infer<typeof ZChatCompletionDeveloperMessageParam>;

export const ZChatCompletionSystemMessageParam = z.lazy(() => z.object({
    content: z.union([z.string(), z.array(ZChatCompletionContentPartTextParam)]),
    role: z.literal("system"),
    name: z.string(),
}));
export type ChatCompletionSystemMessageParam = z.infer<typeof ZChatCompletionSystemMessageParam>;

export const ZChatCompletionUserMessageParam = z.lazy(() => z.object({
    content: z.union([z.string(), z.array(z.union([ZChatCompletionContentPartTextParam, ZChatCompletionContentPartImageParam, ZChatCompletionContentPartInputAudioParam, ZFile]))]),
    role: z.literal("user"),
    name: z.string(),
}));
export type ChatCompletionUserMessageParam = z.infer<typeof ZChatCompletionUserMessageParam>;

export const ZChatCompletionAssistantMessageParam = z.lazy(() => z.object({
    role: z.literal("assistant"),
    audio: ZAudio.optional(),
    content: z.union([z.string(), z.array(z.union([ZChatCompletionContentPartTextParam, ZChatCompletionContentPartRefusalParam]))]).optional(),
    function_call: ZFunctionCall.optional(),
    name: z.string(),
    refusal: z.string().optional(),
    tool_calls: z.array(ZChatCompletionMessageToolCallParam),
}));
export type ChatCompletionAssistantMessageParam = z.infer<typeof ZChatCompletionAssistantMessageParam>;

export const ZChatCompletionToolMessageParam = z.lazy(() => z.object({
    content: z.union([z.string(), z.array(ZChatCompletionContentPartTextParam)]),
    role: z.literal("tool"),
    tool_call_id: z.string(),
}));
export type ChatCompletionToolMessageParam = z.infer<typeof ZChatCompletionToolMessageParam>;

export const ZChatCompletionFunctionMessageParam = z.lazy(() => z.object({
    content: z.string().optional(),
    name: z.string(),
    role: z.literal("function"),
}));
export type ChatCompletionFunctionMessageParam = z.infer<typeof ZChatCompletionFunctionMessageParam>;

export const ZContent = z.lazy(() => z.object({
    parts: z.array(ZPart).optional(),
    role: z.string().optional(),
}));
export type Content = z.infer<typeof ZContent>;

export const ZPart = z.lazy(() => z.object({
    video_metadata: ZVideoMetadata.optional(),
    thought: z.boolean().optional(),
    inline_data: ZBlob.optional(),
    file_data: ZFileData.optional(),
    thought_signature: z.instanceof(Uint8Array).optional(),
    code_execution_result: ZCodeExecutionResult.optional(),
    executable_code: ZExecutableCode.optional(),
    function_call: ZFunctionCall.optional(),
    function_response: ZFunctionResponse.optional(),
    text: z.string().optional(),
}));
export type Part = z.infer<typeof ZPart>;

export const ZContentDict = z.lazy(() => z.object({
    parts: z.array(ZPartDict).optional(),
    role: z.string().optional(),
}));
export type ContentDict = z.infer<typeof ZContentDict>;

export const ZTextBlockParam = z.lazy(() => z.object({
    text: z.string(),
    type: z.literal("text"),
    cache_control: ZCacheControlEphemeralParam.optional(),
    citations: z.array(z.union([ZCitationCharLocationParam, ZCitationPageLocationParam, ZCitationContentBlockLocationParam, ZCitationWebSearchResultLocationParam])).optional(),
}));
export type TextBlockParam = z.infer<typeof ZTextBlockParam>;

export const ZImageBlockParam = z.lazy(() => z.object({
    source: z.union([ZBase64ImageSourceParam, ZURLImageSourceParam]),
    type: z.literal("image"),
    cache_control: ZCacheControlEphemeralParam.optional(),
}));
export type ImageBlockParam = z.infer<typeof ZImageBlockParam>;

export const ZDocumentBlockParam = z.lazy(() => z.object({
    source: z.union([ZBase64PDFSourceParam, ZPlainTextSourceParam, ZContentBlockSourceParam, ZURLPDFSourceParam]),
    type: z.literal("document"),
    cache_control: ZCacheControlEphemeralParam.optional(),
    citations: ZCitationsConfigParam,
    context: z.string().optional(),
    title: z.string().optional(),
}));
export type DocumentBlockParam = z.infer<typeof ZDocumentBlockParam>;

export const ZThinkingBlockParam = z.lazy(() => z.object({
    signature: z.string(),
    thinking: z.string(),
    type: z.literal("thinking"),
}));
export type ThinkingBlockParam = z.infer<typeof ZThinkingBlockParam>;

export const ZRedactedThinkingBlockParam = z.lazy(() => z.object({
    data: z.string(),
    type: z.literal("redacted_thinking"),
}));
export type RedactedThinkingBlockParam = z.infer<typeof ZRedactedThinkingBlockParam>;

export const ZToolUseBlockParam = z.lazy(() => z.object({
    id: z.string(),
    input: z.object({}).passthrough(),
    name: z.string(),
    type: z.literal("tool_use"),
    cache_control: ZCacheControlEphemeralParam.optional(),
}));
export type ToolUseBlockParam = z.infer<typeof ZToolUseBlockParam>;

export const ZToolResultBlockParam = z.lazy(() => z.object({
    tool_use_id: z.string(),
    type: z.literal("tool_result"),
    cache_control: ZCacheControlEphemeralParam.optional(),
    content: z.union([z.string(), z.array(z.union([ZTextBlockParam, ZImageBlockParam]))]),
    is_error: z.boolean(),
}));
export type ToolResultBlockParam = z.infer<typeof ZToolResultBlockParam>;

export const ZServerToolUseBlockParam = z.lazy(() => z.object({
    id: z.string(),
    input: z.object({}).passthrough(),
    name: z.literal("web_search"),
    type: z.literal("server_tool_use"),
    cache_control: ZCacheControlEphemeralParam.optional(),
}));
export type ServerToolUseBlockParam = z.infer<typeof ZServerToolUseBlockParam>;

export const ZWebSearchToolResultBlockParam = z.lazy(() => z.object({
    content: z.union([z.array(ZWebSearchResultBlockParam), ZWebSearchToolRequestErrorParam]),
    tool_use_id: z.string(),
    type: z.literal("web_search_tool_result"),
    cache_control: ZCacheControlEphemeralParam.optional(),
}));
export type WebSearchToolResultBlockParam = z.infer<typeof ZWebSearchToolResultBlockParam>;

export const ZTextBlock = z.lazy(() => z.object({
    citations: z.array(z.union([ZCitationCharLocation, ZCitationPageLocation, ZCitationContentBlockLocation, ZCitationsWebSearchResultLocation])).optional(),
    text: z.string(),
    type: z.literal("text"),
}));
export type TextBlock = z.infer<typeof ZTextBlock>;

export const ZThinkingBlock = z.lazy(() => z.object({
    signature: z.string(),
    thinking: z.string(),
    type: z.literal("thinking"),
}));
export type ThinkingBlock = z.infer<typeof ZThinkingBlock>;

export const ZRedactedThinkingBlock = z.lazy(() => z.object({
    data: z.string(),
    type: z.literal("redacted_thinking"),
}));
export type RedactedThinkingBlock = z.infer<typeof ZRedactedThinkingBlock>;

export const ZToolUseBlock = z.lazy(() => z.object({
    id: z.string(),
    input: z.object({}).passthrough(),
    name: z.string(),
    type: z.literal("tool_use"),
}));
export type ToolUseBlock = z.infer<typeof ZToolUseBlock>;

export const ZServerToolUseBlock = z.lazy(() => z.object({
    id: z.string(),
    input: z.object({}).passthrough(),
    name: z.literal("web_search"),
    type: z.literal("server_tool_use"),
}));
export type ServerToolUseBlock = z.infer<typeof ZServerToolUseBlock>;

export const ZWebSearchToolResultBlock = z.lazy(() => z.object({
    content: z.union([ZWebSearchToolResultError, z.array(ZWebSearchResultBlock)]),
    tool_use_id: z.string(),
    type: z.literal("web_search_tool_result"),
}));
export type WebSearchToolResultBlock = z.infer<typeof ZWebSearchToolResultBlock>;

export const ZChoiceDelta = z.lazy(() => z.object({
    content: z.string().optional(),
    function_call: ZChoiceDeltaFunctionCall.optional(),
    refusal: z.string().optional(),
    role: z.union([z.literal("developer"), z.literal("system"), z.literal("user"), z.literal("assistant"), z.literal("tool")]).optional(),
    tool_calls: z.array(ZChoiceDeltaToolCall).optional(),
}));
export type ChoiceDelta = z.infer<typeof ZChoiceDelta>;

export const ZChoiceLogprobs = z.lazy(() => z.object({
    content: z.array(ZChatCompletionTokenLogprob).optional(),
    refusal: z.array(ZChatCompletionTokenLogprob).optional(),
}));
export type ChoiceLogprobs = z.infer<typeof ZChoiceLogprobs>;

export const ZChoiceDeltaFunctionCall = z.lazy(() => z.object({
    arguments: z.string().optional(),
    name: z.string().optional(),
}));
export type ChoiceDeltaFunctionCall = z.infer<typeof ZChoiceDeltaFunctionCall>;

export const ZChoiceDeltaToolCall = z.lazy(() => z.object({
    index: z.number(),
    id: z.string().optional(),
    function: ZChoiceDeltaToolCallFunction.optional(),
    type: z.literal("function").optional(),
}));
export type ChoiceDeltaToolCall = z.infer<typeof ZChoiceDeltaToolCall>;

export const ZUsage = z.lazy(() => z.object({
    cache_creation_input_tokens: z.number().optional(),
    cache_read_input_tokens: z.number().optional(),
    input_tokens: z.number(),
    output_tokens: z.number(),
    server_tool_use: ZServerToolUsage.optional(),
    service_tier: z.union([z.literal("standard"), z.literal("priority"), z.literal("batch")]).optional(),
}));
export type Usage = z.infer<typeof ZUsage>;

export const ZParsedChatCompletionMessage = z.lazy(() => z.object({
    content: z.string().optional(),
    refusal: z.string().optional(),
    role: z.literal("assistant"),
    annotations: z.array(ZAnnotation).optional(),
    audio: ZChatCompletionAudio.optional(),
    function_call: ZFunctionCall.optional(),
    tool_calls: z.array(ZParsedFunctionToolCall).optional(),
    parsed: z.any().optional(),
}).merge(ZChatCompletionMessage.schema));
export type ParsedChatCompletionMessage = z.infer<typeof ZParsedChatCompletionMessage>;

export const ZResponseError = z.lazy(() => z.object({
    code: z.union([z.literal("server_error"), z.literal("rate_limit_exceeded"), z.literal("invalid_prompt"), z.literal("vector_store_timeout"), z.literal("invalid_image"), z.literal("invalid_image_format"), z.literal("invalid_base64_image"), z.literal("invalid_image_url"), z.literal("image_too_large"), z.literal("image_too_small"), z.literal("image_parse_error"), z.literal("image_content_policy_violation"), z.literal("invalid_image_mode"), z.literal("image_file_too_large"), z.literal("unsupported_image_media_type"), z.literal("empty_image_file"), z.literal("failed_to_download_image"), z.literal("image_file_not_found")]),
    message: z.string(),
}));
export type ResponseError = z.infer<typeof ZResponseError>;

export const ZIncompleteDetails = z.lazy(() => z.object({
    reason: z.union([z.literal("max_output_tokens"), z.literal("content_filter")]).optional(),
}));
export type IncompleteDetails = z.infer<typeof ZIncompleteDetails>;

export const ZEasyInputMessage = z.lazy(() => z.object({
    content: z.union([z.string(), z.array(z.union([ZResponseInputText, ZResponseInputImage, ZResponseInputFile]))]),
    role: z.union([z.literal("user"), z.literal("assistant"), z.literal("system"), z.literal("developer")]),
    type: z.literal("message").optional(),
}));
export type EasyInputMessage = z.infer<typeof ZEasyInputMessage>;

export const ZResponseOutputMessage = z.lazy(() => z.object({
    id: z.string(),
    content: z.array(z.union([ZResponseOutputText, ZResponseOutputRefusal])),
    role: z.literal("assistant"),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
    type: z.literal("message"),
}));
export type ResponseOutputMessage = z.infer<typeof ZResponseOutputMessage>;

export const ZResponseFileSearchToolCall = z.lazy(() => z.object({
    id: z.string(),
    queries: z.array(z.string()),
    status: z.union([z.literal("in_progress"), z.literal("searching"), z.literal("completed"), z.literal("incomplete"), z.literal("failed")]),
    type: z.literal("file_search_call"),
    results: z.array(ZResult).optional(),
}));
export type ResponseFileSearchToolCall = z.infer<typeof ZResponseFileSearchToolCall>;

export const ZResponseComputerToolCall = z.lazy(() => z.object({
    id: z.string(),
    action: z.union([ZActionClick, ZActionDoubleClick, ZActionDrag, ZActionKeypress, ZActionMove, ZActionScreenshot, ZActionScroll, ZActionType, ZActionWait]),
    call_id: z.string(),
    pending_safety_checks: z.array(ZPendingSafetyCheck),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
    type: z.literal("computer_call"),
}));
export type ResponseComputerToolCall = z.infer<typeof ZResponseComputerToolCall>;

export const ZResponseFunctionWebSearch = z.lazy(() => z.object({
    id: z.string(),
    action: z.union([ZActionSearch, ZActionOpenPage, ZActionFind]),
    status: z.union([z.literal("in_progress"), z.literal("searching"), z.literal("completed"), z.literal("failed")]),
    type: z.literal("web_search_call"),
}));
export type ResponseFunctionWebSearch = z.infer<typeof ZResponseFunctionWebSearch>;

export const ZResponseFunctionToolCall = z.lazy(() => z.object({
    arguments: z.string(),
    call_id: z.string(),
    name: z.string(),
    type: z.literal("function_call"),
    id: z.string().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).optional(),
}));
export type ResponseFunctionToolCall = z.infer<typeof ZResponseFunctionToolCall>;

export const ZResponseReasoningItem = z.lazy(() => z.object({
    id: z.string(),
    summary: z.array(ZSummary),
    type: z.literal("reasoning"),
    encrypted_content: z.string().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).optional(),
}));
export type ResponseReasoningItem = z.infer<typeof ZResponseReasoningItem>;

export const ZResponseCodeInterpreterToolCall = z.lazy(() => z.object({
    id: z.string(),
    code: z.string().optional(),
    container_id: z.string(),
    outputs: z.array(z.union([ZOutputLogs, ZOutputImage])).optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete"), z.literal("interpreting"), z.literal("failed")]),
    type: z.literal("code_interpreter_call"),
}));
export type ResponseCodeInterpreterToolCall = z.infer<typeof ZResponseCodeInterpreterToolCall>;

export const ZToolChoiceTypes = z.lazy(() => z.object({
    type: z.union([z.literal("file_search"), z.literal("web_search_preview"), z.literal("computer_use_preview"), z.literal("web_search_preview_2025_03_11"), z.literal("image_generation"), z.literal("code_interpreter")]),
}));
export type ToolChoiceTypes = z.infer<typeof ZToolChoiceTypes>;

export const ZToolChoiceFunction = z.lazy(() => z.object({
    name: z.string(),
    type: z.literal("function"),
}));
export type ToolChoiceFunction = z.infer<typeof ZToolChoiceFunction>;

export const ZToolChoiceMcp = z.lazy(() => z.object({
    server_label: z.string(),
    type: z.literal("mcp"),
    name: z.string().optional(),
}));
export type ToolChoiceMcp = z.infer<typeof ZToolChoiceMcp>;

export const ZFunctionTool = z.lazy(() => z.object({
    name: z.string(),
    parameters: z.record(z.string(), z.object({}).passthrough()).optional(),
    strict: z.boolean().optional(),
    type: z.literal("function"),
    description: z.string().optional(),
}));
export type FunctionTool = z.infer<typeof ZFunctionTool>;

export const ZFileSearchTool = z.lazy(() => z.object({
    type: z.literal("file_search"),
    vector_store_ids: z.array(z.string()),
    filters: z.union([ZComparisonFilter, ZCompoundFilter]).optional(),
    max_num_results: z.number().optional(),
    ranking_options: ZRankingOptions.optional(),
}));
export type FileSearchTool = z.infer<typeof ZFileSearchTool>;

export const ZWebSearchTool = z.lazy(() => z.object({
    type: z.union([z.literal("web_search_preview"), z.literal("web_search_preview_2025_03_11")]),
    search_context_size: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).optional(),
    user_location: ZUserLocation.optional(),
}));
export type WebSearchTool = z.infer<typeof ZWebSearchTool>;

export const ZComputerTool = z.lazy(() => z.object({
    display_height: z.number(),
    display_width: z.number(),
    environment: z.union([z.literal("windows"), z.literal("mac"), z.literal("linux"), z.literal("ubuntu"), z.literal("browser")]),
    type: z.literal("computer_use_preview"),
}));
export type ComputerTool = z.infer<typeof ZComputerTool>;

export const ZMcp = z.lazy(() => z.object({
    server_label: z.string(),
    server_url: z.string(),
    type: z.literal("mcp"),
    allowed_tools: z.union([z.array(z.string()), ZMcpAllowedToolsMcpAllowedToolsFilter]).optional(),
    headers: z.record(z.string(), z.string()).optional(),
    require_approval: z.union([ZMcpRequireApprovalMcpToolApprovalFilter, z.union([z.literal("always"), z.literal("never")])]).optional(),
    server_description: z.string().optional(),
}));
export type Mcp = z.infer<typeof ZMcp>;

export const ZCodeInterpreter = z.lazy(() => z.object({
    container: z.union([z.string(), ZCodeInterpreterContainerCodeInterpreterToolAuto]),
    type: z.literal("code_interpreter"),
}));
export type CodeInterpreter = z.infer<typeof ZCodeInterpreter>;

export const ZImageGeneration = z.lazy(() => z.object({
    type: z.literal("image_generation"),
    background: z.union([z.literal("transparent"), z.literal("opaque"), z.literal("auto")]).optional(),
    input_image_mask: ZImageGenerationInputImageMask.optional(),
    model: z.literal("gpt-image-1").optional(),
    moderation: z.union([z.literal("auto"), z.literal("low")]).optional(),
    output_compression: z.number().optional(),
    output_format: z.union([z.literal("png"), z.literal("webp"), z.literal("jpeg")]).optional(),
    partial_images: z.number().optional(),
    quality: z.union([z.literal("low"), z.literal("medium"), z.literal("high"), z.literal("auto")]).optional(),
    size: z.union([z.literal("1024x1024"), z.literal("1024x1536"), z.literal("1536x1024"), z.literal("auto")]).optional(),
}));
export type ImageGeneration = z.infer<typeof ZImageGeneration>;

export const ZLocalShell = z.lazy(() => z.object({
    type: z.literal("local_shell"),
}));
export type LocalShell = z.infer<typeof ZLocalShell>;

export const ZResponsePrompt = z.lazy(() => z.object({
    id: z.string(),
    variables: z.record(z.string(), z.union([z.string(), ZResponseInputText, ZResponseInputImage, ZResponseInputFile])).optional(),
    version: z.string().optional(),
}));
export type ResponsePrompt = z.infer<typeof ZResponsePrompt>;

export const ZResponseTextConfig = z.lazy(() => z.object({
    format: z.union([ZResponseFormatText, ZResponseFormatTextJSONSchemaConfig, ZResponseFormatJSONObject]).optional(),
}));
export type ResponseTextConfig = z.infer<typeof ZResponseTextConfig>;

export const ZResponseUsage = z.lazy(() => z.object({
    input_tokens: z.number(),
    input_tokens_details: ZInputTokensDetails,
    output_tokens: z.number(),
    output_tokens_details: ZOutputTokensDetails,
    total_tokens: z.number(),
}));
export type ResponseUsage = z.infer<typeof ZResponseUsage>;

export const ZChatCompletionMessage = z.lazy(() => z.object({
    content: z.string().optional(),
    refusal: z.string().optional(),
    role: z.literal("assistant"),
    annotations: z.array(ZAnnotation).optional(),
    audio: ZChatCompletionAudio.optional(),
    function_call: ZFunctionCall.optional(),
    tool_calls: z.array(ZChatCompletionMessageToolCall).optional(),
}));
export type ChatCompletionMessage = z.infer<typeof ZChatCompletionMessage>;

export const ZCompletionTokensDetails = z.lazy(() => z.object({
    accepted_prediction_tokens: z.number().optional(),
    audio_tokens: z.number().optional(),
    reasoning_tokens: z.number().optional(),
    rejected_prediction_tokens: z.number().optional(),
}));
export type CompletionTokensDetails = z.infer<typeof ZCompletionTokensDetails>;

export const ZPromptTokensDetails = z.lazy(() => z.object({
    audio_tokens: z.number().optional(),
    cached_tokens: z.number().optional(),
}));
export type PromptTokensDetails = z.infer<typeof ZPromptTokensDetails>;

export const ZImageURL = z.lazy(() => z.object({
    url: z.string(),
    detail: z.union([z.literal("auto"), z.literal("low"), z.literal("high")]),
}));
export type ImageURL = z.infer<typeof ZImageURL>;

export const ZInputAudio = z.lazy(() => z.object({
    data: z.string(),
    format: z.union([z.literal("wav"), z.literal("mp3")]),
}));
export type InputAudio = z.infer<typeof ZInputAudio>;

export const ZFileFile = z.lazy(() => z.object({
    file_data: z.string(),
    file_id: z.string(),
    filename: z.string(),
}));
export type FileFile = z.infer<typeof ZFileFile>;

export const ZResponseInputTextParam = z.lazy(() => z.object({
    text: z.string(),
    type: z.literal("input_text"),
}));
export type ResponseInputTextParam = z.infer<typeof ZResponseInputTextParam>;

export const ZResponseInputImageParam = z.lazy(() => z.object({
    detail: z.union([z.literal("low"), z.literal("high"), z.literal("auto")]),
    type: z.literal("input_image"),
    file_id: z.string().optional(),
    image_url: z.string().optional(),
}));
export type ResponseInputImageParam = z.infer<typeof ZResponseInputImageParam>;

export const ZResponseInputFileParam = z.lazy(() => z.object({
    type: z.literal("input_file"),
    file_data: z.string(),
    file_id: z.string().optional(),
    file_url: z.string(),
    filename: z.string(),
}));
export type ResponseInputFileParam = z.infer<typeof ZResponseInputFileParam>;

export const ZResponseOutputTextParam = z.lazy(() => z.object({
    annotations: z.array(z.union([ZAnnotationFileCitation, ZAnnotationURLCitation, ZAnnotationContainerFileCitation, ZAnnotationFilePath])),
    text: z.string(),
    type: z.literal("output_text"),
    logprobs: z.array(ZLogprob),
}));
export type ResponseOutputTextParam = z.infer<typeof ZResponseOutputTextParam>;

export const ZResponseOutputRefusalParam = z.lazy(() => z.object({
    refusal: z.string(),
    type: z.literal("refusal"),
}));
export type ResponseOutputRefusalParam = z.infer<typeof ZResponseOutputRefusalParam>;

export const ZResult = z.lazy(() => z.object({
    attributes: z.record(z.string(), z.union([z.string(), z.number(), z.boolean()])).optional(),
    file_id: z.string(),
    filename: z.string(),
    score: z.number(),
    text: z.string(),
}));
export type Result = z.infer<typeof ZResult>;

export const ZActionClick = z.lazy(() => z.object({
    button: z.union([z.literal("left"), z.literal("right"), z.literal("wheel"), z.literal("back"), z.literal("forward")]),
    type: z.literal("click"),
    x: z.number(),
    y: z.number(),
}));
export type ActionClick = z.infer<typeof ZActionClick>;

export const ZActionDoubleClick = z.lazy(() => z.object({
    type: z.literal("double_click"),
    x: z.number(),
    y: z.number(),
}));
export type ActionDoubleClick = z.infer<typeof ZActionDoubleClick>;

export const ZActionDrag = z.lazy(() => z.object({
    path: z.array(ZActionDragPath),
    type: z.literal("drag"),
}));
export type ActionDrag = z.infer<typeof ZActionDrag>;

export const ZActionKeypress = z.lazy(() => z.object({
    keys: z.array(z.string()),
    type: z.literal("keypress"),
}));
export type ActionKeypress = z.infer<typeof ZActionKeypress>;

export const ZActionMove = z.lazy(() => z.object({
    type: z.literal("move"),
    x: z.number(),
    y: z.number(),
}));
export type ActionMove = z.infer<typeof ZActionMove>;

export const ZActionScreenshot = z.lazy(() => z.object({
    type: z.literal("screenshot"),
}));
export type ActionScreenshot = z.infer<typeof ZActionScreenshot>;

export const ZActionScroll = z.lazy(() => z.object({
    scroll_x: z.number(),
    scroll_y: z.number(),
    type: z.literal("scroll"),
    x: z.number(),
    y: z.number(),
}));
export type ActionScroll = z.infer<typeof ZActionScroll>;

export const ZActionType = z.lazy(() => z.object({
    text: z.string(),
    type: z.literal("type"),
}));
export type ActionType = z.infer<typeof ZActionType>;

export const ZActionWait = z.lazy(() => z.object({
    type: z.literal("wait"),
}));
export type ActionWait = z.infer<typeof ZActionWait>;

export const ZPendingSafetyCheck = z.lazy(() => z.object({
    id: z.string(),
    code: z.string(),
    message: z.string(),
}));
export type PendingSafetyCheck = z.infer<typeof ZPendingSafetyCheck>;

export const ZResponseComputerToolCallOutputScreenshotParam = z.lazy(() => z.object({
    type: z.literal("computer_screenshot"),
    file_id: z.string(),
    image_url: z.string(),
}));
export type ResponseComputerToolCallOutputScreenshotParam = z.infer<typeof ZResponseComputerToolCallOutputScreenshotParam>;

export const ZComputerCallOutputAcknowledgedSafetyCheck = z.lazy(() => z.object({
    id: z.string(),
    code: z.string().optional(),
    message: z.string().optional(),
}));
export type ComputerCallOutputAcknowledgedSafetyCheck = z.infer<typeof ZComputerCallOutputAcknowledgedSafetyCheck>;

export const ZActionSearch = z.lazy(() => z.object({
    query: z.string(),
    type: z.literal("search"),
}));
export type ActionSearch = z.infer<typeof ZActionSearch>;

export const ZActionOpenPage = z.lazy(() => z.object({
    type: z.literal("open_page"),
    url: z.string(),
}));
export type ActionOpenPage = z.infer<typeof ZActionOpenPage>;

export const ZActionFind = z.lazy(() => z.object({
    pattern: z.string(),
    type: z.literal("find"),
    url: z.string(),
}));
export type ActionFind = z.infer<typeof ZActionFind>;

export const ZSummary = z.lazy(() => z.object({
    text: z.string(),
    type: z.literal("summary_text"),
}));
export type Summary = z.infer<typeof ZSummary>;

export const ZOutputLogs = z.lazy(() => z.object({
    logs: z.string(),
    type: z.literal("logs"),
}));
export type OutputLogs = z.infer<typeof ZOutputLogs>;

export const ZOutputImage = z.lazy(() => z.object({
    type: z.literal("image"),
    url: z.string(),
}));
export type OutputImage = z.infer<typeof ZOutputImage>;

export const ZLocalShellCallAction = z.lazy(() => z.object({
    command: z.array(z.string()),
    env: z.record(z.string(), z.string()),
    type: z.literal("exec"),
    timeout_ms: z.number().optional(),
    user: z.string().optional(),
    working_directory: z.string().optional(),
}));
export type LocalShellCallAction = z.infer<typeof ZLocalShellCallAction>;

export const ZMcpListToolsTool = z.lazy(() => z.object({
    input_schema: z.object({}).passthrough(),
    name: z.string(),
    annotations: z.object({}).passthrough().optional(),
    description: z.string().optional(),
}));
export type McpListToolsTool = z.infer<typeof ZMcpListToolsTool>;

export const ZAudio = z.lazy(() => z.object({
    id: z.string(),
}));
export type Audio = z.infer<typeof ZAudio>;

export const ZChatCompletionContentPartRefusalParam = z.lazy(() => z.object({
    refusal: z.string(),
    type: z.literal("refusal"),
}));
export type ChatCompletionContentPartRefusalParam = z.infer<typeof ZChatCompletionContentPartRefusalParam>;

export const ZFunctionCall = z.lazy(() => z.object({
    arguments: z.string(),
    name: z.string(),
}));
export type FunctionCall = z.infer<typeof ZFunctionCall>;

export const ZChatCompletionMessageToolCallParam = z.lazy(() => z.object({
    id: z.string(),
    function: ZFunction,
    type: z.literal("function"),
}));
export type ChatCompletionMessageToolCallParam = z.infer<typeof ZChatCompletionMessageToolCallParam>;

export const ZVideoMetadata = z.lazy(() => z.object({
    fps: z.number().optional(),
    end_offset: z.string().optional(),
    start_offset: z.string().optional(),
}));
export type VideoMetadata = z.infer<typeof ZVideoMetadata>;

export const ZBlob = z.lazy(() => z.object({
    display_name: z.string().optional(),
    data: z.instanceof(Uint8Array).optional(),
    mime_type: z.string().optional(),
}));
export type Blob = z.infer<typeof ZBlob>;

export const ZCodeExecutionResult = z.lazy(() => z.object({
    outcome: z.any().optional(),
    output: z.string().optional(),
}));
export type CodeExecutionResult = z.infer<typeof ZCodeExecutionResult>;

export const ZExecutableCode = z.lazy(() => z.object({
    code: z.string().optional(),
    language: z.any().optional(),
}));
export type ExecutableCode = z.infer<typeof ZExecutableCode>;

export const ZFunctionResponse = z.lazy(() => z.object({
    will_continue: z.boolean().optional(),
    scheduling: z.any().optional(),
    id: z.string().optional(),
    name: z.string().optional(),
    response: z.record(z.string(), z.any()).optional(),
}));
export type FunctionResponse = z.infer<typeof ZFunctionResponse>;

export const ZPartDict = z.lazy(() => z.object({
    video_metadata: ZVideoMetadataDict.optional(),
    thought: z.boolean().optional(),
    inline_data: ZBlobDict.optional(),
    file_data: ZFileDataDict.optional(),
    thought_signature: z.instanceof(Uint8Array).optional(),
    code_execution_result: ZCodeExecutionResultDict.optional(),
    executable_code: ZExecutableCodeDict.optional(),
    function_call: ZFunctionCallDict.optional(),
    function_response: ZFunctionResponseDict.optional(),
    text: z.string().optional(),
}));
export type PartDict = z.infer<typeof ZPartDict>;

export const ZCacheControlEphemeralParam = z.lazy(() => z.object({
    type: z.literal("ephemeral"),
}));
export type CacheControlEphemeralParam = z.infer<typeof ZCacheControlEphemeralParam>;

export const ZCitationCharLocationParam = z.lazy(() => z.object({
    cited_text: z.string(),
    document_index: z.number(),
    document_title: z.string().optional(),
    end_char_index: z.number(),
    start_char_index: z.number(),
    type: z.literal("char_location"),
}));
export type CitationCharLocationParam = z.infer<typeof ZCitationCharLocationParam>;

export const ZCitationPageLocationParam = z.lazy(() => z.object({
    cited_text: z.string(),
    document_index: z.number(),
    document_title: z.string().optional(),
    end_page_number: z.number(),
    start_page_number: z.number(),
    type: z.literal("page_location"),
}));
export type CitationPageLocationParam = z.infer<typeof ZCitationPageLocationParam>;

export const ZCitationContentBlockLocationParam = z.lazy(() => z.object({
    cited_text: z.string(),
    document_index: z.number(),
    document_title: z.string().optional(),
    end_block_index: z.number(),
    start_block_index: z.number(),
    type: z.literal("content_block_location"),
}));
export type CitationContentBlockLocationParam = z.infer<typeof ZCitationContentBlockLocationParam>;

export const ZCitationWebSearchResultLocationParam = z.lazy(() => z.object({
    cited_text: z.string(),
    encrypted_index: z.string(),
    title: z.string().optional(),
    type: z.literal("web_search_result_location"),
    url: z.string(),
}));
export type CitationWebSearchResultLocationParam = z.infer<typeof ZCitationWebSearchResultLocationParam>;

export const ZBase64ImageSourceParam = z.lazy(() => z.object({
    data: z.union([z.string(), z.instanceof(Uint8Array), z.string()]),
    media_type: z.union([z.literal("image/jpeg"), z.literal("image/png"), z.literal("image/gif"), z.literal("image/webp")]),
    type: z.literal("base64"),
}));
export type Base64ImageSourceParam = z.infer<typeof ZBase64ImageSourceParam>;

export const ZURLImageSourceParam = z.lazy(() => z.object({
    type: z.literal("url"),
    url: z.string(),
}));
export type URLImageSourceParam = z.infer<typeof ZURLImageSourceParam>;

export const ZBase64PDFSourceParam = z.lazy(() => z.object({
    data: z.union([z.string(), z.instanceof(Uint8Array), z.string()]),
    media_type: z.literal("application/pdf"),
    type: z.literal("base64"),
}));
export type Base64PDFSourceParam = z.infer<typeof ZBase64PDFSourceParam>;

export const ZPlainTextSourceParam = z.lazy(() => z.object({
    data: z.string(),
    media_type: z.literal("text/plain"),
    type: z.literal("text"),
}));
export type PlainTextSourceParam = z.infer<typeof ZPlainTextSourceParam>;

export const ZContentBlockSourceParam = z.lazy(() => z.object({
    content: z.union([z.string(), z.array(z.union([ZTextBlockParam, ZImageBlockParam]))]),
    type: z.literal("content"),
}));
export type ContentBlockSourceParam = z.infer<typeof ZContentBlockSourceParam>;

export const ZURLPDFSourceParam = z.lazy(() => z.object({
    type: z.literal("url"),
    url: z.string(),
}));
export type URLPDFSourceParam = z.infer<typeof ZURLPDFSourceParam>;

export const ZCitationsConfigParam = z.lazy(() => z.object({
    enabled: z.boolean(),
}));
export type CitationsConfigParam = z.infer<typeof ZCitationsConfigParam>;

export const ZWebSearchResultBlockParam = z.lazy(() => z.object({
    encrypted_content: z.string(),
    title: z.string(),
    type: z.literal("web_search_result"),
    url: z.string(),
    page_age: z.string().optional(),
}));
export type WebSearchResultBlockParam = z.infer<typeof ZWebSearchResultBlockParam>;

export const ZWebSearchToolRequestErrorParam = z.lazy(() => z.object({
    error_code: z.union([z.literal("invalid_tool_input"), z.literal("unavailable"), z.literal("max_uses_exceeded"), z.literal("too_many_requests"), z.literal("query_too_long")]),
    type: z.literal("web_search_tool_result_error"),
}));
export type WebSearchToolRequestErrorParam = z.infer<typeof ZWebSearchToolRequestErrorParam>;

export const ZCitationCharLocation = z.lazy(() => z.object({
    cited_text: z.string(),
    document_index: z.number(),
    document_title: z.string().optional(),
    end_char_index: z.number(),
    start_char_index: z.number(),
    type: z.literal("char_location"),
}));
export type CitationCharLocation = z.infer<typeof ZCitationCharLocation>;

export const ZCitationPageLocation = z.lazy(() => z.object({
    cited_text: z.string(),
    document_index: z.number(),
    document_title: z.string().optional(),
    end_page_number: z.number(),
    start_page_number: z.number(),
    type: z.literal("page_location"),
}));
export type CitationPageLocation = z.infer<typeof ZCitationPageLocation>;

export const ZCitationContentBlockLocation = z.lazy(() => z.object({
    cited_text: z.string(),
    document_index: z.number(),
    document_title: z.string().optional(),
    end_block_index: z.number(),
    start_block_index: z.number(),
    type: z.literal("content_block_location"),
}));
export type CitationContentBlockLocation = z.infer<typeof ZCitationContentBlockLocation>;

export const ZCitationsWebSearchResultLocation = z.lazy(() => z.object({
    cited_text: z.string(),
    encrypted_index: z.string(),
    title: z.string().optional(),
    type: z.literal("web_search_result_location"),
    url: z.string(),
}));
export type CitationsWebSearchResultLocation = z.infer<typeof ZCitationsWebSearchResultLocation>;

export const ZWebSearchToolResultError = z.lazy(() => z.object({
    error_code: z.union([z.literal("invalid_tool_input"), z.literal("unavailable"), z.literal("max_uses_exceeded"), z.literal("too_many_requests"), z.literal("query_too_long")]),
    type: z.literal("web_search_tool_result_error"),
}));
export type WebSearchToolResultError = z.infer<typeof ZWebSearchToolResultError>;

export const ZWebSearchResultBlock = z.lazy(() => z.object({
    encrypted_content: z.string(),
    page_age: z.string().optional(),
    title: z.string(),
    type: z.literal("web_search_result"),
    url: z.string(),
}));
export type WebSearchResultBlock = z.infer<typeof ZWebSearchResultBlock>;

export const ZChatCompletionTokenLogprob = z.lazy(() => z.object({
    token: z.string(),
    bytes: z.array(z.number()).optional(),
    logprob: z.number(),
    top_logprobs: z.array(ZTopLogprob),
}));
export type ChatCompletionTokenLogprob = z.infer<typeof ZChatCompletionTokenLogprob>;

export const ZChoiceDeltaToolCallFunction = z.lazy(() => z.object({
    arguments: z.string().optional(),
    name: z.string().optional(),
}));
export type ChoiceDeltaToolCallFunction = z.infer<typeof ZChoiceDeltaToolCallFunction>;

export const ZServerToolUsage = z.lazy(() => z.object({
    web_search_requests: z.number(),
}));
export type ServerToolUsage = z.infer<typeof ZServerToolUsage>;

export const ZChatCompletionAudio = z.lazy(() => z.object({
    id: z.string(),
    data: z.string(),
    expires_at: z.number(),
    transcript: z.string(),
}));
export type ChatCompletionAudio = z.infer<typeof ZChatCompletionAudio>;

export const ZParsedFunctionToolCall = z.lazy(() => z.object({
    id: z.string(),
    function: ZParsedFunction,
    type: z.literal("function"),
}).merge(ZChatCompletionMessageToolCall.schema));
export type ParsedFunctionToolCall = z.infer<typeof ZParsedFunctionToolCall>;

export const ZResponseInputText = z.lazy(() => z.object({
    text: z.string(),
    type: z.literal("input_text"),
}));
export type ResponseInputText = z.infer<typeof ZResponseInputText>;

export const ZResponseInputImage = z.lazy(() => z.object({
    detail: z.union([z.literal("low"), z.literal("high"), z.literal("auto")]),
    type: z.literal("input_image"),
    file_id: z.string().optional(),
    image_url: z.string().optional(),
}));
export type ResponseInputImage = z.infer<typeof ZResponseInputImage>;

export const ZResponseInputFile = z.lazy(() => z.object({
    type: z.literal("input_file"),
    file_data: z.string().optional(),
    file_id: z.string().optional(),
    file_url: z.string().optional(),
    filename: z.string().optional(),
}));
export type ResponseInputFile = z.infer<typeof ZResponseInputFile>;

export const ZResponseOutputText = z.lazy(() => z.object({
    annotations: z.array(z.union([ZAnnotationFileCitation, ZAnnotationURLCitation, ZAnnotationContainerFileCitation, ZAnnotationFilePath])),
    text: z.string(),
    type: z.literal("output_text"),
    logprobs: z.array(ZLogprob).optional(),
}));
export type ResponseOutputText = z.infer<typeof ZResponseOutputText>;

export const ZResponseOutputRefusal = z.lazy(() => z.object({
    refusal: z.string(),
    type: z.literal("refusal"),
}));
export type ResponseOutputRefusal = z.infer<typeof ZResponseOutputRefusal>;

export const ZComparisonFilter = z.lazy(() => z.object({
    key: z.string(),
    type: z.union([z.literal("eq"), z.literal("ne"), z.literal("gt"), z.literal("gte"), z.literal("lt"), z.literal("lte")]),
    value: z.union([z.string(), z.number(), z.boolean()]),
}));
export type ComparisonFilter = z.infer<typeof ZComparisonFilter>;

export const ZCompoundFilter = z.lazy(() => z.object({
    filters: z.array(z.union([ZComparisonFilter, z.object({}).passthrough()])),
    type: z.union([z.literal("and"), z.literal("or")]),
}));
export type CompoundFilter = z.infer<typeof ZCompoundFilter>;

export const ZRankingOptions = z.lazy(() => z.object({
    ranker: z.union([z.literal("auto"), z.literal("default-2024-11-15")]).optional(),
    score_threshold: z.number().optional(),
}));
export type RankingOptions = z.infer<typeof ZRankingOptions>;

export const ZUserLocation = z.lazy(() => z.object({
    type: z.literal("approximate"),
    city: z.string().optional(),
    country: z.string().optional(),
    region: z.string().optional(),
    timezone: z.string().optional(),
}));
export type UserLocation = z.infer<typeof ZUserLocation>;

export const ZMcpAllowedToolsMcpAllowedToolsFilter = z.lazy(() => z.object({
    tool_names: z.array(z.string()).optional(),
}));
export type McpAllowedToolsMcpAllowedToolsFilter = z.infer<typeof ZMcpAllowedToolsMcpAllowedToolsFilter>;

export const ZMcpRequireApprovalMcpToolApprovalFilter = z.lazy(() => z.object({
    always: ZMcpRequireApprovalMcpToolApprovalFilterAlways.optional(),
    never: ZMcpRequireApprovalMcpToolApprovalFilterNever.optional(),
}));
export type McpRequireApprovalMcpToolApprovalFilter = z.infer<typeof ZMcpRequireApprovalMcpToolApprovalFilter>;

export const ZCodeInterpreterContainerCodeInterpreterToolAuto = z.lazy(() => z.object({
    type: z.literal("auto"),
    file_ids: z.array(z.string()).optional(),
}));
export type CodeInterpreterContainerCodeInterpreterToolAuto = z.infer<typeof ZCodeInterpreterContainerCodeInterpreterToolAuto>;

export const ZImageGenerationInputImageMask = z.lazy(() => z.object({
    file_id: z.string().optional(),
    image_url: z.string().optional(),
}));
export type ImageGenerationInputImageMask = z.infer<typeof ZImageGenerationInputImageMask>;

export const ZResponseFormatTextJSONSchemaConfig = z.lazy(() => z.object({
    name: z.string(),
    schema_: z.record(z.string(), z.object({}).passthrough()),
    type: z.literal("json_schema"),
    description: z.string().optional(),
    strict: z.boolean().optional(),
}));
export type ResponseFormatTextJSONSchemaConfig = z.infer<typeof ZResponseFormatTextJSONSchemaConfig>;

export const ZInputTokensDetails = z.lazy(() => z.object({
    cached_tokens: z.number(),
}));
export type InputTokensDetails = z.infer<typeof ZInputTokensDetails>;

export const ZOutputTokensDetails = z.lazy(() => z.object({
    reasoning_tokens: z.number(),
}));
export type OutputTokensDetails = z.infer<typeof ZOutputTokensDetails>;

export const ZChatCompletionMessageToolCall = z.lazy(() => z.object({
    id: z.string(),
    function: ZFunction,
    type: z.literal("function"),
}));
export type ChatCompletionMessageToolCall = z.infer<typeof ZChatCompletionMessageToolCall>;

export const ZAnnotationFileCitation = z.lazy(() => z.object({
    file_id: z.string(),
    filename: z.string(),
    index: z.number(),
    type: z.literal("file_citation"),
}));
export type AnnotationFileCitation = z.infer<typeof ZAnnotationFileCitation>;

export const ZAnnotationURLCitation = z.lazy(() => z.object({
    end_index: z.number(),
    start_index: z.number(),
    title: z.string(),
    type: z.literal("url_citation"),
    url: z.string(),
}));
export type AnnotationURLCitation = z.infer<typeof ZAnnotationURLCitation>;

export const ZAnnotationContainerFileCitation = z.lazy(() => z.object({
    container_id: z.string(),
    end_index: z.number(),
    file_id: z.string(),
    filename: z.string(),
    start_index: z.number(),
    type: z.literal("container_file_citation"),
}));
export type AnnotationContainerFileCitation = z.infer<typeof ZAnnotationContainerFileCitation>;

export const ZAnnotationFilePath = z.lazy(() => z.object({
    file_id: z.string(),
    index: z.number(),
    type: z.literal("file_path"),
}));
export type AnnotationFilePath = z.infer<typeof ZAnnotationFilePath>;

export const ZLogprob = z.lazy(() => z.object({
    token: z.string(),
    bytes: z.array(z.number()),
    logprob: z.number(),
    top_logprobs: z.array(ZLogprobTopLogprob),
}));
export type Logprob = z.infer<typeof ZLogprob>;

export const ZActionDragPath = z.lazy(() => z.object({
    x: z.number(),
    y: z.number(),
}));
export type ActionDragPath = z.infer<typeof ZActionDragPath>;

export const ZFunction = z.lazy(() => z.object({
    arguments: z.string(),
    name: z.string(),
}));
export type Function = z.infer<typeof ZFunction>;

export const ZVideoMetadataDict = z.lazy(() => z.object({
    fps: z.number().optional(),
    end_offset: z.string().optional(),
    start_offset: z.string().optional(),
}));
export type VideoMetadataDict = z.infer<typeof ZVideoMetadataDict>;

export const ZBlobDict = z.lazy(() => z.object({
    display_name: z.string().optional(),
    data: z.instanceof(Uint8Array).optional(),
    mime_type: z.string().optional(),
}));
export type BlobDict = z.infer<typeof ZBlobDict>;

export const ZFileDataDict = z.lazy(() => z.object({
    display_name: z.string().optional(),
    file_uri: z.string().optional(),
    mime_type: z.string().optional(),
}));
export type FileDataDict = z.infer<typeof ZFileDataDict>;

export const ZCodeExecutionResultDict = z.lazy(() => z.object({
    outcome: z.any().optional(),
    output: z.string().optional(),
}));
export type CodeExecutionResultDict = z.infer<typeof ZCodeExecutionResultDict>;

export const ZExecutableCodeDict = z.lazy(() => z.object({
    code: z.string().optional(),
    language: z.any().optional(),
}));
export type ExecutableCodeDict = z.infer<typeof ZExecutableCodeDict>;

export const ZFunctionCallDict = z.lazy(() => z.object({
    id: z.string().optional(),
    args: z.record(z.string(), z.any()).optional(),
    name: z.string().optional(),
}));
export type FunctionCallDict = z.infer<typeof ZFunctionCallDict>;

export const ZFunctionResponseDict = z.lazy(() => z.object({
    will_continue: z.boolean().optional(),
    scheduling: z.any().optional(),
    id: z.string().optional(),
    name: z.string().optional(),
    response: z.record(z.string(), z.any()).optional(),
}));
export type FunctionResponseDict = z.infer<typeof ZFunctionResponseDict>;

export const ZTopLogprob = z.lazy(() => z.object({
    token: z.string(),
    bytes: z.array(z.number()).optional(),
    logprob: z.number(),
}));
export type TopLogprob = z.infer<typeof ZTopLogprob>;

export const ZParsedFunction = z.lazy(() => z.object({
    arguments: z.string(),
    name: z.string(),
    parsed_arguments: z.object({}).passthrough().optional(),
}).merge(ZFunction.schema));
export type ParsedFunction = z.infer<typeof ZParsedFunction>;

export const ZMcpRequireApprovalMcpToolApprovalFilterAlways = z.lazy(() => z.object({
    tool_names: z.array(z.string()).optional(),
}));
export type McpRequireApprovalMcpToolApprovalFilterAlways = z.infer<typeof ZMcpRequireApprovalMcpToolApprovalFilterAlways>;

export const ZMcpRequireApprovalMcpToolApprovalFilterNever = z.lazy(() => z.object({
    tool_names: z.array(z.string()).optional(),
}));
export type McpRequireApprovalMcpToolApprovalFilterNever = z.infer<typeof ZMcpRequireApprovalMcpToolApprovalFilterNever>;

export const ZLogprobTopLogprob = z.lazy(() => z.object({
    token: z.string(),
    bytes: z.array(z.number()),
    logprob: z.number(),
}));
export type LogprobTopLogprob = z.infer<typeof ZLogprobTopLogprob>;

