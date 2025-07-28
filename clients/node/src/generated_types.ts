import * as z from 'zod';

export const ZDistancesResult = z.lazy(() => (z.object({
    distances: z.record(z.string(), z.any()),
    mean_distance: z.number(),
    metric_type: z.union([z.literal("levenshtein"), z.literal("jaccard"), z.literal("hamming")]),
})));
export type DistancesResult = z.infer<typeof ZDistancesResult>;

export const ZItemMetric = z.lazy(() => (z.object({
    id: z.string(),
    name: z.string(),
    similarity: z.number(),
    similarities: z.record(z.string(), z.any()),
    flat_similarities: z.record(z.string(), z.number().nullable().optional()),
    aligned_similarity: z.number(),
    aligned_similarities: z.record(z.string(), z.any()),
    aligned_flat_similarities: z.record(z.string(), z.number().nullable().optional()),
})));
export type ItemMetric = z.infer<typeof ZItemMetric>;

export const ZMetricResult = z.lazy(() => (z.object({
    item_metrics: z.array(ZItemMetric),
    mean_similarity: z.number(),
    aligned_mean_similarity: z.number(),
    metric_type: z.union([z.literal("levenshtein"), z.literal("jaccard"), z.literal("hamming")]),
})));
export type MetricResult = z.infer<typeof ZMetricResult>;

export const ZAttachmentMIMEData = z.lazy(() => (ZMIMEData.schema).merge(z.object({
    metadata: ZAttachmentMetadata.default({}),
})));
export type AttachmentMIMEData = z.infer<typeof ZAttachmentMIMEData>;

export const ZAttachmentMetadata = z.lazy(() => (z.object({
    is_inline: z.boolean().default(false),
    inline_cid: z.string().nullable().optional(),
    source: z.string().nullable().optional(),
})));
export type AttachmentMetadata = z.infer<typeof ZAttachmentMetadata>;

export const ZBaseAttachmentMIMEData = z.lazy(() => (ZBaseMIMEData.schema).merge(z.object({
    metadata: ZAttachmentMetadata.default({}),
})));
export type BaseAttachmentMIMEData = z.infer<typeof ZBaseAttachmentMIMEData>;

export const ZBaseEmailData = z.lazy(() => (z.object({
    id: z.string(),
    tree_id: z.string(),
    subject: z.string().nullable().optional(),
    body_plain: z.string().nullable().optional(),
    body_html: z.string().nullable().optional(),
    sender: ZEmailAddressData,
    recipients_to: z.array(ZEmailAddressData),
    recipients_cc: z.array(ZEmailAddressData).default([]),
    recipients_bcc: z.array(ZEmailAddressData).default([]),
    sent_at: z.string(),
    received_at: z.string().nullable().optional(),
    in_reply_to: z.string().nullable().optional(),
    references: z.array(z.string()).default([]),
    headers: z.record(z.string(), z.string()).default({}),
    url: z.string().nullable().optional(),
    attachments: z.array(ZBaseAttachmentMIMEData).default([]),
})));
export type BaseEmailData = z.infer<typeof ZBaseEmailData>;

export const ZBaseMIMEData = z.lazy(() => (ZMIMEData.schema).merge(z.object({
})));
export type BaseMIMEData = z.infer<typeof ZBaseMIMEData>;

export const ZEmailAddressData = z.lazy(() => (z.object({
    email: z.string(),
    display_name: z.string().nullable().optional(),
})));
export type EmailAddressData = z.infer<typeof ZEmailAddressData>;

export const ZEmailData = z.lazy(() => (ZBaseEmailData.schema).merge(z.object({
    attachments: z.array(ZAttachmentMIMEData).default([]),
})));
export type EmailData = z.infer<typeof ZEmailData>;

export const ZMIMEData = z.lazy(() => (z.object({
    filename: z.string(),
    url: z.string(),
})));
export type MIMEData = z.infer<typeof ZMIMEData>;

export const ZMatrix = z.lazy(() => (z.object({
    rows: z.number(),
    cols: z.number(),
    type_: z.number(),
    data: z.string(),
})));
export type Matrix = z.infer<typeof ZMatrix>;

export const ZOCR = z.lazy(() => (z.object({
    pages: z.array(ZPage),
})));
export type OCR = z.infer<typeof ZOCR>;

export const ZPage = z.lazy(() => (z.object({
    page_number: z.number(),
    width: z.number(),
    height: z.number(),
    unit: z.string().default("pixels"),
    blocks: z.array(ZTextBox),
    lines: z.array(ZTextBox),
    tokens: z.array(ZTextBox),
    transforms: z.array(ZMatrix).default([]),
})));
export type Page = z.infer<typeof ZPage>;

export const ZPoint = z.lazy(() => (z.object({
    x: z.number(),
    y: z.number(),
})));
export type Point = z.infer<typeof ZPoint>;

export const ZTextBox = z.lazy(() => (z.object({
    width: z.number(),
    height: z.number(),
    center: ZPoint,
    vertices: z.tuple([ZPoint, ZPoint, ZPoint, ZPoint]),
    text: z.string(),
})));
export type TextBox = z.infer<typeof ZTextBox>;

export const ZAmount = z.lazy(() => (z.object({
    value: z.number(),
    currency: z.string().default("USD"),
})));
export type Amount = z.infer<typeof ZAmount>;

export const ZPredictionData = z.lazy(() => (z.object({
    prediction: z.record(z.string(), z.any()).default({}),
    metadata: ZPredictionMetadata.nullable().optional(),
    updated_at: z.string().nullable().optional(),
})));
export type PredictionData = z.infer<typeof ZPredictionData>;

export const ZPredictionMetadata = z.lazy(() => (z.object({
    extraction_id: z.string().nullable().optional(),
    likelihoods: z.record(z.string(), z.any()).nullable().optional(),
    field_locations: z.record(z.string(), z.any()).nullable().optional(),
    agentic_field_locations: z.record(z.string(), z.any()).nullable().optional(),
    consensus_details: z.array(z.record(z.string(), z.any())).nullable().optional(),
    api_cost: ZAmount.nullable().optional(),
})));
export type PredictionMetadata = z.infer<typeof ZPredictionMetadata>;

export const ZChatCompletion = z.lazy(() => (z.object({
    id: z.string(),
    choices: z.array(ZChatCompletionChoice),
    created: z.number(),
    model: z.string(),
    object: z.literal("chat.completion"),
    service_tier: z.union([z.literal("auto"), z.literal("default"), z.literal("flex"), z.literal("scale"), z.literal("priority")]).nullable().optional(),
    system_fingerprint: z.string().nullable().optional(),
    usage: ZCompletionUsage.nullable().optional(),
})));
export type ChatCompletion = z.infer<typeof ZChatCompletion>;

export const ZChatCompletionRetabMessage = z.lazy(() => (z.object({
    role: z.union([z.literal("user"), z.literal("system"), z.literal("assistant"), z.literal("developer")]),
    content: z.union([z.string(), z.array(z.union([ZChatCompletionContentPartTextParam, ZChatCompletionContentPartImageParam, ZChatCompletionContentPartInputAudioParam, ZFile]))]),
})));
export type ChatCompletionRetabMessage = z.infer<typeof ZChatCompletionRetabMessage>;

export const ZCostBreakdown = z.lazy(() => (z.object({
    total: ZAmount,
    text_prompt_cost: ZAmount,
    text_cached_cost: ZAmount,
    text_completion_cost: ZAmount,
    text_total_cost: ZAmount,
    audio_prompt_cost: ZAmount.nullable().optional(),
    audio_completion_cost: ZAmount.nullable().optional(),
    audio_total_cost: ZAmount.nullable().optional(),
    token_counts: ZTokenCounts,
    model: z.string(),
    is_fine_tuned: z.boolean().default(false),
})));
export type CostBreakdown = z.infer<typeof ZCostBreakdown>;

export const ZExtraction = z.lazy(() => (z.object({
    id: z.string(),
    messages: z.array(ZChatCompletionRetabMessage),
    messages_gcs: z.string(),
    file_gcs_paths: z.array(z.string()),
    file_ids: z.array(z.string()),
    file_gcs: z.string().default(""),
    file_id: z.string().default(""),
    status: z.union([z.literal("success"), z.literal("failed")]),
    completion: z.union([ZRetabParsedChatCompletion, ZChatCompletion]),
    json_schema: z.any(),
    model: z.string(),
    temperature: z.number().default(0.0),
    source: ZExtractionSource,
    image_resolution_dpi: z.number().default(96),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]).default("A4"),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]).default("native"),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional(),
    n_consensus: z.number().default(1),
    timings: z.array(ZExtractionTimingStep),
    schema_id: z.string(),
    schema_data_id: z.string(),
    created_at: z.string(),
    request_at: z.string().nullable().optional(),
    organization_id: z.string(),
    validation_state: z.union([z.literal("pending"), z.literal("validated"), z.literal("invalid")]).nullable().optional(),
    billed: z.boolean().default(false),
})));
export type Extraction = z.infer<typeof ZExtraction>;

export const ZExtractionSource = z.lazy(() => (z.object({
    type: z.union([z.literal("api"), z.literal("annotation"), z.literal("processor"), z.literal("automation"), z.literal("automation.link"), z.literal("automation.mailbox"), z.literal("automation.cron"), z.literal("automation.outlook"), z.literal("automation.endpoint"), z.literal("schema.extract")]),
    id: z.string().nullable().optional(),
})));
export type ExtractionSource = z.infer<typeof ZExtractionSource>;

export const ZExtractionTimingStep = z.lazy(() => (z.object({
    name: z.union([z.string(), z.union([z.literal("initialization"), z.literal("prepare_messages"), z.literal("yield_first_token"), z.literal("completion")])]),
    duration: z.number(),
    notes: z.string().nullable().optional(),
})));
export type ExtractionTimingStep = z.infer<typeof ZExtractionTimingStep>;

export const ZRetabParsedChatCompletion = z.lazy(() => (ZParsedChatCompletion.schema).merge(z.object({
    choices: z.array(ZRetabParsedChoice),
    extraction_id: z.string().nullable().optional(),
    likelihoods: z.record(z.string(), z.any()).nullable().optional(),
    schema_validation_error: ZErrorDetail.nullable().optional(),
    request_at: z.string().nullable().optional(),
    first_token_at: z.string().nullable().optional(),
    last_token_at: z.string().nullable().optional(),
})));
export type RetabParsedChatCompletion = z.infer<typeof ZRetabParsedChatCompletion>;

export const ZEvent = z.lazy(() => (z.object({
    object: z.literal("event").default("event"),
    id: z.string(),
    event: z.string(),
    created_at: z.string(),
    data: z.record(z.string(), z.any()),
    metadata: z.record(z.union([z.literal("automation"), z.literal("cron"), z.literal("data_structure"), z.literal("dataset"), z.literal("dataset_membership"), z.literal("endpoint"), z.literal("evaluation"), z.literal("extraction"), z.literal("file"), z.literal("files"), z.literal("link"), z.literal("mailbox"), z.literal("organization"), z.literal("outlook"), z.literal("preprocessing"), z.literal("reconciliation"), z.literal("schema"), z.literal("schema_data"), z.literal("template"), z.literal("user"), z.literal("webhook")]), z.string()).nullable().optional(),
})));
export type Event = z.infer<typeof ZEvent>;

export const ZStoredEvent = z.lazy(() => (ZEvent.schema).merge(z.object({
    organization_id: z.string(),
})));
export type StoredEvent = z.infer<typeof ZStoredEvent>;

export const ZDeleteResponse = z.lazy(() => (z.object({
    success: z.boolean(),
    id: z.string(),
})));
export type DeleteResponse = z.infer<typeof ZDeleteResponse>;

export const ZDocumentPreprocessResponseContent = z.lazy(() => (z.object({
    messages: z.array(z.record(z.string(), z.any())),
    json_schema: z.record(z.string(), z.any()),
})));
export type DocumentPreprocessResponseContent = z.infer<typeof ZDocumentPreprocessResponseContent>;

export const ZErrorDetail = z.lazy(() => (z.object({
    code: z.string(),
    message: z.string(),
    details: z.record(z.any()).nullable().optional(),
})));
export type ErrorDetail = z.infer<typeof ZErrorDetail>;

export const ZExportResponse = z.lazy(() => (z.object({
    success: z.boolean(),
    path: z.string(),
})));
export type ExportResponse = z.infer<typeof ZExportResponse>;

export const ZPreparedRequest = z.lazy(() => (z.object({
    method: z.union([z.literal("POST"), z.literal("GET"), z.literal("PUT"), z.literal("PATCH"), z.literal("DELETE"), z.literal("HEAD"), z.literal("OPTIONS"), z.literal("CONNECT"), z.literal("TRACE")]),
    url: z.string(),
    data: z.record(z.any()).nullable().optional(),
    params: z.record(z.any()).nullable().optional(),
    form_data: z.record(z.any()).nullable().optional(),
    files: z.union([z.record(z.any()), z.array(z.tuple([z.string(), z.tuple([z.string(), z.instanceof(Uint8Array), z.string()])]))]).nullable().optional(),
    idempotency_key: z.string().nullable().optional(),
    raise_for_status: z.boolean().default(false),
})));
export type PreparedRequest = z.infer<typeof ZPreparedRequest>;

export const ZStandardErrorResponse = z.lazy(() => (z.object({
    detail: ZErrorDetail,
})));
export type StandardErrorResponse = z.infer<typeof ZStandardErrorResponse>;

export const ZStreamingBaseModel = z.lazy(() => (z.object({
    streaming_error: ZErrorDetail.nullable().optional(),
})));
export type StreamingBaseModel = z.infer<typeof ZStreamingBaseModel>;

export const ZReasoning = z.lazy(() => (z.object({
    effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional(),
    generate_summary: z.union([z.literal("auto"), z.literal("concise"), z.literal("detailed")]).nullable().optional(),
    summary: z.union([z.literal("auto"), z.literal("concise"), z.literal("detailed")]).nullable().optional(),
})));
export type Reasoning = z.infer<typeof ZReasoning>;

export const ZResponseFormatJSONSchema = z.lazy(() => (z.object({
    json_schema: ZJSONSchema,
    type: z.literal("json_schema"),
})));
export type ResponseFormatJSONSchema = z.infer<typeof ZResponseFormatJSONSchema>;

export const ZResponseTextConfigParam = z.lazy(() => (z.object({
    format: z.union([ZResponseFormatText, ZResponseFormatTextJSONSchemaConfigParam, ZResponseFormatJSONObject]),
})));
export type ResponseTextConfigParam = z.infer<typeof ZResponseTextConfigParam>;

export const ZRetabChatCompletionsParseRequest = z.lazy(() => (z.object({
    model: z.string(),
    messages: z.array(ZChatCompletionRetabMessage),
    json_schema: z.record(z.string(), z.any()),
    temperature: z.number().default(0.0),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional().default("medium"),
    stream: z.boolean().default(false),
    seed: z.number().nullable().optional(),
    n_consensus: z.number().default(1),
})));
export type RetabChatCompletionsParseRequest = z.infer<typeof ZRetabChatCompletionsParseRequest>;

export const ZRetabChatCompletionsRequest = z.lazy(() => (z.object({
    model: z.string(),
    messages: z.array(ZChatCompletionRetabMessage),
    response_format: ZResponseFormatJSONSchema,
    temperature: z.number().default(0.0),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional().default("medium"),
    stream: z.boolean().default(false),
    seed: z.number().nullable().optional(),
    n_consensus: z.number().default(1),
})));
export type RetabChatCompletionsRequest = z.infer<typeof ZRetabChatCompletionsRequest>;

export const ZRetabChatResponseCreateRequest = z.lazy(() => (z.object({
    input: z.union([z.string(), z.array(z.union([ZEasyInputMessageParam, ZResponseInputParamMessage, ZResponseOutputMessageParam, ZResponseFileSearchToolCallParam, ZResponseComputerToolCallParam, ZComputerCallOutput, ZResponseFunctionWebSearchParam, ZResponseFunctionToolCallParam, ZFunctionCallOutput, ZResponseReasoningItemParam, ZImageGenerationCall, ZResponseCodeInterpreterToolCallParam, ZLocalShellCall, ZLocalShellCallOutput, ZMcpListTools, ZMcpApprovalRequest, ZMcpApprovalResponse, ZMcpCall, ZItemReference]))]),
    instructions: z.string().nullable().optional(),
    model: z.string(),
    temperature: z.number().nullable().optional().default(0.0),
    reasoning: ZReasoning.nullable().optional(),
    stream: z.boolean().nullable().optional().default(false),
    seed: z.number().nullable().optional(),
    text: ZResponseTextConfigParam.default({"format": {"type": "text"}}),
    n_consensus: z.number().default(1),
})));
export type RetabChatResponseCreateRequest = z.infer<typeof ZRetabChatResponseCreateRequest>;

export const ZReconciliationRequest = z.lazy(() => (z.object({
    list_dicts: z.array(z.record(z.any())),
    reference_schema: z.record(z.string(), z.any()).nullable().optional(),
    mode: z.union([z.literal("direct"), z.literal("aligned")]).default("direct"),
})));
export type ReconciliationRequest = z.infer<typeof ZReconciliationRequest>;

export const ZReconciliationResponse = z.lazy(() => (z.object({
    consensus_dict: z.record(z.any()),
    likelihoods: z.record(z.any()),
})));
export type ReconciliationResponse = z.infer<typeof ZReconciliationResponse>;

export const ZAutomationConfig = z.lazy(() => (z.object({
    id: z.string(),
    name: z.string(),
    processor_id: z.string(),
    updated_at: z.string(),
    default_language: z.string().default("en"),
    webhook_url: z.string(),
    webhook_headers: z.record(z.string(), z.string()),
    need_validation: z.boolean().default(false),
})));
export type AutomationConfig = z.infer<typeof ZAutomationConfig>;

export const ZAutomationLog = z.lazy(() => (z.object({
    object: z.literal("automation_log").default("automation_log"),
    id: z.string(),
    user_email: z.string().email().nullable().optional(),
    organization_id: z.string(),
    created_at: z.string(),
    automation_snapshot: ZAutomationConfig,
    completion: z.union([ZRetabParsedChatCompletion, ZChatCompletion]),
    file_metadata: ZBaseMIMEData.nullable().optional(),
    external_request_log: ZExternalRequestLog.nullable().optional(),
    extraction_id: z.string().nullable().optional(),
})));
export type AutomationLog = z.infer<typeof ZAutomationLog>;

export const ZEmailStr = z.lazy(() => z.string().email());
export type EmailStr = z.infer<typeof ZEmailStr>;

export const ZExternalRequestLog = z.lazy(() => (z.object({
    webhook_url: z.string().nullable().optional(),
    request_body: z.record(z.string(), z.any()),
    request_headers: z.record(z.string(), z.string()),
    request_at: z.string(),
    response_body: z.record(z.string(), z.any()),
    response_headers: z.record(z.string(), z.string()),
    response_at: z.string(),
    status_code: z.number(),
    error: z.string().nullable().optional(),
    duration_ms: z.number(),
})));
export type ExternalRequestLog = z.infer<typeof ZExternalRequestLog>;

export const ZListLogs = z.lazy(() => (z.object({
    data: z.array(ZAutomationLog),
    list_metadata: ZListMetadata,
})));
export type ListLogs = z.infer<typeof ZListLogs>;

export const ZListMetadata = z.lazy(() => (z.object({
    before: z.string().nullable().optional(),
    after: z.string().nullable().optional(),
})));
export type ListMetadata = z.infer<typeof ZListMetadata>;

export const ZLogCompletionRequest = z.lazy(() => (z.object({
    json_schema: z.record(z.string(), z.any()),
    completion: ZChatCompletion,
})));
export type LogCompletionRequest = z.infer<typeof ZLogCompletionRequest>;

export const ZOpenAIRequestConfig = z.lazy(() => (z.object({
    object: z.literal("openai_request").default("openai_request"),
    id: z.string(),
    model: z.string(),
    json_schema: z.record(z.string(), z.any()),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional(),
})));
export type OpenAIRequestConfig = z.infer<typeof ZOpenAIRequestConfig>;

export const ZProcessorConfig = z.lazy(() => (z.object({
    object: z.string().default("processor"),
    id: z.string(),
    updated_at: z.string(),
    name: z.string(),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]),
    image_resolution_dpi: z.number().default(96),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]).default("A4"),
    model: z.string(),
    json_schema: z.record(z.string(), z.any()),
    temperature: z.number().default(0.0),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional().default("medium"),
    n_consensus: z.number().default(1),
})));
export type ProcessorConfig = z.infer<typeof ZProcessorConfig>;

export const ZUpdateAutomationRequest = z.lazy(() => (z.object({
    name: z.string().nullable().optional(),
    default_language: z.string().nullable().optional(),
    webhook_url: z.string().nullable().optional(),
    webhook_headers: z.record(z.string(), z.string()).nullable().optional(),
    need_validation: z.boolean().nullable().optional(),
})));
export type UpdateAutomationRequest = z.infer<typeof ZUpdateAutomationRequest>;

export const ZUpdateProcessorRequest = z.lazy(() => (z.object({
    name: z.string().nullable().optional(),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]).nullable().optional(),
    image_resolution_dpi: z.number().nullable().optional(),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]).nullable().optional(),
    model: z.string().nullable().optional(),
    json_schema: z.record(z.any()).nullable().optional(),
    temperature: z.number().nullable().optional(),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional(),
    n_consensus: z.number().nullable().optional(),
})));
export type UpdateProcessorRequest = z.infer<typeof ZUpdateProcessorRequest>;

export const ZFinetunedModel = z.lazy(() => (z.object({
    object: z.literal("finetuned_model").default("finetuned_model"),
    organization_id: z.string(),
    model: z.string(),
    schema_id: z.string(),
    schema_data_id: z.string(),
    finetuning_props: ZInferenceSettings,
    project_id: z.string().nullable().optional(),
    created_at: z.string(),
})));
export type FinetunedModel = z.infer<typeof ZFinetunedModel>;

export const ZInferenceSettings = z.lazy(() => (z.object({
    model: z.string().default("gpt-4.1-mini"),
    temperature: z.number().default(0.0),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]).default("native"),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional().default("medium"),
    image_resolution_dpi: z.number().default(96),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]).default("A4"),
    n_consensus: z.number().default(1),
})));
export type InferenceSettings = z.infer<typeof ZInferenceSettings>;

export const ZModel = z.lazy(() => (z.object({
    id: z.string(),
    created: z.number(),
    object: z.literal("model"),
    owned_by: z.string(),
})));
export type Model = z.infer<typeof ZModel>;

export const ZModelCapabilities = z.lazy(() => (z.object({
    modalities: z.array(z.union([z.literal("text"), z.literal("audio"), z.literal("image")])),
    endpoints: z.array(z.union([z.literal("chat_completions"), z.literal("responses"), z.literal("assistants"), z.literal("batch"), z.literal("fine_tuning"), z.literal("embeddings"), z.literal("speech_generation"), z.literal("translation"), z.literal("completions_legacy"), z.literal("image_generation"), z.literal("transcription"), z.literal("moderation"), z.literal("realtime")])),
    features: z.array(z.union([z.literal("streaming"), z.literal("function_calling"), z.literal("structured_outputs"), z.literal("distillation"), z.literal("fine_tuning"), z.literal("predicted_outputs"), z.literal("schema_generation")])),
})));
export type ModelCapabilities = z.infer<typeof ZModelCapabilities>;

export const ZModelCard = z.lazy(() => (z.object({
    model: z.union([z.union([z.literal("gpt-4o"), z.literal("gpt-4o-mini"), z.literal("chatgpt-4o-latest"), z.literal("gpt-4.1"), z.literal("gpt-4.1-mini"), z.literal("gpt-4.1-mini-2025-04-14"), z.literal("gpt-4.1-2025-04-14"), z.literal("gpt-4.1-nano"), z.literal("gpt-4.1-nano-2025-04-14"), z.literal("gpt-4o-2024-11-20"), z.literal("gpt-4o-2024-08-06"), z.literal("gpt-4o-2024-05-13"), z.literal("gpt-4o-mini-2024-07-18"), z.literal("o1"), z.literal("o1-2024-12-17"), z.literal("o3"), z.literal("o3-2025-04-16"), z.literal("o4-mini"), z.literal("o4-mini-2025-04-16"), z.literal("gpt-4o-audio-preview-2024-12-17"), z.literal("gpt-4o-audio-preview-2024-10-01"), z.literal("gpt-4o-realtime-preview-2024-12-17"), z.literal("gpt-4o-realtime-preview-2024-10-01"), z.literal("gpt-4o-mini-audio-preview-2024-12-17"), z.literal("gpt-4o-mini-realtime-preview-2024-12-17"), z.literal("claude-3-5-sonnet-latest"), z.literal("claude-3-5-sonnet-20241022"), z.literal("claude-3-opus-20240229"), z.literal("claude-3-sonnet-20240229"), z.literal("claude-3-haiku-20240307"), z.literal("grok-3"), z.literal("grok-3-mini"), z.literal("gemini-2.5-pro"), z.literal("gemini-2.5-flash"), z.literal("gemini-2.5-pro-preview-06-05"), z.literal("gemini-2.5-pro-preview-05-06"), z.literal("gemini-2.5-pro-preview-03-25"), z.literal("gemini-2.5-flash-preview-05-20"), z.literal("gemini-2.5-flash-preview-04-17"), z.literal("gemini-2.5-flash-lite-preview-06-17"), z.literal("gemini-2.5-pro-exp-03-25"), z.literal("gemini-2.0-flash-lite"), z.literal("gemini-2.0-flash"), z.literal("auto-large"), z.literal("auto-small"), z.literal("auto-micro"), z.literal("human")]), z.string()]),
    pricing: ZPricing,
    capabilities: ZModelCapabilities,
    temperature_support: z.boolean().default(true),
    reasoning_effort_support: z.boolean().default(false),
    permissions: ZModelCardPermissions.default({}),
})));
export type ModelCard = z.infer<typeof ZModelCard>;

export const ZModelCardPermissions = z.lazy(() => (z.object({
    show_in_free_picker: z.boolean().default(false),
    show_in_paid_picker: z.boolean().default(false),
})));
export type ModelCardPermissions = z.infer<typeof ZModelCardPermissions>;

export const ZMonthlyUsageResponseContent = z.lazy(() => (z.object({
    credits_count: z.number(),
})));
export type MonthlyUsageResponseContent = z.infer<typeof ZMonthlyUsageResponseContent>;

export const ZPricing = z.lazy(() => (z.object({
    text: ZTokenPrice,
    audio: ZTokenPrice.nullable().optional(),
    ft_price_hike: z.number().default(1.0),
})));
export type Pricing = z.infer<typeof ZPricing>;

export const ZTokenPrice = z.lazy(() => (z.object({
    prompt: z.number(),
    completion: z.number(),
    cached_discount: z.number().default(1.0),
})));
export type TokenPrice = z.infer<typeof ZTokenPrice>;

export const ZAddIterationFromJsonlRequest = z.lazy(() => (z.object({
    jsonl_gcs_path: z.string(),
})));
export type AddIterationFromJsonlRequest = z.infer<typeof ZAddIterationFromJsonlRequest>;

export const ZAnnotatedDocument = z.lazy(() => (z.object({
    mime_data: ZMIMEData,
    annotation: z.record(z.string(), z.any()).default({}),
})));
export type AnnotatedDocument = z.infer<typeof ZAnnotatedDocument>;

export const ZCreateIterationRequest = z.lazy(() => (z.object({
    inference_settings: ZInferenceSettings,
    json_schema: z.record(z.string(), z.any()).nullable().optional(),
})));
export type CreateIterationRequest = z.infer<typeof ZCreateIterationRequest>;

export const ZDeprecatedEvalsDistancesResult = z.lazy(() => (z.object({
    distances: z.record(z.string(), z.any()),
    mean_distance: z.number(),
    metric_type: z.union([z.literal("levenshtein"), z.literal("jaccard"), z.literal("hamming")]),
})));
export type DeprecatedEvalsDistancesResult = z.infer<typeof ZDeprecatedEvalsDistancesResult>;

export const ZDocumentItem = z.lazy(() => (ZAnnotatedDocument.schema).merge(z.object({
    annotation_metadata: ZDeprecatedEvalsPredictionMetadata.nullable().optional(),
})));
export type DocumentItem = z.infer<typeof ZDocumentItem>;

export const ZDeprecatedEvalsItemMetric = z.lazy(() => (z.object({
    id: z.string(),
    name: z.string(),
    similarity: z.number(),
    similarities: z.record(z.string(), z.any()),
    flat_similarities: z.record(z.string(), z.number().nullable().optional()),
    aligned_similarity: z.number(),
    aligned_similarities: z.record(z.string(), z.any()),
    aligned_flat_similarities: z.record(z.string(), z.number().nullable().optional()),
})));
export type DeprecatedEvalsItemMetric = z.infer<typeof ZDeprecatedEvalsItemMetric>;

export const ZIteration = z.lazy(() => (z.object({
    id: z.string(),
    inference_settings: ZInferenceSettings,
    json_schema: z.record(z.string(), z.any()),
    predictions: z.array(ZDeprecatedEvalsPredictionData),
    metric_results: ZDeprecatedEvalsMetricResult.nullable().optional(),
})));
export type Iteration = z.infer<typeof ZIteration>;

export const ZDeprecatedEvalsMetricResult = z.lazy(() => (z.object({
    item_metrics: z.array(ZDeprecatedEvalsItemMetric),
    mean_similarity: z.number(),
    aligned_mean_similarity: z.number(),
    metric_type: z.union([z.literal("levenshtein"), z.literal("jaccard"), z.literal("hamming")]),
})));
export type DeprecatedEvalsMetricResult = z.infer<typeof ZDeprecatedEvalsMetricResult>;

export const ZDeprecatedEvalsPredictionData = z.lazy(() => (z.object({
    prediction: z.record(z.string(), z.any()).default({}),
    metadata: ZDeprecatedEvalsPredictionMetadata.nullable().optional(),
})));
export type DeprecatedEvalsPredictionData = z.infer<typeof ZDeprecatedEvalsPredictionData>;

export const ZDeprecatedEvalsPredictionMetadata = z.lazy(() => (z.object({
    extraction_id: z.string().nullable().optional(),
    likelihoods: z.record(z.string(), z.any()).nullable().optional(),
    field_locations: z.record(z.string(), z.any()).nullable().optional(),
    agentic_field_locations: z.record(z.string(), z.any()).nullable().optional(),
    consensus_details: z.array(z.record(z.string(), z.any())).nullable().optional(),
    api_cost: ZAmount.nullable().optional(),
})));
export type DeprecatedEvalsPredictionMetadata = z.infer<typeof ZDeprecatedEvalsPredictionMetadata>;

export const ZProject = z.lazy(() => (z.object({
    id: z.string(),
    updated_at: z.string(),
    name: z.string(),
    old_documents: z.array(ZProjectDocument).nullable().optional(),
    documents: z.array(ZProjectDocument),
    iterations: z.array(ZIteration),
    json_schema: z.record(z.string(), z.any()),
    project_id: z.string().default("default_spreadsheets"),
    default_inference_settings: ZInferenceSettings.nullable().optional(),
})));
export type Project = z.infer<typeof ZProject>;

export const ZProjectDocument = z.lazy(() => (ZDocumentItem.schema).merge(z.object({
    id: z.string(),
})));
export type ProjectDocument = z.infer<typeof ZProjectDocument>;

export const ZUpdateProjectDocumentRequest = z.lazy(() => (z.object({
    annotation: z.record(z.string(), z.any()).nullable().optional(),
    annotation_metadata: ZDeprecatedEvalsPredictionMetadata.nullable().optional(),
})));
export type UpdateProjectDocumentRequest = z.infer<typeof ZUpdateProjectDocumentRequest>;

export const ZUpdateProjectRequest = z.lazy(() => (z.object({
    name: z.string().nullable().optional(),
    documents: z.array(ZProjectDocument).nullable().optional(),
    iterations: z.array(ZIteration).nullable().optional(),
    json_schema: z.record(z.string(), z.any()).nullable().optional(),
    project_id: z.string().nullable().optional(),
    default_inference_settings: ZInferenceSettings.nullable().optional(),
})));
export type UpdateProjectRequest = z.infer<typeof ZUpdateProjectRequest>;

export const ZExternalAPIKey = z.lazy(() => (z.object({
    provider: z.union([z.literal("OpenAI"), z.literal("Anthropic"), z.literal("Gemini"), z.literal("xAI"), z.literal("Retab")]),
    is_configured: z.boolean(),
    last_updated: z.string().nullable().optional(),
})));
export type ExternalAPIKey = z.infer<typeof ZExternalAPIKey>;

export const ZExternalAPIKeyRequest = z.lazy(() => (z.object({
    provider: z.union([z.literal("OpenAI"), z.literal("Anthropic"), z.literal("Gemini"), z.literal("xAI"), z.literal("Retab")]),
    api_key: z.string(),
})));
export type ExternalAPIKeyRequest = z.infer<typeof ZExternalAPIKeyRequest>;

export const ZIterationsAddIterationFromJsonlRequest = z.lazy(() => (z.object({
    jsonl_gcs_path: z.string(),
})));
export type IterationsAddIterationFromJsonlRequest = z.infer<typeof ZIterationsAddIterationFromJsonlRequest>;

export const ZBaseIteration = z.lazy(() => (z.object({
    id: z.string(),
    inference_settings: ZInferenceSettings,
    json_schema: z.record(z.string(), z.any()),
    updated_at: z.string(),
})));
export type BaseIteration = z.infer<typeof ZBaseIteration>;

export const ZIterationsCreateIterationRequest = z.lazy(() => (z.object({
    inference_settings: ZInferenceSettings,
    json_schema: z.record(z.string(), z.any()).nullable().optional(),
    from_iteration_id: z.string().nullable().optional(),
})));
export type IterationsCreateIterationRequest = z.infer<typeof ZIterationsCreateIterationRequest>;

export const ZDocumentStatus = z.lazy(() => (z.object({
    document_id: z.string(),
    filename: z.string(),
    needs_update: z.boolean(),
    has_prediction: z.boolean(),
    prediction_updated_at: z.string().nullable().optional(),
    iteration_updated_at: z.string(),
})));
export type DocumentStatus = z.infer<typeof ZDocumentStatus>;

export const ZIterationsIteration = z.lazy(() => (ZBaseIteration.schema).merge(z.object({
    predictions: z.record(z.string(), ZPredictionData),
})));
export type IterationsIteration = z.infer<typeof ZIterationsIteration>;

export const ZIterationDocumentStatusResponse = z.lazy(() => (z.object({
    iteration_id: z.string(),
    documents: z.array(ZDocumentStatus),
    total_documents: z.number(),
    documents_needing_update: z.number(),
    documents_up_to_date: z.number(),
})));
export type IterationDocumentStatusResponse = z.infer<typeof ZIterationDocumentStatusResponse>;

export const ZPatchIterationRequest = z.lazy(() => (z.object({
    inference_settings: ZInferenceSettings.nullable().optional(),
    json_schema: z.record(z.string(), z.any()).nullable().optional(),
    version: z.number().nullable().optional(),
})));
export type PatchIterationRequest = z.infer<typeof ZPatchIterationRequest>;

export const ZProcessIterationRequest = z.lazy(() => (z.object({
    document_ids: z.array(z.string()).nullable().optional(),
    only_outdated: z.boolean().default(true),
})));
export type ProcessIterationRequest = z.infer<typeof ZProcessIterationRequest>;

export const ZDocumentsAnnotatedDocument = z.lazy(() => (z.object({
    mime_data: ZMIMEData,
    annotation: z.record(z.string(), z.any()).default({}),
})));
export type DocumentsAnnotatedDocument = z.infer<typeof ZDocumentsAnnotatedDocument>;

export const ZCreateProjectDocumentRequest = z.lazy(() => (ZDocumentsDocumentItem.schema).merge(z.object({
})));
export type CreateProjectDocumentRequest = z.infer<typeof ZCreateProjectDocumentRequest>;

export const ZDocumentsDocumentItem = z.lazy(() => (ZDocumentsAnnotatedDocument.schema).merge(z.object({
    annotation_metadata: ZPredictionMetadata.nullable().optional(),
})));
export type DocumentsDocumentItem = z.infer<typeof ZDocumentsDocumentItem>;

export const ZPatchProjectDocumentRequest = z.lazy(() => (z.object({
    annotation: z.record(z.string(), z.any()).nullable().optional(),
    annotation_metadata: ZPredictionMetadata.nullable().optional(),
    ocr_file_id: z.string().nullable().optional(),
})));
export type PatchProjectDocumentRequest = z.infer<typeof ZPatchProjectDocumentRequest>;

export const ZDocumentsProjectDocument = z.lazy(() => (ZDocumentsDocumentItem.schema).merge(z.object({
    id: z.string(),
    ocr_file_id: z.string().nullable().optional(),
})));
export type DocumentsProjectDocument = z.infer<typeof ZDocumentsProjectDocument>;

export const ZBaseProject = z.lazy(() => (z.object({
    id: z.string(),
    name: z.string().default(""),
    json_schema: z.record(z.string(), z.any()),
    default_inference_settings: ZInferenceSettings.default({}),
    updated_at: z.string(),
})));
export type BaseProject = z.infer<typeof ZBaseProject>;

export const ZCreateProjectRequest = z.lazy(() => (z.object({
    name: z.string(),
    json_schema: z.record(z.string(), z.any()),
    default_inference_settings: ZInferenceSettings,
})));
export type CreateProjectRequest = z.infer<typeof ZCreateProjectRequest>;

export const ZListProjectParams = z.lazy(() => (z.object({
    schema_id: z.string().nullable().optional(),
    schema_data_id: z.string().nullable().optional(),
})));
export type ListProjectParams = z.infer<typeof ZListProjectParams>;

export const ZPatchProjectRequest = z.lazy(() => (z.object({
    name: z.string().nullable().optional(),
    json_schema: z.record(z.string(), z.any()).nullable().optional(),
    default_inference_settings: ZInferenceSettings.nullable().optional(),
})));
export type PatchProjectRequest = z.infer<typeof ZPatchProjectRequest>;

export const ZModelProject = z.lazy(() => (ZBaseProject.schema).merge(z.object({
    documents: z.array(ZDocumentsProjectDocument),
    iterations: z.array(ZIterationsIteration),
})));
export type ModelProject = z.infer<typeof ZModelProject>;

export const ZModelAddIterationFromJsonlRequest = z.lazy(() => (z.object({
    jsonl_gcs_path: z.string(),
})));
export type ModelAddIterationFromJsonlRequest = z.infer<typeof ZModelAddIterationFromJsonlRequest>;

export const ZAnnotationInputData = z.lazy(() => (z.object({
    data_file: z.string(),
    schema_id: z.string(),
    inference_settings: ZInferenceSettings,
})));
export type AnnotationInputData = z.infer<typeof ZAnnotationInputData>;

export const ZFinetuningWorkflowInputData = z.lazy(() => (z.object({
    prepare_dataset_input_data: ZPrepareDatasetInputData,
    annotation_model: z.union([z.literal("human"), z.string()]),
    inference_settings: ZInferenceSettings.nullable().optional(),
    finetuning_props: ZInferenceSettings,
})));
export type FinetuningWorkflowInputData = z.infer<typeof ZFinetuningWorkflowInputData>;

export const ZPrepareDatasetInputData = z.lazy(() => (z.object({
    dataset_id: z.string().nullable().optional(),
    schema_id: z.string().nullable().optional(),
    schema_data_id: z.string().nullable().optional(),
    selection_model: z.union([z.literal("all"), z.literal("manual")]).default("all"),
})));
export type PrepareDatasetInputData = z.infer<typeof ZPrepareDatasetInputData>;

export const ZProjectInputData = z.lazy(() => (z.object({
    eval_data_file: z.string(),
    schema_id: z.string(),
    inference_settings_1: ZInferenceSettings.nullable().optional(),
    inference_settings_2: ZInferenceSettings,
})));
export type ProjectInputData = z.infer<typeof ZProjectInputData>;

export const ZStandaloneAnnotationWorkflowInputData = z.lazy(() => (ZAnnotationInputData.schema).merge(z.object({
})));
export type StandaloneAnnotationWorkflowInputData = z.infer<typeof ZStandaloneAnnotationWorkflowInputData>;

export const ZStandaloneProjectWorkflowInputData = z.lazy(() => (ZProjectInputData.schema).merge(z.object({
})));
export type StandaloneProjectWorkflowInputData = z.infer<typeof ZStandaloneProjectWorkflowInputData>;

export const ZMessageParam = z.lazy(() => (z.object({
    content: z.union([z.string(), z.array(z.union([ZTextBlockParam, ZImageBlockParam, ZDocumentBlockParam, ZThinkingBlockParam, ZRedactedThinkingBlockParam, ZToolUseBlockParam, ZToolResultBlockParam, ZServerToolUseBlockParam, ZWebSearchToolResultBlockParam, z.union([ZTextBlock, ZThinkingBlock, ZRedactedThinkingBlock, ZToolUseBlock, ZServerToolUseBlock, ZWebSearchToolResultBlock])]))]),
    role: z.union([z.literal("user"), z.literal("assistant")]),
})));
export type MessageParam = z.infer<typeof ZMessageParam>;

export const ZPartialSchema = z.lazy(() => (z.object({
    object: z.literal("schema").default("schema"),
    created_at: z.string(),
    json_schema: z.record(z.string(), z.any()).default({}),
    strict: z.boolean().default(true),
})));
export type PartialSchema = z.infer<typeof ZPartialSchema>;

export const ZPartialSchemaChunk = z.lazy(() => (ZStreamingBaseModel.schema).merge(z.object({
    object: z.literal("schema.chunk").default("schema.chunk"),
    created_at: z.string(),
    delta_json_schema_flat: z.record(z.string(), z.any()).default({}),
    delta_flat_deleted_keys: z.array(z.string()).default([]),
})));
export type PartialSchemaChunk = z.infer<typeof ZPartialSchemaChunk>;

export const ZSchema = z.lazy(() => (ZPartialSchema.schema).merge(z.object({
    object: z.literal("schema").default("schema"),
    created_at: z.string(),
    json_schema: z.record(z.string(), z.any()).default({}),
})));
export type Schema = z.infer<typeof ZSchema>;

export const ZGenerateSchemaRequest = z.lazy(() => (z.object({
    documents: z.array(ZMIMEData),
    model: z.string().default("gpt-4o-mini"),
    temperature: z.number().default(0.0),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional().default("medium"),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]),
    instructions: z.string().nullable().optional(),
    image_resolution_dpi: z.number().default(96),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]).default("A4"),
    stream: z.boolean().default(false),
})));
export type GenerateSchemaRequest = z.infer<typeof ZGenerateSchemaRequest>;

export const ZGenerateSystemPromptRequest = z.lazy(() => (ZGenerateSchemaRequest.schema).merge(z.object({
    json_schema: z.record(z.string(), z.any()),
})));
export type GenerateSystemPromptRequest = z.infer<typeof ZGenerateSystemPromptRequest>;

export const ZColumn = z.lazy(() => (z.object({
    type: z.literal("column"),
    size: z.number(),
    items: z.array(z.union([ZRow, ZFieldItem, ZRefObject, ZRowList])),
    name: z.string().nullable().optional(),
})));
export type Column = z.infer<typeof ZColumn>;

export const ZFieldItem = z.lazy(() => (z.object({
    type: z.literal("field"),
    name: z.string(),
    size: z.number().nullable().optional(),
})));
export type FieldItem = z.infer<typeof ZFieldItem>;

export const ZLayout = z.lazy(() => (z.object({
    defs: z.record(z.string(), ZColumn),
    type: z.literal("column"),
    size: z.number(),
    items: z.array(z.union([ZRow, ZRowList, ZFieldItem, ZRefObject])),
})));
export type Layout = z.infer<typeof ZLayout>;

export const ZRefObject = z.lazy(() => (z.object({
    type: z.literal("object"),
    size: z.number().nullable().optional(),
    name: z.string().nullable().optional(),
    ref: z.string(),
})));
export type RefObject = z.infer<typeof ZRefObject>;

export const ZRow = z.lazy(() => (z.object({
    type: z.literal("row"),
    name: z.string().nullable().optional(),
    items: z.array(z.union([ZColumn, ZFieldItem, ZRefObject])),
})));
export type Row = z.infer<typeof ZRow>;

export const ZRowList = z.lazy(() => (z.object({
    type: z.literal("rowList"),
    name: z.string().nullable().optional(),
    items: z.array(z.union([ZColumn, ZFieldItem, ZRefObject])),
})));
export type RowList = z.infer<typeof ZRowList>;

export const ZEnhanceSchemaConfig = z.lazy(() => (z.object({
    allow_reasoning_fields_added: z.boolean().default(true),
    allow_field_description_update: z.boolean().default(false),
    allow_system_prompt_update: z.boolean().default(true),
    allow_field_simple_type_change: z.boolean().default(false),
    allow_field_data_structure_breakdown: z.boolean().default(false),
})));
export type EnhanceSchemaConfig = z.infer<typeof ZEnhanceSchemaConfig>;

export const ZEnhanceSchemaConfigDict = z.lazy(() => (z.object({
    allow_reasoning_fields_added: z.boolean(),
    allow_field_description_update: z.boolean(),
    allow_system_prompt_update: z.boolean(),
    allow_field_simple_type_change: z.boolean(),
    allow_field_data_structure_breakdown: z.boolean(),
})));
export type EnhanceSchemaConfigDict = z.infer<typeof ZEnhanceSchemaConfigDict>;

export const ZEnhanceSchemaRequest = z.lazy(() => (z.object({
    documents: z.array(ZMIMEData),
    ground_truths: z.array(z.record(z.string(), z.any())).nullable().optional(),
    model: z.string().default("gpt-4o-mini"),
    temperature: z.number().default(0.0),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional().default("medium"),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]),
    image_resolution_dpi: z.number().default(96),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]).default("A4"),
    stream: z.boolean().default(false),
    tools_config: ZEnhanceSchemaConfig,
    json_schema: z.record(z.string(), z.any()),
    instructions: z.string().nullable().optional(),
    flat_likelihoods: z.union([z.array(z.record(z.string(), z.number())), z.record(z.string(), z.number())]).nullable().optional(),
})));
export type EnhanceSchemaRequest = z.infer<typeof ZEnhanceSchemaRequest>;

export const ZTemplateSchema = z.lazy(() => (z.object({
    id: z.string(),
    name: z.string(),
    object: z.literal("template").default("template"),
    updated_at: z.string(),
    json_schema: z.record(z.string(), z.any()).default({}),
    python_code: z.string().nullable().optional(),
    sample_document_filename: z.string().nullable().optional(),
})));
export type TemplateSchema = z.infer<typeof ZTemplateSchema>;

export const ZUpdateTemplateRequest = z.lazy(() => (z.object({
    id: z.string(),
    name: z.string().nullable().optional(),
    json_schema: z.record(z.string(), z.any()).nullable().optional(),
    python_code: z.string().nullable().optional(),
    sample_document: ZMIMEData.nullable().optional(),
})));
export type UpdateTemplateRequest = z.infer<typeof ZUpdateTemplateRequest>;

export const ZEvaluateSchemaRequest = z.lazy(() => (z.object({
    documents: z.array(ZMIMEData),
    ground_truths: z.array(z.record(z.string(), z.any())).nullable().optional(),
    model: z.string().default("gpt-4o-mini"),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional().default("medium"),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]),
    image_resolution_dpi: z.number().default(96),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]).default("A4"),
    n_consensus: z.number().default(1),
    json_schema: z.record(z.string(), z.any()),
})));
export type EvaluateSchemaRequest = z.infer<typeof ZEvaluateSchemaRequest>;

export const ZEvaluateSchemaResponse = z.lazy(() => (z.object({
    item_metrics: z.array(ZItemMetric),
})));
export type EvaluateSchemaResponse = z.infer<typeof ZEvaluateSchemaResponse>;

export const ZBinaryIO = z.lazy(() => z.instanceof(Uint8Array));
export type BinaryIO = z.infer<typeof ZBinaryIO>;

export const ZDBFile = z.lazy(() => (z.object({
    object: z.literal("file").default("file"),
    id: z.string(),
    filename: z.string(),
})));
export type DBFile = z.infer<typeof ZDBFile>;

export const ZFileLink = z.lazy(() => (z.object({
    download_url: z.string(),
    expires_in: z.string(),
    filename: z.string(),
})));
export type FileLink = z.infer<typeof ZFileLink>;

export const ZAnnotation = z.lazy(() => (z.object({
    file_id: z.string(),
    parameters: ZAnnotationParameters,
    data: z.record(z.string(), z.any()),
    schema_id: z.string(),
    organization_id: z.string(),
    updated_at: z.string(),
})));
export type Annotation = z.infer<typeof ZAnnotation>;

export const ZAnnotationParameters = z.lazy(() => (z.object({
    model: z.string(),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]).nullable().optional().default("native"),
    image_resolution_dpi: z.number().default(96),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]).default("A4"),
    temperature: z.number().default(0.0),
})));
export type AnnotationParameters = z.infer<typeof ZAnnotationParameters>;

export const ZCronSchedule = z.lazy(() => (z.object({
    second: z.number().nullable().optional().default(0),
    minute: z.number(),
    hour: z.number(),
    day_of_month: z.number().nullable().optional(),
    month: z.number().nullable().optional(),
    day_of_week: z.number().nullable().optional(),
})));
export type CronSchedule = z.infer<typeof ZCronSchedule>;

export const ZScrappingConfig = z.lazy(() => (ZAutomationConfig.schema).merge(z.object({
    id: z.string(),
    updated_at: z.string(),
    webhook_url: z.string(),
    webhook_headers: z.record(z.string(), z.string()),
    link: z.string(),
    schedule: ZCronSchedule,
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]),
    image_resolution_dpi: z.number().default(96),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]).default("A4"),
    model: z.string(),
    json_schema: z.record(z.string(), z.any()),
    temperature: z.number().default(0.0),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional().default("medium"),
})));
export type ScrappingConfig = z.infer<typeof ZScrappingConfig>;

export const ZListMailboxes = z.lazy(() => (z.object({
    data: z.array(ZMailbox),
    list_metadata: ZListMetadata,
})));
export type ListMailboxes = z.infer<typeof ZListMailboxes>;

export const ZMailbox = z.lazy(() => (ZAutomationConfig.schema).merge(z.object({
    id: z.string(),
    email: z.string(),
    authorized_domains: z.array(z.string()),
    authorized_emails: z.array(z.string().email()),
})));
export type Mailbox = z.infer<typeof ZMailbox>;

export const ZUpdateMailboxRequest = z.lazy(() => (ZUpdateAutomationRequest.schema).merge(z.object({
    authorized_domains: z.array(z.string()).nullable().optional(),
    authorized_emails: z.array(z.string().email()).nullable().optional(),
})));
export type UpdateMailboxRequest = z.infer<typeof ZUpdateMailboxRequest>;

export const ZLink = z.lazy(() => (ZAutomationConfig.schema).merge(z.object({
    id: z.string(),
    password: z.string().nullable().optional(),
})));
export type Link = z.infer<typeof ZLink>;

export const ZListLinks = z.lazy(() => (z.object({
    data: z.array(ZLink),
    list_metadata: ZListMetadata,
})));
export type ListLinks = z.infer<typeof ZListLinks>;

export const ZUpdateLinkRequest = z.lazy(() => (ZUpdateAutomationRequest.schema).merge(z.object({
    password: z.string().nullable().optional(),
})));
export type UpdateLinkRequest = z.infer<typeof ZUpdateLinkRequest>;

export const ZAutomationLevel = z.lazy(() => (z.object({
    distance_threshold: z.number().default(0.9),
    score_threshold: z.number().default(0.9),
})));
export type AutomationLevel = z.infer<typeof ZAutomationLevel>;

export const ZFetchParams = z.lazy(() => (z.object({
    endpoint: z.string(),
    headers: z.record(z.string(), z.string()),
    name: z.string(),
})));
export type FetchParams = z.infer<typeof ZFetchParams>;

export const ZListOutlooks = z.lazy(() => (z.object({
    data: z.array(ZOutlook),
    list_metadata: ZListMetadata,
})));
export type ListOutlooks = z.infer<typeof ZListOutlooks>;

export const ZMatchParams = z.lazy(() => (z.object({
    endpoint: z.string(),
    headers: z.record(z.string(), z.string()),
    path: z.string(),
})));
export type MatchParams = z.infer<typeof ZMatchParams>;

export const ZOutlook = z.lazy(() => (ZAutomationConfig.schema).merge(z.object({
    id: z.string(),
    authorized_domains: z.array(z.string()),
    authorized_emails: z.array(z.string().email()),
    layout_schema: z.record(z.string(), z.any()).nullable().optional(),
    match_params: z.array(ZMatchParams),
    fetch_params: z.array(ZFetchParams),
})));
export type Outlook = z.infer<typeof ZOutlook>;

export const ZUpdateOutlookRequest = z.lazy(() => (ZUpdateAutomationRequest.schema).merge(z.object({
    authorized_domains: z.array(z.string()).nullable().optional(),
    authorized_emails: z.array(z.string().email()).nullable().optional(),
    match_params: z.array(ZMatchParams).nullable().optional(),
    fetch_params: z.array(ZFetchParams).nullable().optional(),
    layout_schema: z.record(z.string(), z.any()).nullable().optional(),
})));
export type UpdateOutlookRequest = z.infer<typeof ZUpdateOutlookRequest>;

export const ZBaseWebhookRequest = z.lazy(() => (z.object({
    completion: ZRetabParsedChatCompletion,
    user: z.string().email().nullable().optional(),
    file_payload: ZBaseMIMEData,
    metadata: z.record(z.string(), z.any()).nullable().optional(),
})));
export type BaseWebhookRequest = z.infer<typeof ZBaseWebhookRequest>;

export const ZWebhookRequest = z.lazy(() => (z.object({
    completion: ZRetabParsedChatCompletion,
    user: z.string().email().nullable().optional(),
    file_payload: ZMIMEData,
    metadata: z.record(z.string(), z.any()).nullable().optional(),
})));
export type WebhookRequest = z.infer<typeof ZWebhookRequest>;

export const ZEndpoint = z.lazy(() => (ZAutomationConfig.schema).merge(z.object({
    id: z.string(),
})));
export type Endpoint = z.infer<typeof ZEndpoint>;

export const ZListEndpoints = z.lazy(() => (z.object({
    data: z.array(ZEndpoint),
    list_metadata: ZListMetadata,
})));
export type ListEndpoints = z.infer<typeof ZListEndpoints>;

export const ZUpdateEndpointRequest = z.lazy(() => (ZUpdateAutomationRequest.schema).merge(z.object({
})));
export type UpdateEndpointRequest = z.infer<typeof ZUpdateEndpointRequest>;

export const ZChatCompletionChunk = z.lazy(() => (z.object({
    id: z.string(),
    choices: z.array(ZChoice),
    created: z.number(),
    model: z.string(),
    object: z.literal("chat.completion.chunk"),
    service_tier: z.union([z.literal("auto"), z.literal("default"), z.literal("flex"), z.literal("scale"), z.literal("priority")]).nullable().optional(),
    system_fingerprint: z.string().nullable().optional(),
    usage: ZCompletionUsage.nullable().optional(),
})));
export type ChatCompletionChunk = z.infer<typeof ZChatCompletionChunk>;

export const ZChoice = z.lazy(() => (z.object({
    delta: ZChoiceDelta,
    finish_reason: z.union([z.literal("stop"), z.literal("length"), z.literal("tool_calls"), z.literal("content_filter"), z.literal("function_call")]).nullable().optional(),
    index: z.number(),
    logprobs: ZChoiceLogprobs.nullable().optional(),
})));
export type Choice = z.infer<typeof ZChoice>;

export const ZChoiceDelta = z.lazy(() => (z.object({
    content: z.string().nullable().optional(),
    function_call: ZChoiceDeltaFunctionCall.nullable().optional(),
    refusal: z.string().nullable().optional(),
    role: z.union([z.literal("developer"), z.literal("system"), z.literal("user"), z.literal("assistant"), z.literal("tool")]).nullable().optional(),
    tool_calls: z.array(ZChoiceDeltaToolCall).nullable().optional(),
})));
export type ChoiceDelta = z.infer<typeof ZChoiceDelta>;

export const ZConsensusModel = z.lazy(() => (z.object({
    model: z.string(),
    temperature: z.number().default(0.0),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional().default("medium"),
})));
export type ConsensusModel = z.infer<typeof ZConsensusModel>;

export const ZDocumentExtractRequest = z.lazy(() => (z.object({
    document: ZMIMEData,
    documents: z.array(ZMIMEData).default([]),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]).default("native"),
    image_resolution_dpi: z.number().default(96),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]).default("A4"),
    model: z.string(),
    json_schema: z.record(z.string(), z.any()),
    temperature: z.number().default(0.0),
    reasoning_effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional().default("medium"),
    n_consensus: z.number().default(1),
    stream: z.boolean().default(false),
    seed: z.number().nullable().optional(),
    store: z.boolean().default(true),
    need_validation: z.boolean().default(false),
})));
export type DocumentExtractRequest = z.infer<typeof ZDocumentExtractRequest>;

export const ZFieldLocation = z.lazy(() => (z.object({
    label: z.string(),
    value: z.string(),
    quote: z.string(),
    file_id: z.string().nullable().optional(),
    page: z.number().nullable().optional(),
    bbox_normalized: z.tuple([z.number(), z.number(), z.number(), z.number()]).nullable().optional(),
    score: z.number().nullable().optional(),
    match_level: z.union([z.literal("token"), z.literal("line"), z.literal("block")]).nullable().optional(),
})));
export type FieldLocation = z.infer<typeof ZFieldLocation>;

export const ZLogExtractionRequest = z.lazy(() => (z.object({
    messages: z.array(ZChatCompletionRetabMessage).nullable().optional(),
    openai_messages: z.array(z.union([ZChatCompletionDeveloperMessageParam, ZChatCompletionSystemMessageParam, ZChatCompletionUserMessageParam, ZChatCompletionAssistantMessageParam, ZChatCompletionToolMessageParam, ZChatCompletionFunctionMessageParam])).nullable().optional(),
    openai_responses_input: z.array(z.union([ZEasyInputMessageParam, ZResponseInputParamMessage, ZResponseOutputMessageParam, ZResponseFileSearchToolCallParam, ZResponseComputerToolCallParam, ZComputerCallOutput, ZResponseFunctionWebSearchParam, ZResponseFunctionToolCallParam, ZFunctionCallOutput, ZResponseReasoningItemParam, ZImageGenerationCall, ZResponseCodeInterpreterToolCallParam, ZLocalShellCall, ZLocalShellCallOutput, ZMcpListTools, ZMcpApprovalRequest, ZMcpApprovalResponse, ZMcpCall, ZItemReference])).nullable().optional(),
    anthropic_messages: z.array(ZMessageParam).nullable().optional(),
    anthropic_system_prompt: z.string().nullable().optional(),
    document: ZMIMEData.default({"filename": "dummy.txt", "url": "data:text/plain;base64,Tm8gZG9jdW1lbnQgcHJvdmlkZWQ="}),
    completion: z.union([z.record(z.any()), ZRetabParsedChatCompletion, ZMessage, ZParsedChatCompletion, ZChatCompletion]).nullable().optional(),
    openai_responses_output: ZResponse.nullable().optional(),
    json_schema: z.record(z.string(), z.any()),
    model: z.string(),
    temperature: z.number(),
})));
export type LogExtractionRequest = z.infer<typeof ZLogExtractionRequest>;

export const ZLogExtractionResponse = z.lazy(() => (z.object({
    extraction_id: z.string().nullable().optional(),
    status: z.union([z.literal("success"), z.literal("error")]),
    error_message: z.string().nullable().optional(),
})));
export type LogExtractionResponse = z.infer<typeof ZLogExtractionResponse>;

export const ZMessage = z.lazy(() => (z.object({
    id: z.string(),
    content: z.array(z.union([ZTextBlock, ZThinkingBlock, ZRedactedThinkingBlock, ZToolUseBlock, ZServerToolUseBlock, ZWebSearchToolResultBlock])),
    model: z.union([z.union([z.literal("claude-3-7-sonnet-latest"), z.literal("claude-3-7-sonnet-20250219"), z.literal("claude-3-5-haiku-latest"), z.literal("claude-3-5-haiku-20241022"), z.literal("claude-sonnet-4-20250514"), z.literal("claude-sonnet-4-0"), z.literal("claude-4-sonnet-20250514"), z.literal("claude-3-5-sonnet-latest"), z.literal("claude-3-5-sonnet-20241022"), z.literal("claude-3-5-sonnet-20240620"), z.literal("claude-opus-4-0"), z.literal("claude-opus-4-20250514"), z.literal("claude-4-opus-20250514"), z.literal("claude-3-opus-latest"), z.literal("claude-3-opus-20240229"), z.literal("claude-3-sonnet-20240229"), z.literal("claude-3-haiku-20240307"), z.literal("claude-2.1"), z.literal("claude-2.0")]), z.string()]),
    role: z.literal("assistant"),
    stop_reason: z.union([z.literal("end_turn"), z.literal("max_tokens"), z.literal("stop_sequence"), z.literal("tool_use"), z.literal("pause_turn"), z.literal("refusal")]).nullable().optional(),
    stop_sequence: z.string().nullable().optional(),
    type: z.literal("message"),
    usage: ZUsage,
})));
export type Message = z.infer<typeof ZMessage>;

export const ZParsedChatCompletion = z.lazy(() => (ZChatCompletion.schema).merge(z.object({
    choices: z.array(ZParsedChoice),
})));
export type ParsedChatCompletion = z.infer<typeof ZParsedChatCompletion>;

export const ZParsedChoice = z.lazy(() => (ZChatCompletionChoice.schema).merge(z.object({
    message: ZParsedChatCompletionMessage,
})));
export type ParsedChoice = z.infer<typeof ZParsedChoice>;

export const ZResponse = z.lazy(() => (z.object({
    id: z.string(),
    created_at: z.number(),
    error: ZResponseError.nullable().optional(),
    incomplete_details: ZIncompleteDetails.nullable().optional(),
    instructions: z.union([z.string(), z.array(z.union([ZEasyInputMessage, ZResponseInputItemMessage, ZResponseOutputMessage, ZResponseFileSearchToolCall, ZResponseComputerToolCall, ZResponseInputItemComputerCallOutput, ZResponseFunctionWebSearch, ZResponseFunctionToolCall, ZResponseInputItemFunctionCallOutput, ZResponseReasoningItem, ZResponseInputItemImageGenerationCall, ZResponseCodeInterpreterToolCall, ZResponseInputItemLocalShellCall, ZResponseInputItemLocalShellCallOutput, ZResponseInputItemMcpListTools, ZResponseInputItemMcpApprovalRequest, ZResponseInputItemMcpApprovalResponse, ZResponseInputItemMcpCall, ZResponseInputItemItemReference]))]).nullable().optional(),
    metadata: z.record(z.string(), z.string()).nullable().optional(),
    model: z.union([z.string(), z.union([z.literal("gpt-4.1"), z.literal("gpt-4.1-mini"), z.literal("gpt-4.1-nano"), z.literal("gpt-4.1-2025-04-14"), z.literal("gpt-4.1-mini-2025-04-14"), z.literal("gpt-4.1-nano-2025-04-14"), z.literal("o4-mini"), z.literal("o4-mini-2025-04-16"), z.literal("o3"), z.literal("o3-2025-04-16"), z.literal("o3-mini"), z.literal("o3-mini-2025-01-31"), z.literal("o1"), z.literal("o1-2024-12-17"), z.literal("o1-preview"), z.literal("o1-preview-2024-09-12"), z.literal("o1-mini"), z.literal("o1-mini-2024-09-12"), z.literal("gpt-4o"), z.literal("gpt-4o-2024-11-20"), z.literal("gpt-4o-2024-08-06"), z.literal("gpt-4o-2024-05-13"), z.literal("gpt-4o-audio-preview"), z.literal("gpt-4o-audio-preview-2024-10-01"), z.literal("gpt-4o-audio-preview-2024-12-17"), z.literal("gpt-4o-audio-preview-2025-06-03"), z.literal("gpt-4o-mini-audio-preview"), z.literal("gpt-4o-mini-audio-preview-2024-12-17"), z.literal("gpt-4o-search-preview"), z.literal("gpt-4o-mini-search-preview"), z.literal("gpt-4o-search-preview-2025-03-11"), z.literal("gpt-4o-mini-search-preview-2025-03-11"), z.literal("chatgpt-4o-latest"), z.literal("codex-mini-latest"), z.literal("gpt-4o-mini"), z.literal("gpt-4o-mini-2024-07-18"), z.literal("gpt-4-turbo"), z.literal("gpt-4-turbo-2024-04-09"), z.literal("gpt-4-0125-preview"), z.literal("gpt-4-turbo-preview"), z.literal("gpt-4-1106-preview"), z.literal("gpt-4-vision-preview"), z.literal("gpt-4"), z.literal("gpt-4-0314"), z.literal("gpt-4-0613"), z.literal("gpt-4-32k"), z.literal("gpt-4-32k-0314"), z.literal("gpt-4-32k-0613"), z.literal("gpt-3.5-turbo"), z.literal("gpt-3.5-turbo-16k"), z.literal("gpt-3.5-turbo-0301"), z.literal("gpt-3.5-turbo-0613"), z.literal("gpt-3.5-turbo-1106"), z.literal("gpt-3.5-turbo-0125"), z.literal("gpt-3.5-turbo-16k-0613")]), z.union([z.literal("o1-pro"), z.literal("o1-pro-2025-03-19"), z.literal("o3-pro"), z.literal("o3-pro-2025-06-10"), z.literal("o3-deep-research"), z.literal("o3-deep-research-2025-06-26"), z.literal("o4-mini-deep-research"), z.literal("o4-mini-deep-research-2025-06-26"), z.literal("computer-use-preview"), z.literal("computer-use-preview-2025-03-11")])]),
    object: z.literal("response"),
    output: z.array(z.union([ZResponseOutputMessage, ZResponseFileSearchToolCall, ZResponseFunctionToolCall, ZResponseFunctionWebSearch, ZResponseComputerToolCall, ZResponseReasoningItem, ZResponseOutputItemImageGenerationCall, ZResponseCodeInterpreterToolCall, ZResponseOutputItemLocalShellCall, ZResponseOutputItemMcpCall, ZResponseOutputItemMcpListTools, ZResponseOutputItemMcpApprovalRequest])),
    parallel_tool_calls: z.boolean(),
    temperature: z.number().nullable().optional(),
    tool_choice: z.union([z.union([z.literal("none"), z.literal("auto"), z.literal("required")]), ZToolChoiceTypes, ZToolChoiceFunction, ZToolChoiceMcp]),
    tools: z.array(z.union([ZFunctionTool, ZFileSearchTool, ZWebSearchTool, ZComputerTool, ZMcp, ZCodeInterpreter, ZImageGeneration, ZLocalShell])),
    top_p: z.number().nullable().optional(),
    background: z.boolean().nullable().optional(),
    max_output_tokens: z.number().nullable().optional(),
    max_tool_calls: z.number().nullable().optional(),
    previous_response_id: z.string().nullable().optional(),
    prompt: ZResponsePrompt.nullable().optional(),
    reasoning: ZReasoningReasoning.nullable().optional(),
    service_tier: z.union([z.literal("auto"), z.literal("default"), z.literal("flex"), z.literal("scale"), z.literal("priority")]).nullable().optional(),
    status: z.union([z.literal("completed"), z.literal("failed"), z.literal("in_progress"), z.literal("cancelled"), z.literal("queued"), z.literal("incomplete")]).nullable().optional(),
    text: ZResponseTextConfig.nullable().optional(),
    top_logprobs: z.number().nullable().optional(),
    truncation: z.union([z.literal("auto"), z.literal("disabled")]).nullable().optional(),
    usage: ZResponseUsage.nullable().optional(),
    user: z.string().nullable().optional(),
})));
export type Response = z.infer<typeof ZResponse>;

export const ZRetabParsedChatCompletionChunk = z.lazy(() => (ZStreamingBaseModel.schema).merge(ZChatCompletionChunk.schema).merge(z.object({
    choices: z.array(ZRetabParsedChoiceChunk),
    extraction_id: z.string().nullable().optional(),
    schema_validation_error: ZErrorDetail.nullable().optional(),
    request_at: z.string().nullable().optional(),
    first_token_at: z.string().nullable().optional(),
    last_token_at: z.string().nullable().optional(),
})));
export type RetabParsedChatCompletionChunk = z.infer<typeof ZRetabParsedChatCompletionChunk>;

export const ZRetabParsedChoice = z.lazy(() => (ZParsedChoice.schema).merge(z.object({
    finish_reason: z.union([z.literal("stop"), z.literal("length"), z.literal("tool_calls"), z.literal("content_filter"), z.literal("function_call")]).nullable().optional(),
    field_locations: z.record(z.string(), ZFieldLocation).nullable().optional(),
    key_mapping: z.record(z.string(), z.string().nullable().optional()).nullable().optional(),
})));
export type RetabParsedChoice = z.infer<typeof ZRetabParsedChoice>;

export const ZRetabParsedChoiceChunk = z.lazy(() => (ZChoice.schema).merge(z.object({
    delta: ZRetabParsedChoiceDeltaChunk,
})));
export type RetabParsedChoiceChunk = z.infer<typeof ZRetabParsedChoiceChunk>;

export const ZRetabParsedChoiceDeltaChunk = z.lazy(() => (ZChoiceDelta.schema).merge(z.object({
    flat_likelihoods: z.record(z.string(), z.number()).default({}),
    flat_parsed: z.record(z.string(), z.any()).default({}),
    flat_deleted_keys: z.array(z.string()).default([]),
    field_locations: z.record(z.string(), z.array(ZFieldLocation)).nullable().optional(),
    is_valid_json: z.boolean().default(false),
    key_mapping: z.record(z.string(), z.string().nullable().optional()).nullable().optional(),
})));
export type RetabParsedChoiceDeltaChunk = z.infer<typeof ZRetabParsedChoiceDeltaChunk>;

export const ZUiResponse = z.lazy(() => (ZResponse.schema).merge(z.object({
    extraction_id: z.string().nullable().optional(),
    likelihoods: z.record(z.string(), z.any()).nullable().optional(),
    schema_validation_error: ZErrorDetail.nullable().optional(),
    request_at: z.string().nullable().optional(),
    first_token_at: z.string().nullable().optional(),
    last_token_at: z.string().nullable().optional(),
})));
export type UiResponse = z.infer<typeof ZUiResponse>;

export const ZDocumentTransformRequest = z.lazy(() => (z.object({
    document: ZMIMEData,
})));
export type DocumentTransformRequest = z.infer<typeof ZDocumentTransformRequest>;

export const ZDocumentTransformResponse = z.lazy(() => (z.object({
    document: ZMIMEData,
})));
export type DocumentTransformResponse = z.infer<typeof ZDocumentTransformResponse>;

export const ZParseRequest = z.lazy(() => (z.object({
    document: ZMIMEData,
    model: z.union([z.literal("gpt-4o"), z.literal("gpt-4o-mini"), z.literal("chatgpt-4o-latest"), z.literal("gpt-4.1"), z.literal("gpt-4.1-mini"), z.literal("gpt-4.1-mini-2025-04-14"), z.literal("gpt-4.1-2025-04-14"), z.literal("gpt-4.1-nano"), z.literal("gpt-4.1-nano-2025-04-14"), z.literal("gpt-4o-2024-11-20"), z.literal("gpt-4o-2024-08-06"), z.literal("gpt-4o-2024-05-13"), z.literal("gpt-4o-mini-2024-07-18"), z.literal("o1"), z.literal("o1-2024-12-17"), z.literal("o3"), z.literal("o3-2025-04-16"), z.literal("o4-mini"), z.literal("o4-mini-2025-04-16"), z.literal("gpt-4o-audio-preview-2024-12-17"), z.literal("gpt-4o-audio-preview-2024-10-01"), z.literal("gpt-4o-realtime-preview-2024-12-17"), z.literal("gpt-4o-realtime-preview-2024-10-01"), z.literal("gpt-4o-mini-audio-preview-2024-12-17"), z.literal("gpt-4o-mini-realtime-preview-2024-12-17"), z.literal("claude-3-5-sonnet-latest"), z.literal("claude-3-5-sonnet-20241022"), z.literal("claude-3-opus-20240229"), z.literal("claude-3-sonnet-20240229"), z.literal("claude-3-haiku-20240307"), z.literal("grok-3"), z.literal("grok-3-mini"), z.literal("gemini-2.5-pro"), z.literal("gemini-2.5-flash"), z.literal("gemini-2.5-pro-preview-06-05"), z.literal("gemini-2.5-pro-preview-05-06"), z.literal("gemini-2.5-pro-preview-03-25"), z.literal("gemini-2.5-flash-preview-05-20"), z.literal("gemini-2.5-flash-preview-04-17"), z.literal("gemini-2.5-flash-lite-preview-06-17"), z.literal("gemini-2.5-pro-exp-03-25"), z.literal("gemini-2.0-flash-lite"), z.literal("gemini-2.0-flash"), z.literal("auto-large"), z.literal("auto-small"), z.literal("auto-micro"), z.literal("human")]).default("gemini-2.5-flash"),
    table_parsing_format: z.union([z.literal("markdown"), z.literal("yaml"), z.literal("html"), z.literal("json")]).default("html"),
    image_resolution_dpi: z.number().default(96),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]).default("A4"),
})));
export type ParseRequest = z.infer<typeof ZParseRequest>;

export const ZParseResult = z.lazy(() => (z.object({
    document: ZBaseMIMEData,
    usage: ZRetabUsage,
    pages: z.array(z.string()),
    text: z.string(),
})));
export type ParseResult = z.infer<typeof ZParseResult>;

export const ZRetabUsage = z.lazy(() => (z.object({
    page_count: z.number(),
    credits: z.number(),
})));
export type RetabUsage = z.infer<typeof ZRetabUsage>;

export const ZDocumentCreateInputRequest = z.lazy(() => (ZDocumentCreateMessageRequest.schema).merge(z.object({
    json_schema: z.record(z.string(), z.any()),
})));
export type DocumentCreateInputRequest = z.infer<typeof ZDocumentCreateInputRequest>;

export const ZDocumentCreateMessageRequest = z.lazy(() => (z.object({
    document: ZMIMEData,
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]).default("native"),
    image_resolution_dpi: z.number().default(96),
    browser_canvas: z.union([z.literal("A3"), z.literal("A4"), z.literal("A5")]).default("A4"),
})));
export type DocumentCreateMessageRequest = z.infer<typeof ZDocumentCreateMessageRequest>;

export const ZDocumentMessage = z.lazy(() => (z.object({
    id: z.string(),
    object: z.literal("document_message").default("document_message"),
    messages: z.array(ZChatCompletionRetabMessage),
    created: z.number(),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native"), z.literal("image+text")]).default("native"),
})));
export type DocumentMessage = z.infer<typeof ZDocumentMessage>;

export const ZTokenCount = z.lazy(() => (z.object({
    total_tokens: z.number().default(0),
    developer_tokens: z.number().default(0),
    user_tokens: z.number().default(0),
})));
export type TokenCount = z.infer<typeof ZTokenCount>;

export const ZEvaluationProjectInputData = z.lazy(() => (z.object({
    original_dataset_id: z.string(),
    schema_id: z.string(),
    schema_data_id: z.string(),
    inference_settings_1: ZInferenceSettings,
    inference_settings_2: ZInferenceSettings,
})));
export type EvaluationProjectInputData = z.infer<typeof ZEvaluationProjectInputData>;

export const ZBatchAnnotationAnnotationInputData = z.lazy(() => (z.object({
    dataset_id: z.string(),
    files_ids: z.array(z.string()).nullable().optional(),
    upsert: z.boolean().default(false),
    inference_settings: ZInferenceSettings,
})));
export type BatchAnnotationAnnotationInputData = z.infer<typeof ZBatchAnnotationAnnotationInputData>;

export const ZDatasetSplitInputData = z.lazy(() => (z.object({
    dataset_id: z.string(),
    train_size: z.union([z.number(), z.number()]).nullable().optional(),
    eval_size: z.union([z.number(), z.number()]).nullable().optional(),
})));
export type DatasetSplitInputData = z.infer<typeof ZDatasetSplitInputData>;

export const ZChatCompletionChoice = z.lazy(() => (z.object({
    finish_reason: z.union([z.literal("stop"), z.literal("length"), z.literal("tool_calls"), z.literal("content_filter"), z.literal("function_call")]),
    index: z.number(),
    logprobs: ZChatCompletionChoiceLogprobs.nullable().optional(),
    message: ZChatCompletionMessage,
})));
export type ChatCompletionChoice = z.infer<typeof ZChatCompletionChoice>;

export const ZCompletionUsage = z.lazy(() => (z.object({
    completion_tokens: z.number(),
    prompt_tokens: z.number(),
    total_tokens: z.number(),
    completion_tokens_details: ZCompletionTokensDetails.nullable().optional(),
    prompt_tokens_details: ZPromptTokensDetails.nullable().optional(),
})));
export type CompletionUsage = z.infer<typeof ZCompletionUsage>;

export const ZChatCompletionContentPartTextParam = z.lazy(() => (z.object({
    text: z.string(),
    type: z.literal("text"),
})));
export type ChatCompletionContentPartTextParam = z.infer<typeof ZChatCompletionContentPartTextParam>;

export const ZChatCompletionContentPartImageParam = z.lazy(() => (z.object({
    image_url: ZImageURL,
    type: z.literal("image_url"),
})));
export type ChatCompletionContentPartImageParam = z.infer<typeof ZChatCompletionContentPartImageParam>;

export const ZChatCompletionContentPartInputAudioParam = z.lazy(() => (z.object({
    input_audio: ZInputAudio,
    type: z.literal("input_audio"),
})));
export type ChatCompletionContentPartInputAudioParam = z.infer<typeof ZChatCompletionContentPartInputAudioParam>;

export const ZFile = z.lazy(() => (z.object({
    file: ZFileFile,
    type: z.literal("file"),
})));
export type File = z.infer<typeof ZFile>;

export const ZTokenCounts = z.lazy(() => (z.object({
    prompt_regular_text: z.number(),
    prompt_cached_text: z.number(),
    prompt_audio: z.number(),
    completion_regular_text: z.number(),
    completion_audio: z.number(),
    total_tokens: z.number(),
})));
export type TokenCounts = z.infer<typeof ZTokenCounts>;

export const ZJSONSchema = z.lazy(() => (z.object({
    name: z.string(),
    description: z.string(),
    schema: z.record(z.string(), z.object({}).passthrough()),
    strict: z.boolean().nullable().optional(),
})));
export type JSONSchema = z.infer<typeof ZJSONSchema>;

export const ZResponseFormatText = z.lazy(() => (z.object({
    type: z.literal("text"),
})));
export type ResponseFormatText = z.infer<typeof ZResponseFormatText>;

export const ZResponseFormatTextJSONSchemaConfigParam = z.lazy(() => (z.object({
    name: z.string(),
    schema: z.record(z.string(), z.object({}).passthrough()),
    type: z.literal("json_schema"),
    description: z.string(),
    strict: z.boolean().nullable().optional(),
})));
export type ResponseFormatTextJSONSchemaConfigParam = z.infer<typeof ZResponseFormatTextJSONSchemaConfigParam>;

export const ZResponseFormatJSONObject = z.lazy(() => (z.object({
    type: z.literal("json_object"),
})));
export type ResponseFormatJSONObject = z.infer<typeof ZResponseFormatJSONObject>;

export const ZEasyInputMessageParam = z.lazy(() => (z.object({
    content: z.union([z.string(), z.array(z.union([ZResponseInputTextParam, ZResponseInputImageParam, ZResponseInputFileParam]))]),
    role: z.union([z.literal("user"), z.literal("assistant"), z.literal("system"), z.literal("developer")]),
    type: z.literal("message"),
})));
export type EasyInputMessageParam = z.infer<typeof ZEasyInputMessageParam>;

export const ZResponseInputParamMessage = z.lazy(() => (z.object({
    content: z.array(z.union([ZResponseInputTextParam, ZResponseInputImageParam, ZResponseInputFileParam])),
    role: z.union([z.literal("user"), z.literal("system"), z.literal("developer")]),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
    type: z.literal("message"),
})));
export type ResponseInputParamMessage = z.infer<typeof ZResponseInputParamMessage>;

export const ZResponseOutputMessageParam = z.lazy(() => (z.object({
    id: z.string(),
    content: z.array(z.union([ZResponseOutputTextParam, ZResponseOutputRefusalParam])),
    role: z.literal("assistant"),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
    type: z.literal("message"),
})));
export type ResponseOutputMessageParam = z.infer<typeof ZResponseOutputMessageParam>;

export const ZResponseFileSearchToolCallParam = z.lazy(() => (z.object({
    id: z.string(),
    queries: z.array(z.string()),
    status: z.union([z.literal("in_progress"), z.literal("searching"), z.literal("completed"), z.literal("incomplete"), z.literal("failed")]),
    type: z.literal("file_search_call"),
    results: z.array(ZResult).nullable().optional(),
})));
export type ResponseFileSearchToolCallParam = z.infer<typeof ZResponseFileSearchToolCallParam>;

export const ZResponseComputerToolCallParam = z.lazy(() => (z.object({
    id: z.string(),
    action: z.union([ZActionClick, ZActionDoubleClick, ZActionDrag, ZActionKeypress, ZActionMove, ZActionScreenshot, ZActionScroll, ZActionType, ZActionWait]),
    call_id: z.string(),
    pending_safety_checks: z.array(ZPendingSafetyCheck),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
    type: z.literal("computer_call"),
})));
export type ResponseComputerToolCallParam = z.infer<typeof ZResponseComputerToolCallParam>;

export const ZComputerCallOutput = z.lazy(() => (z.object({
    call_id: z.string(),
    output: ZResponseComputerToolCallOutputScreenshotParam,
    type: z.literal("computer_call_output"),
    id: z.string().nullable().optional(),
    acknowledged_safety_checks: z.array(ZComputerCallOutputAcknowledgedSafetyCheck).nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).nullable().optional(),
})));
export type ComputerCallOutput = z.infer<typeof ZComputerCallOutput>;

export const ZResponseFunctionWebSearchParam = z.lazy(() => (z.object({
    id: z.string(),
    action: z.union([ZActionSearch, ZActionOpenPage, ZActionFind]),
    status: z.union([z.literal("in_progress"), z.literal("searching"), z.literal("completed"), z.literal("failed")]),
    type: z.literal("web_search_call"),
})));
export type ResponseFunctionWebSearchParam = z.infer<typeof ZResponseFunctionWebSearchParam>;

export const ZResponseFunctionToolCallParam = z.lazy(() => (z.object({
    arguments: z.string(),
    call_id: z.string(),
    name: z.string(),
    type: z.literal("function_call"),
    id: z.string(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
})));
export type ResponseFunctionToolCallParam = z.infer<typeof ZResponseFunctionToolCallParam>;

export const ZFunctionCallOutput = z.lazy(() => (z.object({
    call_id: z.string(),
    output: z.string(),
    type: z.literal("function_call_output"),
    id: z.string().nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).nullable().optional(),
})));
export type FunctionCallOutput = z.infer<typeof ZFunctionCallOutput>;

export const ZResponseReasoningItemParam = z.lazy(() => (z.object({
    id: z.string(),
    summary: z.array(ZSummary),
    type: z.literal("reasoning"),
    encrypted_content: z.string().nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
})));
export type ResponseReasoningItemParam = z.infer<typeof ZResponseReasoningItemParam>;

export const ZImageGenerationCall = z.lazy(() => (z.object({
    id: z.string(),
    result: z.string().nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("generating"), z.literal("failed")]),
    type: z.literal("image_generation_call"),
})));
export type ImageGenerationCall = z.infer<typeof ZImageGenerationCall>;

export const ZResponseCodeInterpreterToolCallParam = z.lazy(() => (z.object({
    id: z.string(),
    code: z.string().nullable().optional(),
    container_id: z.string(),
    outputs: z.array(z.union([ZOutputLogs, ZOutputImage])).nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete"), z.literal("interpreting"), z.literal("failed")]),
    type: z.literal("code_interpreter_call"),
})));
export type ResponseCodeInterpreterToolCallParam = z.infer<typeof ZResponseCodeInterpreterToolCallParam>;

export const ZLocalShellCall = z.lazy(() => (z.object({
    id: z.string(),
    action: ZLocalShellCallAction,
    call_id: z.string(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
    type: z.literal("local_shell_call"),
})));
export type LocalShellCall = z.infer<typeof ZLocalShellCall>;

export const ZLocalShellCallOutput = z.lazy(() => (z.object({
    id: z.string(),
    output: z.string(),
    type: z.literal("local_shell_call_output"),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).nullable().optional(),
})));
export type LocalShellCallOutput = z.infer<typeof ZLocalShellCallOutput>;

export const ZMcpListTools = z.lazy(() => (z.object({
    id: z.string(),
    server_label: z.string(),
    tools: z.array(ZMcpListToolsTool),
    type: z.literal("mcp_list_tools"),
    error: z.string().nullable().optional(),
})));
export type McpListTools = z.infer<typeof ZMcpListTools>;

export const ZMcpApprovalRequest = z.lazy(() => (z.object({
    id: z.string(),
    arguments: z.string(),
    name: z.string(),
    server_label: z.string(),
    type: z.literal("mcp_approval_request"),
})));
export type McpApprovalRequest = z.infer<typeof ZMcpApprovalRequest>;

export const ZMcpApprovalResponse = z.lazy(() => (z.object({
    approval_request_id: z.string(),
    approve: z.boolean(),
    type: z.literal("mcp_approval_response"),
    id: z.string().nullable().optional(),
    reason: z.string().nullable().optional(),
})));
export type McpApprovalResponse = z.infer<typeof ZMcpApprovalResponse>;

export const ZMcpCall = z.lazy(() => (z.object({
    id: z.string(),
    arguments: z.string(),
    name: z.string(),
    server_label: z.string(),
    type: z.literal("mcp_call"),
    error: z.string().nullable().optional(),
    output: z.string().nullable().optional(),
})));
export type McpCall = z.infer<typeof ZMcpCall>;

export const ZItemReference = z.lazy(() => (z.object({
    id: z.string(),
    type: z.literal("item_reference").nullable().optional(),
})));
export type ItemReference = z.infer<typeof ZItemReference>;

export const ZTextBlockParam = z.lazy(() => (z.object({
    text: z.string(),
    type: z.literal("text"),
    cache_control: ZCacheControlEphemeralParam.nullable().optional(),
    citations: z.array(z.union([ZCitationCharLocationParam, ZCitationPageLocationParam, ZCitationContentBlockLocationParam, ZCitationWebSearchResultLocationParam])).nullable().optional(),
})));
export type TextBlockParam = z.infer<typeof ZTextBlockParam>;

export const ZImageBlockParam = z.lazy(() => (z.object({
    source: z.union([ZBase64ImageSourceParam, ZURLImageSourceParam]),
    type: z.literal("image"),
    cache_control: ZCacheControlEphemeralParam.nullable().optional(),
})));
export type ImageBlockParam = z.infer<typeof ZImageBlockParam>;

export const ZDocumentBlockParam = z.lazy(() => (z.object({
    source: z.union([ZBase64PDFSourceParam, ZPlainTextSourceParam, ZContentBlockSourceParam, ZURLPDFSourceParam]),
    type: z.literal("document"),
    cache_control: ZCacheControlEphemeralParam.nullable().optional(),
    citations: ZCitationsConfigParam,
    context: z.string().nullable().optional(),
    title: z.string().nullable().optional(),
})));
export type DocumentBlockParam = z.infer<typeof ZDocumentBlockParam>;

export const ZThinkingBlockParam = z.lazy(() => (z.object({
    signature: z.string(),
    thinking: z.string(),
    type: z.literal("thinking"),
})));
export type ThinkingBlockParam = z.infer<typeof ZThinkingBlockParam>;

export const ZRedactedThinkingBlockParam = z.lazy(() => (z.object({
    data: z.string(),
    type: z.literal("redacted_thinking"),
})));
export type RedactedThinkingBlockParam = z.infer<typeof ZRedactedThinkingBlockParam>;

export const ZToolUseBlockParam = z.lazy(() => (z.object({
    id: z.string(),
    input: z.object({}).passthrough(),
    name: z.string(),
    type: z.literal("tool_use"),
    cache_control: ZCacheControlEphemeralParam.nullable().optional(),
})));
export type ToolUseBlockParam = z.infer<typeof ZToolUseBlockParam>;

export const ZToolResultBlockParam = z.lazy(() => (z.object({
    tool_use_id: z.string(),
    type: z.literal("tool_result"),
    cache_control: ZCacheControlEphemeralParam.nullable().optional(),
    content: z.union([z.string(), z.array(z.union([ZTextBlockParam, ZImageBlockParam]))]),
    is_error: z.boolean(),
})));
export type ToolResultBlockParam = z.infer<typeof ZToolResultBlockParam>;

export const ZServerToolUseBlockParam = z.lazy(() => (z.object({
    id: z.string(),
    input: z.object({}).passthrough(),
    name: z.literal("web_search"),
    type: z.literal("server_tool_use"),
    cache_control: ZCacheControlEphemeralParam.nullable().optional(),
})));
export type ServerToolUseBlockParam = z.infer<typeof ZServerToolUseBlockParam>;

export const ZWebSearchToolResultBlockParam = z.lazy(() => (z.object({
    content: z.union([z.array(ZWebSearchResultBlockParam), ZWebSearchToolRequestErrorParam]),
    tool_use_id: z.string(),
    type: z.literal("web_search_tool_result"),
    cache_control: ZCacheControlEphemeralParam.nullable().optional(),
})));
export type WebSearchToolResultBlockParam = z.infer<typeof ZWebSearchToolResultBlockParam>;

export const ZTextBlock = z.lazy(() => (z.object({
    citations: z.array(z.union([ZCitationCharLocation, ZCitationPageLocation, ZCitationContentBlockLocation, ZCitationsWebSearchResultLocation])).nullable().optional(),
    text: z.string(),
    type: z.literal("text"),
})));
export type TextBlock = z.infer<typeof ZTextBlock>;

export const ZThinkingBlock = z.lazy(() => (z.object({
    signature: z.string(),
    thinking: z.string(),
    type: z.literal("thinking"),
})));
export type ThinkingBlock = z.infer<typeof ZThinkingBlock>;

export const ZRedactedThinkingBlock = z.lazy(() => (z.object({
    data: z.string(),
    type: z.literal("redacted_thinking"),
})));
export type RedactedThinkingBlock = z.infer<typeof ZRedactedThinkingBlock>;

export const ZToolUseBlock = z.lazy(() => (z.object({
    id: z.string(),
    input: z.object({}).passthrough(),
    name: z.string(),
    type: z.literal("tool_use"),
})));
export type ToolUseBlock = z.infer<typeof ZToolUseBlock>;

export const ZServerToolUseBlock = z.lazy(() => (z.object({
    id: z.string(),
    input: z.object({}).passthrough(),
    name: z.literal("web_search"),
    type: z.literal("server_tool_use"),
})));
export type ServerToolUseBlock = z.infer<typeof ZServerToolUseBlock>;

export const ZWebSearchToolResultBlock = z.lazy(() => (z.object({
    content: z.union([ZWebSearchToolResultError, z.array(ZWebSearchResultBlock)]),
    tool_use_id: z.string(),
    type: z.literal("web_search_tool_result"),
})));
export type WebSearchToolResultBlock = z.infer<typeof ZWebSearchToolResultBlock>;

export const ZChoiceLogprobs = z.lazy(() => (z.object({
    content: z.array(ZChatCompletionTokenLogprob).nullable().optional(),
    refusal: z.array(ZChatCompletionTokenLogprob).nullable().optional(),
})));
export type ChoiceLogprobs = z.infer<typeof ZChoiceLogprobs>;

export const ZChoiceDeltaFunctionCall = z.lazy(() => (z.object({
    arguments: z.string().nullable().optional(),
    name: z.string().nullable().optional(),
})));
export type ChoiceDeltaFunctionCall = z.infer<typeof ZChoiceDeltaFunctionCall>;

export const ZChoiceDeltaToolCall = z.lazy(() => (z.object({
    index: z.number(),
    id: z.string().nullable().optional(),
    function: ZChoiceDeltaToolCallFunction.nullable().optional(),
    type: z.literal("function").nullable().optional(),
})));
export type ChoiceDeltaToolCall = z.infer<typeof ZChoiceDeltaToolCall>;

export const ZChatCompletionDeveloperMessageParam = z.lazy(() => (z.object({
    content: z.union([z.string(), z.array(ZChatCompletionContentPartTextParam)]),
    role: z.literal("developer"),
    name: z.string(),
})));
export type ChatCompletionDeveloperMessageParam = z.infer<typeof ZChatCompletionDeveloperMessageParam>;

export const ZChatCompletionSystemMessageParam = z.lazy(() => (z.object({
    content: z.union([z.string(), z.array(ZChatCompletionContentPartTextParam)]),
    role: z.literal("system"),
    name: z.string(),
})));
export type ChatCompletionSystemMessageParam = z.infer<typeof ZChatCompletionSystemMessageParam>;

export const ZChatCompletionUserMessageParam = z.lazy(() => (z.object({
    content: z.union([z.string(), z.array(z.union([ZChatCompletionContentPartTextParam, ZChatCompletionContentPartImageParam, ZChatCompletionContentPartInputAudioParam, ZFile]))]),
    role: z.literal("user"),
    name: z.string(),
})));
export type ChatCompletionUserMessageParam = z.infer<typeof ZChatCompletionUserMessageParam>;

export const ZChatCompletionAssistantMessageParam = z.lazy(() => (z.object({
    role: z.literal("assistant"),
    audio: ZAudio.nullable().optional(),
    content: z.union([z.string(), z.array(z.union([ZChatCompletionContentPartTextParam, ZChatCompletionContentPartRefusalParam]))]).nullable().optional(),
    function_call: ZFunctionCall.nullable().optional(),
    name: z.string(),
    refusal: z.string().nullable().optional(),
    tool_calls: z.array(ZChatCompletionMessageToolCallParam),
})));
export type ChatCompletionAssistantMessageParam = z.infer<typeof ZChatCompletionAssistantMessageParam>;

export const ZChatCompletionToolMessageParam = z.lazy(() => (z.object({
    content: z.union([z.string(), z.array(ZChatCompletionContentPartTextParam)]),
    role: z.literal("tool"),
    tool_call_id: z.string(),
})));
export type ChatCompletionToolMessageParam = z.infer<typeof ZChatCompletionToolMessageParam>;

export const ZChatCompletionFunctionMessageParam = z.lazy(() => (z.object({
    content: z.string().nullable().optional(),
    name: z.string(),
    role: z.literal("function"),
})));
export type ChatCompletionFunctionMessageParam = z.infer<typeof ZChatCompletionFunctionMessageParam>;

export const ZUsage = z.lazy(() => (z.object({
    cache_creation_input_tokens: z.number().nullable().optional(),
    cache_read_input_tokens: z.number().nullable().optional(),
    input_tokens: z.number(),
    output_tokens: z.number(),
    server_tool_use: ZServerToolUsage.nullable().optional(),
    service_tier: z.union([z.literal("standard"), z.literal("priority"), z.literal("batch")]).nullable().optional(),
})));
export type Usage = z.infer<typeof ZUsage>;

export const ZParsedChatCompletionMessage = z.lazy(() => (ZChatCompletionMessage.schema).merge(z.object({
    tool_calls: z.array(ZParsedFunctionToolCall).nullable().optional(),
    parsed: z.any().nullable().optional(),
})));
export type ParsedChatCompletionMessage = z.infer<typeof ZParsedChatCompletionMessage>;

export const ZResponseError = z.lazy(() => (z.object({
    code: z.union([z.literal("server_error"), z.literal("rate_limit_exceeded"), z.literal("invalid_prompt"), z.literal("vector_store_timeout"), z.literal("invalid_image"), z.literal("invalid_image_format"), z.literal("invalid_base64_image"), z.literal("invalid_image_url"), z.literal("image_too_large"), z.literal("image_too_small"), z.literal("image_parse_error"), z.literal("image_content_policy_violation"), z.literal("invalid_image_mode"), z.literal("image_file_too_large"), z.literal("unsupported_image_media_type"), z.literal("empty_image_file"), z.literal("failed_to_download_image"), z.literal("image_file_not_found")]),
    message: z.string(),
})));
export type ResponseError = z.infer<typeof ZResponseError>;

export const ZIncompleteDetails = z.lazy(() => (z.object({
    reason: z.union([z.literal("max_output_tokens"), z.literal("content_filter")]).nullable().optional(),
})));
export type IncompleteDetails = z.infer<typeof ZIncompleteDetails>;

export const ZEasyInputMessage = z.lazy(() => (z.object({
    content: z.union([z.string(), z.array(z.union([ZResponseInputText, ZResponseInputImage, ZResponseInputFile]))]),
    role: z.union([z.literal("user"), z.literal("assistant"), z.literal("system"), z.literal("developer")]),
    type: z.literal("message").nullable().optional(),
})));
export type EasyInputMessage = z.infer<typeof ZEasyInputMessage>;

export const ZResponseInputItemMessage = z.lazy(() => (z.object({
    content: z.array(z.union([ZResponseInputText, ZResponseInputImage, ZResponseInputFile])),
    role: z.union([z.literal("user"), z.literal("system"), z.literal("developer")]),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).nullable().optional(),
    type: z.literal("message").nullable().optional(),
})));
export type ResponseInputItemMessage = z.infer<typeof ZResponseInputItemMessage>;

export const ZResponseOutputMessage = z.lazy(() => (z.object({
    id: z.string(),
    content: z.array(z.union([ZResponseOutputText, ZResponseOutputRefusal])),
    role: z.literal("assistant"),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
    type: z.literal("message"),
})));
export type ResponseOutputMessage = z.infer<typeof ZResponseOutputMessage>;

export const ZResponseFileSearchToolCall = z.lazy(() => (z.object({
    id: z.string(),
    queries: z.array(z.string()),
    status: z.union([z.literal("in_progress"), z.literal("searching"), z.literal("completed"), z.literal("incomplete"), z.literal("failed")]),
    type: z.literal("file_search_call"),
    results: z.array(ZResponseFileSearchToolCallResult).nullable().optional(),
})));
export type ResponseFileSearchToolCall = z.infer<typeof ZResponseFileSearchToolCall>;

export const ZResponseComputerToolCall = z.lazy(() => (z.object({
    id: z.string(),
    action: z.union([ZResponseComputerToolCallActionClick, ZResponseComputerToolCallActionDoubleClick, ZResponseComputerToolCallActionDrag, ZResponseComputerToolCallActionKeypress, ZResponseComputerToolCallActionMove, ZResponseComputerToolCallActionScreenshot, ZResponseComputerToolCallActionScroll, ZResponseComputerToolCallActionType, ZResponseComputerToolCallActionWait]),
    call_id: z.string(),
    pending_safety_checks: z.array(ZResponseComputerToolCallPendingSafetyCheck),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
    type: z.literal("computer_call"),
})));
export type ResponseComputerToolCall = z.infer<typeof ZResponseComputerToolCall>;

export const ZResponseInputItemComputerCallOutput = z.lazy(() => (z.object({
    call_id: z.string(),
    output: ZResponseComputerToolCallOutputScreenshot,
    type: z.literal("computer_call_output"),
    id: z.string().nullable().optional(),
    acknowledged_safety_checks: z.array(ZResponseInputItemComputerCallOutputAcknowledgedSafetyCheck).nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).nullable().optional(),
})));
export type ResponseInputItemComputerCallOutput = z.infer<typeof ZResponseInputItemComputerCallOutput>;

export const ZResponseFunctionWebSearch = z.lazy(() => (z.object({
    id: z.string(),
    action: z.union([ZResponseFunctionWebSearchActionSearch, ZResponseFunctionWebSearchActionOpenPage, ZResponseFunctionWebSearchActionFind]),
    status: z.union([z.literal("in_progress"), z.literal("searching"), z.literal("completed"), z.literal("failed")]),
    type: z.literal("web_search_call"),
})));
export type ResponseFunctionWebSearch = z.infer<typeof ZResponseFunctionWebSearch>;

export const ZResponseFunctionToolCall = z.lazy(() => (z.object({
    arguments: z.string(),
    call_id: z.string(),
    name: z.string(),
    type: z.literal("function_call"),
    id: z.string().nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).nullable().optional(),
})));
export type ResponseFunctionToolCall = z.infer<typeof ZResponseFunctionToolCall>;

export const ZResponseInputItemFunctionCallOutput = z.lazy(() => (z.object({
    call_id: z.string(),
    output: z.string(),
    type: z.literal("function_call_output"),
    id: z.string().nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).nullable().optional(),
})));
export type ResponseInputItemFunctionCallOutput = z.infer<typeof ZResponseInputItemFunctionCallOutput>;

export const ZResponseReasoningItem = z.lazy(() => (z.object({
    id: z.string(),
    summary: z.array(ZResponseReasoningItemSummary),
    type: z.literal("reasoning"),
    encrypted_content: z.string().nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).nullable().optional(),
})));
export type ResponseReasoningItem = z.infer<typeof ZResponseReasoningItem>;

export const ZResponseInputItemImageGenerationCall = z.lazy(() => (z.object({
    id: z.string(),
    result: z.string().nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("generating"), z.literal("failed")]),
    type: z.literal("image_generation_call"),
})));
export type ResponseInputItemImageGenerationCall = z.infer<typeof ZResponseInputItemImageGenerationCall>;

export const ZResponseCodeInterpreterToolCall = z.lazy(() => (z.object({
    id: z.string(),
    code: z.string().nullable().optional(),
    container_id: z.string(),
    outputs: z.array(z.union([ZResponseCodeInterpreterToolCallOutputLogs, ZResponseCodeInterpreterToolCallOutputImage])).nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete"), z.literal("interpreting"), z.literal("failed")]),
    type: z.literal("code_interpreter_call"),
})));
export type ResponseCodeInterpreterToolCall = z.infer<typeof ZResponseCodeInterpreterToolCall>;

export const ZResponseInputItemLocalShellCall = z.lazy(() => (z.object({
    id: z.string(),
    action: ZResponseInputItemLocalShellCallAction,
    call_id: z.string(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
    type: z.literal("local_shell_call"),
})));
export type ResponseInputItemLocalShellCall = z.infer<typeof ZResponseInputItemLocalShellCall>;

export const ZResponseInputItemLocalShellCallOutput = z.lazy(() => (z.object({
    id: z.string(),
    output: z.string(),
    type: z.literal("local_shell_call_output"),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).nullable().optional(),
})));
export type ResponseInputItemLocalShellCallOutput = z.infer<typeof ZResponseInputItemLocalShellCallOutput>;

export const ZResponseInputItemMcpListTools = z.lazy(() => (z.object({
    id: z.string(),
    server_label: z.string(),
    tools: z.array(ZResponseInputItemMcpListToolsTool),
    type: z.literal("mcp_list_tools"),
    error: z.string().nullable().optional(),
})));
export type ResponseInputItemMcpListTools = z.infer<typeof ZResponseInputItemMcpListTools>;

export const ZResponseInputItemMcpApprovalRequest = z.lazy(() => (z.object({
    id: z.string(),
    arguments: z.string(),
    name: z.string(),
    server_label: z.string(),
    type: z.literal("mcp_approval_request"),
})));
export type ResponseInputItemMcpApprovalRequest = z.infer<typeof ZResponseInputItemMcpApprovalRequest>;

export const ZResponseInputItemMcpApprovalResponse = z.lazy(() => (z.object({
    approval_request_id: z.string(),
    approve: z.boolean(),
    type: z.literal("mcp_approval_response"),
    id: z.string().nullable().optional(),
    reason: z.string().nullable().optional(),
})));
export type ResponseInputItemMcpApprovalResponse = z.infer<typeof ZResponseInputItemMcpApprovalResponse>;

export const ZResponseInputItemMcpCall = z.lazy(() => (z.object({
    id: z.string(),
    arguments: z.string(),
    name: z.string(),
    server_label: z.string(),
    type: z.literal("mcp_call"),
    error: z.string().nullable().optional(),
    output: z.string().nullable().optional(),
})));
export type ResponseInputItemMcpCall = z.infer<typeof ZResponseInputItemMcpCall>;

export const ZResponseInputItemItemReference = z.lazy(() => (z.object({
    id: z.string(),
    type: z.literal("item_reference").nullable().optional(),
})));
export type ResponseInputItemItemReference = z.infer<typeof ZResponseInputItemItemReference>;

export const ZResponseOutputItemImageGenerationCall = z.lazy(() => (z.object({
    id: z.string(),
    result: z.string().nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("generating"), z.literal("failed")]),
    type: z.literal("image_generation_call"),
})));
export type ResponseOutputItemImageGenerationCall = z.infer<typeof ZResponseOutputItemImageGenerationCall>;

export const ZResponseOutputItemLocalShellCall = z.lazy(() => (z.object({
    id: z.string(),
    action: ZResponseOutputItemLocalShellCallAction,
    call_id: z.string(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
    type: z.literal("local_shell_call"),
})));
export type ResponseOutputItemLocalShellCall = z.infer<typeof ZResponseOutputItemLocalShellCall>;

export const ZResponseOutputItemMcpCall = z.lazy(() => (z.object({
    id: z.string(),
    arguments: z.string(),
    name: z.string(),
    server_label: z.string(),
    type: z.literal("mcp_call"),
    error: z.string().nullable().optional(),
    output: z.string().nullable().optional(),
})));
export type ResponseOutputItemMcpCall = z.infer<typeof ZResponseOutputItemMcpCall>;

export const ZResponseOutputItemMcpListTools = z.lazy(() => (z.object({
    id: z.string(),
    server_label: z.string(),
    tools: z.array(ZResponseOutputItemMcpListToolsTool),
    type: z.literal("mcp_list_tools"),
    error: z.string().nullable().optional(),
})));
export type ResponseOutputItemMcpListTools = z.infer<typeof ZResponseOutputItemMcpListTools>;

export const ZResponseOutputItemMcpApprovalRequest = z.lazy(() => (z.object({
    id: z.string(),
    arguments: z.string(),
    name: z.string(),
    server_label: z.string(),
    type: z.literal("mcp_approval_request"),
})));
export type ResponseOutputItemMcpApprovalRequest = z.infer<typeof ZResponseOutputItemMcpApprovalRequest>;

export const ZToolChoiceTypes = z.lazy(() => (z.object({
    type: z.union([z.literal("file_search"), z.literal("web_search_preview"), z.literal("computer_use_preview"), z.literal("web_search_preview_2025_03_11"), z.literal("image_generation"), z.literal("code_interpreter")]),
})));
export type ToolChoiceTypes = z.infer<typeof ZToolChoiceTypes>;

export const ZToolChoiceFunction = z.lazy(() => (z.object({
    name: z.string(),
    type: z.literal("function"),
})));
export type ToolChoiceFunction = z.infer<typeof ZToolChoiceFunction>;

export const ZToolChoiceMcp = z.lazy(() => (z.object({
    server_label: z.string(),
    type: z.literal("mcp"),
    name: z.string().nullable().optional(),
})));
export type ToolChoiceMcp = z.infer<typeof ZToolChoiceMcp>;

export const ZFunctionTool = z.lazy(() => (z.object({
    name: z.string(),
    parameters: z.record(z.string(), z.object({}).passthrough()).nullable().optional(),
    strict: z.boolean().nullable().optional(),
    type: z.literal("function"),
    description: z.string().nullable().optional(),
})));
export type FunctionTool = z.infer<typeof ZFunctionTool>;

export const ZFileSearchTool = z.lazy(() => (z.object({
    type: z.literal("file_search"),
    vector_store_ids: z.array(z.string()),
    filters: z.union([ZComparisonFilter, ZCompoundFilter]).nullable().optional(),
    max_num_results: z.number().nullable().optional(),
    ranking_options: ZRankingOptions.nullable().optional(),
})));
export type FileSearchTool = z.infer<typeof ZFileSearchTool>;

export const ZWebSearchTool = z.lazy(() => (z.object({
    type: z.union([z.literal("web_search_preview"), z.literal("web_search_preview_2025_03_11")]),
    search_context_size: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional(),
    user_location: ZUserLocation.nullable().optional(),
})));
export type WebSearchTool = z.infer<typeof ZWebSearchTool>;

export const ZComputerTool = z.lazy(() => (z.object({
    display_height: z.number(),
    display_width: z.number(),
    environment: z.union([z.literal("windows"), z.literal("mac"), z.literal("linux"), z.literal("ubuntu"), z.literal("browser")]),
    type: z.literal("computer_use_preview"),
})));
export type ComputerTool = z.infer<typeof ZComputerTool>;

export const ZMcp = z.lazy(() => (z.object({
    server_label: z.string(),
    server_url: z.string(),
    type: z.literal("mcp"),
    allowed_tools: z.union([z.array(z.string()), ZMcpAllowedToolsMcpAllowedToolsFilter]).nullable().optional(),
    headers: z.record(z.string(), z.string()).nullable().optional(),
    require_approval: z.union([ZMcpRequireApprovalMcpToolApprovalFilter, z.union([z.literal("always"), z.literal("never")])]).nullable().optional(),
    server_description: z.string().nullable().optional(),
})));
export type Mcp = z.infer<typeof ZMcp>;

export const ZCodeInterpreter = z.lazy(() => (z.object({
    container: z.union([z.string(), ZCodeInterpreterContainerCodeInterpreterToolAuto]),
    type: z.literal("code_interpreter"),
})));
export type CodeInterpreter = z.infer<typeof ZCodeInterpreter>;

export const ZImageGeneration = z.lazy(() => (z.object({
    type: z.literal("image_generation"),
    background: z.union([z.literal("transparent"), z.literal("opaque"), z.literal("auto")]).nullable().optional(),
    input_fidelity: z.union([z.literal("high"), z.literal("low")]).nullable().optional(),
    input_image_mask: ZImageGenerationInputImageMask.nullable().optional(),
    model: z.literal("gpt-image-1").nullable().optional(),
    moderation: z.union([z.literal("auto"), z.literal("low")]).nullable().optional(),
    output_compression: z.number().nullable().optional(),
    output_format: z.union([z.literal("png"), z.literal("webp"), z.literal("jpeg")]).nullable().optional(),
    partial_images: z.number().nullable().optional(),
    quality: z.union([z.literal("low"), z.literal("medium"), z.literal("high"), z.literal("auto")]).nullable().optional(),
    size: z.union([z.literal("1024x1024"), z.literal("1024x1536"), z.literal("1536x1024"), z.literal("auto")]).nullable().optional(),
})));
export type ImageGeneration = z.infer<typeof ZImageGeneration>;

export const ZLocalShell = z.lazy(() => (z.object({
    type: z.literal("local_shell"),
})));
export type LocalShell = z.infer<typeof ZLocalShell>;

export const ZResponsePrompt = z.lazy(() => (z.object({
    id: z.string(),
    variables: z.record(z.string(), z.union([z.string(), ZResponseInputText, ZResponseInputImage, ZResponseInputFile])).nullable().optional(),
    version: z.string().nullable().optional(),
})));
export type ResponsePrompt = z.infer<typeof ZResponsePrompt>;

export const ZReasoningReasoning = z.lazy(() => (z.object({
    effort: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional(),
    generate_summary: z.union([z.literal("auto"), z.literal("concise"), z.literal("detailed")]).nullable().optional(),
    summary: z.union([z.literal("auto"), z.literal("concise"), z.literal("detailed")]).nullable().optional(),
})));
export type ReasoningReasoning = z.infer<typeof ZReasoningReasoning>;

export const ZResponseTextConfig = z.lazy(() => (z.object({
    format: z.union([ZResponseFormatTextResponseFormatText, ZResponseFormatTextJSONSchemaConfig, ZResponseFormatJsonObjectResponseFormatJSONObject]).nullable().optional(),
})));
export type ResponseTextConfig = z.infer<typeof ZResponseTextConfig>;

export const ZResponseUsage = z.lazy(() => (z.object({
    input_tokens: z.number(),
    input_tokens_details: ZInputTokensDetails,
    output_tokens: z.number(),
    output_tokens_details: ZOutputTokensDetails,
    total_tokens: z.number(),
})));
export type ResponseUsage = z.infer<typeof ZResponseUsage>;

export const ZChatCompletionChoiceLogprobs = z.lazy(() => (z.object({
    content: z.array(ZChatCompletionTokenLogprob).nullable().optional(),
    refusal: z.array(ZChatCompletionTokenLogprob).nullable().optional(),
})));
export type ChatCompletionChoiceLogprobs = z.infer<typeof ZChatCompletionChoiceLogprobs>;

export const ZChatCompletionMessage = z.lazy(() => (z.object({
    content: z.string().nullable().optional(),
    refusal: z.string().nullable().optional(),
    role: z.literal("assistant"),
    annotations: z.array(ZChatCompletionMessageAnnotation).nullable().optional(),
    audio: ZChatCompletionAudio.nullable().optional(),
    function_call: ZChatCompletionMessageFunctionCall.nullable().optional(),
    tool_calls: z.array(ZChatCompletionMessageToolCall).nullable().optional(),
})));
export type ChatCompletionMessage = z.infer<typeof ZChatCompletionMessage>;

export const ZCompletionTokensDetails = z.lazy(() => (z.object({
    accepted_prediction_tokens: z.number().nullable().optional(),
    audio_tokens: z.number().nullable().optional(),
    reasoning_tokens: z.number().nullable().optional(),
    rejected_prediction_tokens: z.number().nullable().optional(),
})));
export type CompletionTokensDetails = z.infer<typeof ZCompletionTokensDetails>;

export const ZPromptTokensDetails = z.lazy(() => (z.object({
    audio_tokens: z.number().nullable().optional(),
    cached_tokens: z.number().nullable().optional(),
})));
export type PromptTokensDetails = z.infer<typeof ZPromptTokensDetails>;

export const ZImageURL = z.lazy(() => (z.object({
    url: z.string(),
    detail: z.union([z.literal("auto"), z.literal("low"), z.literal("high")]),
})));
export type ImageURL = z.infer<typeof ZImageURL>;

export const ZInputAudio = z.lazy(() => (z.object({
    data: z.string(),
    format: z.union([z.literal("wav"), z.literal("mp3")]),
})));
export type InputAudio = z.infer<typeof ZInputAudio>;

export const ZFileFile = z.lazy(() => (z.object({
    file_data: z.string(),
    file_id: z.string(),
    filename: z.string(),
})));
export type FileFile = z.infer<typeof ZFileFile>;

export const ZResponseInputTextParam = z.lazy(() => (z.object({
    text: z.string(),
    type: z.literal("input_text"),
})));
export type ResponseInputTextParam = z.infer<typeof ZResponseInputTextParam>;

export const ZResponseInputImageParam = z.lazy(() => (z.object({
    detail: z.union([z.literal("low"), z.literal("high"), z.literal("auto")]),
    type: z.literal("input_image"),
    file_id: z.string().nullable().optional(),
    image_url: z.string().nullable().optional(),
})));
export type ResponseInputImageParam = z.infer<typeof ZResponseInputImageParam>;

export const ZResponseInputFileParam = z.lazy(() => (z.object({
    type: z.literal("input_file"),
    file_data: z.string(),
    file_id: z.string().nullable().optional(),
    file_url: z.string(),
    filename: z.string(),
})));
export type ResponseInputFileParam = z.infer<typeof ZResponseInputFileParam>;

export const ZResponseOutputTextParam = z.lazy(() => (z.object({
    annotations: z.array(z.union([ZAnnotationFileCitation, ZAnnotationURLCitation, ZAnnotationContainerFileCitation, ZAnnotationFilePath])),
    text: z.string(),
    type: z.literal("output_text"),
    logprobs: z.array(ZLogprob),
})));
export type ResponseOutputTextParam = z.infer<typeof ZResponseOutputTextParam>;

export const ZResponseOutputRefusalParam = z.lazy(() => (z.object({
    refusal: z.string(),
    type: z.literal("refusal"),
})));
export type ResponseOutputRefusalParam = z.infer<typeof ZResponseOutputRefusalParam>;

export const ZResult = z.lazy(() => (z.object({
    attributes: z.record(z.string(), z.union([z.string(), z.number(), z.boolean()])).nullable().optional(),
    file_id: z.string(),
    filename: z.string(),
    score: z.number(),
    text: z.string(),
})));
export type Result = z.infer<typeof ZResult>;

export const ZActionClick = z.lazy(() => (z.object({
    button: z.union([z.literal("left"), z.literal("right"), z.literal("wheel"), z.literal("back"), z.literal("forward")]),
    type: z.literal("click"),
    x: z.number(),
    y: z.number(),
})));
export type ActionClick = z.infer<typeof ZActionClick>;

export const ZActionDoubleClick = z.lazy(() => (z.object({
    type: z.literal("double_click"),
    x: z.number(),
    y: z.number(),
})));
export type ActionDoubleClick = z.infer<typeof ZActionDoubleClick>;

export const ZActionDrag = z.lazy(() => (z.object({
    path: z.array(ZActionDragPath),
    type: z.literal("drag"),
})));
export type ActionDrag = z.infer<typeof ZActionDrag>;

export const ZActionKeypress = z.lazy(() => (z.object({
    keys: z.array(z.string()),
    type: z.literal("keypress"),
})));
export type ActionKeypress = z.infer<typeof ZActionKeypress>;

export const ZActionMove = z.lazy(() => (z.object({
    type: z.literal("move"),
    x: z.number(),
    y: z.number(),
})));
export type ActionMove = z.infer<typeof ZActionMove>;

export const ZActionScreenshot = z.lazy(() => (z.object({
    type: z.literal("screenshot"),
})));
export type ActionScreenshot = z.infer<typeof ZActionScreenshot>;

export const ZActionScroll = z.lazy(() => (z.object({
    scroll_x: z.number(),
    scroll_y: z.number(),
    type: z.literal("scroll"),
    x: z.number(),
    y: z.number(),
})));
export type ActionScroll = z.infer<typeof ZActionScroll>;

export const ZActionType = z.lazy(() => (z.object({
    text: z.string(),
    type: z.literal("type"),
})));
export type ActionType = z.infer<typeof ZActionType>;

export const ZActionWait = z.lazy(() => (z.object({
    type: z.literal("wait"),
})));
export type ActionWait = z.infer<typeof ZActionWait>;

export const ZPendingSafetyCheck = z.lazy(() => (z.object({
    id: z.string(),
    code: z.string(),
    message: z.string(),
})));
export type PendingSafetyCheck = z.infer<typeof ZPendingSafetyCheck>;

export const ZResponseComputerToolCallOutputScreenshotParam = z.lazy(() => (z.object({
    type: z.literal("computer_screenshot"),
    file_id: z.string(),
    image_url: z.string(),
})));
export type ResponseComputerToolCallOutputScreenshotParam = z.infer<typeof ZResponseComputerToolCallOutputScreenshotParam>;

export const ZComputerCallOutputAcknowledgedSafetyCheck = z.lazy(() => (z.object({
    id: z.string(),
    code: z.string().nullable().optional(),
    message: z.string().nullable().optional(),
})));
export type ComputerCallOutputAcknowledgedSafetyCheck = z.infer<typeof ZComputerCallOutputAcknowledgedSafetyCheck>;

export const ZActionSearch = z.lazy(() => (z.object({
    query: z.string(),
    type: z.literal("search"),
})));
export type ActionSearch = z.infer<typeof ZActionSearch>;

export const ZActionOpenPage = z.lazy(() => (z.object({
    type: z.literal("open_page"),
    url: z.string(),
})));
export type ActionOpenPage = z.infer<typeof ZActionOpenPage>;

export const ZActionFind = z.lazy(() => (z.object({
    pattern: z.string(),
    type: z.literal("find"),
    url: z.string(),
})));
export type ActionFind = z.infer<typeof ZActionFind>;

export const ZSummary = z.lazy(() => (z.object({
    text: z.string(),
    type: z.literal("summary_text"),
})));
export type Summary = z.infer<typeof ZSummary>;

export const ZOutputLogs = z.lazy(() => (z.object({
    logs: z.string(),
    type: z.literal("logs"),
})));
export type OutputLogs = z.infer<typeof ZOutputLogs>;

export const ZOutputImage = z.lazy(() => (z.object({
    type: z.literal("image"),
    url: z.string(),
})));
export type OutputImage = z.infer<typeof ZOutputImage>;

export const ZLocalShellCallAction = z.lazy(() => (z.object({
    command: z.array(z.string()),
    env: z.record(z.string(), z.string()),
    type: z.literal("exec"),
    timeout_ms: z.number().nullable().optional(),
    user: z.string().nullable().optional(),
    working_directory: z.string().nullable().optional(),
})));
export type LocalShellCallAction = z.infer<typeof ZLocalShellCallAction>;

export const ZMcpListToolsTool = z.lazy(() => (z.object({
    input_schema: z.object({}).passthrough(),
    name: z.string(),
    annotations: z.object({}).passthrough().nullable().optional(),
    description: z.string().nullable().optional(),
})));
export type McpListToolsTool = z.infer<typeof ZMcpListToolsTool>;

export const ZCacheControlEphemeralParam = z.lazy(() => (z.object({
    type: z.literal("ephemeral"),
})));
export type CacheControlEphemeralParam = z.infer<typeof ZCacheControlEphemeralParam>;

export const ZCitationCharLocationParam = z.lazy(() => (z.object({
    cited_text: z.string(),
    document_index: z.number(),
    document_title: z.string().nullable().optional(),
    end_char_index: z.number(),
    start_char_index: z.number(),
    type: z.literal("char_location"),
})));
export type CitationCharLocationParam = z.infer<typeof ZCitationCharLocationParam>;

export const ZCitationPageLocationParam = z.lazy(() => (z.object({
    cited_text: z.string(),
    document_index: z.number(),
    document_title: z.string().nullable().optional(),
    end_page_number: z.number(),
    start_page_number: z.number(),
    type: z.literal("page_location"),
})));
export type CitationPageLocationParam = z.infer<typeof ZCitationPageLocationParam>;

export const ZCitationContentBlockLocationParam = z.lazy(() => (z.object({
    cited_text: z.string(),
    document_index: z.number(),
    document_title: z.string().nullable().optional(),
    end_block_index: z.number(),
    start_block_index: z.number(),
    type: z.literal("content_block_location"),
})));
export type CitationContentBlockLocationParam = z.infer<typeof ZCitationContentBlockLocationParam>;

export const ZCitationWebSearchResultLocationParam = z.lazy(() => (z.object({
    cited_text: z.string(),
    encrypted_index: z.string(),
    title: z.string().nullable().optional(),
    type: z.literal("web_search_result_location"),
    url: z.string(),
})));
export type CitationWebSearchResultLocationParam = z.infer<typeof ZCitationWebSearchResultLocationParam>;

export const ZBase64ImageSourceParam = z.lazy(() => (z.object({
    data: z.union([z.string(), z.instanceof(Uint8Array), z.string()]),
    media_type: z.union([z.literal("image/jpeg"), z.literal("image/png"), z.literal("image/gif"), z.literal("image/webp")]),
    type: z.literal("base64"),
})));
export type Base64ImageSourceParam = z.infer<typeof ZBase64ImageSourceParam>;

export const ZURLImageSourceParam = z.lazy(() => (z.object({
    type: z.literal("url"),
    url: z.string(),
})));
export type URLImageSourceParam = z.infer<typeof ZURLImageSourceParam>;

export const ZBase64PDFSourceParam = z.lazy(() => (z.object({
    data: z.union([z.string(), z.instanceof(Uint8Array), z.string()]),
    media_type: z.literal("application/pdf"),
    type: z.literal("base64"),
})));
export type Base64PDFSourceParam = z.infer<typeof ZBase64PDFSourceParam>;

export const ZPlainTextSourceParam = z.lazy(() => (z.object({
    data: z.string(),
    media_type: z.literal("text/plain"),
    type: z.literal("text"),
})));
export type PlainTextSourceParam = z.infer<typeof ZPlainTextSourceParam>;

export const ZContentBlockSourceParam = z.lazy(() => (z.object({
    content: z.union([z.string(), z.array(z.union([ZTextBlockParam, ZImageBlockParam]))]),
    type: z.literal("content"),
})));
export type ContentBlockSourceParam = z.infer<typeof ZContentBlockSourceParam>;

export const ZURLPDFSourceParam = z.lazy(() => (z.object({
    type: z.literal("url"),
    url: z.string(),
})));
export type URLPDFSourceParam = z.infer<typeof ZURLPDFSourceParam>;

export const ZCitationsConfigParam = z.lazy(() => (z.object({
    enabled: z.boolean(),
})));
export type CitationsConfigParam = z.infer<typeof ZCitationsConfigParam>;

export const ZWebSearchResultBlockParam = z.lazy(() => (z.object({
    encrypted_content: z.string(),
    title: z.string(),
    type: z.literal("web_search_result"),
    url: z.string(),
    page_age: z.string().nullable().optional(),
})));
export type WebSearchResultBlockParam = z.infer<typeof ZWebSearchResultBlockParam>;

export const ZWebSearchToolRequestErrorParam = z.lazy(() => (z.object({
    error_code: z.union([z.literal("invalid_tool_input"), z.literal("unavailable"), z.literal("max_uses_exceeded"), z.literal("too_many_requests"), z.literal("query_too_long")]),
    type: z.literal("web_search_tool_result_error"),
})));
export type WebSearchToolRequestErrorParam = z.infer<typeof ZWebSearchToolRequestErrorParam>;

export const ZCitationCharLocation = z.lazy(() => (z.object({
    cited_text: z.string(),
    document_index: z.number(),
    document_title: z.string().nullable().optional(),
    end_char_index: z.number(),
    start_char_index: z.number(),
    type: z.literal("char_location"),
})));
export type CitationCharLocation = z.infer<typeof ZCitationCharLocation>;

export const ZCitationPageLocation = z.lazy(() => (z.object({
    cited_text: z.string(),
    document_index: z.number(),
    document_title: z.string().nullable().optional(),
    end_page_number: z.number(),
    start_page_number: z.number(),
    type: z.literal("page_location"),
})));
export type CitationPageLocation = z.infer<typeof ZCitationPageLocation>;

export const ZCitationContentBlockLocation = z.lazy(() => (z.object({
    cited_text: z.string(),
    document_index: z.number(),
    document_title: z.string().nullable().optional(),
    end_block_index: z.number(),
    start_block_index: z.number(),
    type: z.literal("content_block_location"),
})));
export type CitationContentBlockLocation = z.infer<typeof ZCitationContentBlockLocation>;

export const ZCitationsWebSearchResultLocation = z.lazy(() => (z.object({
    cited_text: z.string(),
    encrypted_index: z.string(),
    title: z.string().nullable().optional(),
    type: z.literal("web_search_result_location"),
    url: z.string(),
})));
export type CitationsWebSearchResultLocation = z.infer<typeof ZCitationsWebSearchResultLocation>;

export const ZWebSearchToolResultError = z.lazy(() => (z.object({
    error_code: z.union([z.literal("invalid_tool_input"), z.literal("unavailable"), z.literal("max_uses_exceeded"), z.literal("too_many_requests"), z.literal("query_too_long")]),
    type: z.literal("web_search_tool_result_error"),
})));
export type WebSearchToolResultError = z.infer<typeof ZWebSearchToolResultError>;

export const ZWebSearchResultBlock = z.lazy(() => (z.object({
    encrypted_content: z.string(),
    page_age: z.string().nullable().optional(),
    title: z.string(),
    type: z.literal("web_search_result"),
    url: z.string(),
})));
export type WebSearchResultBlock = z.infer<typeof ZWebSearchResultBlock>;

export const ZChatCompletionTokenLogprob = z.lazy(() => (z.object({
    token: z.string(),
    bytes: z.array(z.number()).nullable().optional(),
    logprob: z.number(),
    top_logprobs: z.array(ZTopLogprob),
})));
export type ChatCompletionTokenLogprob = z.infer<typeof ZChatCompletionTokenLogprob>;

export const ZChoiceDeltaToolCallFunction = z.lazy(() => (z.object({
    arguments: z.string().nullable().optional(),
    name: z.string().nullable().optional(),
})));
export type ChoiceDeltaToolCallFunction = z.infer<typeof ZChoiceDeltaToolCallFunction>;

export const ZAudio = z.lazy(() => (z.object({
    id: z.string(),
})));
export type Audio = z.infer<typeof ZAudio>;

export const ZChatCompletionContentPartRefusalParam = z.lazy(() => (z.object({
    refusal: z.string(),
    type: z.literal("refusal"),
})));
export type ChatCompletionContentPartRefusalParam = z.infer<typeof ZChatCompletionContentPartRefusalParam>;

export const ZFunctionCall = z.lazy(() => (z.object({
    arguments: z.string(),
    name: z.string(),
})));
export type FunctionCall = z.infer<typeof ZFunctionCall>;

export const ZChatCompletionMessageToolCallParam = z.lazy(() => (z.object({
    id: z.string(),
    function: ZFunction,
    type: z.literal("function"),
})));
export type ChatCompletionMessageToolCallParam = z.infer<typeof ZChatCompletionMessageToolCallParam>;

export const ZServerToolUsage = z.lazy(() => (z.object({
    web_search_requests: z.number(),
})));
export type ServerToolUsage = z.infer<typeof ZServerToolUsage>;

export const ZParsedFunctionToolCall = z.lazy(() => (ZChatCompletionMessageToolCall.schema).merge(z.object({
    function: ZParsedFunction,
})));
export type ParsedFunctionToolCall = z.infer<typeof ZParsedFunctionToolCall>;

export const ZResponseInputText = z.lazy(() => (z.object({
    text: z.string(),
    type: z.literal("input_text"),
})));
export type ResponseInputText = z.infer<typeof ZResponseInputText>;

export const ZResponseInputImage = z.lazy(() => (z.object({
    detail: z.union([z.literal("low"), z.literal("high"), z.literal("auto")]),
    type: z.literal("input_image"),
    file_id: z.string().nullable().optional(),
    image_url: z.string().nullable().optional(),
})));
export type ResponseInputImage = z.infer<typeof ZResponseInputImage>;

export const ZResponseInputFile = z.lazy(() => (z.object({
    type: z.literal("input_file"),
    file_data: z.string().nullable().optional(),
    file_id: z.string().nullable().optional(),
    file_url: z.string().nullable().optional(),
    filename: z.string().nullable().optional(),
})));
export type ResponseInputFile = z.infer<typeof ZResponseInputFile>;

export const ZResponseOutputText = z.lazy(() => (z.object({
    annotations: z.array(z.union([ZResponseOutputTextAnnotationFileCitation, ZResponseOutputTextAnnotationURLCitation, ZResponseOutputTextAnnotationContainerFileCitation, ZResponseOutputTextAnnotationFilePath])),
    text: z.string(),
    type: z.literal("output_text"),
    logprobs: z.array(ZResponseOutputTextLogprob).nullable().optional(),
})));
export type ResponseOutputText = z.infer<typeof ZResponseOutputText>;

export const ZResponseOutputRefusal = z.lazy(() => (z.object({
    refusal: z.string(),
    type: z.literal("refusal"),
})));
export type ResponseOutputRefusal = z.infer<typeof ZResponseOutputRefusal>;

export const ZResponseFileSearchToolCallResult = z.lazy(() => (z.object({
    attributes: z.record(z.string(), z.union([z.string(), z.number(), z.boolean()])).nullable().optional(),
    file_id: z.string().nullable().optional(),
    filename: z.string().nullable().optional(),
    score: z.number().nullable().optional(),
    text: z.string().nullable().optional(),
})));
export type ResponseFileSearchToolCallResult = z.infer<typeof ZResponseFileSearchToolCallResult>;

export const ZResponseComputerToolCallActionClick = z.lazy(() => (z.object({
    button: z.union([z.literal("left"), z.literal("right"), z.literal("wheel"), z.literal("back"), z.literal("forward")]),
    type: z.literal("click"),
    x: z.number(),
    y: z.number(),
})));
export type ResponseComputerToolCallActionClick = z.infer<typeof ZResponseComputerToolCallActionClick>;

export const ZResponseComputerToolCallActionDoubleClick = z.lazy(() => (z.object({
    type: z.literal("double_click"),
    x: z.number(),
    y: z.number(),
})));
export type ResponseComputerToolCallActionDoubleClick = z.infer<typeof ZResponseComputerToolCallActionDoubleClick>;

export const ZResponseComputerToolCallActionDrag = z.lazy(() => (z.object({
    path: z.array(ZResponseComputerToolCallActionDragPath),
    type: z.literal("drag"),
})));
export type ResponseComputerToolCallActionDrag = z.infer<typeof ZResponseComputerToolCallActionDrag>;

export const ZResponseComputerToolCallActionKeypress = z.lazy(() => (z.object({
    keys: z.array(z.string()),
    type: z.literal("keypress"),
})));
export type ResponseComputerToolCallActionKeypress = z.infer<typeof ZResponseComputerToolCallActionKeypress>;

export const ZResponseComputerToolCallActionMove = z.lazy(() => (z.object({
    type: z.literal("move"),
    x: z.number(),
    y: z.number(),
})));
export type ResponseComputerToolCallActionMove = z.infer<typeof ZResponseComputerToolCallActionMove>;

export const ZResponseComputerToolCallActionScreenshot = z.lazy(() => (z.object({
    type: z.literal("screenshot"),
})));
export type ResponseComputerToolCallActionScreenshot = z.infer<typeof ZResponseComputerToolCallActionScreenshot>;

export const ZResponseComputerToolCallActionScroll = z.lazy(() => (z.object({
    scroll_x: z.number(),
    scroll_y: z.number(),
    type: z.literal("scroll"),
    x: z.number(),
    y: z.number(),
})));
export type ResponseComputerToolCallActionScroll = z.infer<typeof ZResponseComputerToolCallActionScroll>;

export const ZResponseComputerToolCallActionType = z.lazy(() => (z.object({
    text: z.string(),
    type: z.literal("type"),
})));
export type ResponseComputerToolCallActionType = z.infer<typeof ZResponseComputerToolCallActionType>;

export const ZResponseComputerToolCallActionWait = z.lazy(() => (z.object({
    type: z.literal("wait"),
})));
export type ResponseComputerToolCallActionWait = z.infer<typeof ZResponseComputerToolCallActionWait>;

export const ZResponseComputerToolCallPendingSafetyCheck = z.lazy(() => (z.object({
    id: z.string(),
    code: z.string(),
    message: z.string(),
})));
export type ResponseComputerToolCallPendingSafetyCheck = z.infer<typeof ZResponseComputerToolCallPendingSafetyCheck>;

export const ZResponseComputerToolCallOutputScreenshot = z.lazy(() => (z.object({
    type: z.literal("computer_screenshot"),
    file_id: z.string().nullable().optional(),
    image_url: z.string().nullable().optional(),
})));
export type ResponseComputerToolCallOutputScreenshot = z.infer<typeof ZResponseComputerToolCallOutputScreenshot>;

export const ZResponseInputItemComputerCallOutputAcknowledgedSafetyCheck = z.lazy(() => (z.object({
    id: z.string(),
    code: z.string().nullable().optional(),
    message: z.string().nullable().optional(),
})));
export type ResponseInputItemComputerCallOutputAcknowledgedSafetyCheck = z.infer<typeof ZResponseInputItemComputerCallOutputAcknowledgedSafetyCheck>;

export const ZResponseFunctionWebSearchActionSearch = z.lazy(() => (z.object({
    query: z.string(),
    type: z.literal("search"),
})));
export type ResponseFunctionWebSearchActionSearch = z.infer<typeof ZResponseFunctionWebSearchActionSearch>;

export const ZResponseFunctionWebSearchActionOpenPage = z.lazy(() => (z.object({
    type: z.literal("open_page"),
    url: z.string(),
})));
export type ResponseFunctionWebSearchActionOpenPage = z.infer<typeof ZResponseFunctionWebSearchActionOpenPage>;

export const ZResponseFunctionWebSearchActionFind = z.lazy(() => (z.object({
    pattern: z.string(),
    type: z.literal("find"),
    url: z.string(),
})));
export type ResponseFunctionWebSearchActionFind = z.infer<typeof ZResponseFunctionWebSearchActionFind>;

export const ZResponseReasoningItemSummary = z.lazy(() => (z.object({
    text: z.string(),
    type: z.literal("summary_text"),
})));
export type ResponseReasoningItemSummary = z.infer<typeof ZResponseReasoningItemSummary>;

export const ZResponseCodeInterpreterToolCallOutputLogs = z.lazy(() => (z.object({
    logs: z.string(),
    type: z.literal("logs"),
})));
export type ResponseCodeInterpreterToolCallOutputLogs = z.infer<typeof ZResponseCodeInterpreterToolCallOutputLogs>;

export const ZResponseCodeInterpreterToolCallOutputImage = z.lazy(() => (z.object({
    type: z.literal("image"),
    url: z.string(),
})));
export type ResponseCodeInterpreterToolCallOutputImage = z.infer<typeof ZResponseCodeInterpreterToolCallOutputImage>;

export const ZResponseInputItemLocalShellCallAction = z.lazy(() => (z.object({
    command: z.array(z.string()),
    env: z.record(z.string(), z.string()),
    type: z.literal("exec"),
    timeout_ms: z.number().nullable().optional(),
    user: z.string().nullable().optional(),
    working_directory: z.string().nullable().optional(),
})));
export type ResponseInputItemLocalShellCallAction = z.infer<typeof ZResponseInputItemLocalShellCallAction>;

export const ZResponseInputItemMcpListToolsTool = z.lazy(() => (z.object({
    input_schema: z.object({}).passthrough(),
    name: z.string(),
    annotations: z.object({}).passthrough().nullable().optional(),
    description: z.string().nullable().optional(),
})));
export type ResponseInputItemMcpListToolsTool = z.infer<typeof ZResponseInputItemMcpListToolsTool>;

export const ZResponseOutputItemLocalShellCallAction = z.lazy(() => (z.object({
    command: z.array(z.string()),
    env: z.record(z.string(), z.string()),
    type: z.literal("exec"),
    timeout_ms: z.number().nullable().optional(),
    user: z.string().nullable().optional(),
    working_directory: z.string().nullable().optional(),
})));
export type ResponseOutputItemLocalShellCallAction = z.infer<typeof ZResponseOutputItemLocalShellCallAction>;

export const ZResponseOutputItemMcpListToolsTool = z.lazy(() => (z.object({
    input_schema: z.object({}).passthrough(),
    name: z.string(),
    annotations: z.object({}).passthrough().nullable().optional(),
    description: z.string().nullable().optional(),
})));
export type ResponseOutputItemMcpListToolsTool = z.infer<typeof ZResponseOutputItemMcpListToolsTool>;

export const ZComparisonFilter = z.lazy(() => (z.object({
    key: z.string(),
    type: z.union([z.literal("eq"), z.literal("ne"), z.literal("gt"), z.literal("gte"), z.literal("lt"), z.literal("lte")]),
    value: z.union([z.string(), z.number(), z.boolean()]),
})));
export type ComparisonFilter = z.infer<typeof ZComparisonFilter>;

export const ZCompoundFilter = z.lazy(() => (z.object({
    filters: z.array(z.union([ZComparisonFilter, z.object({}).passthrough()])),
    type: z.union([z.literal("and"), z.literal("or")]),
})));
export type CompoundFilter = z.infer<typeof ZCompoundFilter>;

export const ZRankingOptions = z.lazy(() => (z.object({
    ranker: z.union([z.literal("auto"), z.literal("default-2024-11-15")]).nullable().optional(),
    score_threshold: z.number().nullable().optional(),
})));
export type RankingOptions = z.infer<typeof ZRankingOptions>;

export const ZUserLocation = z.lazy(() => (z.object({
    type: z.literal("approximate"),
    city: z.string().nullable().optional(),
    country: z.string().nullable().optional(),
    region: z.string().nullable().optional(),
    timezone: z.string().nullable().optional(),
})));
export type UserLocation = z.infer<typeof ZUserLocation>;

export const ZMcpAllowedToolsMcpAllowedToolsFilter = z.lazy(() => (z.object({
    tool_names: z.array(z.string()).nullable().optional(),
})));
export type McpAllowedToolsMcpAllowedToolsFilter = z.infer<typeof ZMcpAllowedToolsMcpAllowedToolsFilter>;

export const ZMcpRequireApprovalMcpToolApprovalFilter = z.lazy(() => (z.object({
    always: ZMcpRequireApprovalMcpToolApprovalFilterAlways.nullable().optional(),
    never: ZMcpRequireApprovalMcpToolApprovalFilterNever.nullable().optional(),
})));
export type McpRequireApprovalMcpToolApprovalFilter = z.infer<typeof ZMcpRequireApprovalMcpToolApprovalFilter>;

export const ZCodeInterpreterContainerCodeInterpreterToolAuto = z.lazy(() => (z.object({
    type: z.literal("auto"),
    file_ids: z.array(z.string()).nullable().optional(),
})));
export type CodeInterpreterContainerCodeInterpreterToolAuto = z.infer<typeof ZCodeInterpreterContainerCodeInterpreterToolAuto>;

export const ZImageGenerationInputImageMask = z.lazy(() => (z.object({
    file_id: z.string().nullable().optional(),
    image_url: z.string().nullable().optional(),
})));
export type ImageGenerationInputImageMask = z.infer<typeof ZImageGenerationInputImageMask>;

export const ZResponseFormatTextResponseFormatText = z.lazy(() => (z.object({
    type: z.literal("text"),
})));
export type ResponseFormatTextResponseFormatText = z.infer<typeof ZResponseFormatTextResponseFormatText>;

export const ZResponseFormatTextJSONSchemaConfig = z.lazy(() => (z.object({
    name: z.string(),
    schema_: z.record(z.string(), z.object({}).passthrough()),
    type: z.literal("json_schema"),
    description: z.string().nullable().optional(),
    strict: z.boolean().nullable().optional(),
})));
export type ResponseFormatTextJSONSchemaConfig = z.infer<typeof ZResponseFormatTextJSONSchemaConfig>;

export const ZResponseFormatJsonObjectResponseFormatJSONObject = z.lazy(() => (z.object({
    type: z.literal("json_object"),
})));
export type ResponseFormatJsonObjectResponseFormatJSONObject = z.infer<typeof ZResponseFormatJsonObjectResponseFormatJSONObject>;

export const ZInputTokensDetails = z.lazy(() => (z.object({
    cached_tokens: z.number(),
})));
export type InputTokensDetails = z.infer<typeof ZInputTokensDetails>;

export const ZOutputTokensDetails = z.lazy(() => (z.object({
    reasoning_tokens: z.number(),
})));
export type OutputTokensDetails = z.infer<typeof ZOutputTokensDetails>;

export const ZChatCompletionMessageAnnotation = z.lazy(() => (z.object({
    type: z.literal("url_citation"),
    url_citation: ZChatCompletionMessageAnnotationURLCitation,
})));
export type ChatCompletionMessageAnnotation = z.infer<typeof ZChatCompletionMessageAnnotation>;

export const ZChatCompletionAudio = z.lazy(() => (z.object({
    id: z.string(),
    data: z.string(),
    expires_at: z.number(),
    transcript: z.string(),
})));
export type ChatCompletionAudio = z.infer<typeof ZChatCompletionAudio>;

export const ZChatCompletionMessageFunctionCall = z.lazy(() => (z.object({
    arguments: z.string(),
    name: z.string(),
})));
export type ChatCompletionMessageFunctionCall = z.infer<typeof ZChatCompletionMessageFunctionCall>;

export const ZChatCompletionMessageToolCall = z.lazy(() => (z.object({
    id: z.string(),
    function: ZChatCompletionMessageToolCallFunction,
    type: z.literal("function"),
})));
export type ChatCompletionMessageToolCall = z.infer<typeof ZChatCompletionMessageToolCall>;

export const ZAnnotationFileCitation = z.lazy(() => (z.object({
    file_id: z.string(),
    filename: z.string(),
    index: z.number(),
    type: z.literal("file_citation"),
})));
export type AnnotationFileCitation = z.infer<typeof ZAnnotationFileCitation>;

export const ZAnnotationURLCitation = z.lazy(() => (z.object({
    end_index: z.number(),
    start_index: z.number(),
    title: z.string(),
    type: z.literal("url_citation"),
    url: z.string(),
})));
export type AnnotationURLCitation = z.infer<typeof ZAnnotationURLCitation>;

export const ZAnnotationContainerFileCitation = z.lazy(() => (z.object({
    container_id: z.string(),
    end_index: z.number(),
    file_id: z.string(),
    filename: z.string(),
    start_index: z.number(),
    type: z.literal("container_file_citation"),
})));
export type AnnotationContainerFileCitation = z.infer<typeof ZAnnotationContainerFileCitation>;

export const ZAnnotationFilePath = z.lazy(() => (z.object({
    file_id: z.string(),
    index: z.number(),
    type: z.literal("file_path"),
})));
export type AnnotationFilePath = z.infer<typeof ZAnnotationFilePath>;

export const ZLogprob = z.lazy(() => (z.object({
    token: z.string(),
    bytes: z.array(z.number()),
    logprob: z.number(),
    top_logprobs: z.array(ZLogprobTopLogprob),
})));
export type Logprob = z.infer<typeof ZLogprob>;

export const ZActionDragPath = z.lazy(() => (z.object({
    x: z.number(),
    y: z.number(),
})));
export type ActionDragPath = z.infer<typeof ZActionDragPath>;

export const ZTopLogprob = z.lazy(() => (z.object({
    token: z.string(),
    bytes: z.array(z.number()).nullable().optional(),
    logprob: z.number(),
})));
export type TopLogprob = z.infer<typeof ZTopLogprob>;

export const ZFunction = z.lazy(() => (z.object({
    arguments: z.string(),
    name: z.string(),
})));
export type Function = z.infer<typeof ZFunction>;

export const ZParsedFunction = z.lazy(() => (ZChatCompletionMessageToolCallFunction.schema).merge(z.object({
    parsed_arguments: z.object({}).passthrough().nullable().optional(),
})));
export type ParsedFunction = z.infer<typeof ZParsedFunction>;

export const ZResponseOutputTextAnnotationFileCitation = z.lazy(() => (z.object({
    file_id: z.string(),
    filename: z.string(),
    index: z.number(),
    type: z.literal("file_citation"),
})));
export type ResponseOutputTextAnnotationFileCitation = z.infer<typeof ZResponseOutputTextAnnotationFileCitation>;

export const ZResponseOutputTextAnnotationURLCitation = z.lazy(() => (z.object({
    end_index: z.number(),
    start_index: z.number(),
    title: z.string(),
    type: z.literal("url_citation"),
    url: z.string(),
})));
export type ResponseOutputTextAnnotationURLCitation = z.infer<typeof ZResponseOutputTextAnnotationURLCitation>;

export const ZResponseOutputTextAnnotationContainerFileCitation = z.lazy(() => (z.object({
    container_id: z.string(),
    end_index: z.number(),
    file_id: z.string(),
    filename: z.string(),
    start_index: z.number(),
    type: z.literal("container_file_citation"),
})));
export type ResponseOutputTextAnnotationContainerFileCitation = z.infer<typeof ZResponseOutputTextAnnotationContainerFileCitation>;

export const ZResponseOutputTextAnnotationFilePath = z.lazy(() => (z.object({
    file_id: z.string(),
    index: z.number(),
    type: z.literal("file_path"),
})));
export type ResponseOutputTextAnnotationFilePath = z.infer<typeof ZResponseOutputTextAnnotationFilePath>;

export const ZResponseOutputTextLogprob = z.lazy(() => (z.object({
    token: z.string(),
    bytes: z.array(z.number()),
    logprob: z.number(),
    top_logprobs: z.array(ZResponseOutputTextLogprobTopLogprob),
})));
export type ResponseOutputTextLogprob = z.infer<typeof ZResponseOutputTextLogprob>;

export const ZResponseComputerToolCallActionDragPath = z.lazy(() => (z.object({
    x: z.number(),
    y: z.number(),
})));
export type ResponseComputerToolCallActionDragPath = z.infer<typeof ZResponseComputerToolCallActionDragPath>;

export const ZMcpRequireApprovalMcpToolApprovalFilterAlways = z.lazy(() => (z.object({
    tool_names: z.array(z.string()).nullable().optional(),
})));
export type McpRequireApprovalMcpToolApprovalFilterAlways = z.infer<typeof ZMcpRequireApprovalMcpToolApprovalFilterAlways>;

export const ZMcpRequireApprovalMcpToolApprovalFilterNever = z.lazy(() => (z.object({
    tool_names: z.array(z.string()).nullable().optional(),
})));
export type McpRequireApprovalMcpToolApprovalFilterNever = z.infer<typeof ZMcpRequireApprovalMcpToolApprovalFilterNever>;

export const ZChatCompletionMessageAnnotationURLCitation = z.lazy(() => (z.object({
    end_index: z.number(),
    start_index: z.number(),
    title: z.string(),
    url: z.string(),
})));
export type ChatCompletionMessageAnnotationURLCitation = z.infer<typeof ZChatCompletionMessageAnnotationURLCitation>;

export const ZChatCompletionMessageToolCallFunction = z.lazy(() => (z.object({
    arguments: z.string(),
    name: z.string(),
})));
export type ChatCompletionMessageToolCallFunction = z.infer<typeof ZChatCompletionMessageToolCallFunction>;

export const ZLogprobTopLogprob = z.lazy(() => (z.object({
    token: z.string(),
    bytes: z.array(z.number()),
    logprob: z.number(),
})));
export type LogprobTopLogprob = z.infer<typeof ZLogprobTopLogprob>;

export const ZResponseOutputTextLogprobTopLogprob = z.lazy(() => (z.object({
    token: z.string(),
    bytes: z.array(z.number()),
    logprob: z.number(),
})));
export type ResponseOutputTextLogprobTopLogprob = z.infer<typeof ZResponseOutputTextLogprobTopLogprob>;

