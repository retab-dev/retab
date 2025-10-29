import * as z from 'zod';

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

export const ZChatCompletionMessageFunctionToolCallParam = z.lazy(() => (z.object({
    id: z.string(),
    function: ZChatCompletionMessageFunctionToolCallParamFunction,
    type: z.literal("function"),
})));
export type ChatCompletionMessageFunctionToolCallParam = z.infer<typeof ZChatCompletionMessageFunctionToolCallParam>;

export const ZChatCompletionRetabMessage = z.lazy(() => (z.object({
    role: z.union([z.literal("user"), z.literal("system"), z.literal("assistant"), z.literal("developer"), z.literal("tool")]),
    content: z.union([z.string(), z.array(z.union([ZChatCompletionContentPartTextParam, ZChatCompletionContentPartImageParam, ZChatCompletionContentPartInputAudioParam, ZFile]))]).nullable().optional(),
    tool_call_id: z.string().nullable().optional(),
    tool_calls: z.array(ZChatCompletionMessageFunctionToolCallParam).nullable().optional(),
})));
export type ChatCompletionRetabMessage = z.infer<typeof ZChatCompletionRetabMessage>;

export const ZListMetadata = z.lazy(() => (z.object({
    before: z.string().nullable().optional(),
    after: z.string().nullable().optional(),
})));
export type ListMetadata = z.infer<typeof ZListMetadata>;

export const ZInferenceSettings = z.lazy(() => (z.object({
    model: z.string().default("gpt-5-mini"),
    temperature: z.number().default(0.0),
    reasoning_effort: z.union([z.literal("minimal"), z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional().default("minimal"),
    image_resolution_dpi: z.number().default(192),
    browser_canvas: z.union([z.literal("A1"), z.literal("A2"), z.literal("A3"), z.literal("A4"), z.literal("A5")]).default("A4"),
    n_consensus: z.number().default(1),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native")]).default("native"),
    parallel_ocr_keys: z.record(z.string(), z.string()).nullable().optional(),
})));
export type InferenceSettings = z.infer<typeof ZInferenceSettings>;

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
})));
export type PredictionMetadata = z.infer<typeof ZPredictionMetadata>;

export const ZCreateProjectRequest = z.lazy(() => (z.object({
    name: z.string(),
    json_schema: z.record(z.string(), z.any()),
})));
export type CreateProjectRequest = z.infer<typeof ZCreateProjectRequest>;

export const ZPatchProjectRequest = z.lazy(() => (z.object({
    name: z.string().nullable().optional(),
    published_config: ZPublishedConfig.nullable().optional(),
    draft_config: ZDraftConfig.nullable().optional(),
    is_published: z.boolean().nullable().optional(),
    functions: z.array(ZFunction).nullable().optional(),
})));
export type PatchProjectRequest = z.infer<typeof ZPatchProjectRequest>;

export const ZProject = z.lazy(() => (z.object({
    id: z.string(),
    name: z.string().default(""),
    updated_at: z.string(),
    published_config: ZPublishedConfig,
    draft_config: ZDraftConfig,
    is_published: z.boolean().default(false),
    functions: z.array(ZFunction),
})));
export type Project = z.infer<typeof ZProject>;

export const ZDraftConfig = z.lazy(() => (z.object({
    inference_settings: ZInferenceSettings.default({"model": "auto-small", "temperature": 0.5, "reasoning_effort": "minimal", "image_resolution_dpi": 192, "browser_canvas": "A4", "n_consensus": 1, "modality": "native"}),
    json_schema: z.record(z.string(), z.any()),
    human_in_the_loop_criteria: z.array(ZFunctionHilCriterion),
})));
export type DraftConfig = z.infer<typeof ZDraftConfig>;

export const ZFunction = z.lazy(() => (z.object({
    id: z.string(),
    path: z.string(),
    code: z.string().nullable().optional(),
    function_registry_id: z.string().nullable().optional(),
})));
export type Function = z.infer<typeof ZFunction>;

export const ZFunctionHilCriterion = z.lazy(() => (z.object({
    path: z.string(),
    agentic_fix: z.boolean().default(false),
})));
export type FunctionHilCriterion = z.infer<typeof ZFunctionHilCriterion>;

export const ZHumanInTheLoopParams = z.lazy(() => (z.object({
    enabled: z.boolean().default(false),
    url: z.string().default(""),
    headers: z.record(z.string(), z.string()),
    criteria: z.array(ZFunctionHilCriterion),
})));
export type HumanInTheLoopParams = z.infer<typeof ZHumanInTheLoopParams>;

export const ZPublishedConfig = z.lazy(() => (z.object({
    inference_settings: ZInferenceSettings.default({"model": "auto-small", "temperature": 0.5, "reasoning_effort": "minimal", "image_resolution_dpi": 192, "browser_canvas": "A4", "n_consensus": 1, "modality": "native"}),
    json_schema: z.record(z.string(), z.any()),
    human_in_the_loop_params: ZHumanInTheLoopParams,
    origin: z.string().default("manual"),
})));
export type PublishedConfig = z.infer<typeof ZPublishedConfig>;

export const ZStoredProject = z.lazy(() => (ZProject.schema).merge(z.object({
    organization_id: z.string(),
})));
export type StoredProject = z.infer<typeof ZStoredProject>;

export const ZGenerateSchemaRequest = z.lazy(() => (z.object({
    documents: z.array(ZMIMEData),
    model: z.string().default("gpt-4o-mini"),
    temperature: z.number().default(0.0),
    reasoning_effort: z.union([z.literal("minimal"), z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional().default("minimal"),
    instructions: z.string().nullable().optional(),
    image_resolution_dpi: z.number().default(192),
    browser_canvas: z.union([z.literal("A1"), z.literal("A2"), z.literal("A3"), z.literal("A4"), z.literal("A5")]).default("A4"),
    stream: z.boolean().default(false),
})));
export type GenerateSchemaRequest = z.infer<typeof ZGenerateSchemaRequest>;

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

export const ZSchema = z.lazy(() => (ZPartialSchema.schema).merge(z.object({
    object: z.literal("schema").default("schema"),
    created_at: z.string(),
    json_schema: z.record(z.string(), z.any()).default({}),
})));
export type Schema = z.infer<typeof ZSchema>;

export const ZMessageParam = z.lazy(() => (z.object({
    content: z.union([z.string(), z.array(z.union([ZTextBlockParam, ZImageBlockParam, ZDocumentBlockParam, ZSearchResultBlockParam, ZThinkingBlockParam, ZRedactedThinkingBlockParam, ZToolUseBlockParam, ZToolResultBlockParam, ZServerToolUseBlockParam, ZWebSearchToolResultBlockParam, z.union([ZTextBlock, ZThinkingBlock, ZRedactedThinkingBlock, ZToolUseBlock, ZServerToolUseBlock, ZWebSearchToolResultBlock])]))]),
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

export const ZBlobDict = z.lazy(() => (z.object({
    display_name: z.string().nullable().optional(),
    data: z.instanceof(Uint8Array).nullable().optional(),
    mime_type: z.string().nullable().optional(),
})));
export type BlobDict = z.infer<typeof ZBlobDict>;

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

export const ZChatCompletionContentPartTextParam = z.lazy(() => (z.object({
    text: z.string(),
    type: z.literal("text"),
})));
export type ChatCompletionContentPartTextParam = z.infer<typeof ZChatCompletionContentPartTextParam>;

export const ZCompletionTokensDetails = z.lazy(() => (z.object({
    accepted_prediction_tokens: z.number().nullable().optional(),
    audio_tokens: z.number().nullable().optional(),
    reasoning_tokens: z.number().nullable().optional(),
    rejected_prediction_tokens: z.number().nullable().optional(),
})));
export type CompletionTokensDetails = z.infer<typeof ZCompletionTokensDetails>;

export const ZCompletionUsage = z.lazy(() => (z.object({
    completion_tokens: z.number(),
    prompt_tokens: z.number(),
    total_tokens: z.number(),
    completion_tokens_details: ZCompletionTokensDetails.nullable().optional(),
    prompt_tokens_details: ZPromptTokensDetails.nullable().optional(),
})));
export type CompletionUsage = z.infer<typeof ZCompletionUsage>;

export const ZContentDict = z.lazy(() => (z.object({
    parts: z.array(ZPartDict).nullable().optional(),
    role: z.string().nullable().optional(),
})));
export type ContentDict = z.infer<typeof ZContentDict>;

export const ZEasyInputMessageParam = z.lazy(() => (z.object({
    content: z.union([z.string(), z.array(z.union([ZResponseInputTextParam, ZResponseInputImageParam, ZResponseInputFileParam, ZResponseInputAudioParam]))]),
    role: z.union([z.literal("user"), z.literal("assistant"), z.literal("system"), z.literal("developer")]),
    type: z.literal("message"),
})));
export type EasyInputMessageParam = z.infer<typeof ZEasyInputMessageParam>;

export const ZImageBlockParam = z.lazy(() => (z.object({
    source: z.union([ZBase64ImageSourceParam, ZURLImageSourceParam]),
    type: z.literal("image"),
    cache_control: ZCacheControlEphemeralParam.nullable().optional(),
})));
export type ImageBlockParam = z.infer<typeof ZImageBlockParam>;

export const ZImageURL = z.lazy(() => (z.object({
    url: z.string(),
    detail: z.union([z.literal("auto"), z.literal("low"), z.literal("high")]),
})));
export type ImageURL = z.infer<typeof ZImageURL>;

export const ZParsedChatCompletionMessage = z.lazy(() => (ZChatCompletionMessage.schema).merge(z.object({
    tool_calls: z.array(ZParsedFunctionToolCall).nullable().optional(),
    parsed: z.any().nullable().optional(),
})));
export type ParsedChatCompletionMessage = z.infer<typeof ZParsedChatCompletionMessage>;

export const ZPartDict = z.lazy(() => (z.object({
    video_metadata: ZVideoMetadataDict.nullable().optional(),
    thought: z.boolean().nullable().optional(),
    inline_data: ZBlobDict.nullable().optional(),
    file_data: ZFileDataDict.nullable().optional(),
    thought_signature: z.instanceof(Uint8Array).nullable().optional(),
    function_call: ZFunctionCallDict.nullable().optional(),
    code_execution_result: ZCodeExecutionResultDict.nullable().optional(),
    executable_code: ZExecutableCodeDict.nullable().optional(),
    function_response: ZFunctionResponseDict.nullable().optional(),
    text: z.string().nullable().optional(),
})));
export type PartDict = z.infer<typeof ZPartDict>;

export const ZPromptTokensDetails = z.lazy(() => (z.object({
    audio_tokens: z.number().nullable().optional(),
    cached_tokens: z.number().nullable().optional(),
})));
export type PromptTokensDetails = z.infer<typeof ZPromptTokensDetails>;

export const ZResponse = z.lazy(() => (z.object({
    id: z.string(),
    created_at: z.number(),
    error: ZResponseError.nullable().optional(),
    incomplete_details: ZIncompleteDetails.nullable().optional(),
    instructions: z.union([z.string(), z.array(z.union([ZEasyInputMessage, ZMessage, ZResponseOutputMessage, ZResponseFileSearchToolCall, ZResponseComputerToolCall, ZComputerCallOutput, ZResponseFunctionWebSearch, ZResponseFunctionToolCall, ZFunctionCallOutput, ZResponseReasoningItem, ZImageGenerationCall, ZResponseCodeInterpreterToolCall, ZLocalShellCall, ZLocalShellCallOutput, ZMcpListTools, ZMcpApprovalRequest, ZMcpApprovalResponse, ZMcpCall, ZResponseCustomToolCallOutput, ZResponseCustomToolCall, ZItemReference]))]).nullable().optional(),
    metadata: z.record(z.string(), z.string()).nullable().optional(),
    model: z.union([z.string(), z.union([z.literal("gpt-5"), z.literal("gpt-5-mini"), z.literal("gpt-5-nano"), z.literal("gpt-5-2025-08-07"), z.literal("gpt-5-mini-2025-08-07"), z.literal("gpt-5-nano-2025-08-07"), z.literal("gpt-5-chat-latest"), z.literal("gpt-4.1"), z.literal("gpt-4.1-mini"), z.literal("gpt-4.1-nano"), z.literal("gpt-4.1-2025-04-14"), z.literal("gpt-4.1-mini-2025-04-14"), z.literal("gpt-4.1-nano-2025-04-14"), z.literal("o4-mini"), z.literal("o4-mini-2025-04-16"), z.literal("o3"), z.literal("o3-2025-04-16"), z.literal("o3-mini"), z.literal("o3-mini-2025-01-31"), z.literal("o1"), z.literal("o1-2024-12-17"), z.literal("o1-preview"), z.literal("o1-preview-2024-09-12"), z.literal("o1-mini"), z.literal("o1-mini-2024-09-12"), z.literal("gpt-4o"), z.literal("gpt-4o-2024-11-20"), z.literal("gpt-4o-2024-08-06"), z.literal("gpt-4o-2024-05-13"), z.literal("gpt-4o-audio-preview"), z.literal("gpt-4o-audio-preview-2024-10-01"), z.literal("gpt-4o-audio-preview-2024-12-17"), z.literal("gpt-4o-audio-preview-2025-06-03"), z.literal("gpt-4o-mini-audio-preview"), z.literal("gpt-4o-mini-audio-preview-2024-12-17"), z.literal("gpt-4o-search-preview"), z.literal("gpt-4o-mini-search-preview"), z.literal("gpt-4o-search-preview-2025-03-11"), z.literal("gpt-4o-mini-search-preview-2025-03-11"), z.literal("chatgpt-4o-latest"), z.literal("codex-mini-latest"), z.literal("gpt-4o-mini"), z.literal("gpt-4o-mini-2024-07-18"), z.literal("gpt-4-turbo"), z.literal("gpt-4-turbo-2024-04-09"), z.literal("gpt-4-0125-preview"), z.literal("gpt-4-turbo-preview"), z.literal("gpt-4-1106-preview"), z.literal("gpt-4-vision-preview"), z.literal("gpt-4"), z.literal("gpt-4-0314"), z.literal("gpt-4-0613"), z.literal("gpt-4-32k"), z.literal("gpt-4-32k-0314"), z.literal("gpt-4-32k-0613"), z.literal("gpt-3.5-turbo"), z.literal("gpt-3.5-turbo-16k"), z.literal("gpt-3.5-turbo-0301"), z.literal("gpt-3.5-turbo-0613"), z.literal("gpt-3.5-turbo-1106"), z.literal("gpt-3.5-turbo-0125"), z.literal("gpt-3.5-turbo-16k-0613")]), z.union([z.literal("o1-pro"), z.literal("o1-pro-2025-03-19"), z.literal("o3-pro"), z.literal("o3-pro-2025-06-10"), z.literal("o3-deep-research"), z.literal("o3-deep-research-2025-06-26"), z.literal("o4-mini-deep-research"), z.literal("o4-mini-deep-research-2025-06-26"), z.literal("computer-use-preview"), z.literal("computer-use-preview-2025-03-11"), z.literal("gpt-5-codex"), z.literal("gpt-5-pro"), z.literal("gpt-5-pro-2025-10-06")])]),
    object: z.literal("response"),
    output: z.array(z.union([ZResponseOutputMessage, ZResponseFileSearchToolCall, ZResponseFunctionToolCall, ZResponseFunctionWebSearch, ZResponseComputerToolCall, ZResponseReasoningItem, ZResponseOutputItemImageGenerationCall, ZResponseCodeInterpreterToolCall, ZResponseOutputItemLocalShellCall, ZResponseOutputItemMcpCall, ZResponseOutputItemMcpListTools, ZResponseOutputItemMcpApprovalRequest, ZResponseCustomToolCall])),
    parallel_tool_calls: z.boolean(),
    temperature: z.number().nullable().optional(),
    tool_choice: z.union([z.union([z.literal("none"), z.literal("auto"), z.literal("required")]), ZToolChoiceAllowed, ZToolChoiceTypes, ZToolChoiceFunction, ZToolChoiceMcp, ZToolChoiceCustom]),
    tools: z.array(z.union([ZFunctionTool, ZFileSearchTool, ZComputerTool, ZWebSearchTool, ZMcp, ZCodeInterpreter, ZImageGeneration, ZLocalShell, ZCustomTool, ZWebSearchPreviewTool])),
    top_p: z.number().nullable().optional(),
    background: z.boolean().nullable().optional(),
    conversation: ZConversation.nullable().optional(),
    max_output_tokens: z.number().nullable().optional(),
    max_tool_calls: z.number().nullable().optional(),
    previous_response_id: z.string().nullable().optional(),
    prompt: ZResponsePrompt.nullable().optional(),
    prompt_cache_key: z.string().nullable().optional(),
    reasoning: ZReasoning.nullable().optional(),
    safety_identifier: z.string().nullable().optional(),
    service_tier: z.union([z.literal("auto"), z.literal("default"), z.literal("flex"), z.literal("scale"), z.literal("priority")]).nullable().optional(),
    status: z.union([z.literal("completed"), z.literal("failed"), z.literal("in_progress"), z.literal("cancelled"), z.literal("queued"), z.literal("incomplete")]).nullable().optional(),
    text: ZResponseTextConfig.nullable().optional(),
    top_logprobs: z.number().nullable().optional(),
    truncation: z.union([z.literal("auto"), z.literal("disabled")]).nullable().optional(),
    usage: ZResponseUsage.nullable().optional(),
    user: z.string().nullable().optional(),
})));
export type Response = z.infer<typeof ZResponse>;

export const ZResponseInputImageParam = z.lazy(() => (z.object({
    detail: z.union([z.literal("low"), z.literal("high"), z.literal("auto")]),
    type: z.literal("input_image"),
    file_id: z.string().nullable().optional(),
    image_url: z.string().nullable().optional(),
})));
export type ResponseInputImageParam = z.infer<typeof ZResponseInputImageParam>;

export const ZResponseInputTextParam = z.lazy(() => (z.object({
    text: z.string(),
    type: z.literal("input_text"),
})));
export type ResponseInputTextParam = z.infer<typeof ZResponseInputTextParam>;

export const ZRetabParsedChatCompletion = z.lazy(() => (ZParsedChatCompletion.schema).merge(z.object({
    choices: z.array(ZRetabParsedChoice),
    extraction_id: z.string().nullable().optional(),
    likelihoods: z.record(z.string(), z.any()).nullable().optional(),
    requires_human_review: z.boolean().default(false),
    schema_validation_error: ZErrorDetail.nullable().optional(),
    request_at: z.string().nullable().optional(),
    first_token_at: z.string().nullable().optional(),
    last_token_at: z.string().nullable().optional(),
})));
export type RetabParsedChatCompletion = z.infer<typeof ZRetabParsedChatCompletion>;

export const ZRetabParsedChoice = z.lazy(() => (ZParsedChoice.schema).merge(z.object({
    finish_reason: z.union([z.literal("stop"), z.literal("length"), z.literal("tool_calls"), z.literal("content_filter"), z.literal("function_call")]).nullable().optional(),
    field_locations: z.record(z.string(), ZFieldLocation).nullable().optional(),
    key_mapping: z.record(z.string(), z.string().nullable().optional()).nullable().optional(),
})));
export type RetabParsedChoice = z.infer<typeof ZRetabParsedChoice>;

export const ZTextBlockParam = z.lazy(() => (z.object({
    text: z.string(),
    type: z.literal("text"),
    cache_control: ZCacheControlEphemeralParam.nullable().optional(),
    citations: z.array(z.union([ZCitationCharLocationParam, ZCitationPageLocationParam, ZCitationContentBlockLocationParam, ZCitationWebSearchResultLocationParam, ZCitationSearchResultLocationParam])).nullable().optional(),
})));
export type TextBlockParam = z.infer<typeof ZTextBlockParam>;

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
    model: z.string().default("gemini-2.5-flash"),
    table_parsing_format: z.union([z.literal("markdown"), z.literal("yaml"), z.literal("html"), z.literal("json")]).default("html"),
    image_resolution_dpi: z.number().default(192),
    browser_canvas: z.union([z.literal("A1"), z.literal("A2"), z.literal("A3"), z.literal("A4"), z.literal("A5")]).default("A4"),
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
    image_resolution_dpi: z.number().default(192),
    browser_canvas: z.union([z.literal("A1"), z.literal("A2"), z.literal("A3"), z.literal("A4"), z.literal("A5")]).default("A4"),
    model: z.string().default("gemini-2.5-flash"),
})));
export type DocumentCreateMessageRequest = z.infer<typeof ZDocumentCreateMessageRequest>;

export const ZDocumentMessage = z.lazy(() => (z.object({
    id: z.string(),
    object: z.literal("document_message").default("document_message"),
    messages: z.array(ZChatCompletionRetabMessage),
    created: z.number(),
})));
export type DocumentMessage = z.infer<typeof ZDocumentMessage>;

export const ZTokenCount = z.lazy(() => (z.object({
    total_tokens: z.number().default(0),
    developer_tokens: z.number().default(0),
    user_tokens: z.number().default(0),
})));
export type TokenCount = z.infer<typeof ZTokenCount>;

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
    reasoning_effort: z.union([z.literal("minimal"), z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional().default("minimal"),
})));
export type ConsensusModel = z.infer<typeof ZConsensusModel>;

export const ZDocumentExtractRequest = z.lazy(() => (z.object({
    document: ZMIMEData,
    documents: z.array(ZMIMEData).default([]),
    image_resolution_dpi: z.number().default(192),
    model: z.string(),
    json_schema: z.record(z.string(), z.any()),
    temperature: z.number().default(0.0),
    reasoning_effort: z.union([z.literal("minimal"), z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional().default("minimal"),
    n_consensus: z.number().default(1),
    stream: z.boolean().default(false),
    seed: z.number().nullable().optional(),
    store: z.boolean().default(true),
    need_validation: z.boolean().default(false),
    modality: z.union([z.literal("text"), z.literal("image"), z.literal("native")]).default("native"),
    parallel_ocr_keys: z.record(z.string(), z.string()).nullable().optional(),
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
    match_level: z.union([z.literal("token"), z.literal("line"), z.literal("block"), z.literal("token-windows")]).nullable().optional(),
})));
export type FieldLocation = z.infer<typeof ZFieldLocation>;

export const ZLogExtractionRequest = z.lazy(() => (z.object({
    messages: z.array(ZChatCompletionRetabMessage).nullable().optional(),
    openai_messages: z.array(z.union([ZChatCompletionDeveloperMessageParam, ZChatCompletionSystemMessageParam, ZChatCompletionUserMessageParam, ZChatCompletionAssistantMessageParam, ZChatCompletionToolMessageParam, ZChatCompletionFunctionMessageParam])).nullable().optional(),
    openai_responses_input: z.array(z.union([ZEasyInputMessageParam, ZResponseInputParamMessage, ZResponseOutputMessageParam, ZResponseFileSearchToolCallParam, ZResponseComputerToolCallParam, ZResponseInputParamComputerCallOutput, ZResponseFunctionWebSearchParam, ZResponseFunctionToolCallParam, ZResponseInputParamFunctionCallOutput, ZResponseReasoningItemParam, ZResponseInputParamImageGenerationCall, ZResponseCodeInterpreterToolCallParam, ZResponseInputParamLocalShellCall, ZResponseInputParamLocalShellCallOutput, ZResponseInputParamMcpListTools, ZResponseInputParamMcpApprovalRequest, ZResponseInputParamMcpApprovalResponse, ZResponseInputParamMcpCall, ZResponseCustomToolCallOutputParam, ZResponseCustomToolCallParam, ZResponseInputParamItemReference])).nullable().optional(),
    document: ZMIMEData.default({"filename": "dummy.txt", "url": "data:text/plain;base64,Tm8gZG9jdW1lbnQgcHJvdmlkZWQ="}),
    completion: z.union([z.record(z.any()), ZRetabParsedChatCompletion, ZParsedChatCompletion, ZChatCompletion]).nullable().optional(),
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

export const ZParsedChatCompletion = z.lazy(() => (ZChatCompletion.schema).merge(z.object({
    choices: z.array(ZParsedChoice),
})));
export type ParsedChatCompletion = z.infer<typeof ZParsedChatCompletion>;

export const ZParsedChoice = z.lazy(() => (ZChatCompletionChoice.schema).merge(z.object({
    message: ZParsedChatCompletionMessage,
})));
export type ParsedChoice = z.infer<typeof ZParsedChoice>;

export const ZRetabParsedChatCompletionChunk = z.lazy(() => (ZStreamingBaseModel.schema).merge(ZChatCompletionChunk.schema).merge(z.object({
    choices: z.array(ZRetabParsedChoiceChunk),
    extraction_id: z.string().nullable().optional(),
    schema_validation_error: ZErrorDetail.nullable().optional(),
    request_at: z.string().nullable().optional(),
    first_token_at: z.string().nullable().optional(),
    last_token_at: z.string().nullable().optional(),
})));
export type RetabParsedChatCompletionChunk = z.infer<typeof ZRetabParsedChatCompletionChunk>;

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

export const ZChatCompletionMessageFunctionToolCallParamFunction = z.lazy(() => (z.object({
    arguments: z.string(),
    name: z.string(),
})));
export type ChatCompletionMessageFunctionToolCallParamFunction = z.infer<typeof ZChatCompletionMessageFunctionToolCallParamFunction>;

export const ZFile = z.lazy(() => (z.object({
    file: ZFileFile,
    type: z.literal("file"),
})));
export type File = z.infer<typeof ZFile>;

export const ZDocumentBlockParam = z.lazy(() => (z.object({
    source: z.union([ZBase64PDFSourceParam, ZPlainTextSourceParam, ZContentBlockSourceParam, ZURLPDFSourceParam]),
    type: z.literal("document"),
    cache_control: ZCacheControlEphemeralParam.nullable().optional(),
    citations: ZCitationsConfigParam.nullable().optional(),
    context: z.string().nullable().optional(),
    title: z.string().nullable().optional(),
})));
export type DocumentBlockParam = z.infer<typeof ZDocumentBlockParam>;

export const ZSearchResultBlockParam = z.lazy(() => (z.object({
    content: z.array(ZTextBlockParam),
    source: z.string(),
    title: z.string(),
    type: z.literal("search_result"),
    cache_control: ZCacheControlEphemeralParam.nullable().optional(),
    citations: ZCitationsConfigParam,
})));
export type SearchResultBlockParam = z.infer<typeof ZSearchResultBlockParam>;

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
    input: z.record(z.string(), z.object({}).passthrough()),
    name: z.string(),
    type: z.literal("tool_use"),
    cache_control: ZCacheControlEphemeralParam.nullable().optional(),
})));
export type ToolUseBlockParam = z.infer<typeof ZToolUseBlockParam>;

export const ZToolResultBlockParam = z.lazy(() => (z.object({
    tool_use_id: z.string(),
    type: z.literal("tool_result"),
    cache_control: ZCacheControlEphemeralParam.nullable().optional(),
    content: z.union([z.string(), z.array(z.union([ZTextBlockParam, ZImageBlockParam, ZSearchResultBlockParam, ZDocumentBlockParam]))]),
    is_error: z.boolean(),
})));
export type ToolResultBlockParam = z.infer<typeof ZToolResultBlockParam>;

export const ZServerToolUseBlockParam = z.lazy(() => (z.object({
    id: z.string(),
    input: z.record(z.string(), z.object({}).passthrough()),
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
    citations: z.array(z.union([ZCitationCharLocation, ZCitationPageLocation, ZCitationContentBlockLocation, ZCitationsWebSearchResultLocation, ZCitationsSearchResultLocation])).nullable().optional(),
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
    input: z.record(z.string(), z.object({}).passthrough()),
    name: z.string(),
    type: z.literal("tool_use"),
})));
export type ToolUseBlock = z.infer<typeof ZToolUseBlock>;

export const ZServerToolUseBlock = z.lazy(() => (z.object({
    id: z.string(),
    input: z.record(z.string(), z.object({}).passthrough()),
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

export const ZInputAudio = z.lazy(() => (z.object({
    data: z.string(),
    format: z.union([z.literal("wav"), z.literal("mp3")]),
})));
export type InputAudio = z.infer<typeof ZInputAudio>;

export const ZResponseInputFileParam = z.lazy(() => (z.object({
    type: z.literal("input_file"),
    file_data: z.string(),
    file_id: z.string().nullable().optional(),
    file_url: z.string(),
    filename: z.string(),
})));
export type ResponseInputFileParam = z.infer<typeof ZResponseInputFileParam>;

export const ZResponseInputAudioParam = z.lazy(() => (z.object({
    input_audio: ZResponseInputAudioParamInputAudio,
    type: z.literal("input_audio"),
})));
export type ResponseInputAudioParam = z.infer<typeof ZResponseInputAudioParam>;

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

export const ZCacheControlEphemeralParam = z.lazy(() => (z.object({
    type: z.literal("ephemeral"),
    ttl: z.union([z.literal("5m"), z.literal("1h")]),
})));
export type CacheControlEphemeralParam = z.infer<typeof ZCacheControlEphemeralParam>;

export const ZParsedFunctionToolCall = z.lazy(() => (ZChatCompletionMessageFunctionToolCall.schema).merge(z.object({
    function: ZParsedFunction,
})));
export type ParsedFunctionToolCall = z.infer<typeof ZParsedFunctionToolCall>;

export const ZVideoMetadataDict = z.lazy(() => (z.object({
    fps: z.number().nullable().optional(),
    end_offset: z.string().nullable().optional(),
    start_offset: z.string().nullable().optional(),
})));
export type VideoMetadataDict = z.infer<typeof ZVideoMetadataDict>;

export const ZFileDataDict = z.lazy(() => (z.object({
    display_name: z.string().nullable().optional(),
    file_uri: z.string().nullable().optional(),
    mime_type: z.string().nullable().optional(),
})));
export type FileDataDict = z.infer<typeof ZFileDataDict>;

export const ZFunctionCallDict = z.lazy(() => (z.object({
    id: z.string().nullable().optional(),
    args: z.record(z.string(), z.any()).nullable().optional(),
    name: z.string().nullable().optional(),
})));
export type FunctionCallDict = z.infer<typeof ZFunctionCallDict>;

export const ZCodeExecutionResultDict = z.lazy(() => (z.object({
    outcome: z.any().nullable().optional(),
    output: z.string().nullable().optional(),
})));
export type CodeExecutionResultDict = z.infer<typeof ZCodeExecutionResultDict>;

export const ZExecutableCodeDict = z.lazy(() => (z.object({
    code: z.string().nullable().optional(),
    language: z.any().nullable().optional(),
})));
export type ExecutableCodeDict = z.infer<typeof ZExecutableCodeDict>;

export const ZFunctionResponseDict = z.lazy(() => (z.object({
    will_continue: z.boolean().nullable().optional(),
    scheduling: z.any().nullable().optional(),
    parts: z.array(ZFunctionResponsePartDict).nullable().optional(),
    id: z.string().nullable().optional(),
    name: z.string().nullable().optional(),
    response: z.record(z.string(), z.any()).nullable().optional(),
})));
export type FunctionResponseDict = z.infer<typeof ZFunctionResponseDict>;

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
    content: z.union([z.string(), z.array(z.union([ZResponseInputText, ZResponseInputImage, ZResponseInputFile, ZResponseInputAudio]))]),
    role: z.union([z.literal("user"), z.literal("assistant"), z.literal("system"), z.literal("developer")]),
    type: z.literal("message").nullable().optional(),
})));
export type EasyInputMessage = z.infer<typeof ZEasyInputMessage>;

export const ZMessage = z.lazy(() => (z.object({
    content: z.array(z.union([ZResponseInputText, ZResponseInputImage, ZResponseInputFile, ZResponseInputAudio])),
    role: z.union([z.literal("user"), z.literal("system"), z.literal("developer")]),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).nullable().optional(),
    type: z.literal("message").nullable().optional(),
})));
export type Message = z.infer<typeof ZMessage>;

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
    results: z.array(ZResult).nullable().optional(),
})));
export type ResponseFileSearchToolCall = z.infer<typeof ZResponseFileSearchToolCall>;

export const ZResponseComputerToolCall = z.lazy(() => (z.object({
    id: z.string(),
    action: z.union([ZActionClick, ZActionDoubleClick, ZActionDrag, ZActionKeypress, ZActionMove, ZActionScreenshot, ZActionScroll, ZActionType, ZActionWait]),
    call_id: z.string(),
    pending_safety_checks: z.array(ZPendingSafetyCheck),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
    type: z.literal("computer_call"),
})));
export type ResponseComputerToolCall = z.infer<typeof ZResponseComputerToolCall>;

export const ZComputerCallOutput = z.lazy(() => (z.object({
    call_id: z.string(),
    output: ZResponseComputerToolCallOutputScreenshot,
    type: z.literal("computer_call_output"),
    id: z.string().nullable().optional(),
    acknowledged_safety_checks: z.array(ZComputerCallOutputAcknowledgedSafetyCheck).nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).nullable().optional(),
})));
export type ComputerCallOutput = z.infer<typeof ZComputerCallOutput>;

export const ZResponseFunctionWebSearch = z.lazy(() => (z.object({
    id: z.string(),
    action: z.union([ZActionSearch, ZActionOpenPage, ZActionFind]),
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

export const ZFunctionCallOutput = z.lazy(() => (z.object({
    call_id: z.string(),
    output: z.union([z.string(), z.array(z.union([ZResponseInputTextContent, ZResponseInputImageContent, ZResponseInputFileContent]))]),
    type: z.literal("function_call_output"),
    id: z.string().nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).nullable().optional(),
})));
export type FunctionCallOutput = z.infer<typeof ZFunctionCallOutput>;

export const ZResponseReasoningItem = z.lazy(() => (z.object({
    id: z.string(),
    summary: z.array(ZSummary),
    type: z.literal("reasoning"),
    content: z.array(ZContent).nullable().optional(),
    encrypted_content: z.string().nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).nullable().optional(),
})));
export type ResponseReasoningItem = z.infer<typeof ZResponseReasoningItem>;

export const ZImageGenerationCall = z.lazy(() => (z.object({
    id: z.string(),
    result: z.string().nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("generating"), z.literal("failed")]),
    type: z.literal("image_generation_call"),
})));
export type ImageGenerationCall = z.infer<typeof ZImageGenerationCall>;

export const ZResponseCodeInterpreterToolCall = z.lazy(() => (z.object({
    id: z.string(),
    code: z.string().nullable().optional(),
    container_id: z.string(),
    outputs: z.array(z.union([ZOutputLogs, ZOutputImage])).nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete"), z.literal("interpreting"), z.literal("failed")]),
    type: z.literal("code_interpreter_call"),
})));
export type ResponseCodeInterpreterToolCall = z.infer<typeof ZResponseCodeInterpreterToolCall>;

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
    approval_request_id: z.string().nullable().optional(),
    error: z.string().nullable().optional(),
    output: z.string().nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete"), z.literal("calling"), z.literal("failed")]).nullable().optional(),
})));
export type McpCall = z.infer<typeof ZMcpCall>;

export const ZResponseCustomToolCallOutput = z.lazy(() => (z.object({
    call_id: z.string(),
    output: z.union([z.string(), z.array(z.union([ZResponseInputText, ZResponseInputImage, ZResponseInputFile]))]),
    type: z.literal("custom_tool_call_output"),
    id: z.string().nullable().optional(),
})));
export type ResponseCustomToolCallOutput = z.infer<typeof ZResponseCustomToolCallOutput>;

export const ZResponseCustomToolCall = z.lazy(() => (z.object({
    call_id: z.string(),
    input: z.string(),
    name: z.string(),
    type: z.literal("custom_tool_call"),
    id: z.string().nullable().optional(),
})));
export type ResponseCustomToolCall = z.infer<typeof ZResponseCustomToolCall>;

export const ZItemReference = z.lazy(() => (z.object({
    id: z.string(),
    type: z.literal("item_reference").nullable().optional(),
})));
export type ItemReference = z.infer<typeof ZItemReference>;

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
    approval_request_id: z.string().nullable().optional(),
    error: z.string().nullable().optional(),
    output: z.string().nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete"), z.literal("calling"), z.literal("failed")]).nullable().optional(),
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

export const ZToolChoiceAllowed = z.lazy(() => (z.object({
    mode: z.union([z.literal("auto"), z.literal("required")]),
    tools: z.array(z.record(z.string(), z.object({}).passthrough())),
    type: z.literal("allowed_tools"),
})));
export type ToolChoiceAllowed = z.infer<typeof ZToolChoiceAllowed>;

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

export const ZToolChoiceCustom = z.lazy(() => (z.object({
    name: z.string(),
    type: z.literal("custom"),
})));
export type ToolChoiceCustom = z.infer<typeof ZToolChoiceCustom>;

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

export const ZComputerTool = z.lazy(() => (z.object({
    display_height: z.number(),
    display_width: z.number(),
    environment: z.union([z.literal("windows"), z.literal("mac"), z.literal("linux"), z.literal("ubuntu"), z.literal("browser")]),
    type: z.literal("computer_use_preview"),
})));
export type ComputerTool = z.infer<typeof ZComputerTool>;

export const ZWebSearchTool = z.lazy(() => (z.object({
    type: z.union([z.literal("web_search"), z.literal("web_search_2025_08_26")]),
    filters: ZFilters.nullable().optional(),
    search_context_size: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional(),
    user_location: ZUserLocation.nullable().optional(),
})));
export type WebSearchTool = z.infer<typeof ZWebSearchTool>;

export const ZMcp = z.lazy(() => (z.object({
    server_label: z.string(),
    type: z.literal("mcp"),
    allowed_tools: z.union([z.array(z.string()), ZMcpAllowedToolsMcpToolFilter]).nullable().optional(),
    authorization: z.string().nullable().optional(),
    connector_id: z.union([z.literal("connector_dropbox"), z.literal("connector_gmail"), z.literal("connector_googlecalendar"), z.literal("connector_googledrive"), z.literal("connector_microsoftteams"), z.literal("connector_outlookcalendar"), z.literal("connector_outlookemail"), z.literal("connector_sharepoint")]).nullable().optional(),
    headers: z.record(z.string(), z.string()).nullable().optional(),
    require_approval: z.union([ZMcpRequireApprovalMcpToolApprovalFilter, z.union([z.literal("always"), z.literal("never")])]).nullable().optional(),
    server_description: z.string().nullable().optional(),
    server_url: z.string().nullable().optional(),
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
    model: z.union([z.literal("gpt-image-1"), z.literal("gpt-image-1-mini")]).nullable().optional(),
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

export const ZCustomTool = z.lazy(() => (z.object({
    name: z.string(),
    type: z.literal("custom"),
    description: z.string().nullable().optional(),
    format: z.union([ZText, ZGrammar]).nullable().optional(),
})));
export type CustomTool = z.infer<typeof ZCustomTool>;

export const ZWebSearchPreviewTool = z.lazy(() => (z.object({
    type: z.union([z.literal("web_search_preview"), z.literal("web_search_preview_2025_03_11")]),
    search_context_size: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional(),
    user_location: ZWebSearchPreviewToolUserLocation.nullable().optional(),
})));
export type WebSearchPreviewTool = z.infer<typeof ZWebSearchPreviewTool>;

export const ZConversation = z.lazy(() => (z.object({
    id: z.string(),
})));
export type Conversation = z.infer<typeof ZConversation>;

export const ZResponsePrompt = z.lazy(() => (z.object({
    id: z.string(),
    variables: z.record(z.string(), z.union([z.string(), ZResponseInputText, ZResponseInputImage, ZResponseInputFile])).nullable().optional(),
    version: z.string().nullable().optional(),
})));
export type ResponsePrompt = z.infer<typeof ZResponsePrompt>;

export const ZReasoning = z.lazy(() => (z.object({
    effort: z.union([z.literal("minimal"), z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional(),
    generate_summary: z.union([z.literal("auto"), z.literal("concise"), z.literal("detailed")]).nullable().optional(),
    summary: z.union([z.literal("auto"), z.literal("concise"), z.literal("detailed")]).nullable().optional(),
})));
export type Reasoning = z.infer<typeof ZReasoning>;

export const ZResponseTextConfig = z.lazy(() => (z.object({
    format: z.union([ZResponseFormatText, ZResponseFormatTextJSONSchemaConfig, ZResponseFormatJSONObject]).nullable().optional(),
    verbosity: z.union([z.literal("low"), z.literal("medium"), z.literal("high")]).nullable().optional(),
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

export const ZCitationSearchResultLocationParam = z.lazy(() => (z.object({
    cited_text: z.string(),
    end_block_index: z.number(),
    search_result_index: z.number(),
    source: z.string(),
    start_block_index: z.number(),
    title: z.string().nullable().optional(),
    type: z.literal("search_result_location"),
})));
export type CitationSearchResultLocationParam = z.infer<typeof ZCitationSearchResultLocationParam>;

export const ZChatCompletionChoice = z.lazy(() => (z.object({
    finish_reason: z.union([z.literal("stop"), z.literal("length"), z.literal("tool_calls"), z.literal("content_filter"), z.literal("function_call")]),
    index: z.number(),
    logprobs: ZChatCompletionChoiceLogprobs.nullable().optional(),
    message: ZChatCompletionMessage,
})));
export type ChatCompletionChoice = z.infer<typeof ZChatCompletionChoice>;

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
    tool_calls: z.array(z.union([ZChatCompletionMessageFunctionToolCallParam, ZChatCompletionMessageCustomToolCallParam])),
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

export const ZResponseInputParamMessage = z.lazy(() => (z.object({
    content: z.array(z.union([ZResponseInputTextParam, ZResponseInputImageParam, ZResponseInputFileParam, ZResponseInputAudioParam])),
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
    results: z.array(ZResponseFileSearchToolCallParamResult).nullable().optional(),
})));
export type ResponseFileSearchToolCallParam = z.infer<typeof ZResponseFileSearchToolCallParam>;

export const ZResponseComputerToolCallParam = z.lazy(() => (z.object({
    id: z.string(),
    action: z.union([ZResponseComputerToolCallParamActionClick, ZResponseComputerToolCallParamActionDoubleClick, ZResponseComputerToolCallParamActionDrag, ZResponseComputerToolCallParamActionKeypress, ZResponseComputerToolCallParamActionMove, ZResponseComputerToolCallParamActionScreenshot, ZResponseComputerToolCallParamActionScroll, ZResponseComputerToolCallParamActionType, ZResponseComputerToolCallParamActionWait]),
    call_id: z.string(),
    pending_safety_checks: z.array(ZResponseComputerToolCallParamPendingSafetyCheck),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
    type: z.literal("computer_call"),
})));
export type ResponseComputerToolCallParam = z.infer<typeof ZResponseComputerToolCallParam>;

export const ZResponseInputParamComputerCallOutput = z.lazy(() => (z.object({
    call_id: z.string(),
    output: ZResponseComputerToolCallOutputScreenshotParam,
    type: z.literal("computer_call_output"),
    id: z.string().nullable().optional(),
    acknowledged_safety_checks: z.array(ZResponseInputParamComputerCallOutputAcknowledgedSafetyCheck).nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).nullable().optional(),
})));
export type ResponseInputParamComputerCallOutput = z.infer<typeof ZResponseInputParamComputerCallOutput>;

export const ZResponseFunctionWebSearchParam = z.lazy(() => (z.object({
    id: z.string(),
    action: z.union([ZResponseFunctionWebSearchParamActionSearch, ZResponseFunctionWebSearchParamActionOpenPage, ZResponseFunctionWebSearchParamActionFind]),
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

export const ZResponseInputParamFunctionCallOutput = z.lazy(() => (z.object({
    call_id: z.string(),
    output: z.union([z.string(), z.array(z.union([ZResponseInputTextContentParam, ZResponseInputImageContentParam, ZResponseInputFileContentParam]))]),
    type: z.literal("function_call_output"),
    id: z.string().nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).nullable().optional(),
})));
export type ResponseInputParamFunctionCallOutput = z.infer<typeof ZResponseInputParamFunctionCallOutput>;

export const ZResponseReasoningItemParam = z.lazy(() => (z.object({
    id: z.string(),
    summary: z.array(ZResponseReasoningItemParamSummary),
    type: z.literal("reasoning"),
    content: z.array(ZResponseReasoningItemParamContent),
    encrypted_content: z.string().nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
})));
export type ResponseReasoningItemParam = z.infer<typeof ZResponseReasoningItemParam>;

export const ZResponseInputParamImageGenerationCall = z.lazy(() => (z.object({
    id: z.string(),
    result: z.string().nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("generating"), z.literal("failed")]),
    type: z.literal("image_generation_call"),
})));
export type ResponseInputParamImageGenerationCall = z.infer<typeof ZResponseInputParamImageGenerationCall>;

export const ZResponseCodeInterpreterToolCallParam = z.lazy(() => (z.object({
    id: z.string(),
    code: z.string().nullable().optional(),
    container_id: z.string(),
    outputs: z.array(z.union([ZResponseCodeInterpreterToolCallParamOutputLogs, ZResponseCodeInterpreterToolCallParamOutputImage])).nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete"), z.literal("interpreting"), z.literal("failed")]),
    type: z.literal("code_interpreter_call"),
})));
export type ResponseCodeInterpreterToolCallParam = z.infer<typeof ZResponseCodeInterpreterToolCallParam>;

export const ZResponseInputParamLocalShellCall = z.lazy(() => (z.object({
    id: z.string(),
    action: ZResponseInputParamLocalShellCallAction,
    call_id: z.string(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]),
    type: z.literal("local_shell_call"),
})));
export type ResponseInputParamLocalShellCall = z.infer<typeof ZResponseInputParamLocalShellCall>;

export const ZResponseInputParamLocalShellCallOutput = z.lazy(() => (z.object({
    id: z.string(),
    output: z.string(),
    type: z.literal("local_shell_call_output"),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete")]).nullable().optional(),
})));
export type ResponseInputParamLocalShellCallOutput = z.infer<typeof ZResponseInputParamLocalShellCallOutput>;

export const ZResponseInputParamMcpListTools = z.lazy(() => (z.object({
    id: z.string(),
    server_label: z.string(),
    tools: z.array(ZResponseInputParamMcpListToolsTool),
    type: z.literal("mcp_list_tools"),
    error: z.string().nullable().optional(),
})));
export type ResponseInputParamMcpListTools = z.infer<typeof ZResponseInputParamMcpListTools>;

export const ZResponseInputParamMcpApprovalRequest = z.lazy(() => (z.object({
    id: z.string(),
    arguments: z.string(),
    name: z.string(),
    server_label: z.string(),
    type: z.literal("mcp_approval_request"),
})));
export type ResponseInputParamMcpApprovalRequest = z.infer<typeof ZResponseInputParamMcpApprovalRequest>;

export const ZResponseInputParamMcpApprovalResponse = z.lazy(() => (z.object({
    approval_request_id: z.string(),
    approve: z.boolean(),
    type: z.literal("mcp_approval_response"),
    id: z.string().nullable().optional(),
    reason: z.string().nullable().optional(),
})));
export type ResponseInputParamMcpApprovalResponse = z.infer<typeof ZResponseInputParamMcpApprovalResponse>;

export const ZResponseInputParamMcpCall = z.lazy(() => (z.object({
    id: z.string(),
    arguments: z.string(),
    name: z.string(),
    server_label: z.string(),
    type: z.literal("mcp_call"),
    approval_request_id: z.string().nullable().optional(),
    error: z.string().nullable().optional(),
    output: z.string().nullable().optional(),
    status: z.union([z.literal("in_progress"), z.literal("completed"), z.literal("incomplete"), z.literal("calling"), z.literal("failed")]),
})));
export type ResponseInputParamMcpCall = z.infer<typeof ZResponseInputParamMcpCall>;

export const ZResponseCustomToolCallOutputParam = z.lazy(() => (z.object({
    call_id: z.string(),
    output: z.union([z.string(), z.array(z.union([ZResponseInputTextParam, ZResponseInputImageParam, ZResponseInputFileParam]))]),
    type: z.literal("custom_tool_call_output"),
    id: z.string(),
})));
export type ResponseCustomToolCallOutputParam = z.infer<typeof ZResponseCustomToolCallOutputParam>;

export const ZResponseCustomToolCallParam = z.lazy(() => (z.object({
    call_id: z.string(),
    input: z.string(),
    name: z.string(),
    type: z.literal("custom_tool_call"),
    id: z.string(),
})));
export type ResponseCustomToolCallParam = z.infer<typeof ZResponseCustomToolCallParam>;

export const ZResponseInputParamItemReference = z.lazy(() => (z.object({
    id: z.string(),
    type: z.literal("item_reference").nullable().optional(),
})));
export type ResponseInputParamItemReference = z.infer<typeof ZResponseInputParamItemReference>;

export const ZFileFile = z.lazy(() => (z.object({
    file_data: z.string(),
    file_id: z.string(),
    filename: z.string(),
})));
export type FileFile = z.infer<typeof ZFileFile>;

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
    file_id: z.string().nullable().optional(),
    start_char_index: z.number(),
    type: z.literal("char_location"),
})));
export type CitationCharLocation = z.infer<typeof ZCitationCharLocation>;

export const ZCitationPageLocation = z.lazy(() => (z.object({
    cited_text: z.string(),
    document_index: z.number(),
    document_title: z.string().nullable().optional(),
    end_page_number: z.number(),
    file_id: z.string().nullable().optional(),
    start_page_number: z.number(),
    type: z.literal("page_location"),
})));
export type CitationPageLocation = z.infer<typeof ZCitationPageLocation>;

export const ZCitationContentBlockLocation = z.lazy(() => (z.object({
    cited_text: z.string(),
    document_index: z.number(),
    document_title: z.string().nullable().optional(),
    end_block_index: z.number(),
    file_id: z.string().nullable().optional(),
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

export const ZCitationsSearchResultLocation = z.lazy(() => (z.object({
    cited_text: z.string(),
    end_block_index: z.number(),
    search_result_index: z.number(),
    source: z.string(),
    start_block_index: z.number(),
    title: z.string().nullable().optional(),
    type: z.literal("search_result_location"),
})));
export type CitationsSearchResultLocation = z.infer<typeof ZCitationsSearchResultLocation>;

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

export const ZResponseInputAudioParamInputAudio = z.lazy(() => (z.object({
    data: z.string(),
    format: z.union([z.literal("wav"), z.literal("mp3")]),
})));
export type ResponseInputAudioParamInputAudio = z.infer<typeof ZResponseInputAudioParamInputAudio>;

export const ZParsedFunction = z.lazy(() => (ZChatCompletionMessageFunctionToolCallFunction.schema).merge(z.object({
    parsed_arguments: z.object({}).passthrough().nullable().optional(),
})));
export type ParsedFunction = z.infer<typeof ZParsedFunction>;

export const ZFunctionResponsePartDict = z.lazy(() => (z.object({
    inline_data: ZFunctionResponseBlobDict.nullable().optional(),
    file_data: ZFunctionResponseFileDataDict.nullable().optional(),
})));
export type FunctionResponsePartDict = z.infer<typeof ZFunctionResponsePartDict>;

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

export const ZResponseInputAudio = z.lazy(() => (z.object({
    input_audio: ZResponseInputAudioInputAudio,
    type: z.literal("input_audio"),
})));
export type ResponseInputAudio = z.infer<typeof ZResponseInputAudio>;

export const ZResponseOutputText = z.lazy(() => (z.object({
    annotations: z.array(z.union([ZAnnotationFileCitation, ZAnnotationURLCitation, ZAnnotationContainerFileCitation, ZAnnotationFilePath])),
    text: z.string(),
    type: z.literal("output_text"),
    logprobs: z.array(ZLogprob).nullable().optional(),
})));
export type ResponseOutputText = z.infer<typeof ZResponseOutputText>;

export const ZResponseOutputRefusal = z.lazy(() => (z.object({
    refusal: z.string(),
    type: z.literal("refusal"),
})));
export type ResponseOutputRefusal = z.infer<typeof ZResponseOutputRefusal>;

export const ZResult = z.lazy(() => (z.object({
    attributes: z.record(z.string(), z.union([z.string(), z.number(), z.boolean()])).nullable().optional(),
    file_id: z.string().nullable().optional(),
    filename: z.string().nullable().optional(),
    score: z.number().nullable().optional(),
    text: z.string().nullable().optional(),
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
    code: z.string().nullable().optional(),
    message: z.string().nullable().optional(),
})));
export type PendingSafetyCheck = z.infer<typeof ZPendingSafetyCheck>;

export const ZResponseComputerToolCallOutputScreenshot = z.lazy(() => (z.object({
    type: z.literal("computer_screenshot"),
    file_id: z.string().nullable().optional(),
    image_url: z.string().nullable().optional(),
})));
export type ResponseComputerToolCallOutputScreenshot = z.infer<typeof ZResponseComputerToolCallOutputScreenshot>;

export const ZComputerCallOutputAcknowledgedSafetyCheck = z.lazy(() => (z.object({
    id: z.string(),
    code: z.string().nullable().optional(),
    message: z.string().nullable().optional(),
})));
export type ComputerCallOutputAcknowledgedSafetyCheck = z.infer<typeof ZComputerCallOutputAcknowledgedSafetyCheck>;

export const ZActionSearch = z.lazy(() => (z.object({
    query: z.string(),
    type: z.literal("search"),
    sources: z.array(ZActionSearchSource).nullable().optional(),
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

export const ZResponseInputTextContent = z.lazy(() => (z.object({
    text: z.string(),
    type: z.literal("input_text"),
})));
export type ResponseInputTextContent = z.infer<typeof ZResponseInputTextContent>;

export const ZResponseInputImageContent = z.lazy(() => (z.object({
    type: z.literal("input_image"),
    detail: z.union([z.literal("low"), z.literal("high"), z.literal("auto")]).nullable().optional(),
    file_id: z.string().nullable().optional(),
    image_url: z.string().nullable().optional(),
})));
export type ResponseInputImageContent = z.infer<typeof ZResponseInputImageContent>;

export const ZResponseInputFileContent = z.lazy(() => (z.object({
    type: z.literal("input_file"),
    file_data: z.string().nullable().optional(),
    file_id: z.string().nullable().optional(),
    file_url: z.string().nullable().optional(),
    filename: z.string().nullable().optional(),
})));
export type ResponseInputFileContent = z.infer<typeof ZResponseInputFileContent>;

export const ZSummary = z.lazy(() => (z.object({
    text: z.string(),
    type: z.literal("summary_text"),
})));
export type Summary = z.infer<typeof ZSummary>;

export const ZContent = z.lazy(() => (z.object({
    text: z.string(),
    type: z.literal("reasoning_text"),
})));
export type Content = z.infer<typeof ZContent>;

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
    value: z.union([z.string(), z.number(), z.boolean(), z.array(z.union([z.string(), z.number()]))]),
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

export const ZFilters = z.lazy(() => (z.object({
    allowed_domains: z.array(z.string()).nullable().optional(),
})));
export type Filters = z.infer<typeof ZFilters>;

export const ZUserLocation = z.lazy(() => (z.object({
    city: z.string().nullable().optional(),
    country: z.string().nullable().optional(),
    region: z.string().nullable().optional(),
    timezone: z.string().nullable().optional(),
    type: z.literal("approximate").nullable().optional(),
})));
export type UserLocation = z.infer<typeof ZUserLocation>;

export const ZMcpAllowedToolsMcpToolFilter = z.lazy(() => (z.object({
    read_only: z.boolean().nullable().optional(),
    tool_names: z.array(z.string()).nullable().optional(),
})));
export type McpAllowedToolsMcpToolFilter = z.infer<typeof ZMcpAllowedToolsMcpToolFilter>;

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

export const ZText = z.lazy(() => (z.object({
    type: z.literal("text"),
})));
export type Text = z.infer<typeof ZText>;

export const ZGrammar = z.lazy(() => (z.object({
    definition: z.string(),
    syntax: z.union([z.literal("lark"), z.literal("regex")]),
    type: z.literal("grammar"),
})));
export type Grammar = z.infer<typeof ZGrammar>;

export const ZWebSearchPreviewToolUserLocation = z.lazy(() => (z.object({
    type: z.literal("approximate"),
    city: z.string().nullable().optional(),
    country: z.string().nullable().optional(),
    region: z.string().nullable().optional(),
    timezone: z.string().nullable().optional(),
})));
export type WebSearchPreviewToolUserLocation = z.infer<typeof ZWebSearchPreviewToolUserLocation>;

export const ZResponseFormatText = z.lazy(() => (z.object({
    type: z.literal("text"),
})));
export type ResponseFormatText = z.infer<typeof ZResponseFormatText>;

export const ZResponseFormatTextJSONSchemaConfig = z.lazy(() => (z.object({
    name: z.string(),
    schema_: z.record(z.string(), z.object({}).passthrough()),
    type: z.literal("json_schema"),
    description: z.string().nullable().optional(),
    strict: z.boolean().nullable().optional(),
})));
export type ResponseFormatTextJSONSchemaConfig = z.infer<typeof ZResponseFormatTextJSONSchemaConfig>;

export const ZResponseFormatJSONObject = z.lazy(() => (z.object({
    type: z.literal("json_object"),
})));
export type ResponseFormatJSONObject = z.infer<typeof ZResponseFormatJSONObject>;

export const ZInputTokensDetails = z.lazy(() => (z.object({
    cached_tokens: z.number(),
})));
export type InputTokensDetails = z.infer<typeof ZInputTokensDetails>;

export const ZOutputTokensDetails = z.lazy(() => (z.object({
    reasoning_tokens: z.number(),
})));
export type OutputTokensDetails = z.infer<typeof ZOutputTokensDetails>;

export const ZChatCompletionChoiceLogprobs = z.lazy(() => (z.object({
    content: z.array(ZChatCompletionTokenLogprob).nullable().optional(),
    refusal: z.array(ZChatCompletionTokenLogprob).nullable().optional(),
})));
export type ChatCompletionChoiceLogprobs = z.infer<typeof ZChatCompletionChoiceLogprobs>;

export const ZChatCompletionMessage = z.lazy(() => (z.object({
    content: z.string().nullable().optional(),
    refusal: z.string().nullable().optional(),
    role: z.literal("assistant"),
    annotations: z.array(ZAnnotation).nullable().optional(),
    audio: ZChatCompletionAudio.nullable().optional(),
    function_call: ZChatCompletionMessageFunctionCall.nullable().optional(),
    tool_calls: z.array(z.union([ZChatCompletionMessageFunctionToolCall, ZChatCompletionMessageCustomToolCall])).nullable().optional(),
})));
export type ChatCompletionMessage = z.infer<typeof ZChatCompletionMessage>;

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

export const ZChatCompletionMessageCustomToolCallParam = z.lazy(() => (z.object({
    id: z.string(),
    custom: ZCustom,
    type: z.literal("custom"),
})));
export type ChatCompletionMessageCustomToolCallParam = z.infer<typeof ZChatCompletionMessageCustomToolCallParam>;

export const ZResponseOutputTextParam = z.lazy(() => (z.object({
    annotations: z.array(z.union([ZResponseOutputTextParamAnnotationFileCitation, ZResponseOutputTextParamAnnotationURLCitation, ZResponseOutputTextParamAnnotationContainerFileCitation, ZResponseOutputTextParamAnnotationFilePath])),
    text: z.string(),
    type: z.literal("output_text"),
    logprobs: z.array(ZResponseOutputTextParamLogprob),
})));
export type ResponseOutputTextParam = z.infer<typeof ZResponseOutputTextParam>;

export const ZResponseOutputRefusalParam = z.lazy(() => (z.object({
    refusal: z.string(),
    type: z.literal("refusal"),
})));
export type ResponseOutputRefusalParam = z.infer<typeof ZResponseOutputRefusalParam>;

export const ZResponseFileSearchToolCallParamResult = z.lazy(() => (z.object({
    attributes: z.record(z.string(), z.union([z.string(), z.number(), z.boolean()])).nullable().optional(),
    file_id: z.string(),
    filename: z.string(),
    score: z.number(),
    text: z.string(),
})));
export type ResponseFileSearchToolCallParamResult = z.infer<typeof ZResponseFileSearchToolCallParamResult>;

export const ZResponseComputerToolCallParamActionClick = z.lazy(() => (z.object({
    button: z.union([z.literal("left"), z.literal("right"), z.literal("wheel"), z.literal("back"), z.literal("forward")]),
    type: z.literal("click"),
    x: z.number(),
    y: z.number(),
})));
export type ResponseComputerToolCallParamActionClick = z.infer<typeof ZResponseComputerToolCallParamActionClick>;

export const ZResponseComputerToolCallParamActionDoubleClick = z.lazy(() => (z.object({
    type: z.literal("double_click"),
    x: z.number(),
    y: z.number(),
})));
export type ResponseComputerToolCallParamActionDoubleClick = z.infer<typeof ZResponseComputerToolCallParamActionDoubleClick>;

export const ZResponseComputerToolCallParamActionDrag = z.lazy(() => (z.object({
    path: z.array(ZResponseComputerToolCallParamActionDragPath),
    type: z.literal("drag"),
})));
export type ResponseComputerToolCallParamActionDrag = z.infer<typeof ZResponseComputerToolCallParamActionDrag>;

export const ZResponseComputerToolCallParamActionKeypress = z.lazy(() => (z.object({
    keys: z.array(z.string()),
    type: z.literal("keypress"),
})));
export type ResponseComputerToolCallParamActionKeypress = z.infer<typeof ZResponseComputerToolCallParamActionKeypress>;

export const ZResponseComputerToolCallParamActionMove = z.lazy(() => (z.object({
    type: z.literal("move"),
    x: z.number(),
    y: z.number(),
})));
export type ResponseComputerToolCallParamActionMove = z.infer<typeof ZResponseComputerToolCallParamActionMove>;

export const ZResponseComputerToolCallParamActionScreenshot = z.lazy(() => (z.object({
    type: z.literal("screenshot"),
})));
export type ResponseComputerToolCallParamActionScreenshot = z.infer<typeof ZResponseComputerToolCallParamActionScreenshot>;

export const ZResponseComputerToolCallParamActionScroll = z.lazy(() => (z.object({
    scroll_x: z.number(),
    scroll_y: z.number(),
    type: z.literal("scroll"),
    x: z.number(),
    y: z.number(),
})));
export type ResponseComputerToolCallParamActionScroll = z.infer<typeof ZResponseComputerToolCallParamActionScroll>;

export const ZResponseComputerToolCallParamActionType = z.lazy(() => (z.object({
    text: z.string(),
    type: z.literal("type"),
})));
export type ResponseComputerToolCallParamActionType = z.infer<typeof ZResponseComputerToolCallParamActionType>;

export const ZResponseComputerToolCallParamActionWait = z.lazy(() => (z.object({
    type: z.literal("wait"),
})));
export type ResponseComputerToolCallParamActionWait = z.infer<typeof ZResponseComputerToolCallParamActionWait>;

export const ZResponseComputerToolCallParamPendingSafetyCheck = z.lazy(() => (z.object({
    id: z.string(),
    code: z.string().nullable().optional(),
    message: z.string().nullable().optional(),
})));
export type ResponseComputerToolCallParamPendingSafetyCheck = z.infer<typeof ZResponseComputerToolCallParamPendingSafetyCheck>;

export const ZResponseComputerToolCallOutputScreenshotParam = z.lazy(() => (z.object({
    type: z.literal("computer_screenshot"),
    file_id: z.string(),
    image_url: z.string(),
})));
export type ResponseComputerToolCallOutputScreenshotParam = z.infer<typeof ZResponseComputerToolCallOutputScreenshotParam>;

export const ZResponseInputParamComputerCallOutputAcknowledgedSafetyCheck = z.lazy(() => (z.object({
    id: z.string(),
    code: z.string().nullable().optional(),
    message: z.string().nullable().optional(),
})));
export type ResponseInputParamComputerCallOutputAcknowledgedSafetyCheck = z.infer<typeof ZResponseInputParamComputerCallOutputAcknowledgedSafetyCheck>;

export const ZResponseFunctionWebSearchParamActionSearch = z.lazy(() => (z.object({
    query: z.string(),
    type: z.literal("search"),
    sources: z.array(ZResponseFunctionWebSearchParamActionSearchSource),
})));
export type ResponseFunctionWebSearchParamActionSearch = z.infer<typeof ZResponseFunctionWebSearchParamActionSearch>;

export const ZResponseFunctionWebSearchParamActionOpenPage = z.lazy(() => (z.object({
    type: z.literal("open_page"),
    url: z.string(),
})));
export type ResponseFunctionWebSearchParamActionOpenPage = z.infer<typeof ZResponseFunctionWebSearchParamActionOpenPage>;

export const ZResponseFunctionWebSearchParamActionFind = z.lazy(() => (z.object({
    pattern: z.string(),
    type: z.literal("find"),
    url: z.string(),
})));
export type ResponseFunctionWebSearchParamActionFind = z.infer<typeof ZResponseFunctionWebSearchParamActionFind>;

export const ZResponseInputTextContentParam = z.lazy(() => (z.object({
    text: z.string(),
    type: z.literal("input_text"),
})));
export type ResponseInputTextContentParam = z.infer<typeof ZResponseInputTextContentParam>;

export const ZResponseInputImageContentParam = z.lazy(() => (z.object({
    type: z.literal("input_image"),
    detail: z.union([z.literal("low"), z.literal("high"), z.literal("auto")]).nullable().optional(),
    file_id: z.string().nullable().optional(),
    image_url: z.string().nullable().optional(),
})));
export type ResponseInputImageContentParam = z.infer<typeof ZResponseInputImageContentParam>;

export const ZResponseInputFileContentParam = z.lazy(() => (z.object({
    type: z.literal("input_file"),
    file_data: z.string().nullable().optional(),
    file_id: z.string().nullable().optional(),
    file_url: z.string().nullable().optional(),
    filename: z.string().nullable().optional(),
})));
export type ResponseInputFileContentParam = z.infer<typeof ZResponseInputFileContentParam>;

export const ZResponseReasoningItemParamSummary = z.lazy(() => (z.object({
    text: z.string(),
    type: z.literal("summary_text"),
})));
export type ResponseReasoningItemParamSummary = z.infer<typeof ZResponseReasoningItemParamSummary>;

export const ZResponseReasoningItemParamContent = z.lazy(() => (z.object({
    text: z.string(),
    type: z.literal("reasoning_text"),
})));
export type ResponseReasoningItemParamContent = z.infer<typeof ZResponseReasoningItemParamContent>;

export const ZResponseCodeInterpreterToolCallParamOutputLogs = z.lazy(() => (z.object({
    logs: z.string(),
    type: z.literal("logs"),
})));
export type ResponseCodeInterpreterToolCallParamOutputLogs = z.infer<typeof ZResponseCodeInterpreterToolCallParamOutputLogs>;

export const ZResponseCodeInterpreterToolCallParamOutputImage = z.lazy(() => (z.object({
    type: z.literal("image"),
    url: z.string(),
})));
export type ResponseCodeInterpreterToolCallParamOutputImage = z.infer<typeof ZResponseCodeInterpreterToolCallParamOutputImage>;

export const ZResponseInputParamLocalShellCallAction = z.lazy(() => (z.object({
    command: z.array(z.string()),
    env: z.record(z.string(), z.string()),
    type: z.literal("exec"),
    timeout_ms: z.number().nullable().optional(),
    user: z.string().nullable().optional(),
    working_directory: z.string().nullable().optional(),
})));
export type ResponseInputParamLocalShellCallAction = z.infer<typeof ZResponseInputParamLocalShellCallAction>;

export const ZResponseInputParamMcpListToolsTool = z.lazy(() => (z.object({
    input_schema: z.object({}).passthrough(),
    name: z.string(),
    annotations: z.object({}).passthrough().nullable().optional(),
    description: z.string().nullable().optional(),
})));
export type ResponseInputParamMcpListToolsTool = z.infer<typeof ZResponseInputParamMcpListToolsTool>;

export const ZFunctionResponseBlobDict = z.lazy(() => (z.object({
    mime_type: z.string().nullable().optional(),
    data: z.instanceof(Uint8Array).nullable().optional(),
})));
export type FunctionResponseBlobDict = z.infer<typeof ZFunctionResponseBlobDict>;

export const ZFunctionResponseFileDataDict = z.lazy(() => (z.object({
    file_uri: z.string().nullable().optional(),
    mime_type: z.string().nullable().optional(),
})));
export type FunctionResponseFileDataDict = z.infer<typeof ZFunctionResponseFileDataDict>;

export const ZResponseInputAudioInputAudio = z.lazy(() => (z.object({
    data: z.string(),
    format: z.union([z.literal("mp3"), z.literal("wav")]),
})));
export type ResponseInputAudioInputAudio = z.infer<typeof ZResponseInputAudioInputAudio>;

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

export const ZActionSearchSource = z.lazy(() => (z.object({
    type: z.literal("url"),
    url: z.string(),
})));
export type ActionSearchSource = z.infer<typeof ZActionSearchSource>;

export const ZMcpRequireApprovalMcpToolApprovalFilterAlways = z.lazy(() => (z.object({
    read_only: z.boolean().nullable().optional(),
    tool_names: z.array(z.string()).nullable().optional(),
})));
export type McpRequireApprovalMcpToolApprovalFilterAlways = z.infer<typeof ZMcpRequireApprovalMcpToolApprovalFilterAlways>;

export const ZMcpRequireApprovalMcpToolApprovalFilterNever = z.lazy(() => (z.object({
    read_only: z.boolean().nullable().optional(),
    tool_names: z.array(z.string()).nullable().optional(),
})));
export type McpRequireApprovalMcpToolApprovalFilterNever = z.infer<typeof ZMcpRequireApprovalMcpToolApprovalFilterNever>;

export const ZAnnotation = z.lazy(() => (z.object({
    type: z.literal("url_citation"),
    url_citation: ZChatCompletionMessageAnnotationURLCitation,
})));
export type Annotation = z.infer<typeof ZAnnotation>;

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

export const ZChatCompletionMessageFunctionToolCall = z.lazy(() => (z.object({
    id: z.string(),
    function: ZChatCompletionMessageFunctionToolCallFunction,
    type: z.literal("function"),
})));
export type ChatCompletionMessageFunctionToolCall = z.infer<typeof ZChatCompletionMessageFunctionToolCall>;

export const ZChatCompletionMessageCustomToolCall = z.lazy(() => (z.object({
    id: z.string(),
    custom: ZChatCompletionMessageCustomToolCallCustom,
    type: z.literal("custom"),
})));
export type ChatCompletionMessageCustomToolCall = z.infer<typeof ZChatCompletionMessageCustomToolCall>;

export const ZTopLogprob = z.lazy(() => (z.object({
    token: z.string(),
    bytes: z.array(z.number()).nullable().optional(),
    logprob: z.number(),
})));
export type TopLogprob = z.infer<typeof ZTopLogprob>;

export const ZCustom = z.lazy(() => (z.object({
    input: z.string(),
    name: z.string(),
})));
export type Custom = z.infer<typeof ZCustom>;

export const ZResponseOutputTextParamAnnotationFileCitation = z.lazy(() => (z.object({
    file_id: z.string(),
    filename: z.string(),
    index: z.number(),
    type: z.literal("file_citation"),
})));
export type ResponseOutputTextParamAnnotationFileCitation = z.infer<typeof ZResponseOutputTextParamAnnotationFileCitation>;

export const ZResponseOutputTextParamAnnotationURLCitation = z.lazy(() => (z.object({
    end_index: z.number(),
    start_index: z.number(),
    title: z.string(),
    type: z.literal("url_citation"),
    url: z.string(),
})));
export type ResponseOutputTextParamAnnotationURLCitation = z.infer<typeof ZResponseOutputTextParamAnnotationURLCitation>;

export const ZResponseOutputTextParamAnnotationContainerFileCitation = z.lazy(() => (z.object({
    container_id: z.string(),
    end_index: z.number(),
    file_id: z.string(),
    filename: z.string(),
    start_index: z.number(),
    type: z.literal("container_file_citation"),
})));
export type ResponseOutputTextParamAnnotationContainerFileCitation = z.infer<typeof ZResponseOutputTextParamAnnotationContainerFileCitation>;

export const ZResponseOutputTextParamAnnotationFilePath = z.lazy(() => (z.object({
    file_id: z.string(),
    index: z.number(),
    type: z.literal("file_path"),
})));
export type ResponseOutputTextParamAnnotationFilePath = z.infer<typeof ZResponseOutputTextParamAnnotationFilePath>;

export const ZResponseOutputTextParamLogprob = z.lazy(() => (z.object({
    token: z.string(),
    bytes: z.array(z.number()),
    logprob: z.number(),
    top_logprobs: z.array(ZResponseOutputTextParamLogprobTopLogprob),
})));
export type ResponseOutputTextParamLogprob = z.infer<typeof ZResponseOutputTextParamLogprob>;

export const ZResponseComputerToolCallParamActionDragPath = z.lazy(() => (z.object({
    x: z.number(),
    y: z.number(),
})));
export type ResponseComputerToolCallParamActionDragPath = z.infer<typeof ZResponseComputerToolCallParamActionDragPath>;

export const ZResponseFunctionWebSearchParamActionSearchSource = z.lazy(() => (z.object({
    type: z.literal("url"),
    url: z.string(),
})));
export type ResponseFunctionWebSearchParamActionSearchSource = z.infer<typeof ZResponseFunctionWebSearchParamActionSearchSource>;

export const ZLogprobTopLogprob = z.lazy(() => (z.object({
    token: z.string(),
    bytes: z.array(z.number()),
    logprob: z.number(),
})));
export type LogprobTopLogprob = z.infer<typeof ZLogprobTopLogprob>;

export const ZChatCompletionMessageAnnotationURLCitation = z.lazy(() => (z.object({
    end_index: z.number(),
    start_index: z.number(),
    title: z.string(),
    url: z.string(),
})));
export type ChatCompletionMessageAnnotationURLCitation = z.infer<typeof ZChatCompletionMessageAnnotationURLCitation>;

export const ZChatCompletionMessageFunctionToolCallFunction = z.lazy(() => (z.object({
    arguments: z.string(),
    name: z.string(),
})));
export type ChatCompletionMessageFunctionToolCallFunction = z.infer<typeof ZChatCompletionMessageFunctionToolCallFunction>;

export const ZChatCompletionMessageCustomToolCallCustom = z.lazy(() => (z.object({
    input: z.string(),
    name: z.string(),
})));
export type ChatCompletionMessageCustomToolCallCustom = z.infer<typeof ZChatCompletionMessageCustomToolCallCustom>;

export const ZResponseOutputTextParamLogprobTopLogprob = z.lazy(() => (z.object({
    token: z.string(),
    bytes: z.array(z.number()),
    logprob: z.number(),
})));
export type ResponseOutputTextParamLogprobTopLogprob = z.infer<typeof ZResponseOutputTextParamLogprobTopLogprob>;

