export type APIKeyCreate = {
  name: string,
  description?: string | null,
};

export type APIKeyInfo = {
  id: string,
  name: string,
  created_at: Date,
  is_active: boolean,
};

export type APIKeyResponse = {
  key: string,
  name: string | null,
  created_at: Date,
  organization_id: string,
};

export type AddDomainRequest = {
  domain: string,
};

export type AlignDictsRequest = {
  list_dicts: object[],
  min_support_ratio?: number,
  reference_idx?: number | null,
};

export type AlignDictsResponse = {
  aligned_dicts: object[],
  key_mapping_all: object,
};

export type Amount = {
  value: number,
  currency: string,
};

export type AnalysesChartResponse = {
  date: Date,
  count: number,
};

export type AnnotationInput = {
  type: string,
  url_citation: OpenaiTypesChatChatCompletionMessageAnnotationURLCitation,
};

export type AnnotationOutput = {
  type: string,
  url_citation: AnnotationURLCitationOutput,
};

export type AnnotationURLCitationOutput = {
  end_index: number,
  start_index: number,
  title: string,
  url: string,
};

export type Article = {
  id?: string,
  title: string,
  type: string,
  slug: string,
  status: string,
  tags: string[],
  summary: string,
  coverImage: string,
  author_id: "louis-de-benoist" | "sacha-ichbiah" | "victor-plaisance",
  date?: string,
};

export type AttachmentMIMEDataInput = {
  filename: string,
  url: string,
  metadata?: AttachmentMetadataInput,
};

export type AttachmentMIMEDataOutput = {
  filename: string,
  url: string,
  metadata?: RetabTypesMimeAttachmentMetadata,
};

export type AttachmentMetadataInput = {
  is_inline?: boolean,
  inline_cid?: string | null,
  source?: string | null,
};

export type Audio = {
  id: string,
};

export type AutomationConfig = {
  id?: string,
  name: string,
  processor_id: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  need_validation?: boolean,
  object: string,
};

export type AutomationDecisionRequest = {
  email_data: EmailDataInput,
  schema_id: string,
};

export type AutomationDecisionResponse = {
  decision: "automated" | "human_validation",
  cosine_threshold: number,
  score_threshold: number,
  results_threshold: number,
  score_std_penalty: number,
  average_cosine: number,
  average_score: number,
  std_score: number,
  total_results: number,
};

export type AutomationLog = {
  object?: string,
  id?: string,
  user_email: string | null,
  organization_id: string,
  created_at?: Date,
  automation_snapshot: AutomationConfig,
  completion: RetabParsedChatCompletionOutput | ChatCompletionOutput,
  file_metadata: RetabTypesMimeBaseMIMEData | null,
  external_request_log: ExternalRequestLog | null,
  extraction_id?: string | null,
  api_cost: Amount | null,
  cost_breakdown: CostBreakdown | null,
};

export type Base64ImageSourceParam = {
  data: string | string,
  media_type: "image/jpeg" | "image/png" | "image/gif" | "image/webp",
  type: string,
};

export type Base64PDFSourceParam = {
  data: string | string,
  media_type: string,
  type: string,
};

export type BaseEmailData = {
  id: string,
  tree_id: string,
  subject?: string | null,
  body_plain?: string | null,
  body_html?: string | null,
  sender: EmailAddressData,
  recipients_to: EmailAddressData[],
  recipients_cc?: EmailAddressData[],
  recipients_bcc?: EmailAddressData[],
  sent_at: Date,
  received_at?: Date | null,
  in_reply_to?: string | null,
  references?: string[],
  headers?: object,
  url?: string | null,
  attachments?: MainServerServicesCustomBertCubemimedataBaseMIMEData[],
};

export type BodyConvertMsgToEmailModelV1ProcessorsAutomationsOutlookConvertMsgToEmailModelPost = {
  file: File,
};

export type BodyConvertToEmailDataAndUploadFileV1ProcessorsAutomationsOutlookConvertToEmailDataAndUploadFilePost = {
  file: File,
};

export type BodyCreateFileInternalDbFilesPost = {
  file: File,
};

export type BodyCreateFilesInternalDbFilesBatchPost = {
  files: File[],
};

export type BodyHandleEndpointProcessingV1EndpointsEndpointIdPost = {
  document: File,
};

export type BodyHandleEndpointProcessingV1ProcessorsAutomationsEndpointsProcessEndpointIdPost = {
  file: File,
  identity?: Identity,
};

export type BodyHandleLinkWebhookV1ProcessorsAutomationsLinksParseLinkIdPost = {
  file: File,
};

export type BodyImportAnnotationsCsvV1EvalsIoEvaluationIdImportAnnotationsCsvPost = {
  csv_file: File,
};

export type BodyImportAnnotationsCsvV1EvaluationsIoEvaluationIdImportAnnotationsCsvPost = {
  csv_file: File,
};

export type BodyImportDocumentsV1EvalsIoEvaluationIdImportDocumentsPost = {
  jsonl_file: File,
};

export type BodyImportDocumentsV1EvaluationsIoEvaluationIdImportDocumentsPost = {
  jsonl_file: File,
};

export type BodySubmitToProcessorV1ProcessorsProcessorIdSubmitPost = {
  document?: File | null,
  documents?: File[] | null,
  temperature?: number | null,
  stream?: boolean,
  seed?: number | null,
  store?: boolean,
  test_exception?: string | null,
};

export type BodySubmitToProcessorV1ProcessorsProcessorIdSubmitStreamPost = {
  document?: File | null,
  documents?: File[] | null,
  temperature?: number | null,
  stream?: boolean,
  seed?: number | null,
  store?: boolean,
  test_exception?: string | null,
};

export type BodyTestDocumentUploadV1ProcessorsAutomationsTestsUploadAutomationIdPost = {
  request?: DocumentUploadRequest | null,
  file?: File,
};

export type Branding = {
  button_color?: string,
  text_color?: string,
  button_text_color?: string,
  page_background_color?: string,
  logo_url?: string,
  icon_url?: string,
  company_name?: string,
  website_url?: string,
};

export type BrandingUpdateRequest = {
  button_color?: string | null,
  text_color?: string | null,
  button_text_color?: string | null,
  page_background_color?: string | null,
  logo_url?: string | null,
  icon_url?: string | null,
  company_name?: string | null,
  website_url?: string | null,
};

export type CacheControlEphemeralParam = {
  type: string,
};

export type CachePreloadRequest = {
  model?: string,
  temperature?: number,
  modality?: "text" | "image" | "native" | "image+text",
  stream?: boolean,
};

export type ChatCompletionInput = {
  id: string,
  choices: ChoiceInput[],
  created: number,
  model: string,
  object: string,
  service_tier?: "auto" | "default" | "flex" | "scale" | "priority" | null,
  system_fingerprint?: string | null,
  usage?: CompletionUsage | null,
};

export type ChatCompletionOutput = {
  id: string,
  choices: ChoiceOutput[],
  created: number,
  model: string,
  object: string,
  service_tier?: "auto" | "default" | "flex" | "scale" | "priority" | null,
  system_fingerprint?: string | null,
  usage?: CompletionUsage | null,
};

export type ChatCompletionAssistantMessageParam = {
  role: string,
  audio?: Audio | null,
  content?: string | ChatCompletionContentPartTextParam | ChatCompletionContentPartRefusalParam[] | null,
  function_call?: OpenaiTypesChatChatCompletionAssistantMessageParamFunctionCall | null,
  name?: string,
  refusal?: string | null,
  tool_calls?: ChatCompletionMessageToolCallParam[],
};

export type ChatCompletionAudio = {
  id: string,
  data: string,
  expires_at: number,
  transcript: string,
};

export type ChatCompletionContentPartImageParam = {
  image_url: ImageURL,
  type: string,
};

export type ChatCompletionContentPartInputAudioParam = {
  input_audio: InputAudio,
  type: string,
};

export type ChatCompletionContentPartRefusalParam = {
  refusal: string,
  type: string,
};

export type ChatCompletionContentPartTextParam = {
  text: string,
  type: string,
};

export type ChatCompletionDeveloperMessageParam = {
  content: string | ChatCompletionContentPartTextParam[],
  role: string,
  name?: string,
};

export type ChatCompletionFunctionMessageParam = {
  content: string | null,
  name: string,
  role: string,
};

export type ChatCompletionMessageInput = {
  content?: string | null,
  refusal?: string | null,
  role: string,
  annotations?: AnnotationInput[] | null,
  audio?: ChatCompletionAudio | null,
  function_call?: OpenaiTypesChatChatCompletionMessageFunctionCall | null,
  tool_calls?: ChatCompletionMessageToolCallInput[] | null,
};

export type ChatCompletionMessageOutput = {
  content?: string | null,
  refusal?: string | null,
  role: string,
  annotations?: AnnotationOutput[] | null,
  audio?: ChatCompletionAudio | null,
  function_call?: FunctionCallOutput | null,
  tool_calls?: ChatCompletionMessageToolCallOutput[] | null,
};

export type ChatCompletionMessageToolCallInput = {
  id: string,
  function: OpenaiTypesChatChatCompletionMessageToolCallFunction,
  type: string,
};

export type ChatCompletionMessageToolCallOutput = {
  id: string,
  function: FunctionOutput,
  type: string,
};

export type ChatCompletionMessageToolCallParam = {
  id: string,
  function: OpenaiTypesChatChatCompletionMessageToolCallParamFunction,
  type: string,
};

export type ChatCompletionRetabMessageInput = {
  role: "user" | "system" | "assistant" | "developer",
  content: string | ChatCompletionContentPartTextParam | ChatCompletionContentPartImageParam | ChatCompletionContentPartInputAudioParam | File[],
};

export type ChatCompletionRetabMessageOutput = {
  role: "user" | "system" | "assistant" | "developer",
  content: string | ChatCompletionContentPartTextParam | ChatCompletionContentPartImageParam | ChatCompletionContentPartInputAudioParam | File[],
};

export type ChatCompletionSystemMessageParam = {
  content: string | ChatCompletionContentPartTextParam[],
  role: string,
  name?: string,
};

export type ChatCompletionTokenLogprob = {
  token: string,
  bytes?: number[] | null,
  logprob: number,
  top_logprobs: TopLogprob[],
};

export type ChatCompletionToolMessageParam = {
  content: string | ChatCompletionContentPartTextParam[],
  role: string,
  tool_call_id: string,
};

export type ChatCompletionUserMessageParam = {
  content: string | ChatCompletionContentPartTextParam | ChatCompletionContentPartImageParam | ChatCompletionContentPartInputAudioParam | File[],
  role: string,
  name?: string,
};

export type ChoiceInput = {
  finish_reason: "stop" | "length" | "tool_calls" | "content_filter" | "function_call",
  index: number,
  logprobs?: ChoiceLogprobsInput | null,
  message: ChatCompletionMessageInput,
};

export type ChoiceOutput = {
  finish_reason: "stop" | "length" | "tool_calls" | "content_filter" | "function_call",
  index: number,
  logprobs?: ChoiceLogprobsOutput | null,
  message: ChatCompletionMessageOutput,
};

export type ChoiceDeltaFunctionCall = {
  arguments?: string | null,
  name?: string | null,
};

export type ChoiceDeltaToolCall = {
  index: number,
  id?: string | null,
  function?: ChoiceDeltaToolCallFunction | null,
  type?: string | null,
};

export type ChoiceDeltaToolCallFunction = {
  arguments?: string | null,
  name?: string | null,
};

export type ChoiceLogprobsInput = {
  content?: ChatCompletionTokenLogprob[] | null,
  refusal?: ChatCompletionTokenLogprob[] | null,
};

export type ChoiceLogprobsOutput = {
  content?: ChatCompletionTokenLogprob[] | null,
  refusal?: ChatCompletionTokenLogprob[] | null,
};

export type CitationCharLocation = {
  cited_text: string,
  document_index: number,
  document_title?: string | null,
  end_char_index: number,
  start_char_index: number,
  type: string,
};

export type CitationCharLocationParam = {
  cited_text: string,
  document_index: number,
  document_title: string | null,
  end_char_index: number,
  start_char_index: number,
  type: string,
};

export type CitationContentBlockLocation = {
  cited_text: string,
  document_index: number,
  document_title?: string | null,
  end_block_index: number,
  start_block_index: number,
  type: string,
};

export type CitationContentBlockLocationParam = {
  cited_text: string,
  document_index: number,
  document_title: string | null,
  end_block_index: number,
  start_block_index: number,
  type: string,
};

export type CitationPageLocation = {
  cited_text: string,
  document_index: number,
  document_title?: string | null,
  end_page_number: number,
  start_page_number: number,
  type: string,
};

export type CitationPageLocationParam = {
  cited_text: string,
  document_index: number,
  document_title: string | null,
  end_page_number: number,
  start_page_number: number,
  type: string,
};

export type CitationWebSearchResultLocationParam = {
  cited_text: string,
  encrypted_index: string,
  title: string | null,
  type: string,
  url: string,
};

export type CitationsConfigParam = {
  enabled?: boolean,
};

export type CitationsWebSearchResultLocation = {
  cited_text: string,
  encrypted_index: string,
  title?: string | null,
  type: string,
  url: string,
};

export type CodeInterpreter = {
  container: string | CodeInterpreterContainerCodeInterpreterToolAuto,
  type: string,
};

export type CodeInterpreterContainerCodeInterpreterToolAuto = {
  type: string,
  file_ids?: string[] | null,
};

export type ComparisonFilter = {
  key: string,
  type: "eq" | "ne" | "gt" | "gte" | "lt" | "lte",
  value: string | number | boolean,
};

export type ComparisonRequest = {
  dict1: object,
  dict2: object,
  metric?: "levenshtein_similarity" | "jaccard_similarity" | "hamming_similarity",
};

export type ComparisonResponse = {
  comparison_results: object,
  metric_used: "levenshtein_similarity" | "jaccard_similarity" | "hamming_similarity",
};

export type CompletionTokensDetails = {
  accepted_prediction_tokens?: number | null,
  audio_tokens?: number | null,
  reasoning_tokens?: number | null,
  rejected_prediction_tokens?: number | null,
};

export type CompletionUsage = {
  completion_tokens: number,
  prompt_tokens: number,
  total_tokens: number,
  completion_tokens_details?: CompletionTokensDetails | null,
  prompt_tokens_details?: PromptTokensDetails | null,
};

export type CompoundFilter = {
  filters: ComparisonFilter | any[],
  type: "and" | "or",
};

export type ComputeDictSimilarityRequest = {
  dict1: object,
  dict2: object,
  string_similarity_method: "levenshtein" | "jaccard" | "hamming" | "embeddings",
  min_support_ratio?: number,
};

export type ComputeDictSimilarityResponse = {
  flat_reference_elements: object,
  per_element_similarity: object,
  total_similarity: number,
  aligned_flat_reference_elements: object,
  aligned_per_element_similarity: object,
  aligned_total_similarity: number,
};

export type ComputeFieldLocationsRequest = {
  ocr_file_id: string,
  ocr_result: OCRInput,
  data: object,
  data_type: "ground_truth" | "extraction",
};

export type ComputeFieldLocationsResponse = {
  status: "success" | "error",
  message: string,
  field_locations: object,
};

export type ComputerTool = {
  display_height: number,
  display_width: number,
  environment: "windows" | "mac" | "linux" | "ubuntu" | "browser",
  type: string,
};

export type ConsensusDictRequest = {
  list_dicts: object[],
  reference_schema?: object | null,
  mode?: "direct" | "aligned",
};

export type ContentBlockSourceParam = {
  content: string | TextBlockParam | ImageBlockParam[],
  type: string,
};

export type CostBreakdown = {
  total: Amount,
  text_prompt_cost: Amount,
  text_cached_cost: Amount,
  text_completion_cost: Amount,
  text_total_cost: Amount,
  audio_prompt_cost?: Amount | null,
  audio_completion_cost?: Amount | null,
  audio_total_cost?: Amount | null,
  token_counts: TokenCounts,
  model: string,
  is_fine_tuned?: boolean,
};

export type CreateAndLinkOrganizationRequest = {
  organization_name: string,
};

export type CreateEvaluation = {
  name: string,
  json_schema: object,
  project_id?: string,
  default_inference_settings?: InferenceSettings,
};

export type CreateOrganizationResponse = {
  success: boolean,
  workos_organization: Organization,
};

export type CreateSpreadsheetWithStoredTokenRequest = {
  spreadsheet_name: string,
};

export type Credits = {
  credits: number,
};

export type CreditsDataPoint = {
  date: string,
  credits: number,
};

export type CreditsTimeSeries = {
  data: CreditsDataPoint[],
};

export type CustomDomain = {
  id: string,
  domain: string,
  status: "active" | "pending" | "active_redeploying" | "moved" | "pending_deletion" | "deleted" | "pending_blocked" | "pending_migration" | "pending_provisioned" | "test_pending" | "test_active" | "test_active_apex" | "test_blocked" | "test_failed" | "provisioned" | "blocked" | null,
  default?: boolean,
};

export type DBFile = {
  object?: string,
  id: string,
  filename: string,
};

export type DataPoint = {
  date: string,
  value: number,
};

export type DetachProcessorRequest = {
  new_processor_name: string,
};

export type DisplayMetadata = {
  url: string,
  type: "image" | "pdf" | "txt",
};

export type DistancesResult = {
  distances: object,
  mean_distance: number,
  metric_type: "levenshtein" | "jaccard" | "hamming",
};

export type DocumentBlockParam = {
  source: Base64PDFSourceParam | PlainTextSourceParam | ContentBlockSourceParam | URLPDFSourceParam,
  type: string,
  cache_control?: CacheControlEphemeralParam | null,
  citations?: CitationsConfigParam,
  context?: string | null,
  title?: string | null,
};

export type DocumentCreateInputRequest = {
  document: MIMEDataInput,
  modality: "text" | "image" | "native" | "image+text",
  image_resolution_dpi?: number,
  browser_canvas?: "A3" | "A4" | "A5",
  json_schema: object,
};

export type DocumentCreateMessageRequest = {
  document: MIMEDataInput,
  modality: "text" | "image" | "native" | "image+text",
  image_resolution_dpi?: number,
  browser_canvas?: "A3" | "A4" | "A5",
};

export type DocumentItem = {
  mime_data: MIMEDataInput,
  annotation?: object,
  annotation_metadata?: PredictionMetadata | null,
};

export type DocumentMessage = {
  id: string,
  object?: string,
  messages: ChatCompletionRetabMessageOutput[],
  created: number,
  modality: "text" | "image" | "native" | "image+text",
  token_count: TokenCount,
};

export type DocumentStatus = {
  document_id: string,
  filename: string,
  needs_update: boolean,
  has_prediction: boolean,
  prediction_updated_at: Date | null,
  iteration_updated_at: Date,
};

export type DocumentTransformRequest = {
  document: MIMEDataInput,
};

export type DocumentTransformResponse = {
  document: RetabTypesMimeMIMEData,
};

export type DocumentUploadRequest = {
  document: MIMEDataInput,
};

export type DuplicateEvaluationRequest = {
  project_id?: string | null,
  name?: string | null,
};

export type EasyInputMessage = {
  content: string | ResponseInputText | ResponseInputImage | ResponseInputFile[],
  role: "user" | "assistant" | "system" | "developer",
  type?: string | null,
};

export type EasyInputMessageParam = {
  content: string | ResponseInputTextParam | ResponseInputImageParam | ResponseInputFileParam[],
  role: "user" | "assistant" | "system" | "developer",
  type?: string,
};

export type EmailAddressData = {
  email: string,
  display_name?: string | null,
};

export type EmailConversionRequest = {
  bytes: string,
};

export type EmailDataInput = {
  id: string,
  tree_id: string,
  subject?: string | null,
  body_plain?: string | null,
  body_html?: string | null,
  sender: EmailAddressData,
  recipients_to: EmailAddressData[],
  recipients_cc?: EmailAddressData[],
  recipients_bcc?: EmailAddressData[],
  sent_at: Date,
  received_at?: Date | null,
  in_reply_to?: string | null,
  references?: string[],
  headers?: object,
  url?: string | null,
  attachments?: AttachmentMIMEDataInput[],
};

export type EmailDataOutput = {
  id: string,
  tree_id: string,
  subject?: string | null,
  body_plain?: string | null,
  body_html?: string | null,
  sender: EmailAddressData,
  recipients_to: EmailAddressData[],
  recipients_cc?: EmailAddressData[],
  recipients_bcc?: EmailAddressData[],
  sent_at: Date,
  received_at?: Date | null,
  in_reply_to?: string | null,
  references?: string[],
  headers?: object,
  url?: string | null,
  attachments?: AttachmentMIMEDataOutput[],
};

export type EmailExtractRequest = {
  automation_id: string,
  email_data: EmailDataInput,
  modality?: "text" | "image" | "native" | "image+text" | null,
  image_resolution_dpi?: number | null,
  browser_canvas?: "A3" | "A4" | "A5" | null,
  model?: string | null,
  json_schema?: object | null,
  temperature?: number | null,
  n_consensus?: number | null,
  stream?: boolean,
  seed?: number | null,
  store?: boolean,
};

export type EndpointInput = {
  id?: string,
  name: string,
  processor_id: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  need_validation?: boolean,
};

export type EndpointOutput = {
  id?: string,
  name: string,
  processor_id: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  need_validation?: boolean,
  object: string,
};

export type EnhanceSchemaConfig = {
  allow_reasoning_fields_added?: boolean,
  allow_field_description_update?: boolean,
  allow_system_prompt_update?: boolean,
  allow_field_simple_type_change?: boolean,
  allow_field_data_structure_breakdown?: boolean,
};

export type EnhanceSchemaRequest = {
  documents: MIMEDataInput[],
  ground_truths?: object[] | null,
  model?: string,
  temperature?: number,
  reasoning_effort?: "low" | "medium" | "high" | null,
  modality: "text" | "image" | "native" | "image+text",
  image_resolution_dpi?: number,
  browser_canvas?: "A3" | "A4" | "A5",
  stream?: boolean,
  tools_config?: EnhanceSchemaConfig,
  json_schema: object,
  instructions?: string | null,
  flat_likelihoods?: object[] | object | null,
};

export type ErrorDetail = {
  code: string,
  message: string,
  details?: object | null,
};

export type EvaluateSchemaRequest = {
  documents: MIMEDataInput[],
  ground_truths?: object[] | null,
  model?: string,
  reasoning_effort?: "low" | "medium" | "high" | null,
  modality: "text" | "image" | "native" | "image+text",
  image_resolution_dpi?: number,
  browser_canvas?: "A3" | "A4" | "A5",
  n_consensus?: number,
  json_schema: object,
};

export type EvaluateSchemaResponse = {
  item_metrics: ItemMetric[],
};

export type EvaluationDocumentInput = {
  mime_data: MIMEDataInput,
  annotation?: object,
  annotation_metadata?: PredictionMetadata | null,
  id: string,
};

export type EvaluationDocumentOutput = {
  mime_data: RetabTypesMimeMIMEData,
  annotation?: object,
  annotation_metadata?: PredictionMetadata | null,
  id: string,
};

export type Event = {
  object?: string,
  id?: string,
  event: string,
  created_at?: Date,
  data: object,
  metadata?: object | null,
};

export type ExportToCsvRequest = {
  json_data: any,
  json_schema: object,
  delimiter?: string,
  line_delimiter?: string,
  quote?: string,
};

export type ExternalAPIKey = {
  provider: "OpenAI" | "Anthropic" | "Gemini" | "xAI" | "Retab",
  is_configured: boolean,
  last_updated: Date | null,
};

export type ExternalAPIKeyRequest = {
  provider: "OpenAI" | "Anthropic" | "Gemini" | "xAI" | "Retab",
  api_key: string,
};

export type ExternalRequestLog = {
  webhook_url: string | null,
  request_body: object,
  request_headers: object,
  request_at: Date,
  response_body: object,
  response_headers: object,
  response_at: Date,
  status_code: number,
  error?: string | null,
  duration_ms: number,
};

export type Extraction = {
  id?: string,
  messages?: ChatCompletionRetabMessageOutput[],
  messages_gcs: string,
  file_gcs_paths: string[],
  file_ids: string[],
  file_gcs?: string,
  file_id?: string,
  status: "success" | "failed",
  completion: RetabParsedChatCompletionOutput | ChatCompletionOutput,
  json_schema: any,
  model: string,
  temperature?: number,
  source: ExtractionSource,
  image_resolution_dpi?: number,
  browser_canvas?: "A3" | "A4" | "A5",
  modality?: "text" | "image" | "native" | "image+text",
  reasoning_effort?: "low" | "medium" | "high" | null,
  n_consensus?: number,
  timings?: ExtractionTimingStep[],
  schema_id: string,
  schema_data_id: string,
  created_at?: Date,
  request_at?: Date | null,
  organization_id: string,
  validation_state?: "pending" | "validated" | "invalid" | null,
  billed?: boolean,
  api_cost: Amount | null,
  cost_breakdown: CostBreakdown | null,
};

export type ExtractionCount = {
  total: number,
};

export type ExtractionSource = {
  type: "api" | "annotation" | "processor" | "automation" | "automation.link" | "automation.mailbox" | "automation.cron" | "automation.outlook" | "automation.endpoint" | "schema.extract",
  id?: string | null,
};

export type ExtractionTimingStep = {
  name: string | "initialization" | "prepare_messages" | "yield_first_token" | "completion",
  duration: number,
  notes?: string | null,
};

export type FetchParams = {
  endpoint: string,
  headers: object,
  name: string,
};

export type FieldLocation = {
  label: string,
  value: string,
  quote: string,
  file_id?: string | null,
  page?: number | null,
  bbox_normalized?: [number, number, number, number] | null,
  score?: number | null,
  match_level?: "token" | "line" | "block" | null,
};

export type FieldLocationsResult = {
  choices: object[],
};

export type File = {
  file: FileFile,
  type: string,
};

export type FileFile = {
  file_data?: string,
  file_id?: string,
  filename?: string,
};

export type FileLink = {
  download_url: string,
  expires_in: string,
  filename: string,
};

export type FileScoreIndex = {
  file_id: string,
  created_at: Date,
  file_embedding: number[],
  llm_output: object,
  hil_output: object,
  schema_id: string,
  levenshtein_similarity: object,
  jaccard_similarity: object,
  hamming_similarity: object,
  schema_data_id: string,
  organization_id: string,
};

export type FileSearchTool = {
  type: string,
  vector_store_ids: string[],
  filters?: ComparisonFilter | CompoundFilter | null,
  max_num_results?: number | null,
  ranking_options?: RankingOptions | null,
};

export type FinetunedModel = {
  object?: string,
  organization_id: string,
  model: string,
  schema_id: string,
  schema_data_id: string,
  finetuning_props: InferenceSettings,
  evaluation_id?: string | null,
  created_at?: Date,
};

export type FreightDocumentAnalysisResponse = {
  analyses: FreightEmailAnalysisAPIResponse[],
  has_more: boolean,
};

export type FreightEmailAnalysisAPIResponse = {
  id: string,
  is_demo?: boolean,
  extraction_source: "mailbox" | "plugin",
  extraction_status: "success" | "pending" | "failed" | "sent_to_tms" | "need_review",
  user?: MainServerServicesCustomBertfakeRoutesUser | null,
  request_at: Date,
  extraction_type: "RoadBookingConfirmation" | "RoadTransportOrder" | "AirBookingConfirmation" | "RoadBookingConfirmationBert" | "RoadBookingConfirmationGroussard" | "RoadBookingConfirmationJourdan" | "RoadBookingConfirmationMGE" | "RoadBookingConfirmationSuus" | "RoadBookingConfirmationThevenon" | "RoadQuoteRequest" | "RoadPickupMazet" | "RoadCMR",
  extraction: any,
  uncertainties: any,
  mappings: object,
  action_type: "Creation" | "Modification" | "Deletion",
  documents: MainServerServicesCustomBertCubemimedataBaseMIMEData | MainServerServicesCustomBertCubemimedataMIMEData[],
  email_data: BaseEmailData,
};

export type FunctionOutput = {
  arguments: string,
  name: string,
};

export type FunctionCallOutput = {
  arguments: string,
  name: string,
};

export type FunctionTool = {
  name: string,
  parameters?: object | null,
  strict?: boolean | null,
  type: string,
  description?: string | null,
};

export type GenerateSchemaRequest = {
  documents: MIMEDataInput[],
  model?: string,
  temperature?: number,
  reasoning_effort?: "low" | "medium" | "high" | null,
  modality: "text" | "image" | "native" | "image+text",
  instructions?: string | null,
  image_resolution_dpi?: number,
  browser_canvas?: "A3" | "A4" | "A5",
  stream?: boolean,
};

export type GenerateSystemPromptRequest = {
  documents: MIMEDataInput[],
  model?: string,
  temperature?: number,
  reasoning_effort?: "low" | "medium" | "high" | null,
  modality: "text" | "image" | "native" | "image+text",
  instructions?: string | null,
  image_resolution_dpi?: number,
  browser_canvas?: "A3" | "A4" | "A5",
  stream?: boolean,
  json_schema: object,
};

export type GoogleSpreadsheet = {
  id: string,
  name: string,
  url?: string | null,
};

export type GoogleWorksheet = {
  id: number,
  title: string,
};

export type HTTPValidationError = {
  detail?: ValidationError[],
};

export type Identity = {
  user_id: string,
  organization_id?: string | null,
  tier?: number,
  auth_method?: "api_key" | "bearer_token" | "master_key" | "outlook_auth",
};

export type ImageBlockParam = {
  source: Base64ImageSourceParam | URLImageSourceParam,
  type: string,
  cache_control?: CacheControlEphemeralParam | null,
};

export type ImageGeneration = {
  type: string,
  background?: "transparent" | "opaque" | "auto" | null,
  input_image_mask?: ImageGenerationInputImageMask | null,
  model?: string | null,
  moderation?: "auto" | "low" | null,
  output_compression?: number | null,
  output_format?: "png" | "webp" | "jpeg" | null,
  partial_images?: number | null,
  quality?: "low" | "medium" | "high" | "auto" | null,
  size?: "1024x1024" | "1024x1536" | "1536x1024" | "auto" | null,
};

export type ImageGenerationInputImageMask = {
  file_id?: string | null,
  image_url?: string | null,
};

export type ImageURL = {
  url: string,
  detail?: "auto" | "low" | "high",
};

export type ImportAnnotationsCsvResponse = {
  success: boolean,
  evaluation_id: string,
  file_data: object,
  total_files: number,
  message: string,
};

export type IncompleteDetails = {
  reason?: "max_output_tokens" | "content_filter" | null,
};

export type InferenceSettings = {
  model?: string,
  temperature?: number,
  modality?: "text" | "image" | "native" | "image+text",
  reasoning_effort?: "low" | "medium" | "high" | null,
  image_resolution_dpi?: number,
  browser_canvas?: "A3" | "A4" | "A5",
  n_consensus?: number,
};

export type InputAudio = {
  data: string,
  format: "wav" | "mp3",
};

export type InputTokensDetails = {
  cached_tokens: number,
};

export type ItemMetric = {
  id: string,
  name: string,
  similarity: number,
  similarities: object,
  flat_similarities: object,
  aligned_similarity: number,
  aligned_similarities: object,
  aligned_flat_similarities: object,
};

export type IterationDocumentStatusResponse = {
  iteration_id: string,
  documents: DocumentStatus[],
  total_documents: number,
  documents_needing_update: number,
  documents_up_to_date: number,
};

export type JSONSchema = {
  name: string,
  description?: string,
  schema?: object,
  strict?: boolean | null,
};

export type KeyValidationResponse = {
  is_valid: boolean,
  message: string,
};

export type LLMAnnotateDocumentRequest = {
  stream?: boolean,
};

export type LinkInput = {
  id?: string,
  name: string,
  processor_id: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  need_validation?: boolean,
  password?: string | null,
};

export type LinkOutput = {
  id?: string,
  name: string,
  processor_id: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  need_validation?: boolean,
  password?: string | null,
  object: string,
};

export type ListAutomations = {
  data: AutomationConfig[],
  list_metadata: ListMetadata,
};

export type ListDomainsResponse = {
  domains: CustomDomain[],
};

export type ListEndpoints = {
  data: EndpointOutput[],
  list_metadata: ListMetadata,
};

export type ListEvaluationDocumentsResponse = {
  data: EvaluationDocumentOutput[],
};

export type ListFiles = {
  data: StoredDBFile[],
  list_metadata: ListMetadata,
};

export type ListFinetunedModels = {
  data: FinetunedModel[],
  list_metadata: ListMetadata,
};

export type ListLinks = {
  data: LinkOutput[],
  list_metadata: ListMetadata,
};

export type ListLogs = {
  data: AutomationLog[],
  list_metadata: ListMetadata,
};

export type ListMetadata = {
  before: string | null,
  after: string | null,
};

export type ListTemplates = {
  data: TemplateSchema[],
  list_metadata: ListMetadata,
};

export type LocalShell = {
  type: string,
};

export type LogExtractionRequest = {
  messages?: ChatCompletionRetabMessageInput[] | null,
  openai_messages?: ChatCompletionDeveloperMessageParam | ChatCompletionSystemMessageParam | ChatCompletionUserMessageParam | ChatCompletionAssistantMessageParam | ChatCompletionToolMessageParam | ChatCompletionFunctionMessageParam[] | null,
  openai_responses_input?: EasyInputMessageParam | OpenaiTypesResponsesResponseInputParamMessage | ResponseOutputMessageParam | ResponseFileSearchToolCallParam | ResponseComputerToolCallParam | OpenaiTypesResponsesResponseInputParamComputerCallOutput | ResponseFunctionWebSearchParam | ResponseFunctionToolCallParam | OpenaiTypesResponsesResponseInputParamFunctionCallOutput | ResponseReasoningItemParam | OpenaiTypesResponsesResponseInputParamImageGenerationCall | ResponseCodeInterpreterToolCallParam | OpenaiTypesResponsesResponseInputParamLocalShellCall | OpenaiTypesResponsesResponseInputParamLocalShellCallOutput | OpenaiTypesResponsesResponseInputParamMcpListTools | OpenaiTypesResponsesResponseInputParamMcpApprovalRequest | OpenaiTypesResponsesResponseInputParamMcpApprovalResponse | OpenaiTypesResponsesResponseInputParamMcpCall | OpenaiTypesResponsesResponseInputParamItemReference[] | null,
  anthropic_messages?: MessageParam[] | null,
  anthropic_system_prompt?: string | null,
  document?: MIMEDataInput,
  completion?: object | RetabParsedChatCompletionInput | AnthropicTypesMessageMessage | ParsedChatCompletion | ChatCompletionInput | null,
  openai_responses_output?: Response | null,
  json_schema: object,
  model: string,
  temperature: number,
};

export type LogExtractionResponse = {
  extraction_id?: string | null,
  status: "success" | "error",
  error_message?: string | null,
};

export type MIMEDataInput = {
  filename: string,
  url: string,
};

export type MailboxInput = {
  id?: string,
  name: string,
  processor_id: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  need_validation?: boolean,
  email: string,
  authorized_domains?: string[],
  authorized_emails?: string[],
};

export type MailboxOutput = {
  id?: string,
  name: string,
  processor_id: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  need_validation?: boolean,
  email: string,
  authorized_domains?: string[],
  authorized_emails?: string[],
  object: string,
};

export type MappingObject = {
  internal_code: string | null,
  extracted_object?: any,
  mapped_object?: any,
};

export type MatchParams = {
  endpoint: string,
  headers: object,
  path: string,
};

export type MatchResultModel = {
  record: object,
  similarity: number,
};

export type Matrix = {
  rows: number,
  cols: number,
  type_: number,
  data: string,
};

export type Mcp = {
  server_label: string,
  server_url: string,
  type: string,
  allowed_tools?: string[] | McpAllowedToolsMcpAllowedToolsFilter | null,
  headers?: object | null,
  require_approval?: McpRequireApprovalMcpToolApprovalFilter | "always" | "never" | null,
};

export type McpAllowedToolsMcpAllowedToolsFilter = {
  tool_names?: string[] | null,
};

export type McpRequireApprovalMcpToolApprovalFilter = {
  always?: McpRequireApprovalMcpToolApprovalFilterAlways | null,
  never?: McpRequireApprovalMcpToolApprovalFilterNever | null,
};

export type McpRequireApprovalMcpToolApprovalFilterAlways = {
  tool_names?: string[] | null,
};

export type McpRequireApprovalMcpToolApprovalFilterNever = {
  tool_names?: string[] | null,
};

export type MessageParam = {
  content: string | TextBlockParam | ImageBlockParam | DocumentBlockParam | ThinkingBlockParam | RedactedThinkingBlockParam | ToolUseBlockParam | ToolResultBlockParam | ServerToolUseBlockParam | WebSearchToolResultBlockParam | TextBlock | ThinkingBlock | RedactedThinkingBlock | ToolUseBlock | ServerToolUseBlock | WebSearchToolResultBlock[],
  role: "user" | "assistant",
};

export type MetricResult = {
  item_metrics: ItemMetric[],
  mean_similarity: number,
  aligned_mean_similarity: number,
  metric_type: "levenshtein" | "jaccard" | "hamming",
};

export type Model = {
  id: string,
  created: number,
  object: string,
  owned_by: string,
};

export type ModelCapabilities = {
  modalities: "text" | "audio" | "image"[],
  endpoints: "chat_completions" | "responses" | "assistants" | "batch" | "fine_tuning" | "embeddings" | "speech_generation" | "translation" | "completions_legacy" | "image_generation" | "transcription" | "moderation" | "realtime"[],
  features: "streaming" | "function_calling" | "structured_outputs" | "distillation" | "fine_tuning" | "predicted_outputs" | "schema_generation"[],
};

export type ModelCard = {
  model: "gpt-4o" | "gpt-4o-mini" | "chatgpt-4o-latest" | "gpt-4.1" | "gpt-4.1-mini" | "gpt-4.1-mini-2025-04-14" | "gpt-4.1-2025-04-14" | "gpt-4.1-nano" | "gpt-4.1-nano-2025-04-14" | "gpt-4o-2024-11-20" | "gpt-4o-2024-08-06" | "gpt-4o-2024-05-13" | "gpt-4o-mini-2024-07-18" | "o1" | "o1-2024-12-17" | "o3" | "o3-2025-04-16" | "o4-mini" | "o4-mini-2025-04-16" | "gpt-4o-audio-preview-2024-12-17" | "gpt-4o-audio-preview-2024-10-01" | "gpt-4o-realtime-preview-2024-12-17" | "gpt-4o-realtime-preview-2024-10-01" | "gpt-4o-mini-audio-preview-2024-12-17" | "gpt-4o-mini-realtime-preview-2024-12-17" | "claude-3-5-sonnet-latest" | "claude-3-5-sonnet-20241022" | "claude-3-opus-20240229" | "claude-3-sonnet-20240229" | "claude-3-haiku-20240307" | "grok-3" | "grok-3-mini" | "gemini-2.5-pro" | "gemini-2.5-flash" | "gemini-2.5-pro-preview-06-05" | "gemini-2.5-pro-preview-05-06" | "gemini-2.5-pro-preview-03-25" | "gemini-2.5-flash-preview-05-20" | "gemini-2.5-flash-preview-04-17" | "gemini-2.5-flash-lite-preview-06-17" | "gemini-2.5-pro-exp-03-25" | "gemini-2.0-flash-lite" | "gemini-2.0-flash" | "auto-large" | "auto-small" | "auto-micro" | "human" | string,
  pricing: Pricing,
  capabilities: ModelCapabilities,
  temperature_support?: boolean,
  reasoning_effort_support?: boolean,
  permissions?: ModelCardPermissions,
  is_finetuned: boolean,
  model_credit_usage_per_page: number,
};

export type ModelCardPermissions = {
  show_in_free_picker?: boolean,
  show_in_paid_picker?: boolean,
};

export type ModelCardsResponse = {
  data: ModelCard[],
  object?: string,
};

export type ModelsResponse = {
  data: Model[],
  object?: string,
};

export type MonthlyUsageResponseContent = {
  credits_count: number,
};

export type MultipleUploadResponse = {
  files: DBFile[],
};

export type OCRInput = {
  pages: PageInput[],
};

export type OCRMetadata = {
  result?: RetabTypesMimeOCROutput | null,
  file_gcs?: string | null,
  file_page_count?: number | null,
};

export type OpenAIRateLimits = {
  limit_requests?: number,
  limit_tokens?: number,
  remaining_requests?: number,
  remaining_tokens?: number,
  reset_requests?: string,
  reset_tokens?: string,
};

export type OpenAITierResponse = {
  rate_limits: OpenAIRateLimits,
};

export type OptimizedDocumentMetrics = {
  document_id: string,
  filename: string,
  true_positives: object[],
  true_negatives: object[],
  false_positives: object[],
  false_negatives: object[],
  mismatched_values: object[],
  field_similarities: object,
};

export type OptimizedIterationMetrics = {
  overall_metrics: OptimizedOverallMetrics,
  document_metrics: OptimizedDocumentMetrics[],
};

export type OptimizedOverallMetrics = {
  accuracy: number,
  similarity: number,
  total_error_rate: number,
  true_positive_rate: number,
  true_negative_rate: number,
  false_positive_rate: number,
  false_negative_rate: number,
  mismatched_value_rate: number,
  accuracy_per_field: object,
  similarity_per_field: object,
  total_documents: number,
  total_fields_compared: number,
};

export type Organization = {
  id: string,
  object: string,
  name: string,
  domains: OrganizationDomain[],
  created_at: string,
  updated_at: string,
  allow_profiles_outside_organization: boolean,
  stripe_customer_id?: string | null,
  external_id?: string | null,
  metadata?: object,
};

export type OrganizationDomain = {
  id: string,
  organization_id: string,
  object: string,
  domain: string,
  state?: "failed" | "pending" | "legacy_verified" | "verified" | null,
  verification_strategy?: "manual" | "dns" | null,
  verification_token?: string | null,
};

export type OutlookInput = {
  id?: string,
  name: string,
  processor_id: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  need_validation?: boolean,
  authorized_domains?: string[],
  authorized_emails?: string[],
  layout_schema?: object | null,
  match_params?: MatchParams[],
  fetch_params?: FetchParams[],
};

export type OutlookOutput = {
  id?: string,
  name: string,
  processor_id: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  need_validation?: boolean,
  authorized_domains?: string[],
  authorized_emails?: string[],
  layout_schema?: object | null,
  match_params?: MatchParams[],
  fetch_params?: FetchParams[],
  object: string,
};

export type OutlookSubmitRequest = {
  email_data: EmailDataInput,
  completion: RetabParsedChatCompletionInput,
  user_email: string,
  metadata: object,
  store?: boolean,
};

export type OutputTokensDetails = {
  reasoning_tokens: number,
};

export type PageInput = {
  page_number: number,
  width: number,
  height: number,
  unit?: string,
  blocks: TextBox[],
  lines: TextBox[],
  tokens: TextBox[],
  transforms?: Matrix[],
};

export type PaginatedList = {
  data: any[],
  list_metadata: ListMetadata,
};

export type ParseRequest = {
  document: MIMEDataInput,
  model?: "gpt-4o" | "gpt-4o-mini" | "chatgpt-4o-latest" | "gpt-4.1" | "gpt-4.1-mini" | "gpt-4.1-mini-2025-04-14" | "gpt-4.1-2025-04-14" | "gpt-4.1-nano" | "gpt-4.1-nano-2025-04-14" | "gpt-4o-2024-11-20" | "gpt-4o-2024-08-06" | "gpt-4o-2024-05-13" | "gpt-4o-mini-2024-07-18" | "o1" | "o1-2024-12-17" | "o3" | "o3-2025-04-16" | "o4-mini" | "o4-mini-2025-04-16" | "gpt-4o-audio-preview-2024-12-17" | "gpt-4o-audio-preview-2024-10-01" | "gpt-4o-realtime-preview-2024-12-17" | "gpt-4o-realtime-preview-2024-10-01" | "gpt-4o-mini-audio-preview-2024-12-17" | "gpt-4o-mini-realtime-preview-2024-12-17" | "claude-3-5-sonnet-latest" | "claude-3-5-sonnet-20241022" | "claude-3-opus-20240229" | "claude-3-sonnet-20240229" | "claude-3-haiku-20240307" | "grok-3" | "grok-3-mini" | "gemini-2.5-pro" | "gemini-2.5-flash" | "gemini-2.5-pro-preview-06-05" | "gemini-2.5-pro-preview-05-06" | "gemini-2.5-pro-preview-03-25" | "gemini-2.5-flash-preview-05-20" | "gemini-2.5-flash-preview-04-17" | "gemini-2.5-flash-lite-preview-06-17" | "gemini-2.5-pro-exp-03-25" | "gemini-2.0-flash-lite" | "gemini-2.0-flash" | "auto-large" | "auto-small" | "auto-micro" | "human",
  table_parsing_format?: "markdown" | "yaml" | "html" | "json",
  image_resolution_dpi?: number,
  browser_canvas?: "A3" | "A4" | "A5",
};

export type ParseResult = {
  document: RetabTypesMimeBaseMIMEData,
  usage: RetabUsage,
  pages: string[],
  text: string,
};

export type ParsedChatCompletion = {
  id: string,
  choices: ParsedChoice[],
  created: number,
  model: string,
  object: string,
  service_tier?: "auto" | "default" | "flex" | "scale" | "priority" | null,
  system_fingerprint?: string | null,
  usage?: CompletionUsage | null,
};

export type ParsedChatCompletionMessageInput = {
  content?: string | null,
  refusal?: string | null,
  role: string,
  annotations?: AnnotationInput[] | null,
  audio?: ChatCompletionAudio | null,
  function_call?: OpenaiTypesChatChatCompletionMessageFunctionCall | null,
  tool_calls?: ParsedFunctionToolCall[] | null,
  parsed?: any | null,
};

export type ParsedChatCompletionMessageOutput = {
  content?: string | null,
  refusal?: string | null,
  role: string,
  annotations?: AnnotationOutput[] | null,
  audio?: ChatCompletionAudio | null,
  function_call?: FunctionCallOutput | null,
  tool_calls?: ParsedFunctionToolCall[] | null,
  parsed?: any | null,
};

export type ParsedChoice = {
  finish_reason: "stop" | "length" | "tool_calls" | "content_filter" | "function_call",
  index: number,
  logprobs?: ChoiceLogprobsInput | null,
  message: ParsedChatCompletionMessageInput,
};

export type ParsedFunction = {
  arguments: string,
  name: string,
  parsed_arguments?: any | null,
};

export type ParsedFunctionToolCall = {
  id: string,
  function: ParsedFunction,
  type: string,
};

export type PartialSchema = {
  object?: string,
  created_at?: Date,
  json_schema?: object,
  strict?: boolean,
};

export type PatchEvaluationDocumentRequest = {
  annotation?: object | null,
  annotation_metadata?: PredictionMetadata | null,
};

export type PatchIterationDocumentPredictionRequest = {
  metadata: PredictionMetadata,
};

export type PatchIterationRequest = {
  inference_settings?: InferenceSettings | null,
  json_schema?: object | null,
};

export type PerformOCROnlyRequest = {
  extraction_id: string,
};

export type PerformOCROnlyResponse = {
  status: "success" | "error",
  message: string,
  extraction_id: string,
  ocr_file_id: string,
  ocr_file_url: string,
  ocr_result?: RetabTypesMimeOCROutput | null,
};

export type PerformOCRRequest = {
  extraction_id: string,
};

export type PerformOCRResponse = {
  status: "success" | "error",
  message: string,
  extraction_id: string,
  ocr_file_id: string,
  ocr_file_url: string,
  field_locations_result?: FieldLocationsResult | null,
};

export type PlainTextSourceParam = {
  data: string,
  media_type: string,
  type: string,
};

export type Point = {
  x: number,
  y: number,
};

export type PredictionMetadata = {
  extraction_id?: string | null,
  likelihoods?: object | null,
  field_locations?: object | null,
  agentic_field_locations?: object | null,
  consensus_details?: object[] | null,
  api_cost?: Amount | null,
};

export type PreprocessingLogResponse = {
  id: string,
  credits_count: number,
  page_count?: number | null,
  filename: string,
  operation: string,
};

export type Pricing = {
  text: TokenPrice,
  audio?: TokenPrice | null,
  ft_price_hike?: number,
};

export type ProcessIterationDocument = {
  stream?: boolean,
};

export type ProcessIterationRequest = {
  document_ids?: string[] | null,
  only_outdated?: boolean,
};

export type ProcessorConfig = {
  object?: string,
  id?: string,
  updated_at?: Date,
  name: string,
  modality: "text" | "image" | "native" | "image+text",
  image_resolution_dpi?: number,
  browser_canvas?: "A3" | "A4" | "A5",
  model: string,
  json_schema: object,
  temperature?: number,
  reasoning_effort?: "low" | "medium" | "high" | null,
  n_consensus?: number,
};

export type Project = {
  id?: string,
  name: string,
  updated_at?: Date,
};

export type PromptTokensDetails = {
  audio_tokens?: number | null,
  cached_tokens?: number | null,
};

export type RankingOptions = {
  ranker?: "auto" | "default-2024-11-15" | null,
  score_threshold?: number | null,
};

export type ReconciliationResponse = {
  consensus_dict: object,
  likelihoods: object,
};

export type RedactedThinkingBlock = {
  data: string,
  type: string,
};

export type RedactedThinkingBlockParam = {
  data: string,
  type: string,
};

export type Response = {
  id: string,
  created_at: number,
  error?: ResponseError | null,
  incomplete_details?: IncompleteDetails | null,
  instructions?: string | EasyInputMessage | OpenaiTypesResponsesResponseInputItemMessage | ResponseOutputMessage | ResponseFileSearchToolCall | ResponseComputerToolCall | OpenaiTypesResponsesResponseInputItemComputerCallOutput | ResponseFunctionWebSearch | ResponseFunctionToolCall | OpenaiTypesResponsesResponseInputItemFunctionCallOutput | ResponseReasoningItem | OpenaiTypesResponsesResponseInputItemImageGenerationCall | ResponseCodeInterpreterToolCall | OpenaiTypesResponsesResponseInputItemLocalShellCall | OpenaiTypesResponsesResponseInputItemLocalShellCallOutput | OpenaiTypesResponsesResponseInputItemMcpListTools | OpenaiTypesResponsesResponseInputItemMcpApprovalRequest | OpenaiTypesResponsesResponseInputItemMcpApprovalResponse | OpenaiTypesResponsesResponseInputItemMcpCall | OpenaiTypesResponsesResponseInputItemItemReference[] | null,
  metadata?: object | null,
  model: string | "gpt-4.1" | "gpt-4.1-mini" | "gpt-4.1-nano" | "gpt-4.1-2025-04-14" | "gpt-4.1-mini-2025-04-14" | "gpt-4.1-nano-2025-04-14" | "o4-mini" | "o4-mini-2025-04-16" | "o3" | "o3-2025-04-16" | "o3-mini" | "o3-mini-2025-01-31" | "o1" | "o1-2024-12-17" | "o1-preview" | "o1-preview-2024-09-12" | "o1-mini" | "o1-mini-2024-09-12" | "gpt-4o" | "gpt-4o-2024-11-20" | "gpt-4o-2024-08-06" | "gpt-4o-2024-05-13" | "gpt-4o-audio-preview" | "gpt-4o-audio-preview-2024-10-01" | "gpt-4o-audio-preview-2024-12-17" | "gpt-4o-audio-preview-2025-06-03" | "gpt-4o-mini-audio-preview" | "gpt-4o-mini-audio-preview-2024-12-17" | "gpt-4o-search-preview" | "gpt-4o-mini-search-preview" | "gpt-4o-search-preview-2025-03-11" | "gpt-4o-mini-search-preview-2025-03-11" | "chatgpt-4o-latest" | "codex-mini-latest" | "gpt-4o-mini" | "gpt-4o-mini-2024-07-18" | "gpt-4-turbo" | "gpt-4-turbo-2024-04-09" | "gpt-4-0125-preview" | "gpt-4-turbo-preview" | "gpt-4-1106-preview" | "gpt-4-vision-preview" | "gpt-4" | "gpt-4-0314" | "gpt-4-0613" | "gpt-4-32k" | "gpt-4-32k-0314" | "gpt-4-32k-0613" | "gpt-3.5-turbo" | "gpt-3.5-turbo-16k" | "gpt-3.5-turbo-0301" | "gpt-3.5-turbo-0613" | "gpt-3.5-turbo-1106" | "gpt-3.5-turbo-0125" | "gpt-3.5-turbo-16k-0613" | "o1-pro" | "o1-pro-2025-03-19" | "o3-pro" | "o3-pro-2025-06-10" | "o3-deep-research" | "o3-deep-research-2025-06-26" | "o4-mini-deep-research" | "o4-mini-deep-research-2025-06-26" | "computer-use-preview" | "computer-use-preview-2025-03-11",
  object: string,
  output: ResponseOutputMessage | ResponseFileSearchToolCall | ResponseFunctionToolCall | ResponseFunctionWebSearch | ResponseComputerToolCall | ResponseReasoningItem | OpenaiTypesResponsesResponseOutputItemImageGenerationCall | ResponseCodeInterpreterToolCall | OpenaiTypesResponsesResponseOutputItemLocalShellCall | OpenaiTypesResponsesResponseOutputItemMcpCall | OpenaiTypesResponsesResponseOutputItemMcpListTools | OpenaiTypesResponsesResponseOutputItemMcpApprovalRequest[],
  parallel_tool_calls: boolean,
  temperature?: number | null,
  tool_choice: "none" | "auto" | "required" | ToolChoiceTypes | ToolChoiceFunction | ToolChoiceMcp,
  tools: FunctionTool | FileSearchTool | WebSearchTool | ComputerTool | Mcp | CodeInterpreter | ImageGeneration | LocalShell[],
  top_p?: number | null,
  background?: boolean | null,
  max_output_tokens?: number | null,
  max_tool_calls?: number | null,
  previous_response_id?: string | null,
  prompt?: ResponsePrompt | null,
  reasoning?: OpenaiTypesSharedReasoningReasoning | null,
  service_tier?: "auto" | "default" | "flex" | "scale" | "priority" | null,
  status?: "completed" | "failed" | "in_progress" | "cancelled" | "queued" | "incomplete" | null,
  text?: ResponseTextConfig | null,
  top_logprobs?: number | null,
  truncation?: "auto" | "disabled" | null,
  usage?: ResponseUsage | null,
  user?: string | null,
};

export type ResponseCodeInterpreterToolCall = {
  id: string,
  code?: string | null,
  container_id: string,
  outputs?: OpenaiTypesResponsesResponseCodeInterpreterToolCallOutputLogs | OpenaiTypesResponsesResponseCodeInterpreterToolCallOutputImage[] | null,
  status: "in_progress" | "completed" | "incomplete" | "interpreting" | "failed",
  type: string,
};

export type ResponseCodeInterpreterToolCallParam = {
  id: string,
  code: string | null,
  container_id: string,
  outputs: OpenaiTypesResponsesResponseCodeInterpreterToolCallParamOutputLogs | OpenaiTypesResponsesResponseCodeInterpreterToolCallParamOutputImage[] | null,
  status: "in_progress" | "completed" | "incomplete" | "interpreting" | "failed",
  type: string,
};

export type ResponseComputerToolCall = {
  id: string,
  action: OpenaiTypesResponsesResponseComputerToolCallActionClick | OpenaiTypesResponsesResponseComputerToolCallActionDoubleClick | OpenaiTypesResponsesResponseComputerToolCallActionDrag | OpenaiTypesResponsesResponseComputerToolCallActionKeypress | OpenaiTypesResponsesResponseComputerToolCallActionMove | OpenaiTypesResponsesResponseComputerToolCallActionScreenshot | OpenaiTypesResponsesResponseComputerToolCallActionScroll | OpenaiTypesResponsesResponseComputerToolCallActionType | OpenaiTypesResponsesResponseComputerToolCallActionWait,
  call_id: string,
  pending_safety_checks: OpenaiTypesResponsesResponseComputerToolCallPendingSafetyCheck[],
  status: "in_progress" | "completed" | "incomplete",
  type: string,
};

export type ResponseComputerToolCallOutputScreenshot = {
  type: string,
  file_id?: string | null,
  image_url?: string | null,
};

export type ResponseComputerToolCallOutputScreenshotParam = {
  type: string,
  file_id?: string,
  image_url?: string,
};

export type ResponseComputerToolCallParam = {
  id: string,
  action: OpenaiTypesResponsesResponseComputerToolCallParamActionClick | OpenaiTypesResponsesResponseComputerToolCallParamActionDoubleClick | OpenaiTypesResponsesResponseComputerToolCallParamActionDrag | OpenaiTypesResponsesResponseComputerToolCallParamActionKeypress | OpenaiTypesResponsesResponseComputerToolCallParamActionMove | OpenaiTypesResponsesResponseComputerToolCallParamActionScreenshot | OpenaiTypesResponsesResponseComputerToolCallParamActionScroll | OpenaiTypesResponsesResponseComputerToolCallParamActionType | OpenaiTypesResponsesResponseComputerToolCallParamActionWait,
  call_id: string,
  pending_safety_checks: OpenaiTypesResponsesResponseComputerToolCallParamPendingSafetyCheck[],
  status: "in_progress" | "completed" | "incomplete",
  type: string,
};

export type ResponseError = {
  code: "server_error" | "rate_limit_exceeded" | "invalid_prompt" | "vector_store_timeout" | "invalid_image" | "invalid_image_format" | "invalid_base64_image" | "invalid_image_url" | "image_too_large" | "image_too_small" | "image_parse_error" | "image_content_policy_violation" | "invalid_image_mode" | "image_file_too_large" | "unsupported_image_media_type" | "empty_image_file" | "failed_to_download_image" | "image_file_not_found",
  message: string,
};

export type ResponseFileSearchToolCall = {
  id: string,
  queries: string[],
  status: "in_progress" | "searching" | "completed" | "incomplete" | "failed",
  type: string,
  results?: OpenaiTypesResponsesResponseFileSearchToolCallResult[] | null,
};

export type ResponseFileSearchToolCallParam = {
  id: string,
  queries: string[],
  status: "in_progress" | "searching" | "completed" | "incomplete" | "failed",
  type: string,
  results?: OpenaiTypesResponsesResponseFileSearchToolCallParamResult[] | null,
};

export type ResponseFormatJSONSchema = {
  json_schema: JSONSchema,
  type: string,
};

export type ResponseFormatTextJSONSchemaConfig = {
  name: string,
  schema: object,
  type: string,
  description?: string | null,
  strict?: boolean | null,
};

export type ResponseFormatTextJSONSchemaConfigParam = {
  name: string,
  schema: object,
  type: string,
  description?: string,
  strict?: boolean | null,
};

export type ResponseFunctionToolCall = {
  arguments: string,
  call_id: string,
  name: string,
  type: string,
  id?: string | null,
  status?: "in_progress" | "completed" | "incomplete" | null,
};

export type ResponseFunctionToolCallParam = {
  arguments: string,
  call_id: string,
  name: string,
  type: string,
  id?: string,
  status?: "in_progress" | "completed" | "incomplete",
};

export type ResponseFunctionWebSearch = {
  id: string,
  action: OpenaiTypesResponsesResponseFunctionWebSearchActionSearch | OpenaiTypesResponsesResponseFunctionWebSearchActionOpenPage | OpenaiTypesResponsesResponseFunctionWebSearchActionFind,
  status: "in_progress" | "searching" | "completed" | "failed",
  type: string,
};

export type ResponseFunctionWebSearchParam = {
  id: string,
  action: OpenaiTypesResponsesResponseFunctionWebSearchParamActionSearch | OpenaiTypesResponsesResponseFunctionWebSearchParamActionOpenPage | OpenaiTypesResponsesResponseFunctionWebSearchParamActionFind,
  status: "in_progress" | "searching" | "completed" | "failed",
  type: string,
};

export type ResponseInputFile = {
  type: string,
  file_data?: string | null,
  file_id?: string | null,
  filename?: string | null,
};

export type ResponseInputFileParam = {
  type: string,
  file_data?: string,
  file_id?: string | null,
  filename?: string,
};

export type ResponseInputImage = {
  detail: "low" | "high" | "auto",
  type: string,
  file_id?: string | null,
  image_url?: string | null,
};

export type ResponseInputImageParam = {
  detail: "low" | "high" | "auto",
  type: string,
  file_id?: string | null,
  image_url?: string | null,
};

export type ResponseInputText = {
  text: string,
  type: string,
};

export type ResponseInputTextParam = {
  text: string,
  type: string,
};

export type ResponseOutputMessage = {
  id: string,
  content: ResponseOutputText | ResponseOutputRefusal[],
  role: string,
  status: "in_progress" | "completed" | "incomplete",
  type: string,
};

export type ResponseOutputMessageParam = {
  id: string,
  content: ResponseOutputTextParam | ResponseOutputRefusalParam[],
  role: string,
  status: "in_progress" | "completed" | "incomplete",
  type: string,
};

export type ResponseOutputRefusal = {
  refusal: string,
  type: string,
};

export type ResponseOutputRefusalParam = {
  refusal: string,
  type: string,
};

export type ResponseOutputText = {
  annotations: OpenaiTypesResponsesResponseOutputTextAnnotationFileCitation | OpenaiTypesResponsesResponseOutputTextAnnotationURLCitation | OpenaiTypesResponsesResponseOutputTextAnnotationContainerFileCitation | OpenaiTypesResponsesResponseOutputTextAnnotationFilePath[],
  text: string,
  type: string,
  logprobs?: OpenaiTypesResponsesResponseOutputTextLogprob[] | null,
};

export type ResponseOutputTextParam = {
  annotations: OpenaiTypesResponsesResponseOutputTextParamAnnotationFileCitation | OpenaiTypesResponsesResponseOutputTextParamAnnotationURLCitation | OpenaiTypesResponsesResponseOutputTextParamAnnotationContainerFileCitation | OpenaiTypesResponsesResponseOutputTextParamAnnotationFilePath[],
  text: string,
  type: string,
  logprobs?: OpenaiTypesResponsesResponseOutputTextParamLogprob[],
};

export type ResponsePrompt = {
  id: string,
  variables?: object | null,
  version?: string | null,
};

export type ResponseReasoningItem = {
  id: string,
  summary: OpenaiTypesResponsesResponseReasoningItemSummary[],
  type: string,
  encrypted_content?: string | null,
  status?: "in_progress" | "completed" | "incomplete" | null,
};

export type ResponseReasoningItemParam = {
  id: string,
  summary: OpenaiTypesResponsesResponseReasoningItemParamSummary[],
  type: string,
  encrypted_content?: string | null,
  status?: "in_progress" | "completed" | "incomplete",
};

export type ResponseTextConfig = {
  format?: OpenaiTypesSharedResponseFormatTextResponseFormatText | ResponseFormatTextJSONSchemaConfig | OpenaiTypesSharedResponseFormatJsonObjectResponseFormatJSONObject | null,
};

export type ResponseTextConfigParam = {
  format?: OpenaiTypesSharedParamsResponseFormatTextResponseFormatText | ResponseFormatTextJSONSchemaConfigParam | OpenaiTypesSharedParamsResponseFormatJsonObjectResponseFormatJSONObject,
};

export type ResponseUsage = {
  input_tokens: number,
  input_tokens_details: InputTokensDetails,
  output_tokens: number,
  output_tokens_details: OutputTokensDetails,
  total_tokens: number,
};

export type RetabChatCompletionsRequest = {
  model: string,
  messages: ChatCompletionRetabMessageInput[],
  response_format: ResponseFormatJSONSchema,
  temperature?: number,
  reasoning_effort?: "low" | "medium" | "high" | null,
  stream?: boolean,
  seed?: number | null,
  n_consensus?: number,
};

export type RetabChatResponseCreateRequest = {
  input: string | EasyInputMessageParam | OpenaiTypesResponsesResponseInputParamMessage | ResponseOutputMessageParam | ResponseFileSearchToolCallParam | ResponseComputerToolCallParam | OpenaiTypesResponsesResponseInputParamComputerCallOutput | ResponseFunctionWebSearchParam | ResponseFunctionToolCallParam | OpenaiTypesResponsesResponseInputParamFunctionCallOutput | ResponseReasoningItemParam | OpenaiTypesResponsesResponseInputParamImageGenerationCall | ResponseCodeInterpreterToolCallParam | OpenaiTypesResponsesResponseInputParamLocalShellCall | OpenaiTypesResponsesResponseInputParamLocalShellCallOutput | OpenaiTypesResponsesResponseInputParamMcpListTools | OpenaiTypesResponsesResponseInputParamMcpApprovalRequest | OpenaiTypesResponsesResponseInputParamMcpApprovalResponse | OpenaiTypesResponsesResponseInputParamMcpCall | OpenaiTypesResponsesResponseInputParamItemReference[],
  instructions?: string | null,
  model: string,
  temperature?: number | null,
  reasoning?: OpenaiTypesSharedParamsReasoningReasoning | null,
  stream?: boolean | null,
  seed?: number | null,
  text?: ResponseTextConfigParam,
  n_consensus?: number,
};

export type RetabParsedChatCompletionInput = {
  id: string,
  choices: RetabParsedChoiceInput[],
  created: number,
  model: string,
  object: string,
  service_tier?: "auto" | "default" | "flex" | "scale" | "priority" | null,
  system_fingerprint?: string | null,
  usage?: CompletionUsage | null,
  extraction_id?: string | null,
  likelihoods?: object | null,
  schema_validation_error?: ErrorDetail | null,
  request_at?: Date | null,
  first_token_at?: Date | null,
  last_token_at?: Date | null,
};

export type RetabParsedChatCompletionOutput = {
  id: string,
  choices: RetabParsedChoiceOutput[],
  created: number,
  model: string,
  object: string,
  service_tier?: "auto" | "default" | "flex" | "scale" | "priority" | null,
  system_fingerprint?: string | null,
  usage?: CompletionUsage | null,
  extraction_id?: string | null,
  likelihoods?: object | null,
  schema_validation_error?: ErrorDetail | null,
  request_at?: Date | null,
  first_token_at?: Date | null,
  last_token_at?: Date | null,
  api_cost: Amount | null,
};

export type RetabParsedChatCompletionChunk = {
  id: string,
  choices: RetabParsedChoiceChunk[],
  created: number,
  model: string,
  object: string,
  service_tier?: "auto" | "default" | "flex" | "scale" | "priority" | null,
  system_fingerprint?: string | null,
  usage?: CompletionUsage | null,
  streaming_error?: ErrorDetail | null,
  extraction_id?: string | null,
  schema_validation_error?: ErrorDetail | null,
  request_at?: Date | null,
  first_token_at?: Date | null,
  last_token_at?: Date | null,
  api_cost: Amount | null,
  cost_breakdown: CostBreakdown | null,
};

export type RetabParsedChoiceInput = {
  finish_reason?: "stop" | "length" | "tool_calls" | "content_filter" | "function_call" | null,
  index: number,
  logprobs?: ChoiceLogprobsInput | null,
  message: ParsedChatCompletionMessageInput,
  field_locations?: object | null,
  key_mapping?: object | null,
};

export type RetabParsedChoiceOutput = {
  finish_reason?: "stop" | "length" | "tool_calls" | "content_filter" | "function_call" | null,
  index: number,
  logprobs?: ChoiceLogprobsOutput | null,
  message: ParsedChatCompletionMessageOutput,
  field_locations?: object | null,
  key_mapping?: object | null,
};

export type RetabParsedChoiceChunk = {
  delta: RetabParsedChoiceDeltaChunk,
  finish_reason?: "stop" | "length" | "tool_calls" | "content_filter" | "function_call" | null,
  index: number,
  logprobs?: ChoiceLogprobsOutput | null,
};

export type RetabParsedChoiceDeltaChunk = {
  content?: string | null,
  function_call?: ChoiceDeltaFunctionCall | null,
  refusal?: string | null,
  role?: "developer" | "system" | "user" | "assistant" | "tool" | null,
  tool_calls?: ChoiceDeltaToolCall[] | null,
  flat_likelihoods?: object,
  flat_parsed?: object,
  flat_deleted_keys?: string[],
  field_locations?: object | null,
  is_valid_json?: boolean,
  key_mapping?: object | null,
};

export type RetabUsage = {
  page_count: number,
  credits: number,
};

export type ReviewExtractionRequest = {
  extraction?: object | null,
};

export type Schema = {
  object?: string,
  created_at?: Date,
  json_schema?: object,
  strict?: boolean,
  data_id: string,
  id: string,
};

export type ServerToolUsage = {
  web_search_requests: number,
};

export type ServerToolUseBlock = {
  id: string,
  input: any,
  name: string,
  type: string,
};

export type ServerToolUseBlockParam = {
  id: string,
  input: any,
  name: string,
  type: string,
  cache_control?: CacheControlEphemeralParam | null,
};

export type SpreadsheetDetails = {
  accessible: boolean,
  spreadsheet_id: string,
  spreadsheet_name: string,
  worksheets: GoogleWorksheet[],
  spreadsheet_url: string,
  error?: string | null,
};

export type StoredDBFile = {
  object?: string,
  id: string,
  filename: string,
  organization_id: string,
  created_at?: Date,
  page_count?: number | null,
  ocr?: OCRMetadata | null,
};

export type StoredProcessor = {
  object?: string,
  id?: string,
  updated_at?: Date,
  name: string,
  modality: "text" | "image" | "native" | "image+text",
  image_resolution_dpi?: number,
  browser_canvas?: "A3" | "A4" | "A5",
  model: string,
  json_schema: object,
  temperature?: number,
  reasoning_effort?: "low" | "medium" | "high" | null,
  n_consensus?: number,
  organization_id: string,
  schema_data_id: string,
  schema_id: string,
};

export type SubscriptionStatus = {
  status?: "active" | "canceled" | "incomplete" | "incomplete_expired" | "past_due" | "paused" | "trialing" | "unpaid",
  plan?: string,
  current_period_end?: number,
  plan_name?: string,
  cancel_at_period_end?: boolean,
};

export type TemplateSchema = {
  id?: string,
  name: string,
  object?: string,
  updated_at?: Date,
  json_schema?: object,
  python_code?: string | null,
  sample_document_filename?: string | null,
  schema_data_id: string,
  schema_id: string,
};

export type TextBlock = {
  citations?: CitationCharLocation | CitationPageLocation | CitationContentBlockLocation | CitationsWebSearchResultLocation[] | null,
  text: string,
  type: string,
};

export type TextBlockParam = {
  text: string,
  type: string,
  cache_control?: CacheControlEphemeralParam | null,
  citations?: CitationCharLocationParam | CitationPageLocationParam | CitationContentBlockLocationParam | CitationWebSearchResultLocationParam[] | null,
};

export type TextBox = {
  width: number,
  height: number,
  center: Point,
  vertices: [Point, Point, Point, Point],
  text: string,
};

export type ThinkingBlock = {
  signature: string,
  thinking: string,
  type: string,
};

export type ThinkingBlockParam = {
  signature: string,
  thinking: string,
  type: string,
};

export type TimeRange = "day" | "week" | "month" | "three_months";

export type TokenCount = {
  total_tokens?: number,
  developer_tokens?: number,
  user_tokens?: number,
};

export type TokenCounts = {
  prompt_regular_text: number,
  prompt_cached_text: number,
  prompt_audio: number,
  completion_regular_text: number,
  completion_audio: number,
  total_tokens: number,
};

export type TokenPrice = {
  prompt: number,
  completion: number,
  cached_discount?: number,
};

export type ToolChoiceFunction = {
  name: string,
  type: string,
};

export type ToolChoiceMcp = {
  server_label: string,
  type: string,
  name?: string | null,
};

export type ToolChoiceTypes = {
  type: "file_search" | "web_search_preview" | "computer_use_preview" | "web_search_preview_2025_03_11" | "image_generation" | "code_interpreter",
};

export type ToolResultBlockParam = {
  tool_use_id: string,
  type: string,
  cache_control?: CacheControlEphemeralParam | null,
  content?: string | TextBlockParam | ImageBlockParam[],
  is_error?: boolean,
};

export type ToolUseBlock = {
  id: string,
  input: any,
  name: string,
  type: string,
};

export type ToolUseBlockParam = {
  id: string,
  input: any,
  name: string,
  type: string,
  cache_control?: CacheControlEphemeralParam | null,
};

export type TopLogprob = {
  token: string,
  bytes?: number[] | null,
  logprob: number,
};

export type URLImageSourceParam = {
  type: string,
  url: string,
};

export type URLPDFSourceParam = {
  type: string,
  url: string,
};

export type UpdateEmailDataRequest = {
  email_data: EmailDataInput,
  additional_documents?: AttachmentMIMEDataInput[] | null,
};

export type UpdateEndpointRequest = {
  name?: string | null,
  default_language?: string | null,
  webhook_url?: string | null,
  webhook_headers?: object | null,
  need_validation?: boolean | null,
};

export type UpdateLinkRequest = {
  name?: string | null,
  default_language?: string | null,
  webhook_url?: string | null,
  webhook_headers?: object | null,
  need_validation?: boolean | null,
  password?: string | null,
};

export type UpdateMailboxRequest = {
  name?: string | null,
  default_language?: string | null,
  webhook_url?: string | null,
  webhook_headers?: object | null,
  need_validation?: boolean | null,
  authorized_domains?: string[] | null,
  authorized_emails?: string[] | null,
};

export type UpdateOutlookRequest = {
  name?: string | null,
  default_language?: string | null,
  webhook_url?: string | null,
  webhook_headers?: object | null,
  need_validation?: boolean | null,
  authorized_domains?: string[] | null,
  authorized_emails?: string[] | null,
  match_params?: MatchParams[] | null,
  fetch_params?: FetchParams[] | null,
  layout_schema?: object | null,
};

export type UpdateProcessorRequest = {
  name?: string | null,
  modality?: "text" | "image" | "native" | "image+text" | null,
  image_resolution_dpi?: number | null,
  browser_canvas?: "A3" | "A4" | "A5" | null,
  model?: string | null,
  json_schema?: object | null,
  temperature?: number | null,
  reasoning_effort?: "low" | "medium" | "high" | null,
  n_consensus?: number | null,
};

export type UpdateTemplateRequest = {
  id: string,
  name?: string | null,
  json_schema?: object | null,
  python_code?: string | null,
  sample_document?: MIMEDataInput | null,
};

export type Usage = {
  cache_creation_input_tokens?: number | null,
  cache_read_input_tokens?: number | null,
  input_tokens: number,
  output_tokens: number,
  server_tool_use?: ServerToolUsage | null,
  service_tier?: "standard" | "priority" | "batch" | null,
};

export type UsageTimeSeries = {
  data: DataPoint[],
};

export type UserLocation = {
  type: string,
  city?: string | null,
  country?: string | null,
  region?: string | null,
  timezone?: string | null,
};

export type UserParameters = {
  language: "fr" | "en" | "de" | "es" | "it" | "nl" | "pt" | "pl",
  agency_id: string,
  organization_id: string,
  phone_number?: string | null,
};

export type ValidationError = {
  loc: string | number[],
  msg: string,
  type: string,
};

export type VectorSearchRequest = {
  query_vector: number[],
  organization_id: string,
  schema_id: string,
  limit?: number,
  num_candidates?: number,
  similarity_metric?: "cosine" | "euclidean" | "dotProduct",
};

export type VectorSearchResponse = {
  results: VectorSearchResult[],
  average_search_score: number,
};

export type VectorSearchResult = {
  file_id: string,
  search_score: number,
  llm_output: object,
  hil_output: object,
  schema_id: string,
  levenshtein_similarity: object,
  created_at: Date,
};

export type WebSearchResultBlock = {
  encrypted_content: string,
  page_age?: string | null,
  title: string,
  type: string,
  url: string,
};

export type WebSearchResultBlockParam = {
  encrypted_content: string,
  title: string,
  type: string,
  url: string,
  page_age?: string | null,
};

export type WebSearchTool = {
  type: "web_search_preview" | "web_search_preview_2025_03_11",
  search_context_size?: "low" | "medium" | "high" | null,
  user_location?: UserLocation | null,
};

export type WebSearchToolRequestErrorParam = {
  error_code: "invalid_tool_input" | "unavailable" | "max_uses_exceeded" | "too_many_requests" | "query_too_long",
  type: string,
};

export type WebSearchToolResultBlock = {
  content: WebSearchToolResultError | WebSearchResultBlock[],
  tool_use_id: string,
  type: string,
};

export type WebSearchToolResultBlockParam = {
  content: WebSearchResultBlockParam[] | WebSearchToolRequestErrorParam,
  tool_use_id: string,
  type: string,
  cache_control?: CacheControlEphemeralParam | null,
};

export type WebSearchToolResultError = {
  error_code: "invalid_tool_input" | "unavailable" | "max_uses_exceeded" | "too_many_requests" | "query_too_long",
  type: string,
};

export type WebhookRequest = {
  completion: RetabParsedChatCompletionInput,
  user?: string | null,
  file_payload: MIMEDataInput,
  metadata?: object | null,
};

export type WebhookSignature = {
  organization_id: string,
  updated_at: Date,
  signature: string,
};

export type DocumentExtractRequest = {
  document?: MIMEDataInput,
  documents: MIMEDataInput[],
  modality: "text" | "image" | "native" | "image+text",
  image_resolution_dpi?: number,
  browser_canvas?: "A3" | "A4" | "A5",
  model: string,
  json_schema: object,
  temperature?: number,
  reasoning_effort?: "low" | "medium" | "high" | null,
  n_consensus?: number,
  stream?: boolean,
  seed?: number | null,
  store?: boolean,
  need_validation?: boolean,
  test_exception?: "before_handle_extraction" | "within_extraction_parse_or_stream" | "after_handle_extraction" | "within_process_document_stream_generator" | null,
};

export type AnthropicTypesMessageMessage = {
  id: string,
  content: TextBlock | ThinkingBlock | RedactedThinkingBlock | ToolUseBlock | ServerToolUseBlock | WebSearchToolResultBlock[],
  model: "claude-3-7-sonnet-latest" | "claude-3-7-sonnet-20250219" | "claude-3-5-haiku-latest" | "claude-3-5-haiku-20241022" | "claude-sonnet-4-20250514" | "claude-sonnet-4-0" | "claude-4-sonnet-20250514" | "claude-3-5-sonnet-latest" | "claude-3-5-sonnet-20241022" | "claude-3-5-sonnet-20240620" | "claude-opus-4-0" | "claude-opus-4-20250514" | "claude-4-opus-20250514" | "claude-3-opus-latest" | "claude-3-opus-20240229" | "claude-3-sonnet-20240229" | "claude-3-haiku-20240307" | "claude-2.1" | "claude-2.0" | string,
  role: string,
  stop_reason?: "end_turn" | "max_tokens" | "stop_sequence" | "tool_use" | "pause_turn" | "refusal" | null,
  stop_sequence?: string | null,
  type: string,
  usage: Usage,
};

export type MainServerServicesCustomBertCubemimedataAttachmentMetadata = {
  is_inline?: boolean,
  inline_cid?: string | null,
  url?: string | null,
  display_metadata?: DisplayMetadata | null,
  ocr?: MainServerServicesCustomBertCubemimedataOCR | null,
  source?: string | null,
};

export type MainServerServicesCustomBertCubemimedataBaseMIMEData = {
  id: string,
  name: string,
  size: number,
  mime_type: string,
  metadata: MainServerServicesCustomBertCubemimedataAttachmentMetadata,
};

export type MainServerServicesCustomBertCubemimedataMIMEData = {
  id: string,
  name: string,
  size: number,
  mime_type: string,
  metadata: MainServerServicesCustomBertCubemimedataAttachmentMetadata,
  content: string,
};

export type MainServerServicesCustomBertCubemimedataOCR = {
  pages: MainServerServicesCustomBertCubemimedataPage[],
};

export type MainServerServicesCustomBertCubemimedataPage = {
  page_number: number,
  width: number,
  height: number,
  blocks: TextBox[],
  lines: TextBox[],
};

export type MainServerServicesCustomBertfakeRoutesUser = {
  object: string,
  id: string,
  email: string,
  first_name?: string | null,
  last_name?: string | null,
  email_verified: boolean,
  profile_picture_url?: string | null,
  last_sign_in_at?: string | null,
  created_at: string,
  updated_at: string,
  external_id?: string | null,
  metadata?: object,
  parameters: UserParameters,
};

export type MainServerServicesInternalBlogModelsUser = {
  id: "louis-de-benoist" | "sacha-ichbiah" | "victor-plaisance",
  name: string,
  avatarUrl?: string | null,
  bio?: string | null,
  organization_id?: string | null,
};

export type MainServerServicesV1EvalsDistancesRoutesIterationMetricsFromEvaluationRequest = {
  evaluation: RetabTypesEvalsEvaluationInput,
  iteration_id: string,
};

export type MainServerServicesV1EvalsIoRoutesExportToCsvResponse = {
  csv_data: string,
  rows: number,
  columns: number,
};

export type MainServerServicesV1EvalsIterationsRoutesListIterationsResponse = {
  data: RetabTypesEvalsIterationOutput[],
};

export type MainServerServicesV1EvalsRoutesListEvaluations = {
  data: RetabTypesEvalsEvaluationOutput[],
  list_metadata: ListMetadata,
};

export type MainServerServicesV1EvalsRoutesPatchEvaluationRequest = {
  name?: string | null,
  project_id?: string | null,
  json_schema?: object | null,
};

export type MainServerServicesV1EvaluationsDistancesRoutesIterationMetricsFromEvaluationRequest = {
  evaluation: RetabTypesEvaluationsModelEvaluationInput,
  iteration_id: string,
};

export type MainServerServicesV1EvaluationsIoRoutesExportToCsvResponse = {
  csv_data: string,
  rows: number,
  columns: number,
};

export type MainServerServicesV1EvaluationsIterationsRoutesListIterationsResponse = {
  data: RetabTypesEvaluationsIterationsIterationOutput[],
};

export type MainServerServicesV1EvaluationsRoutesListEvaluations = {
  data: RetabTypesEvaluationsModelEvaluationOutput[],
  list_metadata: ListMetadata,
};

export type MainServerServicesV1IntegrationsGoogleSheetsRoutesExportToCsvResponse = {
  csv_data: string,
  rows: number,
  columns: number,
};

export type MainServerServicesV1SchemasDefaultTemplatesRoutesCreateTemplateRequest = {
  id?: string,
  name: string,
  json_schema: object,
  python_code?: string | null,
  sample_document?: MIMEDataInput | null,
};

export type MainServerServicesV1SchemasTemplatesRoutesCreateTemplateRequest = {
  name: string,
  json_schema: object,
  python_code?: string | null,
  sample_document?: MIMEDataInput | null,
};

export type OpenaiTypesChatChatCompletionAssistantMessageParamFunctionCall = {
  arguments: string,
  name: string,
};

export type OpenaiTypesChatChatCompletionMessageAnnotationURLCitation = {
  end_index: number,
  start_index: number,
  title: string,
  url: string,
};

export type OpenaiTypesChatChatCompletionMessageFunctionCall = {
  arguments: string,
  name: string,
};

export type OpenaiTypesChatChatCompletionMessageToolCallFunction = {
  arguments: string,
  name: string,
};

export type OpenaiTypesChatChatCompletionMessageToolCallParamFunction = {
  arguments: string,
  name: string,
};

export type OpenaiTypesResponsesResponseCodeInterpreterToolCallOutputImage = {
  type: string,
  url: string,
};

export type OpenaiTypesResponsesResponseCodeInterpreterToolCallOutputLogs = {
  logs: string,
  type: string,
};

export type OpenaiTypesResponsesResponseCodeInterpreterToolCallParamOutputImage = {
  type: string,
  url: string,
};

export type OpenaiTypesResponsesResponseCodeInterpreterToolCallParamOutputLogs = {
  logs: string,
  type: string,
};

export type OpenaiTypesResponsesResponseComputerToolCallActionClick = {
  button: "left" | "right" | "wheel" | "back" | "forward",
  type: string,
  x: number,
  y: number,
};

export type OpenaiTypesResponsesResponseComputerToolCallActionDoubleClick = {
  type: string,
  x: number,
  y: number,
};

export type OpenaiTypesResponsesResponseComputerToolCallActionDrag = {
  path: OpenaiTypesResponsesResponseComputerToolCallActionDragPath[],
  type: string,
};

export type OpenaiTypesResponsesResponseComputerToolCallActionDragPath = {
  x: number,
  y: number,
};

export type OpenaiTypesResponsesResponseComputerToolCallActionKeypress = {
  keys: string[],
  type: string,
};

export type OpenaiTypesResponsesResponseComputerToolCallActionMove = {
  type: string,
  x: number,
  y: number,
};

export type OpenaiTypesResponsesResponseComputerToolCallActionScreenshot = {
  type: string,
};

export type OpenaiTypesResponsesResponseComputerToolCallActionScroll = {
  scroll_x: number,
  scroll_y: number,
  type: string,
  x: number,
  y: number,
};

export type OpenaiTypesResponsesResponseComputerToolCallActionType = {
  text: string,
  type: string,
};

export type OpenaiTypesResponsesResponseComputerToolCallActionWait = {
  type: string,
};

export type OpenaiTypesResponsesResponseComputerToolCallPendingSafetyCheck = {
  id: string,
  code: string,
  message: string,
};

export type OpenaiTypesResponsesResponseComputerToolCallParamActionClick = {
  button: "left" | "right" | "wheel" | "back" | "forward",
  type: string,
  x: number,
  y: number,
};

export type OpenaiTypesResponsesResponseComputerToolCallParamActionDoubleClick = {
  type: string,
  x: number,
  y: number,
};

export type OpenaiTypesResponsesResponseComputerToolCallParamActionDrag = {
  path: OpenaiTypesResponsesResponseComputerToolCallParamActionDragPath[],
  type: string,
};

export type OpenaiTypesResponsesResponseComputerToolCallParamActionDragPath = {
  x: number,
  y: number,
};

export type OpenaiTypesResponsesResponseComputerToolCallParamActionKeypress = {
  keys: string[],
  type: string,
};

export type OpenaiTypesResponsesResponseComputerToolCallParamActionMove = {
  type: string,
  x: number,
  y: number,
};

export type OpenaiTypesResponsesResponseComputerToolCallParamActionScreenshot = {
  type: string,
};

export type OpenaiTypesResponsesResponseComputerToolCallParamActionScroll = {
  scroll_x: number,
  scroll_y: number,
  type: string,
  x: number,
  y: number,
};

export type OpenaiTypesResponsesResponseComputerToolCallParamActionType = {
  text: string,
  type: string,
};

export type OpenaiTypesResponsesResponseComputerToolCallParamActionWait = {
  type: string,
};

export type OpenaiTypesResponsesResponseComputerToolCallParamPendingSafetyCheck = {
  id: string,
  code: string,
  message: string,
};

export type OpenaiTypesResponsesResponseFileSearchToolCallResult = {
  attributes?: object | null,
  file_id?: string | null,
  filename?: string | null,
  score?: number | null,
  text?: string | null,
};

export type OpenaiTypesResponsesResponseFileSearchToolCallParamResult = {
  attributes?: object | null,
  file_id?: string,
  filename?: string,
  score?: number,
  text?: string,
};

export type OpenaiTypesResponsesResponseFunctionWebSearchActionFind = {
  pattern: string,
  type: string,
  url: string,
};

export type OpenaiTypesResponsesResponseFunctionWebSearchActionOpenPage = {
  type: string,
  url: string,
};

export type OpenaiTypesResponsesResponseFunctionWebSearchActionSearch = {
  query: string,
  type: string,
};

export type OpenaiTypesResponsesResponseFunctionWebSearchParamActionFind = {
  pattern: string,
  type: string,
  url: string,
};

export type OpenaiTypesResponsesResponseFunctionWebSearchParamActionOpenPage = {
  type: string,
  url: string,
};

export type OpenaiTypesResponsesResponseFunctionWebSearchParamActionSearch = {
  query: string,
  type: string,
};

export type OpenaiTypesResponsesResponseInputItemComputerCallOutput = {
  call_id: string,
  output: ResponseComputerToolCallOutputScreenshot,
  type: string,
  id?: string | null,
  acknowledged_safety_checks?: OpenaiTypesResponsesResponseInputItemComputerCallOutputAcknowledgedSafetyCheck[] | null,
  status?: "in_progress" | "completed" | "incomplete" | null,
};

export type OpenaiTypesResponsesResponseInputItemComputerCallOutputAcknowledgedSafetyCheck = {
  id: string,
  code?: string | null,
  message?: string | null,
};

export type OpenaiTypesResponsesResponseInputItemFunctionCallOutput = {
  call_id: string,
  output: string,
  type: string,
  id?: string | null,
  status?: "in_progress" | "completed" | "incomplete" | null,
};

export type OpenaiTypesResponsesResponseInputItemImageGenerationCall = {
  id: string,
  result?: string | null,
  status: "in_progress" | "completed" | "generating" | "failed",
  type: string,
};

export type OpenaiTypesResponsesResponseInputItemItemReference = {
  id: string,
  type?: string | null,
};

export type OpenaiTypesResponsesResponseInputItemLocalShellCall = {
  id: string,
  action: OpenaiTypesResponsesResponseInputItemLocalShellCallAction,
  call_id: string,
  status: "in_progress" | "completed" | "incomplete",
  type: string,
};

export type OpenaiTypesResponsesResponseInputItemLocalShellCallAction = {
  command: string[],
  env: object,
  type: string,
  timeout_ms?: number | null,
  user?: string | null,
  working_directory?: string | null,
};

export type OpenaiTypesResponsesResponseInputItemLocalShellCallOutput = {
  id: string,
  output: string,
  type: string,
  status?: "in_progress" | "completed" | "incomplete" | null,
};

export type OpenaiTypesResponsesResponseInputItemMcpApprovalRequest = {
  id: string,
  arguments: string,
  name: string,
  server_label: string,
  type: string,
};

export type OpenaiTypesResponsesResponseInputItemMcpApprovalResponse = {
  approval_request_id: string,
  approve: boolean,
  type: string,
  id?: string | null,
  reason?: string | null,
};

export type OpenaiTypesResponsesResponseInputItemMcpCall = {
  id: string,
  arguments: string,
  name: string,
  server_label: string,
  type: string,
  error?: string | null,
  output?: string | null,
};

export type OpenaiTypesResponsesResponseInputItemMcpListTools = {
  id: string,
  server_label: string,
  tools: OpenaiTypesResponsesResponseInputItemMcpListToolsTool[],
  type: string,
  error?: string | null,
};

export type OpenaiTypesResponsesResponseInputItemMcpListToolsTool = {
  input_schema: any,
  name: string,
  annotations?: any | null,
  description?: string | null,
};

export type OpenaiTypesResponsesResponseInputItemMessage = {
  content: ResponseInputText | ResponseInputImage | ResponseInputFile[],
  role: "user" | "system" | "developer",
  status?: "in_progress" | "completed" | "incomplete" | null,
  type?: string | null,
};

export type OpenaiTypesResponsesResponseInputParamComputerCallOutput = {
  call_id: string,
  output: ResponseComputerToolCallOutputScreenshotParam,
  type: string,
  id?: string | null,
  acknowledged_safety_checks?: OpenaiTypesResponsesResponseInputParamComputerCallOutputAcknowledgedSafetyCheck[] | null,
  status?: "in_progress" | "completed" | "incomplete" | null,
};

export type OpenaiTypesResponsesResponseInputParamComputerCallOutputAcknowledgedSafetyCheck = {
  id: string,
  code?: string | null,
  message?: string | null,
};

export type OpenaiTypesResponsesResponseInputParamFunctionCallOutput = {
  call_id: string,
  output: string,
  type: string,
  id?: string | null,
  status?: "in_progress" | "completed" | "incomplete" | null,
};

export type OpenaiTypesResponsesResponseInputParamImageGenerationCall = {
  id: string,
  result: string | null,
  status: "in_progress" | "completed" | "generating" | "failed",
  type: string,
};

export type OpenaiTypesResponsesResponseInputParamItemReference = {
  id: string,
  type?: string | null,
};

export type OpenaiTypesResponsesResponseInputParamLocalShellCall = {
  id: string,
  action: OpenaiTypesResponsesResponseInputParamLocalShellCallAction,
  call_id: string,
  status: "in_progress" | "completed" | "incomplete",
  type: string,
};

export type OpenaiTypesResponsesResponseInputParamLocalShellCallAction = {
  command: string[],
  env: object,
  type: string,
  timeout_ms?: number | null,
  user?: string | null,
  working_directory?: string | null,
};

export type OpenaiTypesResponsesResponseInputParamLocalShellCallOutput = {
  id: string,
  output: string,
  type: string,
  status?: "in_progress" | "completed" | "incomplete" | null,
};

export type OpenaiTypesResponsesResponseInputParamMcpApprovalRequest = {
  id: string,
  arguments: string,
  name: string,
  server_label: string,
  type: string,
};

export type OpenaiTypesResponsesResponseInputParamMcpApprovalResponse = {
  approval_request_id: string,
  approve: boolean,
  type: string,
  id?: string | null,
  reason?: string | null,
};

export type OpenaiTypesResponsesResponseInputParamMcpCall = {
  id: string,
  arguments: string,
  name: string,
  server_label: string,
  type: string,
  error?: string | null,
  output?: string | null,
};

export type OpenaiTypesResponsesResponseInputParamMcpListTools = {
  id: string,
  server_label: string,
  tools: OpenaiTypesResponsesResponseInputParamMcpListToolsTool[],
  type: string,
  error?: string | null,
};

export type OpenaiTypesResponsesResponseInputParamMcpListToolsTool = {
  input_schema: any,
  name: string,
  annotations?: any | null,
  description?: string | null,
};

export type OpenaiTypesResponsesResponseInputParamMessage = {
  content: ResponseInputTextParam | ResponseInputImageParam | ResponseInputFileParam[],
  role: "user" | "system" | "developer",
  status?: "in_progress" | "completed" | "incomplete",
  type?: string,
};

export type OpenaiTypesResponsesResponseOutputItemImageGenerationCall = {
  id: string,
  result?: string | null,
  status: "in_progress" | "completed" | "generating" | "failed",
  type: string,
};

export type OpenaiTypesResponsesResponseOutputItemLocalShellCall = {
  id: string,
  action: OpenaiTypesResponsesResponseOutputItemLocalShellCallAction,
  call_id: string,
  status: "in_progress" | "completed" | "incomplete",
  type: string,
};

export type OpenaiTypesResponsesResponseOutputItemLocalShellCallAction = {
  command: string[],
  env: object,
  type: string,
  timeout_ms?: number | null,
  user?: string | null,
  working_directory?: string | null,
};

export type OpenaiTypesResponsesResponseOutputItemMcpApprovalRequest = {
  id: string,
  arguments: string,
  name: string,
  server_label: string,
  type: string,
};

export type OpenaiTypesResponsesResponseOutputItemMcpCall = {
  id: string,
  arguments: string,
  name: string,
  server_label: string,
  type: string,
  error?: string | null,
  output?: string | null,
};

export type OpenaiTypesResponsesResponseOutputItemMcpListTools = {
  id: string,
  server_label: string,
  tools: OpenaiTypesResponsesResponseOutputItemMcpListToolsTool[],
  type: string,
  error?: string | null,
};

export type OpenaiTypesResponsesResponseOutputItemMcpListToolsTool = {
  input_schema: any,
  name: string,
  annotations?: any | null,
  description?: string | null,
};

export type OpenaiTypesResponsesResponseOutputTextAnnotationContainerFileCitation = {
  container_id: string,
  end_index: number,
  file_id: string,
  filename: string,
  start_index: number,
  type: string,
};

export type OpenaiTypesResponsesResponseOutputTextAnnotationFileCitation = {
  file_id: string,
  filename: string,
  index: number,
  type: string,
};

export type OpenaiTypesResponsesResponseOutputTextAnnotationFilePath = {
  file_id: string,
  index: number,
  type: string,
};

export type OpenaiTypesResponsesResponseOutputTextAnnotationURLCitation = {
  end_index: number,
  start_index: number,
  title: string,
  type: string,
  url: string,
};

export type OpenaiTypesResponsesResponseOutputTextLogprob = {
  token: string,
  bytes: number[],
  logprob: number,
  top_logprobs: OpenaiTypesResponsesResponseOutputTextLogprobTopLogprob[],
};

export type OpenaiTypesResponsesResponseOutputTextLogprobTopLogprob = {
  token: string,
  bytes: number[],
  logprob: number,
};

export type OpenaiTypesResponsesResponseOutputTextParamAnnotationContainerFileCitation = {
  container_id: string,
  end_index: number,
  file_id: string,
  filename: string,
  start_index: number,
  type: string,
};

export type OpenaiTypesResponsesResponseOutputTextParamAnnotationFileCitation = {
  file_id: string,
  filename: string,
  index: number,
  type: string,
};

export type OpenaiTypesResponsesResponseOutputTextParamAnnotationFilePath = {
  file_id: string,
  index: number,
  type: string,
};

export type OpenaiTypesResponsesResponseOutputTextParamAnnotationURLCitation = {
  end_index: number,
  start_index: number,
  title: string,
  type: string,
  url: string,
};

export type OpenaiTypesResponsesResponseOutputTextParamLogprob = {
  token: string,
  bytes: number[],
  logprob: number,
  top_logprobs: OpenaiTypesResponsesResponseOutputTextParamLogprobTopLogprob[],
};

export type OpenaiTypesResponsesResponseOutputTextParamLogprobTopLogprob = {
  token: string,
  bytes: number[],
  logprob: number,
};

export type OpenaiTypesResponsesResponseReasoningItemSummary = {
  text: string,
  type: string,
};

export type OpenaiTypesResponsesResponseReasoningItemParamSummary = {
  text: string,
  type: string,
};

export type OpenaiTypesSharedReasoningReasoning = {
  effort?: "low" | "medium" | "high" | null,
  generate_summary?: "auto" | "concise" | "detailed" | null,
  summary?: "auto" | "concise" | "detailed" | null,
};

export type OpenaiTypesSharedResponseFormatJsonObjectResponseFormatJSONObject = {
  type: string,
};

export type OpenaiTypesSharedResponseFormatTextResponseFormatText = {
  type: string,
};

export type OpenaiTypesSharedParamsReasoningReasoning = {
  effort?: "low" | "medium" | "high" | null,
  generate_summary?: "auto" | "concise" | "detailed" | null,
  summary?: "auto" | "concise" | "detailed" | null,
};

export type OpenaiTypesSharedParamsResponseFormatJsonObjectResponseFormatJSONObject = {
  type: string,
};

export type OpenaiTypesSharedParamsResponseFormatTextResponseFormatText = {
  type: string,
};

export type RetabTypesEvalsCreateIterationRequest = {
  inference_settings: InferenceSettings,
  json_schema?: object | null,
};

export type RetabTypesEvalsEvaluationInput = {
  id?: string,
  updated_at?: Date,
  name: string,
  old_documents?: EvaluationDocumentInput[] | null,
  documents: EvaluationDocumentInput[],
  iterations: RetabTypesEvalsIterationInput[],
  json_schema: object,
  project_id?: string,
  default_inference_settings?: InferenceSettings | null,
};

export type RetabTypesEvalsEvaluationOutput = {
  id?: string,
  updated_at?: Date,
  name: string,
  old_documents?: EvaluationDocumentOutput[] | null,
  documents: EvaluationDocumentOutput[],
  iterations: RetabTypesEvalsIterationOutput[],
  json_schema: object,
  project_id?: string,
  default_inference_settings?: InferenceSettings | null,
  schema_data_id: string,
  schema_id: string,
};

export type RetabTypesEvalsIterationInput = {
  id?: string,
  inference_settings: InferenceSettings,
  json_schema: object,
  predictions?: RetabTypesEvalsPredictionDataInput[],
  metric_results?: MetricResult | null,
};

export type RetabTypesEvalsIterationOutput = {
  id?: string,
  inference_settings: InferenceSettings,
  json_schema: object,
  predictions?: RetabTypesEvalsPredictionDataOutput[],
  metric_results?: MetricResult | null,
  schema_data_id: string,
  schema_id: string,
};

export type RetabTypesEvalsPredictionDataInput = {
  prediction?: object,
  metadata?: PredictionMetadata | null,
};

export type RetabTypesEvalsPredictionDataOutput = {
  prediction?: object,
  metadata?: PredictionMetadata | null,
};

export type RetabTypesEvaluationsIterationsCreateIterationRequest = {
  inference_settings: InferenceSettings,
  json_schema?: object | null,
  from_iteration_id?: string | null,
};

export type RetabTypesEvaluationsIterationsIterationInput = {
  id?: string,
  updated_at?: Date,
  inference_settings: InferenceSettings,
  json_schema: object,
  predictions?: object,
  metric_results?: MetricResult | null,
};

export type RetabTypesEvaluationsIterationsIterationOutput = {
  id?: string,
  updated_at?: Date,
  inference_settings: InferenceSettings,
  json_schema: object,
  predictions?: object,
  metric_results?: MetricResult | null,
  schema_data_id: string,
  schema_id: string,
};

export type RetabTypesEvaluationsModelEvaluationInput = {
  id?: string,
  updated_at?: Date,
  name: string,
  documents?: EvaluationDocumentInput[],
  iterations?: RetabTypesEvaluationsIterationsIterationInput[],
  json_schema: object,
  project_id?: string,
  default_inference_settings?: InferenceSettings,
};

export type RetabTypesEvaluationsModelEvaluationOutput = {
  id?: string,
  updated_at?: Date,
  name: string,
  documents?: EvaluationDocumentOutput[],
  iterations?: RetabTypesEvaluationsIterationsIterationOutput[],
  json_schema: object,
  project_id?: string,
  default_inference_settings?: InferenceSettings,
  schema_data_id: string,
  schema_id: string,
};

export type RetabTypesEvaluationsModelPatchEvaluationRequest = {
  name?: string | null,
  json_schema?: object | null,
  project_id?: string | null,
  default_inference_settings?: InferenceSettings | null,
};

export type RetabTypesMimeAttachmentMetadata = {
  is_inline?: boolean,
  inline_cid?: string | null,
  source?: string | null,
};

export type RetabTypesMimeBaseMIMEData = {
  filename: string,
  url: string,
};

export type RetabTypesMimeMIMEData = {
  filename: string,
  url: string,
};

export type RetabTypesMimeOCROutput = {
  pages: RetabTypesMimePageOutput[],
};

export type RetabTypesMimePageOutput = {
  page_number: number,
  width: number,
  height: number,
  unit?: string,
  blocks: TextBox[],
  lines: TextBox[],
  tokens: TextBox[],
  transforms?: Matrix[],
};

export type RetabTypesPredictionsPredictionDataInput = {
  prediction?: object,
  metadata?: PredictionMetadata | null,
  updated_at?: Date | null,
};

export type RetabTypesPredictionsPredictionDataOutput = {
  prediction?: object,
  metadata?: PredictionMetadata | null,
  updated_at?: Date | null,
};

