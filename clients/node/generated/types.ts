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

export type ActionDefinition = {
  name: string,
  input_model: string | null,
  output_model: string | null,
  action_function: string | null,
  poller_model?: string | null,
  poller_function?: string | null,
};

export type AddDomainRequest = {
  domain: string,
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

export type AnnotationParameters = {
  model: string,
  modality?: "text" | "image" | "native" | "image+text" | null,
  image_settings?: ImageSettings | null,
  temperature?: number,
};

export type AnnotationProps = {
  model?: string,
  temperature?: number,
  modality?: "text" | "image" | "native" | "image+text",
  reasoning_effort?: "low" | "medium" | "high" | null,
  image_settings?: ImageSettings,
};

export type AnnotationURLCitationOutput = {
  end_index: number,
  start_index: number,
  title: string,
  url: string,
};

export type ApplicationIdentity = {
  application_id: string,
  organization_id: string,
  tier?: number,
};

export type AstreDocumentAPIRequest = {
  document: MIMEDataInput,
  document_type?: string,
};

export type AstreDocumentAPIResponse = {
  extraction: object,
  created: number,
  likelihoods: any,
  schema_validation_error?: ErrorDetail | null,
};

export type AttachmentMIMEDataInput = {
  filename: string,
  url: string,
  metadata?: AttachmentMetadataInput,
};

export type AttachmentMIMEDataOutput = {
  filename: string,
  url: string,
  metadata?: UiformTypesMimeAttachmentMetadata,
};

export type AttachmentMetadataInput = {
  is_inline?: boolean,
  inline_cid?: string | null,
  url?: string | null,
  display_metadata?: DisplayMetadata | null,
  source?: string | null,
};

export type Audio = {
  id: string,
};

export type AutomationConfig = {
  object: string,
  id: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  modality: "text" | "image" | "native" | "image+text",
  image_settings?: ImageSettings,
  model: string,
  json_schema: object,
  temperature?: number,
  reasoning_effort?: "low" | "medium" | "high" | null,
  need_validation?: boolean,
  n_consensus?: number,
  schema_data_id: string,
  schema_id: string,
};

export type AutomationConfigWithName = {
  object: string,
  id: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  modality: "text" | "image" | "native" | "image+text",
  image_settings?: ImageSettings,
  model: string,
  json_schema: object,
  temperature?: number,
  reasoning_effort?: "low" | "medium" | "high" | null,
  need_validation?: boolean,
  n_consensus?: number,
  name?: string | null,
  email?: string | null,
  schema_data_id: string,
  schema_id: string,
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
  completion: UiParsedChatCompletionOutput | ChatCompletionOutput,
  file_metadata: UiformTypesMimeBaseMIMEData | null,
  external_request_log: ExternalRequestLog | null,
  extraction_id?: string | null,
  api_cost: Amount | null,
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
  attachments?: CubeServerServicesCustomBertCubemimedataBaseMIMEData[],
};

export type BodyConvertMsgToEmailModelV1AutomationsOutlookConvertMsgToEmailModelPost = {
  file: File,
};

export type BodyConvertToEmailDataAndUploadFileV1AutomationsOutlookConvertToEmailDataAndUploadFilePost = {
  file: File,
};

export type BodyCreateFileV1DbFilesPost = {
  file: File,
};

export type BodyCreateFilesV1DbFilesBatchPost = {
  files: File[],
};

export type BodyHandleEndpointProcessingV1EndpointIdPost = {
  file: File,
  identity?: Identity | ApplicationIdentity,
};

export type BodyHandleEndpointProcessingV1AutomationsEndpointsProcessEndpointIdPost = {
  file: File,
  identity?: Identity | ApplicationIdentity,
};

export type BodyHandleLinkWebhookV1AutomationsLinksParseLinkIdPost = {
  file: File,
  identity?: Identity | ApplicationIdentity,
};

export type BodyTestDocumentUploadV1AutomationsTestsUploadAutomationIdPost = {
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

export type ChatCompletionInput = {
  id: string,
  choices: ChoiceInput[],
  created: number,
  model: string,
  object: string,
  service_tier?: "scale" | "default" | null,
  system_fingerprint?: string | null,
  usage?: CompletionUsage | null,
};

export type ChatCompletionOutput = {
  id: string,
  choices: ChoiceOutput[],
  created: number,
  model: string,
  object: string,
  service_tier?: "scale" | "default" | null,
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

export type ChatCompletionUiformMessageInput = {
  role: "user" | "system" | "assistant" | "developer",
  content: string | ChatCompletionContentPartTextParam | ChatCompletionContentPartImageParam | ChatCompletionContentPartInputAudioParam | File[],
};

export type ChatCompletionUiformMessageOutput = {
  role: "user" | "system" | "assistant" | "developer",
  content: string | ChatCompletionContentPartTextParam | ChatCompletionContentPartImageParam | ChatCompletionContentPartInputAudioParam | File[],
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

export type CitationsConfigParam = {
  enabled?: boolean,
};

export type ClusterClusteringMetrics = {
  dispersion: number,
  avg_dist_to_centroid: number,
};

export type ClusteringExtractionRequest = {
  extraction_ids: string[],
};

export type ClusteringExtractionResponse = {
  download_url: string,
  gcs_path: string,
  filename: string,
  expires_in: string,
};

export type ClusteringMetrics = {
  silhouette: number,
  calinski: number,
  negative_davies: number,
  frac_clustered: number,
  per_cluster_metrics: object,
};

export type ColumnInput = {
  type: string,
  size: number,
  items?: RowInput | FieldItem | RefObject | RowListInput[],
  name?: string | null,
};

export type ColumnOutput = {
  type: string,
  size: number,
  items?: RowOutput | FieldItem | RefObject | RowListOutput[],
  name?: string | null,
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

export type ComputerCallOutput = {
  call_id: string,
  output: ResponseComputerToolCallOutputScreenshotParam,
  type: string,
  id?: string,
  acknowledged_safety_checks?: ComputerCallOutputAcknowledgedSafetyCheck[],
  status?: "in_progress" | "completed" | "incomplete",
};

export type ComputerCallOutputAcknowledgedSafetyCheck = {
  id: string,
  code: string,
  message: string,
};

export type ComputerTool = {
  display_height: number,
  display_width: number,
  environment: "mac" | "windows" | "ubuntu" | "browser",
  type: string,
};

export type ContentBlockSourceParam = {
  content: string | TextBlockParam | ImageBlockParam[],
  type: string,
};

export type CreateAndLinkOrganizationRequest = {
  organization_name: string,
};

export type CreateJobExecutionRequest = {
  job_name: string,
  context: object,
};

export type CreateOrganizationResponse = {
  success: boolean,
  workos_organization: Organization,
};

export type CreateSchemaEntry = {
  prompted_json_schema: object,
  is_current?: boolean,
  is_active?: boolean,
};

export type CreateSheetWithStoredTokenRequest = {
  sheet_name: string,
};

export type CreateWorkflowExecutionRequest = {
  workflow_name: string,
  context: object,
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

export type DatasetClusteringResponse = {
  points: DatasetClusteringResponsePoint[],
  metrics: ClusteringMetrics | null,
};

export type DatasetClusteringResponsePoint = {
  extraction_id: string,
  file_id: string,
  point: [number, number],
  label: number,
};

export type DisplayMetadata = {
  url: string,
  type: "image" | "pdf" | "txt",
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
  image_settings?: ImageSettings,
  json_schema: object,
};

export type DocumentCreateMessageRequest = {
  document: MIMEDataInput,
  modality: "text" | "image" | "native" | "image+text",
  image_settings?: ImageSettings,
};

export type DocumentMessage = {
  id: string,
  object?: string,
  messages: ChatCompletionUiformMessageOutput[],
  created: number,
  modality: "text" | "image" | "native" | "image+text",
};

export type DocumentTransformRequest = {
  document: MIMEDataInput,
};

export type DocumentTransformResponse = {
  document: UiformTypesMimeMIMEData,
};

export type DocumentUploadRequest = {
  document: MIMEDataInput,
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
  image_settings?: ImageSettings | null,
  model?: string | null,
  json_schema?: object | null,
  temperature?: number | null,
  n_consensus?: number | null,
  stream?: boolean,
  seed?: number | null,
  store?: boolean,
};

export type EndpointInput = {
  object?: string,
  id?: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  modality: "text" | "image" | "native" | "image+text",
  image_settings?: ImageSettings,
  model: string,
  json_schema: object,
  temperature?: number,
  reasoning_effort?: "low" | "medium" | "high" | null,
  need_validation?: boolean,
  n_consensus?: number,
  name: string,
};

export type EndpointOutput = {
  object?: string,
  id?: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  modality: "text" | "image" | "native" | "image+text",
  image_settings?: ImageSettings,
  model: string,
  json_schema: object,
  temperature?: number,
  reasoning_effort?: "low" | "medium" | "high" | null,
  need_validation?: boolean,
  n_consensus?: number,
  name: string,
  schema_data_id: string,
  schema_id: string,
};

export type ErrorDetail = {
  code: string,
  message: string,
  details?: object | null,
};

export type Event = {
  object?: string,
  id?: string,
  event: string,
  created_at?: Date,
  data: object,
  metadata?: object | null,
};

export type Experiment = {
  id?: string,
  name: string,
  updated_at: Date,
  documents: ExperimentDocument[],
  iterations: IterationOutput[],
  json_schema: object,
  organization_id: string,
  schema_data_id: string,
  schema_id: string,
};

export type ExperimentCreateRequest = {
  name: string,
  documents: ExperimentDocumentRequest[],
  iterations?: IterationInput[],
  json_schema: object,
};

export type ExperimentDocument = {
  id: string,
  name: string,
  ground_truth?: object | null,
  file_id?: string | null,
  mime_data: UiformTypesMimeBaseMIMEData,
};

export type ExperimentDocumentRequest = {
  id: string,
  name: string,
  ground_truth?: object | null,
  file_id?: string | null,
  mime_data: MIMEDataInput,
};

export type ExternalAPIKey = {
  provider: "OpenAI" | "Gemini",
  is_configured: boolean,
  last_updated: Date | null,
};

export type ExternalAPIKeyRequest = {
  provider: "OpenAI" | "Gemini",
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
  messages?: ChatCompletionUiformMessageOutput[],
  messages_gcs: string,
  file_gcs: string,
  file_id: string,
  status: "success" | "failed",
  completion: UiParsedChatCompletionOutput | ChatCompletionOutput,
  json_schema: any,
  model: string,
  temperature?: number,
  source: ExtractionSource,
  image_settings?: ImageSettings,
  modality?: "text" | "image" | "native" | "image+text",
  reasoning_effort?: "low" | "medium" | "high" | null,
  schema_id: string,
  schema_data_id: string,
  created_at?: Date,
  organization_id: string,
  validation_state?: "pending" | "validated" | "invalid" | null,
  api_cost: Amount | null,
};

export type ExtractionCount = {
  total: number,
};

export type ExtractionSource = {
  type: "api" | "annotation" | "automation.link" | "automation.email" | "automation.cron" | "automation.outlook" | "automation.endpoint" | "schema.extract",
  id?: string | null,
};

export type FetchParams = {
  endpoint: string,
  headers: object,
  name: string,
};

export type FieldItem = {
  type: string,
  name: string,
  size?: number | null,
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
  finetuning_props: AnnotationProps,
  eval_id?: string | null,
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
  user?: User | null,
  request_at: Date,
  extraction_type: "RoadBookingConfirmation" | "RoadTransportOrder" | "AirBookingConfirmation" | "RoadBookingConfirmationBert" | "RoadBookingConfirmationGroussard" | "RoadBookingConfirmationJourdan" | "RoadBookingConfirmationMGE" | "RoadBookingConfirmationSuus" | "RoadBookingConfirmationThevenon" | "RoadQuoteRequest" | "RoadPickupMazet" | "RoadCMR",
  extraction: any,
  uncertainties: any,
  mappings: object,
  action_type: "Creation" | "Modification" | "Deletion",
  documents: CubeServerServicesCustomBertCubemimedataBaseMIMEData | CubeServerServicesCustomBertCubemimedataMIMEData[],
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

export type FunctionCallOutput = {
  call_id: string,
  output: string,
  type: string,
  id?: string,
  status?: "in_progress" | "completed" | "incomplete",
};

export type FunctionTool = {
  name: string,
  parameters: object,
  strict: boolean,
  type: string,
  description?: string | null,
};

export type GenerateSchemaRequest = {
  documents: MIMEDataInput[],
  model?: string,
  temperature?: number,
  reasoning_effort?: "low" | "medium" | "high" | null,
  modality: "text" | "image" | "native" | "image+text",
  image_settings?: ImageSettings,
  flat?: boolean,
  stream?: boolean,
};

export type HTTPValidationError = {
  detail?: ValidationError[],
};

export type Identity = {
  user_id: string,
  organization_id?: string | null,
  tier?: number,
};

export type ImageBlockParam = {
  source: Base64ImageSourceParam | URLImageSourceParam,
  type: string,
  cache_control?: CacheControlEphemeralParam | null,
};

export type ImageSettings = {
  correct_image_orientation?: boolean,
  dpi?: number,
  image_to_text?: "ocr" | "llm_description",
  browser_canvas?: "A3" | "A4" | "A5",
};

export type ImageURL = {
  url: string,
  detail?: "auto" | "low" | "high",
};

export type IncompleteDetails = {
  reason?: "max_output_tokens" | "content_filter" | null,
};

export type InputAudio = {
  data: string,
  format: "wav" | "mp3",
};

export type InputTokensDetails = {
  cached_tokens: number,
};

export type InvitationRequest = {
  email: string,
  role?: string,
};

export type ItemMetric = {
  id: string,
  name: string,
  similarity: number,
  similarities: object,
  flat_similarities: object,
};

export type ItemReference = {
  id: string,
  type: string,
};

export type IterationInput = {
  id?: string,
  annotation_props: AnnotationProps,
  json_schema: object,
  extraction_results?: object[],
  metric_results?: MetricResult | null,
};

export type IterationOutput = {
  id?: string,
  annotation_props: AnnotationProps,
  json_schema: object,
  extraction_results?: object[],
  metric_results?: MetricResult | null,
  schema_data_id: string,
  schema_id: string,
};

export type JobDefinition = {
  name: string,
  description: string,
  actions: ActionDefinition[],
  execution_policy?: JobExecutionPolicy,
  input_model: string | null,
  output_model: string | null,
};

export type JobExecution = {
  id?: string,
  job: JobDefinition,
  organization_id: string,
  workflow_execution_id?: string | null,
  status?: JobStatus,
  current_action_index?: number,
  waiting_for?: "human" | "third-party" | null,
  action_results?: object[],
  last_action_poller_results?: object,
  context?: object,
  result?: object | null,
  error?: string | null,
  retry_count?: number,
  last_heartbeat?: Date | null,
  created_at?: Date,
  updated_at?: Date,
  started_at?: Date | null,
  completed_at?: Date | null,
};

export type JobExecutionPolicy = {
  auto_resume?: boolean,
  stall_threshold_minutes?: number,
  allow_external_wait?: boolean,
};

export type JobStatus = "not_started" | "pending" | "running" | "waiting" | "stalled" | "succeeded" | "failed" | "skipped";

export type LayoutInput = {
  $defs?: object,
  type: string,
  size: number,
  items?: RowInput | RowListInput | FieldItem | RefObject[],
};

export type LayoutOutput = {
  $defs?: object,
  type: string,
  size: number,
  items?: RowOutput | RowListOutput | FieldItem | RefObject[],
};

export type LinkInput = {
  object?: string,
  id?: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  modality: "text" | "image" | "native" | "image+text",
  image_settings?: ImageSettings,
  model: string,
  json_schema: object,
  temperature?: number,
  reasoning_effort?: "low" | "medium" | "high" | null,
  need_validation?: boolean,
  n_consensus?: number,
  name: string,
  password?: string | null,
};

export type LinkOutput = {
  object?: string,
  id?: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  modality: "text" | "image" | "native" | "image+text",
  image_settings?: ImageSettings,
  model: string,
  json_schema: object,
  temperature?: number,
  reasoning_effort?: "low" | "medium" | "high" | null,
  need_validation?: boolean,
  n_consensus?: number,
  name: string,
  password?: string | null,
  schema_data_id: string,
  schema_id: string,
};

export type ListAutomations = {
  data: AutomationConfigWithName[],
  list_metadata: ListMetadata,
};

export type ListDomainsResponse = {
  domains: CustomDomain[],
};

export type ListEndpoints = {
  data: EndpointOutput[],
  list_metadata: ListMetadata,
};

export type ListEvals = {
  data: SingleFileEval[],
  list_metadata: ListMetadata,
};

export type ListFieldMetrics = {
  data: MetricsResponse[],
  list_metadata: ListMetadata,
};

export type ListFiles = {
  data: StoredDBFile[],
  list_metadata: ListMetadata,
};

export type ListFinetunedModels = {
  data: FinetunedModel[],
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

export type ListSchemas = {
  data: StoredSchema[],
  list_metadata: ListMetadata,
};

export type ListTemplates = {
  data: TemplateSchema[],
  list_metadata: ListMetadata,
};

export type LogCompletionRequest = {
  json_schema: object,
  completion: ChatCompletionInput,
};

export type LogExtractionRequest = {
  messages?: ChatCompletionUiformMessageInput[] | null,
  openai_messages?: ChatCompletionDeveloperMessageParam | ChatCompletionSystemMessageParam | ChatCompletionUserMessageParam | ChatCompletionAssistantMessageParam | ChatCompletionToolMessageParam | ChatCompletionFunctionMessageParam[] | null,
  openai_responses_input?: EasyInputMessageParam | OpenaiTypesResponsesResponseInputParamMessage | ResponseOutputMessageParam | ResponseFileSearchToolCallParam | ResponseComputerToolCallParam | ComputerCallOutput | ResponseFunctionWebSearchParam | ResponseFunctionToolCallParam | FunctionCallOutput | ResponseReasoningItemParam | ItemReference[] | null,
  anthropic_messages?: MessageParam[] | null,
  anthropic_system_prompt?: string | null,
  document?: MIMEDataInput,
  completion?: object | UiParsedChatCompletionInput | AnthropicTypesMessageMessage | ParsedChatCompletion | ChatCompletionInput | null,
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
  object?: string,
  id?: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  modality: "text" | "image" | "native" | "image+text",
  image_settings?: ImageSettings,
  model: string,
  json_schema: object,
  temperature?: number,
  reasoning_effort?: "low" | "medium" | "high" | null,
  need_validation?: boolean,
  n_consensus?: number,
  email: string,
  authorized_domains?: string[],
  authorized_emails?: string[],
};

export type MailboxOutput = {
  object?: string,
  id?: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  modality: "text" | "image" | "native" | "image+text",
  image_settings?: ImageSettings,
  model: string,
  json_schema: object,
  temperature?: number,
  reasoning_effort?: "low" | "medium" | "high" | null,
  need_validation?: boolean,
  n_consensus?: number,
  email: string,
  authorized_domains?: string[],
  authorized_emails?: string[],
  schema_data_id: string,
  schema_id: string,
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

export type MatchRequest = {
  record: object,
  k?: number,
};

export type MessageParam = {
  content: string | TextBlockParam | ImageBlockParam | ToolUseBlockParam | ToolResultBlockParam | DocumentBlockParam | ThinkingBlockParam | RedactedThinkingBlockParam | TextBlock | ToolUseBlock | ThinkingBlock | RedactedThinkingBlock[],
  role: "user" | "assistant",
};

export type MetricResult = {
  item_metrics: ItemMetric[],
  mean_similarity: number,
  metric_type: "levenshtein_similarity" | "jaccard_similarity" | "hamming_similarity",
};

export type MetricsResponse = {
  field_path: string,
  file_id: string,
  dataset_membership_id_1: string,
  dataset_membership_id_2: string,
  schema_id: string,
  schema_data_id: string,
  hamming_similarity?: number | null,
  jaccard_similarity?: number | null,
  levenshtein_similarity?: number | null,
  created_at: string,
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
  features: "streaming" | "function_calling" | "structured_outputs" | "distillation" | "fine_tuning" | "predicted_outputs"[],
};

export type ModelCard = {
  model: "gpt-4o" | "gpt-4o-mini" | "chatgpt-4o-latest" | "gpt-4o-2024-11-20" | "gpt-4o-2024-08-06" | "gpt-4o-2024-05-13" | "gpt-4o-mini-2024-07-18" | "o3-mini" | "o3-mini-2025-01-31" | "o1" | "o1-2024-12-17" | "o1-preview-2024-09-12" | "o1-mini" | "o1-mini-2024-09-12" | "gpt-4.5-preview" | "gpt-4.5-preview-2025-02-27" | "gpt-4o-audio-preview-2024-12-17" | "gpt-4o-audio-preview-2024-10-01" | "gpt-4o-realtime-preview-2024-12-17" | "gpt-4o-realtime-preview-2024-10-01" | "gpt-4o-mini-audio-preview-2024-12-17" | "gpt-4o-mini-realtime-preview-2024-12-17" | "human" | "claude-3-5-sonnet-latest" | "claude-3-5-sonnet-20241022" | "claude-3-5-haiku-20241022" | "claude-3-opus-20240229" | "claude-3-sonnet-20240229" | "claude-3-haiku-20240307" | "grok-2-vision-1212" | "grok-2-1212" | "gemini-2.5-pro-exp-03-25" | "gemini-2.0-flash" | "gemini-2.0-flash-lite" | "gemini-1.5-pro" | string,
  pricing: Pricing,
  capabilities: ModelCapabilities,
  logprobs_support?: boolean,
  temperature_support?: boolean,
  reasoning_effort_support?: boolean,
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
  request_count: number,
};

export type MultipleUploadResponse = {
  files: DBFile[],
};

export type OCR = {
  pages: Page[],
};

export type OpenAIKeyValidationResponse = {
  is_valid: boolean,
  message: string,
};

export type Organization = {
  id: string,
  object: string,
  name: string,
  domains: OrganizationDomain[],
  created_at: string,
  updated_at: string,
  allow_profiles_outside_organization: boolean,
  lookup_key?: string | null,
  stripe_customer_id?: string | null,
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

export type OrganizationSchemaEntry = {
  schema_id: string,
  data_structure_version: string,
  is_current: boolean,
  is_active: boolean,
  tag?: string | null,
  created_at: Date,
  prompted_json_schema: object,
};

export type OutlookInput = {
  object?: string,
  id?: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  modality: "text" | "image" | "native" | "image+text",
  image_settings?: ImageSettings,
  model: string,
  json_schema: object,
  temperature?: number,
  reasoning_effort?: "low" | "medium" | "high" | null,
  need_validation?: boolean,
  n_consensus?: number,
  name: string,
  authorized_domains?: string[],
  authorized_emails?: string[],
  layout_schema?: LayoutInput | null,
  match_params?: MatchParams[],
  fetch_params?: FetchParams[],
};

export type OutlookOutput = {
  object?: string,
  id?: string,
  updated_at?: Date,
  default_language?: string,
  webhook_url: string,
  webhook_headers?: object,
  modality: "text" | "image" | "native" | "image+text",
  image_settings?: ImageSettings,
  model: string,
  json_schema: object,
  temperature?: number,
  reasoning_effort?: "low" | "medium" | "high" | null,
  need_validation?: boolean,
  n_consensus?: number,
  name: string,
  authorized_domains?: string[],
  authorized_emails?: string[],
  layout_schema?: LayoutOutput | null,
  match_params?: MatchParams[],
  fetch_params?: FetchParams[],
  schema_data_id: string,
  schema_id: string,
};

export type OutlookSubmitRequest = {
  email_data: EmailDataInput,
  completion: any,
  user_email: string,
  metadata: object,
  store?: boolean,
};

export type OutputTokensDetails = {
  reasoning_tokens: number,
};

export type Page = {
  page_number: number,
  width: number,
  height: number,
  blocks: TextBox[],
  lines: TextBox[],
};

export type PaginatedList = {
  data: any[],
  list_metadata: ListMetadata,
};

export type ParsedChatCompletion = {
  id: string,
  choices: ParsedChoice[],
  created: number,
  model: string,
  object: string,
  service_tier?: "scale" | "default" | null,
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

export type PlainTextSourceParam = {
  data: string,
  media_type: string,
  type: string,
};

export type Point = {
  x: number,
  y: number,
};

export type Pricing = {
  text: TokenPrice,
  audio?: TokenPrice | null,
  ft_price_hike?: number,
};

export type PromptTokensDetails = {
  audio_tokens?: number | null,
  cached_tokens?: number | null,
};

export type PromptifyRequest = {
  model?: string,
  temperature?: number,
  modality?: "text" | "image" | "native" | "image+text",
  stream?: boolean,
  reasoning_effort?: "low" | "medium" | "high" | null,
  raw_schema: object,
  documents: MIMEDataInput[],
};

export type RankingOptions = {
  ranker?: "auto" | "default-2024-11-15" | null,
  score_threshold?: number | null,
};

export type Reasoning = {
  effort?: "low" | "medium" | "high" | null,
  generate_summary?: "concise" | "detailed" | null,
};

export type RedactedThinkingBlock = {
  data: string,
  type: string,
};

export type RedactedThinkingBlockParam = {
  data: string,
  type: string,
};

export type RefObject = {
  type: string,
  size?: number | null,
  name?: string | null,
  $ref: string,
};

export type Response = {
  id: string,
  created_at: number,
  error?: ResponseError | null,
  incomplete_details?: IncompleteDetails | null,
  instructions?: string | null,
  metadata?: object | null,
  model: string | "o3-mini" | "o3-mini-2025-01-31" | "o1" | "o1-2024-12-17" | "o1-preview" | "o1-preview-2024-09-12" | "o1-mini" | "o1-mini-2024-09-12" | "gpt-4o" | "gpt-4o-2024-11-20" | "gpt-4o-2024-08-06" | "gpt-4o-2024-05-13" | "gpt-4o-audio-preview" | "gpt-4o-audio-preview-2024-10-01" | "gpt-4o-audio-preview-2024-12-17" | "gpt-4o-mini-audio-preview" | "gpt-4o-mini-audio-preview-2024-12-17" | "gpt-4o-search-preview" | "gpt-4o-mini-search-preview" | "gpt-4o-search-preview-2025-03-11" | "gpt-4o-mini-search-preview-2025-03-11" | "chatgpt-4o-latest" | "gpt-4o-mini" | "gpt-4o-mini-2024-07-18" | "gpt-4-turbo" | "gpt-4-turbo-2024-04-09" | "gpt-4-0125-preview" | "gpt-4-turbo-preview" | "gpt-4-1106-preview" | "gpt-4-vision-preview" | "gpt-4" | "gpt-4-0314" | "gpt-4-0613" | "gpt-4-32k" | "gpt-4-32k-0314" | "gpt-4-32k-0613" | "gpt-3.5-turbo" | "gpt-3.5-turbo-16k" | "gpt-3.5-turbo-0301" | "gpt-3.5-turbo-0613" | "gpt-3.5-turbo-1106" | "gpt-3.5-turbo-0125" | "gpt-3.5-turbo-16k-0613" | "o1-pro" | "o1-pro-2025-03-19" | "computer-use-preview" | "computer-use-preview-2025-03-11",
  object: string,
  output: ResponseOutputMessage | ResponseFileSearchToolCall | ResponseFunctionToolCall | ResponseFunctionWebSearch | ResponseComputerToolCall | ResponseReasoningItem[],
  parallel_tool_calls: boolean,
  temperature?: number | null,
  tool_choice: "none" | "auto" | "required" | ToolChoiceTypes | ToolChoiceFunction,
  tools: FileSearchTool | FunctionTool | ComputerTool | WebSearchTool[],
  top_p?: number | null,
  max_output_tokens?: number | null,
  previous_response_id?: string | null,
  reasoning?: Reasoning | null,
  status?: "completed" | "failed" | "in_progress" | "incomplete" | null,
  text?: ResponseTextConfig | null,
  truncation?: "auto" | "disabled" | null,
  usage?: ResponseUsage | null,
  user?: string | null,
};

export type ResponseComputerToolCall = {
  id: string,
  action: OpenaiTypesResponsesResponseComputerToolCallActionClick | OpenaiTypesResponsesResponseComputerToolCallActionDoubleClick | OpenaiTypesResponsesResponseComputerToolCallActionDrag | OpenaiTypesResponsesResponseComputerToolCallActionKeypress | OpenaiTypesResponsesResponseComputerToolCallActionMove | OpenaiTypesResponsesResponseComputerToolCallActionScreenshot | OpenaiTypesResponsesResponseComputerToolCallActionScroll | OpenaiTypesResponsesResponseComputerToolCallActionType | OpenaiTypesResponsesResponseComputerToolCallActionWait,
  call_id: string,
  pending_safety_checks: OpenaiTypesResponsesResponseComputerToolCallPendingSafetyCheck[],
  status: "in_progress" | "completed" | "incomplete",
  type: string,
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

export type ResponseFormatJSONObject = {
  type: string,
};

export type ResponseFormatText = {
  type: string,
};

export type ResponseFormatTextJSONSchemaConfig = {
  name: string,
  schema: object,
  type: string,
  description?: string | null,
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
  status: "in_progress" | "searching" | "completed" | "failed",
  type: string,
};

export type ResponseFunctionWebSearchParam = {
  id: string,
  status: "in_progress" | "searching" | "completed" | "failed",
  type: string,
};

export type ResponseInputFileParam = {
  type: string,
  file_data?: string,
  file_id?: string,
  filename?: string,
};

export type ResponseInputImageParam = {
  detail: "high" | "low" | "auto",
  type: string,
  file_id?: string | null,
  image_url?: string | null,
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
  annotations: OpenaiTypesResponsesResponseOutputTextAnnotationFileCitation | OpenaiTypesResponsesResponseOutputTextAnnotationURLCitation | OpenaiTypesResponsesResponseOutputTextAnnotationFilePath[],
  text: string,
  type: string,
};

export type ResponseOutputTextParam = {
  annotations: OpenaiTypesResponsesResponseOutputTextParamAnnotationFileCitation | OpenaiTypesResponsesResponseOutputTextParamAnnotationURLCitation | OpenaiTypesResponsesResponseOutputTextParamAnnotationFilePath[],
  text: string,
  type: string,
};

export type ResponseReasoningItem = {
  id: string,
  summary: OpenaiTypesResponsesResponseReasoningItemSummary[],
  type: string,
  status?: "in_progress" | "completed" | "incomplete" | null,
};

export type ResponseReasoningItemParam = {
  id: string,
  summary: OpenaiTypesResponsesResponseReasoningItemParamSummary[],
  type: string,
  status?: "in_progress" | "completed" | "incomplete",
};

export type ResponseTextConfig = {
  format?: ResponseFormatText | ResponseFormatTextJSONSchemaConfig | ResponseFormatJSONObject | null,
};

export type ResponseUsage = {
  input_tokens: number,
  input_tokens_details: InputTokensDetails,
  output_tokens: number,
  output_tokens_details: OutputTokensDetails,
  total_tokens: number,
};

export type ReviewExtractionRequest = {
  extraction?: object | null,
};

export type RowInput = {
  type: string,
  name?: string | null,
  items: ColumnInput | FieldItem | RefObject[],
};

export type RowOutput = {
  type: string,
  name?: string | null,
  items: ColumnOutput | FieldItem | RefObject[],
};

export type RowListInput = {
  type: string,
  name?: string | null,
  items?: ColumnInput | FieldItem | RefObject[],
};

export type RowListOutput = {
  type: string,
  name?: string | null,
  items?: ColumnOutput | FieldItem | RefObject[],
};

export type SchemaCost = {
  schema_id: string,
  api_cost: number,
  currency: string,
  request_count: number,
  automation_snapshot_json_schema: object,
  automation_snapshot_email?: string[] | null,
  automation_snapshot_name?: string[] | null,
};

export type SchemaExtractionRequest = {
  model?: string,
  temperature?: number,
  modality?: "text" | "image" | "native" | "image+text",
  image_settings?: ImageSettings,
  document: MIMEDataInput,
};

export type SingleFileEval = {
  eval_id: string,
  file_id: string,
  schema_id: string,
  schema_data_id?: string | null,
  dict_1: object,
  dict_2: object,
  annotation_props_1: AnnotationParameters,
  annotation_props_2: AnnotationParameters,
  created_at: Date,
  organization_id: string,
  hamming_similarity: object,
  jaccard_similarity: object,
  levenshtein_similarity: object,
};

export type StoredDBFile = {
  object?: string,
  id: string,
  filename: string,
  organization_id: string,
  embeddings?: number[] | null,
  created_at?: Date,
};

export type StoredSchema = {
  object?: string,
  created_at?: Date,
  json_schema?: object,
  strict?: boolean,
  organization_id: string,
  name: string,
  updated_at?: Date,
  data_id: string,
  id: string,
};

export type SubscriptionStatus = {
  status?: "active" | "canceled" | "incomplete" | "incomplete_expired" | "past_due" | "paused" | "trialing" | "unpaid",
  plan?: string,
  current_period_end?: number,
  plan_name?: string,
  cancel_at_period_end?: boolean,
};

export type TeamInvitation = {
  id: string,
  email: string,
  accepted: boolean,
};

export type TeamMember = {
  id: string,
  email: string,
  role: string,
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
  citations?: CitationCharLocation | CitationPageLocation | CitationContentBlockLocation[] | null,
  text: string,
  type: string,
};

export type TextBlockParam = {
  text: string,
  type: string,
  cache_control?: CacheControlEphemeralParam | null,
  citations?: CitationCharLocationParam | CitationPageLocationParam | CitationContentBlockLocationParam[] | null,
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

export type TokenPrice = {
  prompt: number,
  completion: number,
  cached_discount?: number,
};

export type ToolChoiceFunction = {
  name: string,
  type: string,
};

export type ToolChoiceTypes = {
  type: "file_search" | "web_search_preview" | "computer_use_preview" | "web_search_preview_2025_03_11",
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

export type UiParsedChatCompletionInput = {
  id: string,
  choices: UiParsedChoiceInput[],
  created: number,
  model: string,
  object: string,
  service_tier?: "scale" | "default" | null,
  system_fingerprint?: string | null,
  usage?: CompletionUsage | null,
  likelihoods: any,
  schema_validation_error?: ErrorDetail | null,
  likelihoods_source?: "consensus" | "log_probs",
  request_at?: Date | null,
  first_token_at?: Date | null,
  last_token_at?: Date | null,
};

export type UiParsedChatCompletionOutput = {
  id: string,
  choices: UiParsedChoiceOutput[],
  created: number,
  model: string,
  object: string,
  service_tier?: "scale" | "default" | null,
  system_fingerprint?: string | null,
  usage?: CompletionUsage | null,
  likelihoods: any,
  schema_validation_error?: ErrorDetail | null,
  likelihoods_source?: "consensus" | "log_probs",
  request_at?: Date | null,
  first_token_at?: Date | null,
  last_token_at?: Date | null,
};

export type UiParsedChoiceInput = {
  finish_reason?: "stop" | "length" | "tool_calls" | "content_filter" | "function_call" | null,
  index: number,
  logprobs?: ChoiceLogprobsInput | null,
  message: ParsedChatCompletionMessageInput,
};

export type UiParsedChoiceOutput = {
  finish_reason?: "stop" | "length" | "tool_calls" | "content_filter" | "function_call" | null,
  index: number,
  logprobs?: ChoiceLogprobsOutput | null,
  message: ParsedChatCompletionMessageOutput,
};

export type UpdateEmailDataRequest = {
  email_data: EmailDataInput,
  additional_documents?: AttachmentMIMEDataInput[] | null,
};

export type UpdateEndpointRequest = {
  webhook_url?: string | null,
  webhook_headers?: object | null,
  image_settings?: ImageSettings | null,
  modality?: "text" | "image" | "native" | "image+text" | null,
  model?: string | null,
  temperature?: number | null,
  json_schema?: object | null,
  reasoning_effort?: "low" | "medium" | "high" | null,
  need_validation?: boolean | null,
  n_consensus?: number | null,
  name?: string | null,
};

export type UpdateJobExecutionRequest = {
  status?: string | null,
  error?: string | null,
  additional_context?: object | null,
  job_executions_dict?: object | null,
  metadata?: object | null,
};

export type UpdateLinkRequest = {
  webhook_url?: string | null,
  webhook_headers?: object | null,
  image_settings?: ImageSettings | null,
  modality?: "text" | "image" | "native" | "image+text" | null,
  model?: string | null,
  temperature?: number | null,
  json_schema?: object | null,
  reasoning_effort?: "low" | "medium" | "high" | null,
  need_validation?: boolean | null,
  n_consensus?: number | null,
  name?: string | null,
  password?: string | null,
};

export type UpdateMailboxRequest = {
  webhook_url?: string | null,
  webhook_headers?: object | null,
  image_settings?: ImageSettings | null,
  modality?: "text" | "image" | "native" | "image+text" | null,
  model?: string | null,
  temperature?: number | null,
  json_schema?: object | null,
  reasoning_effort?: "low" | "medium" | "high" | null,
  need_validation?: boolean | null,
  n_consensus?: number | null,
  authorized_domains?: string[] | null,
  authorized_emails?: string[] | null,
};

export type UpdateOutlookRequest = {
  webhook_url?: string | null,
  webhook_headers?: object | null,
  image_settings?: ImageSettings | null,
  modality?: "text" | "image" | "native" | "image+text" | null,
  model?: string | null,
  temperature?: number | null,
  json_schema?: object | null,
  reasoning_effort?: "low" | "medium" | "high" | null,
  need_validation?: boolean | null,
  n_consensus?: number | null,
  name?: string | null,
  authorized_domains?: string[] | null,
  authorized_emails?: string[] | null,
  match_params?: MatchParams[] | null,
  fetch_params?: FetchParams[] | null,
  layout_schema?: LayoutInput | null,
};

export type UpdateSchemaEntry = {
  tag?: string | null,
  is_active?: boolean | null,
  is_current?: boolean | null,
};

export type UpdateTemplateRequest = {
  id: string,
  name?: string | null,
  json_schema?: object | null,
  python_code?: string | null,
  sample_document?: MIMEDataInput | null,
};

export type UpdateWorkflowExecutionRequest = {
  status?: string | null,
  error?: string | null,
  additional_context?: object | null,
  job_executions_dict?: object | null,
};

export type Usage = {
  cache_creation_input_tokens?: number | null,
  cache_read_input_tokens?: number | null,
  input_tokens: number,
  output_tokens: number,
};

export type UsageTimeSeries = {
  data: DataPoint[],
};

export type User = {
  object: string,
  id: string,
  email: string,
  first_name?: string | null,
  last_name?: string | null,
  email_verified: boolean,
  profile_picture_url?: string | null,
  created_at: string,
  updated_at: string,
  parameters: UserParameters,
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

export type WebSearchTool = {
  type: "web_search_preview" | "web_search_preview_2025_03_11",
  search_context_size?: "low" | "medium" | "high" | null,
  user_location?: UserLocation | null,
};

export type WebhookRequest = {
  completion: UiParsedChatCompletionInput,
  user?: string | null,
  file_payload: MIMEDataInput,
  metadata?: object | null,
};

export type WebhookSignature = {
  organization_id: string,
  updated_at: Date,
  signature: string,
};

export type WorkflowDefinition = {
  name: string,
  description: string,
  input_model: string | null,
  output_model: string | null,
  jobs: object,
  dependencies: object,
  transitive_dependency_closure?: object,
  execution_plan?: string[][],
};

export type WorkflowExecution = {
  id?: string,
  workflow: WorkflowDefinition,
  organization_id: string,
  status?: WorkflowStatus,
  job_executions_dict?: object,
  context?: object,
  error?: string | null,
  created_at?: Date,
  updated_at?: Date,
  started_at?: Date | null,
  completed_at?: Date | null,
};

export type WorkflowStatus = "pending" | "running" | "waiting" | "stalled" | "succeeded" | "failed";

export type DocumentExtractRequest = {
  document: MIMEDataInput,
  modality: "text" | "image" | "native" | "image+text",
  image_settings?: ImageSettings,
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
  content: TextBlock | ToolUseBlock | ThinkingBlock | RedactedThinkingBlock[],
  model: "claude-3-7-sonnet-latest" | "claude-3-7-sonnet-20250219" | "claude-3-5-haiku-latest" | "claude-3-5-haiku-20241022" | "claude-3-5-sonnet-latest" | "claude-3-5-sonnet-20241022" | "claude-3-5-sonnet-20240620" | "claude-3-opus-latest" | "claude-3-opus-20240229" | "claude-3-sonnet-20240229" | "claude-3-haiku-20240307" | "claude-2.1" | "claude-2.0" | string,
  role: string,
  stop_reason?: "end_turn" | "max_tokens" | "stop_sequence" | "tool_use" | null,
  stop_sequence?: string | null,
  type: string,
  usage: Usage,
};

export type CubeServerServicesCustomBertCubemimedataAttachmentMetadata = {
  is_inline?: boolean,
  inline_cid?: string | null,
  url?: string | null,
  display_metadata?: DisplayMetadata | null,
  ocr?: OCR | null,
  source?: string | null,
};

export type CubeServerServicesCustomBertCubemimedataBaseMIMEData = {
  id: string,
  name: string,
  size: number,
  mime_type: string,
  metadata: CubeServerServicesCustomBertCubemimedataAttachmentMetadata,
};

export type CubeServerServicesCustomBertCubemimedataMIMEData = {
  id: string,
  name: string,
  size: number,
  mime_type: string,
  metadata: CubeServerServicesCustomBertCubemimedataAttachmentMetadata,
  content: string,
};

export type CubeServerServicesCustomBertRoutesMatchResultModel = {
  record: object,
  similarity: number,
};

export type CubeServerServicesCustomGroussardRoutesMatchResultModel = {
  record: object,
  similarity: number,
};

export type CubeServerServicesCustomSampleRoutesMatchResultModel = {
  record: object,
  similarity: number,
};

export type CubeServerServicesCustomSinariRoutesMatchResultModel = {
  record: object,
  similarity: number,
};

export type CubeServerServicesV1SchemasDefaultTemplatesRoutesCreateTemplateRequest = {
  id?: string,
  name: string,
  json_schema: object,
  python_code?: string | null,
  sample_document?: MIMEDataInput | null,
};

export type CubeServerServicesV1SchemasTemplatesRoutesCreateTemplateRequest = {
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

export type OpenaiTypesResponsesResponseInputParamMessage = {
  content: ResponseInputTextParam | ResponseInputImageParam | ResponseInputFileParam[],
  role: "user" | "system" | "developer",
  status?: "in_progress" | "completed" | "incomplete",
  type?: string,
};

export type OpenaiTypesResponsesResponseOutputTextAnnotationFileCitation = {
  file_id: string,
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

export type OpenaiTypesResponsesResponseOutputTextParamAnnotationFileCitation = {
  file_id: string,
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

export type OpenaiTypesResponsesResponseReasoningItemSummary = {
  text: string,
  type: string,
};

export type OpenaiTypesResponsesResponseReasoningItemParamSummary = {
  text: string,
  type: string,
};

export type UiformTypesMimeAttachmentMetadata = {
  is_inline?: boolean,
  inline_cid?: string | null,
  url?: string | null,
  display_metadata?: DisplayMetadata | null,
  source?: string | null,
};

export type UiformTypesMimeBaseMIMEData = {
  filename: string,
  url: string,
};

export type UiformTypesMimeMIMEData = {
  filename: string,
  url: string,
};

