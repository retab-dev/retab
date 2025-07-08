import * as z from 'zod'
import { DateOrISO } from '@/client';

export const ZAPIKeyCreate = z.lazy(() => z.object({
  name: z.string(),
  description: z.union([z.string(), z.null()]).optional(),
}));
export type APIKeyCreate = z.infer<typeof ZAPIKeyCreate>;

export const ZAPIKeyInfo = z.lazy(() => z.object({
  id: z.string(),
  name: z.string(),
  created_at: DateOrISO,
  is_active: z.boolean(),
}));
export type APIKeyInfo = z.infer<typeof ZAPIKeyInfo>;

export const ZAPIKeyResponse = z.lazy(() => z.object({
  key: z.string(),
  name: z.union([z.string(), z.null()]),
  created_at: DateOrISO,
  organization_id: z.string(),
}));
export type APIKeyResponse = z.infer<typeof ZAPIKeyResponse>;

export const ZAddDomainRequest = z.lazy(() => z.object({
  domain: z.string(),
}));
export type AddDomainRequest = z.infer<typeof ZAddDomainRequest>;

export const ZAlignDictsRequest = z.lazy(() => z.object({
  list_dicts: z.array(z.object({})),
  min_support_ratio: z.number().optional(),
  reference_idx: z.union([z.number(), z.null()]).optional(),
}));
export type AlignDictsRequest = z.infer<typeof ZAlignDictsRequest>;

export const ZAlignDictsResponse = z.lazy(() => z.object({
  aligned_dicts: z.array(z.object({})),
  key_mapping_all: z.object({}),
}));
export type AlignDictsResponse = z.infer<typeof ZAlignDictsResponse>;

export const ZAmount = z.lazy(() => z.object({
  value: z.number(),
  currency: z.string(),
}));
export type Amount = z.infer<typeof ZAmount>;

export const ZAnalysesChartResponse = z.lazy(() => z.object({
  date: DateOrISO,
  count: z.number(),
}));
export type AnalysesChartResponse = z.infer<typeof ZAnalysesChartResponse>;

export const ZAnnotationInput = z.lazy(() => z.object({
  type: z.string(),
  url_citation: ZOpenaiTypesChatChatCompletionMessageAnnotationURLCitation,
}));
export type AnnotationInput = z.infer<typeof ZAnnotationInput>;

export const ZAnnotationOutput = z.lazy(() => z.object({
  type: z.string(),
  url_citation: ZAnnotationURLCitationOutput,
}));
export type AnnotationOutput = z.infer<typeof ZAnnotationOutput>;

export const ZAnnotationURLCitationOutput = z.lazy(() => z.object({
  end_index: z.number(),
  start_index: z.number(),
  title: z.string(),
  url: z.string(),
}));
export type AnnotationURLCitationOutput = z.infer<typeof ZAnnotationURLCitationOutput>;

export const ZArticle = z.lazy(() => z.object({
  id: z.string().optional(),
  title: z.string(),
  type: z.string(),
  slug: z.string(),
  status: z.string(),
  tags: z.array(z.string()),
  summary: z.string(),
  coverImage: z.string(),
  author_id: z.enum(["louis-de-benoist", "sacha-ichbiah", "victor-plaisance"]),
  date: z.string().optional(),
}));
export type Article = z.infer<typeof ZArticle>;

export const ZAttachmentMIMEDataInput = z.lazy(() => z.object({
  filename: z.string(),
  url: z.string(),
  metadata: ZAttachmentMetadataInput.optional(),
}));
export type AttachmentMIMEDataInput = z.infer<typeof ZAttachmentMIMEDataInput>;

export const ZAttachmentMIMEDataOutput = z.lazy(() => z.object({
  filename: z.string(),
  url: z.string(),
  metadata: ZRetabTypesMimeAttachmentMetadata.optional(),
}));
export type AttachmentMIMEDataOutput = z.infer<typeof ZAttachmentMIMEDataOutput>;

export const ZAttachmentMetadataInput = z.lazy(() => z.object({
  is_inline: z.boolean().optional(),
  inline_cid: z.union([z.string(), z.null()]).optional(),
  source: z.union([z.string(), z.null()]).optional(),
}));
export type AttachmentMetadataInput = z.infer<typeof ZAttachmentMetadataInput>;

export const ZAudio = z.lazy(() => z.object({
  id: z.string(),
}));
export type Audio = z.infer<typeof ZAudio>;

export const ZAutomationConfig = z.lazy(() => z.object({
  id: z.string().optional(),
  name: z.string(),
  processor_id: z.string(),
  updated_at: DateOrISO.optional(),
  default_language: z.string().optional(),
  webhook_url: z.string(),
  webhook_headers: z.object({}).optional(),
  need_validation: z.boolean().optional(),
  object: z.string(),
}));
export type AutomationConfig = z.infer<typeof ZAutomationConfig>;

export const ZAutomationDecisionRequest = z.lazy(() => z.object({
  email_data: ZEmailDataInput,
  schema_id: z.string(),
}));
export type AutomationDecisionRequest = z.infer<typeof ZAutomationDecisionRequest>;

export const ZAutomationDecisionResponse = z.lazy(() => z.object({
  decision: z.enum(["automated", "human_validation"]),
  cosine_threshold: z.number(),
  score_threshold: z.number(),
  results_threshold: z.number(),
  score_std_penalty: z.number(),
  average_cosine: z.number(),
  average_score: z.number(),
  std_score: z.number(),
  total_results: z.number(),
}));
export type AutomationDecisionResponse = z.infer<typeof ZAutomationDecisionResponse>;

export const ZAutomationLog = z.lazy(() => z.object({
  object: z.string().optional(),
  id: z.string().optional(),
  user_email: z.union([z.string(), z.null()]),
  organization_id: z.string(),
  created_at: DateOrISO.optional(),
  automation_snapshot: ZAutomationConfig,
  completion: z.union([ZRetabParsedChatCompletionOutput, ZChatCompletionOutput]),
  file_metadata: z.union([ZRetabTypesMimeBaseMIMEData, z.null()]),
  external_request_log: z.union([ZExternalRequestLog, z.null()]),
  extraction_id: z.union([z.string(), z.null()]).optional(),
  api_cost: z.union([ZAmount, z.null()]),
  cost_breakdown: z.union([ZCostBreakdown, z.null()]),
}));
export type AutomationLog = z.infer<typeof ZAutomationLog>;

export const ZBase64ImageSourceParam = z.lazy(() => z.object({
  data: z.union([z.string(), z.string()]),
  media_type: z.enum(["image/jpeg", "image/png", "image/gif", "image/webp"]),
  type: z.string(),
}));
export type Base64ImageSourceParam = z.infer<typeof ZBase64ImageSourceParam>;

export const ZBase64PDFSourceParam = z.lazy(() => z.object({
  data: z.union([z.string(), z.string()]),
  media_type: z.string(),
  type: z.string(),
}));
export type Base64PDFSourceParam = z.infer<typeof ZBase64PDFSourceParam>;

export const ZBaseEmailData = z.lazy(() => z.object({
  id: z.string(),
  tree_id: z.string(),
  subject: z.union([z.string(), z.null()]).optional(),
  body_plain: z.union([z.string(), z.null()]).optional(),
  body_html: z.union([z.string(), z.null()]).optional(),
  sender: ZEmailAddressData,
  recipients_to: z.array(ZEmailAddressData),
  recipients_cc: z.array(ZEmailAddressData).optional(),
  recipients_bcc: z.array(ZEmailAddressData).optional(),
  sent_at: DateOrISO,
  received_at: z.union([DateOrISO, z.null()]).optional(),
  in_reply_to: z.union([z.string(), z.null()]).optional(),
  references: z.array(z.string()).optional(),
  headers: z.object({}).optional(),
  url: z.union([z.string(), z.null()]).optional(),
  attachments: z.array(ZMainServerServicesCustomBertCubemimedataBaseMIMEData).optional(),
}));
export type BaseEmailData = z.infer<typeof ZBaseEmailData>;

export const ZBodyConvertMsgToEmailModelV1ProcessorsAutomationsOutlookConvertMsgToEmailModelPost = z.lazy(() => z.object({
  file: z.instanceof(File),
}));
export type BodyConvertMsgToEmailModelV1ProcessorsAutomationsOutlookConvertMsgToEmailModelPost = z.infer<typeof ZBodyConvertMsgToEmailModelV1ProcessorsAutomationsOutlookConvertMsgToEmailModelPost>;

export const ZBodyConvertToEmailDataAndUploadFileV1ProcessorsAutomationsOutlookConvertToEmailDataAndUploadFilePost = z.lazy(() => z.object({
  file: z.instanceof(File),
}));
export type BodyConvertToEmailDataAndUploadFileV1ProcessorsAutomationsOutlookConvertToEmailDataAndUploadFilePost = z.infer<typeof ZBodyConvertToEmailDataAndUploadFileV1ProcessorsAutomationsOutlookConvertToEmailDataAndUploadFilePost>;

export const ZBodyCreateFileInternalDbFilesPost = z.lazy(() => z.object({
  file: z.instanceof(File),
}));
export type BodyCreateFileInternalDbFilesPost = z.infer<typeof ZBodyCreateFileInternalDbFilesPost>;

export const ZBodyCreateFilesInternalDbFilesBatchPost = z.lazy(() => z.object({
  files: z.array(z.instanceof(File)),
}));
export type BodyCreateFilesInternalDbFilesBatchPost = z.infer<typeof ZBodyCreateFilesInternalDbFilesBatchPost>;

export const ZBodyHandleEndpointProcessingV1EndpointsEndpointIdPost = z.lazy(() => z.object({
  document: z.instanceof(File),
}));
export type BodyHandleEndpointProcessingV1EndpointsEndpointIdPost = z.infer<typeof ZBodyHandleEndpointProcessingV1EndpointsEndpointIdPost>;

export const ZBodyHandleEndpointProcessingV1ProcessorsAutomationsEndpointsProcessEndpointIdPost = z.lazy(() => z.object({
  file: z.instanceof(File),
  identity: ZIdentity.optional(),
}));
export type BodyHandleEndpointProcessingV1ProcessorsAutomationsEndpointsProcessEndpointIdPost = z.infer<typeof ZBodyHandleEndpointProcessingV1ProcessorsAutomationsEndpointsProcessEndpointIdPost>;

export const ZBodyHandleLinkWebhookV1ProcessorsAutomationsLinksParseLinkIdPost = z.lazy(() => z.object({
  file: z.instanceof(File),
}));
export type BodyHandleLinkWebhookV1ProcessorsAutomationsLinksParseLinkIdPost = z.infer<typeof ZBodyHandleLinkWebhookV1ProcessorsAutomationsLinksParseLinkIdPost>;

export const ZBodyImportAnnotationsCsvV1EvalsIoEvaluationIdImportAnnotationsCsvPost = z.lazy(() => z.object({
  csv_file: z.instanceof(File),
}));
export type BodyImportAnnotationsCsvV1EvalsIoEvaluationIdImportAnnotationsCsvPost = z.infer<typeof ZBodyImportAnnotationsCsvV1EvalsIoEvaluationIdImportAnnotationsCsvPost>;

export const ZBodyImportAnnotationsCsvV1EvaluationsIoEvaluationIdImportAnnotationsCsvPost = z.lazy(() => z.object({
  csv_file: z.instanceof(File),
}));
export type BodyImportAnnotationsCsvV1EvaluationsIoEvaluationIdImportAnnotationsCsvPost = z.infer<typeof ZBodyImportAnnotationsCsvV1EvaluationsIoEvaluationIdImportAnnotationsCsvPost>;

export const ZBodyImportDocumentsV1EvalsIoEvaluationIdImportDocumentsPost = z.lazy(() => z.object({
  jsonl_file: z.instanceof(File),
}));
export type BodyImportDocumentsV1EvalsIoEvaluationIdImportDocumentsPost = z.infer<typeof ZBodyImportDocumentsV1EvalsIoEvaluationIdImportDocumentsPost>;

export const ZBodyImportDocumentsV1EvaluationsIoEvaluationIdImportDocumentsPost = z.lazy(() => z.object({
  jsonl_file: z.instanceof(File),
}));
export type BodyImportDocumentsV1EvaluationsIoEvaluationIdImportDocumentsPost = z.infer<typeof ZBodyImportDocumentsV1EvaluationsIoEvaluationIdImportDocumentsPost>;

export const ZBodySubmitToProcessorV1ProcessorsProcessorIdSubmitPost = z.lazy(() => z.object({
  document: z.union([z.instanceof(File), z.null()]).optional(),
  documents: z.union([z.array(z.instanceof(File)), z.null()]).optional(),
  temperature: z.union([z.number(), z.null()]).optional(),
  stream: z.boolean().optional(),
  seed: z.union([z.number(), z.null()]).optional(),
  store: z.boolean().optional(),
  test_exception: z.union([z.string(), z.null()]).optional(),
}));
export type BodySubmitToProcessorV1ProcessorsProcessorIdSubmitPost = z.infer<typeof ZBodySubmitToProcessorV1ProcessorsProcessorIdSubmitPost>;

export const ZBodySubmitToProcessorV1ProcessorsProcessorIdSubmitStreamPost = z.lazy(() => z.object({
  document: z.union([z.instanceof(File), z.null()]).optional(),
  documents: z.union([z.array(z.instanceof(File)), z.null()]).optional(),
  temperature: z.union([z.number(), z.null()]).optional(),
  stream: z.boolean().optional(),
  seed: z.union([z.number(), z.null()]).optional(),
  store: z.boolean().optional(),
  test_exception: z.union([z.string(), z.null()]).optional(),
}));
export type BodySubmitToProcessorV1ProcessorsProcessorIdSubmitStreamPost = z.infer<typeof ZBodySubmitToProcessorV1ProcessorsProcessorIdSubmitStreamPost>;

export const ZBodyTestDocumentUploadV1ProcessorsAutomationsTestsUploadAutomationIdPost = z.lazy(() => z.object({
  request: z.union([ZDocumentUploadRequest, z.null()]).optional(),
  file: z.instanceof(File).optional(),
}));
export type BodyTestDocumentUploadV1ProcessorsAutomationsTestsUploadAutomationIdPost = z.infer<typeof ZBodyTestDocumentUploadV1ProcessorsAutomationsTestsUploadAutomationIdPost>;

export const ZBranding = z.lazy(() => z.object({
  button_color: z.string().optional(),
  text_color: z.string().optional(),
  button_text_color: z.string().optional(),
  page_background_color: z.string().optional(),
  logo_url: z.string().optional(),
  icon_url: z.string().optional(),
  company_name: z.string().optional(),
  website_url: z.string().optional(),
}));
export type Branding = z.infer<typeof ZBranding>;

export const ZBrandingUpdateRequest = z.lazy(() => z.object({
  button_color: z.union([z.string(), z.null()]).optional(),
  text_color: z.union([z.string(), z.null()]).optional(),
  button_text_color: z.union([z.string(), z.null()]).optional(),
  page_background_color: z.union([z.string(), z.null()]).optional(),
  logo_url: z.union([z.string(), z.null()]).optional(),
  icon_url: z.union([z.string(), z.null()]).optional(),
  company_name: z.union([z.string(), z.null()]).optional(),
  website_url: z.union([z.string(), z.null()]).optional(),
}));
export type BrandingUpdateRequest = z.infer<typeof ZBrandingUpdateRequest>;

export const ZCacheControlEphemeralParam = z.lazy(() => z.object({
  type: z.string(),
}));
export type CacheControlEphemeralParam = z.infer<typeof ZCacheControlEphemeralParam>;

export const ZCachePreloadRequest = z.lazy(() => z.object({
  model: z.string().optional(),
  temperature: z.number().optional(),
  modality: z.enum(["text", "image", "native", "image+text"]).optional(),
  stream: z.boolean().optional(),
}));
export type CachePreloadRequest = z.infer<typeof ZCachePreloadRequest>;

export const ZChatCompletionInput = z.lazy(() => z.object({
  id: z.string(),
  choices: z.array(ZChoiceInput),
  created: z.number(),
  model: z.string(),
  object: z.string(),
  service_tier: z.union([z.enum(["auto", "default", "flex", "scale", "priority"]), z.null()]).optional(),
  system_fingerprint: z.union([z.string(), z.null()]).optional(),
  usage: z.union([ZCompletionUsage, z.null()]).optional(),
}));
export type ChatCompletionInput = z.infer<typeof ZChatCompletionInput>;

export const ZChatCompletionOutput = z.lazy(() => z.object({
  id: z.string(),
  choices: z.array(ZChoiceOutput),
  created: z.number(),
  model: z.string(),
  object: z.string(),
  service_tier: z.union([z.enum(["auto", "default", "flex", "scale", "priority"]), z.null()]).optional(),
  system_fingerprint: z.union([z.string(), z.null()]).optional(),
  usage: z.union([ZCompletionUsage, z.null()]).optional(),
}));
export type ChatCompletionOutput = z.infer<typeof ZChatCompletionOutput>;

export const ZChatCompletionAssistantMessageParam = z.lazy(() => z.object({
  role: z.string(),
  audio: z.union([ZAudio, z.null()]).optional(),
  content: z.union([z.string(), z.array(z.union([ZChatCompletionContentPartTextParam, ZChatCompletionContentPartRefusalParam])), z.null()]).optional(),
  function_call: z.union([ZOpenaiTypesChatChatCompletionAssistantMessageParamFunctionCall, z.null()]).optional(),
  name: z.string().optional(),
  refusal: z.union([z.string(), z.null()]).optional(),
  tool_calls: z.array(ZChatCompletionMessageToolCallParam).optional(),
}));
export type ChatCompletionAssistantMessageParam = z.infer<typeof ZChatCompletionAssistantMessageParam>;

export const ZChatCompletionAudio = z.lazy(() => z.object({
  id: z.string(),
  data: z.string(),
  expires_at: z.number(),
  transcript: z.string(),
}));
export type ChatCompletionAudio = z.infer<typeof ZChatCompletionAudio>;

export const ZChatCompletionContentPartImageParam = z.lazy(() => z.object({
  image_url: ZImageURL,
  type: z.string(),
}));
export type ChatCompletionContentPartImageParam = z.infer<typeof ZChatCompletionContentPartImageParam>;

export const ZChatCompletionContentPartInputAudioParam = z.lazy(() => z.object({
  input_audio: ZInputAudio,
  type: z.string(),
}));
export type ChatCompletionContentPartInputAudioParam = z.infer<typeof ZChatCompletionContentPartInputAudioParam>;

export const ZChatCompletionContentPartRefusalParam = z.lazy(() => z.object({
  refusal: z.string(),
  type: z.string(),
}));
export type ChatCompletionContentPartRefusalParam = z.infer<typeof ZChatCompletionContentPartRefusalParam>;

export const ZChatCompletionContentPartTextParam = z.lazy(() => z.object({
  text: z.string(),
  type: z.string(),
}));
export type ChatCompletionContentPartTextParam = z.infer<typeof ZChatCompletionContentPartTextParam>;

export const ZChatCompletionDeveloperMessageParam = z.lazy(() => z.object({
  content: z.union([z.string(), z.array(ZChatCompletionContentPartTextParam)]),
  role: z.string(),
  name: z.string().optional(),
}));
export type ChatCompletionDeveloperMessageParam = z.infer<typeof ZChatCompletionDeveloperMessageParam>;

export const ZChatCompletionFunctionMessageParam = z.lazy(() => z.object({
  content: z.union([z.string(), z.null()]),
  name: z.string(),
  role: z.string(),
}));
export type ChatCompletionFunctionMessageParam = z.infer<typeof ZChatCompletionFunctionMessageParam>;

export const ZChatCompletionMessageInput = z.lazy(() => z.object({
  content: z.union([z.string(), z.null()]).optional(),
  refusal: z.union([z.string(), z.null()]).optional(),
  role: z.string(),
  annotations: z.union([z.array(ZAnnotationInput), z.null()]).optional(),
  audio: z.union([ZChatCompletionAudio, z.null()]).optional(),
  function_call: z.union([ZOpenaiTypesChatChatCompletionMessageFunctionCall, z.null()]).optional(),
  tool_calls: z.union([z.array(ZChatCompletionMessageToolCallInput), z.null()]).optional(),
}));
export type ChatCompletionMessageInput = z.infer<typeof ZChatCompletionMessageInput>;

export const ZChatCompletionMessageOutput = z.lazy(() => z.object({
  content: z.union([z.string(), z.null()]).optional(),
  refusal: z.union([z.string(), z.null()]).optional(),
  role: z.string(),
  annotations: z.union([z.array(ZAnnotationOutput), z.null()]).optional(),
  audio: z.union([ZChatCompletionAudio, z.null()]).optional(),
  function_call: z.union([ZFunctionCallOutput, z.null()]).optional(),
  tool_calls: z.union([z.array(ZChatCompletionMessageToolCallOutput), z.null()]).optional(),
}));
export type ChatCompletionMessageOutput = z.infer<typeof ZChatCompletionMessageOutput>;

export const ZChatCompletionMessageToolCallInput = z.lazy(() => z.object({
  id: z.string(),
  function: ZOpenaiTypesChatChatCompletionMessageToolCallFunction,
  type: z.string(),
}));
export type ChatCompletionMessageToolCallInput = z.infer<typeof ZChatCompletionMessageToolCallInput>;

export const ZChatCompletionMessageToolCallOutput = z.lazy(() => z.object({
  id: z.string(),
  function: ZFunctionOutput,
  type: z.string(),
}));
export type ChatCompletionMessageToolCallOutput = z.infer<typeof ZChatCompletionMessageToolCallOutput>;

export const ZChatCompletionMessageToolCallParam = z.lazy(() => z.object({
  id: z.string(),
  function: ZOpenaiTypesChatChatCompletionMessageToolCallParamFunction,
  type: z.string(),
}));
export type ChatCompletionMessageToolCallParam = z.infer<typeof ZChatCompletionMessageToolCallParam>;

export const ZChatCompletionRetabMessageInput = z.lazy(() => z.object({
  role: z.enum(["user", "system", "assistant", "developer"]),
  content: z.union([z.string(), z.array(z.union([ZChatCompletionContentPartTextParam, ZChatCompletionContentPartImageParam, ZChatCompletionContentPartInputAudioParam, ZFile]))]),
}));
export type ChatCompletionRetabMessageInput = z.infer<typeof ZChatCompletionRetabMessageInput>;

export const ZChatCompletionRetabMessageOutput = z.lazy(() => z.object({
  role: z.enum(["user", "system", "assistant", "developer"]),
  content: z.union([z.string(), z.array(z.union([ZChatCompletionContentPartTextParam, ZChatCompletionContentPartImageParam, ZChatCompletionContentPartInputAudioParam, ZFile]))]),
}));
export type ChatCompletionRetabMessageOutput = z.infer<typeof ZChatCompletionRetabMessageOutput>;

export const ZChatCompletionSystemMessageParam = z.lazy(() => z.object({
  content: z.union([z.string(), z.array(ZChatCompletionContentPartTextParam)]),
  role: z.string(),
  name: z.string().optional(),
}));
export type ChatCompletionSystemMessageParam = z.infer<typeof ZChatCompletionSystemMessageParam>;

export const ZChatCompletionTokenLogprob = z.lazy(() => z.object({
  token: z.string(),
  bytes: z.union([z.array(z.number()), z.null()]).optional(),
  logprob: z.number(),
  top_logprobs: z.array(ZTopLogprob),
}));
export type ChatCompletionTokenLogprob = z.infer<typeof ZChatCompletionTokenLogprob>;

export const ZChatCompletionToolMessageParam = z.lazy(() => z.object({
  content: z.union([z.string(), z.array(ZChatCompletionContentPartTextParam)]),
  role: z.string(),
  tool_call_id: z.string(),
}));
export type ChatCompletionToolMessageParam = z.infer<typeof ZChatCompletionToolMessageParam>;

export const ZChatCompletionUserMessageParam = z.lazy(() => z.object({
  content: z.union([z.string(), z.array(z.union([ZChatCompletionContentPartTextParam, ZChatCompletionContentPartImageParam, ZChatCompletionContentPartInputAudioParam, ZFile]))]),
  role: z.string(),
  name: z.string().optional(),
}));
export type ChatCompletionUserMessageParam = z.infer<typeof ZChatCompletionUserMessageParam>;

export const ZChoiceInput = z.lazy(() => z.object({
  finish_reason: z.enum(["stop", "length", "tool_calls", "content_filter", "function_call"]),
  index: z.number(),
  logprobs: z.union([ZChoiceLogprobsInput, z.null()]).optional(),
  message: ZChatCompletionMessageInput,
}));
export type ChoiceInput = z.infer<typeof ZChoiceInput>;

export const ZChoiceOutput = z.lazy(() => z.object({
  finish_reason: z.enum(["stop", "length", "tool_calls", "content_filter", "function_call"]),
  index: z.number(),
  logprobs: z.union([ZChoiceLogprobsOutput, z.null()]).optional(),
  message: ZChatCompletionMessageOutput,
}));
export type ChoiceOutput = z.infer<typeof ZChoiceOutput>;

export const ZChoiceDeltaFunctionCall = z.lazy(() => z.object({
  arguments: z.union([z.string(), z.null()]).optional(),
  name: z.union([z.string(), z.null()]).optional(),
}));
export type ChoiceDeltaFunctionCall = z.infer<typeof ZChoiceDeltaFunctionCall>;

export const ZChoiceDeltaToolCall = z.lazy(() => z.object({
  index: z.number(),
  id: z.union([z.string(), z.null()]).optional(),
  function: z.union([ZChoiceDeltaToolCallFunction, z.null()]).optional(),
  type: z.union([z.string(), z.null()]).optional(),
}));
export type ChoiceDeltaToolCall = z.infer<typeof ZChoiceDeltaToolCall>;

export const ZChoiceDeltaToolCallFunction = z.lazy(() => z.object({
  arguments: z.union([z.string(), z.null()]).optional(),
  name: z.union([z.string(), z.null()]).optional(),
}));
export type ChoiceDeltaToolCallFunction = z.infer<typeof ZChoiceDeltaToolCallFunction>;

export const ZChoiceLogprobsInput = z.lazy(() => z.object({
  content: z.union([z.array(ZChatCompletionTokenLogprob), z.null()]).optional(),
  refusal: z.union([z.array(ZChatCompletionTokenLogprob), z.null()]).optional(),
}));
export type ChoiceLogprobsInput = z.infer<typeof ZChoiceLogprobsInput>;

export const ZChoiceLogprobsOutput = z.lazy(() => z.object({
  content: z.union([z.array(ZChatCompletionTokenLogprob), z.null()]).optional(),
  refusal: z.union([z.array(ZChatCompletionTokenLogprob), z.null()]).optional(),
}));
export type ChoiceLogprobsOutput = z.infer<typeof ZChoiceLogprobsOutput>;

export const ZCitationCharLocation = z.lazy(() => z.object({
  cited_text: z.string(),
  document_index: z.number(),
  document_title: z.union([z.string(), z.null()]).optional(),
  end_char_index: z.number(),
  start_char_index: z.number(),
  type: z.string(),
}));
export type CitationCharLocation = z.infer<typeof ZCitationCharLocation>;

export const ZCitationCharLocationParam = z.lazy(() => z.object({
  cited_text: z.string(),
  document_index: z.number(),
  document_title: z.union([z.string(), z.null()]),
  end_char_index: z.number(),
  start_char_index: z.number(),
  type: z.string(),
}));
export type CitationCharLocationParam = z.infer<typeof ZCitationCharLocationParam>;

export const ZCitationContentBlockLocation = z.lazy(() => z.object({
  cited_text: z.string(),
  document_index: z.number(),
  document_title: z.union([z.string(), z.null()]).optional(),
  end_block_index: z.number(),
  start_block_index: z.number(),
  type: z.string(),
}));
export type CitationContentBlockLocation = z.infer<typeof ZCitationContentBlockLocation>;

export const ZCitationContentBlockLocationParam = z.lazy(() => z.object({
  cited_text: z.string(),
  document_index: z.number(),
  document_title: z.union([z.string(), z.null()]),
  end_block_index: z.number(),
  start_block_index: z.number(),
  type: z.string(),
}));
export type CitationContentBlockLocationParam = z.infer<typeof ZCitationContentBlockLocationParam>;

export const ZCitationPageLocation = z.lazy(() => z.object({
  cited_text: z.string(),
  document_index: z.number(),
  document_title: z.union([z.string(), z.null()]).optional(),
  end_page_number: z.number(),
  start_page_number: z.number(),
  type: z.string(),
}));
export type CitationPageLocation = z.infer<typeof ZCitationPageLocation>;

export const ZCitationPageLocationParam = z.lazy(() => z.object({
  cited_text: z.string(),
  document_index: z.number(),
  document_title: z.union([z.string(), z.null()]),
  end_page_number: z.number(),
  start_page_number: z.number(),
  type: z.string(),
}));
export type CitationPageLocationParam = z.infer<typeof ZCitationPageLocationParam>;

export const ZCitationWebSearchResultLocationParam = z.lazy(() => z.object({
  cited_text: z.string(),
  encrypted_index: z.string(),
  title: z.union([z.string(), z.null()]),
  type: z.string(),
  url: z.string(),
}));
export type CitationWebSearchResultLocationParam = z.infer<typeof ZCitationWebSearchResultLocationParam>;

export const ZCitationsConfigParam = z.lazy(() => z.object({
  enabled: z.boolean().optional(),
}));
export type CitationsConfigParam = z.infer<typeof ZCitationsConfigParam>;

export const ZCitationsWebSearchResultLocation = z.lazy(() => z.object({
  cited_text: z.string(),
  encrypted_index: z.string(),
  title: z.union([z.string(), z.null()]).optional(),
  type: z.string(),
  url: z.string(),
}));
export type CitationsWebSearchResultLocation = z.infer<typeof ZCitationsWebSearchResultLocation>;

export const ZCodeInterpreter = z.lazy(() => z.object({
  container: z.union([z.string(), ZCodeInterpreterContainerCodeInterpreterToolAuto]),
  type: z.string(),
}));
export type CodeInterpreter = z.infer<typeof ZCodeInterpreter>;

export const ZCodeInterpreterContainerCodeInterpreterToolAuto = z.lazy(() => z.object({
  type: z.string(),
  file_ids: z.union([z.array(z.string()), z.null()]).optional(),
}));
export type CodeInterpreterContainerCodeInterpreterToolAuto = z.infer<typeof ZCodeInterpreterContainerCodeInterpreterToolAuto>;

export const ZComparisonFilter = z.lazy(() => z.object({
  key: z.string(),
  type: z.enum(["eq", "ne", "gt", "gte", "lt", "lte"]),
  value: z.union([z.string(), z.number(), z.boolean()]),
}));
export type ComparisonFilter = z.infer<typeof ZComparisonFilter>;

export const ZComparisonRequest = z.lazy(() => z.object({
  dict1: z.object({}),
  dict2: z.object({}),
  metric: z.enum(["levenshtein_similarity", "jaccard_similarity", "hamming_similarity"]).optional(),
}));
export type ComparisonRequest = z.infer<typeof ZComparisonRequest>;

export const ZComparisonResponse = z.lazy(() => z.object({
  comparison_results: z.object({}),
  metric_used: z.enum(["levenshtein_similarity", "jaccard_similarity", "hamming_similarity"]),
}));
export type ComparisonResponse = z.infer<typeof ZComparisonResponse>;

export const ZCompletionTokensDetails = z.lazy(() => z.object({
  accepted_prediction_tokens: z.union([z.number(), z.null()]).optional(),
  audio_tokens: z.union([z.number(), z.null()]).optional(),
  reasoning_tokens: z.union([z.number(), z.null()]).optional(),
  rejected_prediction_tokens: z.union([z.number(), z.null()]).optional(),
}));
export type CompletionTokensDetails = z.infer<typeof ZCompletionTokensDetails>;

export const ZCompletionUsage = z.lazy(() => z.object({
  completion_tokens: z.number(),
  prompt_tokens: z.number(),
  total_tokens: z.number(),
  completion_tokens_details: z.union([ZCompletionTokensDetails, z.null()]).optional(),
  prompt_tokens_details: z.union([ZPromptTokensDetails, z.null()]).optional(),
}));
export type CompletionUsage = z.infer<typeof ZCompletionUsage>;

export const ZCompoundFilter = z.lazy(() => z.object({
  filters: z.array(z.union([ZComparisonFilter, z.any()])),
  type: z.enum(["and", "or"]),
}));
export type CompoundFilter = z.infer<typeof ZCompoundFilter>;

export const ZComputeDictSimilarityRequest = z.lazy(() => z.object({
  dict1: z.object({}),
  dict2: z.object({}),
  string_similarity_method: z.enum(["levenshtein", "jaccard", "hamming", "embeddings"]),
  min_support_ratio: z.number().optional(),
}));
export type ComputeDictSimilarityRequest = z.infer<typeof ZComputeDictSimilarityRequest>;

export const ZComputeDictSimilarityResponse = z.lazy(() => z.object({
  flat_reference_elements: z.object({}),
  per_element_similarity: z.object({}),
  total_similarity: z.number(),
  aligned_flat_reference_elements: z.object({}),
  aligned_per_element_similarity: z.object({}),
  aligned_total_similarity: z.number(),
}));
export type ComputeDictSimilarityResponse = z.infer<typeof ZComputeDictSimilarityResponse>;

export const ZComputeFieldLocationsRequest = z.lazy(() => z.object({
  ocr_file_id: z.string(),
  ocr_result: ZOCRInput,
  data: z.object({}),
  data_type: z.enum(["ground_truth", "extraction"]),
}));
export type ComputeFieldLocationsRequest = z.infer<typeof ZComputeFieldLocationsRequest>;

export const ZComputeFieldLocationsResponse = z.lazy(() => z.object({
  status: z.enum(["success", "error"]),
  message: z.string(),
  field_locations: z.object({}),
}));
export type ComputeFieldLocationsResponse = z.infer<typeof ZComputeFieldLocationsResponse>;

export const ZComputerTool = z.lazy(() => z.object({
  display_height: z.number(),
  display_width: z.number(),
  environment: z.enum(["windows", "mac", "linux", "ubuntu", "browser"]),
  type: z.string(),
}));
export type ComputerTool = z.infer<typeof ZComputerTool>;

export const ZConsensusDictRequest = z.lazy(() => z.object({
  list_dicts: z.array(z.object({})),
  reference_schema: z.union([z.object({}), z.null()]).optional(),
  mode: z.enum(["direct", "aligned"]).optional(),
}));
export type ConsensusDictRequest = z.infer<typeof ZConsensusDictRequest>;

export const ZContentBlockSourceParam = z.lazy(() => z.object({
  content: z.union([z.string(), z.array(z.union([ZTextBlockParam, ZImageBlockParam]))]),
  type: z.string(),
}));
export type ContentBlockSourceParam = z.infer<typeof ZContentBlockSourceParam>;

export const ZCostBreakdown = z.lazy(() => z.object({
  total: ZAmount,
  text_prompt_cost: ZAmount,
  text_cached_cost: ZAmount,
  text_completion_cost: ZAmount,
  text_total_cost: ZAmount,
  audio_prompt_cost: z.union([ZAmount, z.null()]).optional(),
  audio_completion_cost: z.union([ZAmount, z.null()]).optional(),
  audio_total_cost: z.union([ZAmount, z.null()]).optional(),
  token_counts: ZTokenCounts,
  model: z.string(),
  is_fine_tuned: z.boolean().optional(),
}));
export type CostBreakdown = z.infer<typeof ZCostBreakdown>;

export const ZCreateAndLinkOrganizationRequest = z.lazy(() => z.object({
  organization_name: z.string(),
}));
export type CreateAndLinkOrganizationRequest = z.infer<typeof ZCreateAndLinkOrganizationRequest>;

export const ZCreateEvaluation = z.lazy(() => z.object({
  name: z.string(),
  json_schema: z.object({}),
  project_id: z.string().optional(),
  default_inference_settings: ZInferenceSettings.optional(),
}));
export type CreateEvaluation = z.infer<typeof ZCreateEvaluation>;

export const ZCreateOrganizationResponse = z.lazy(() => z.object({
  success: z.boolean(),
  workos_organization: ZOrganization,
}));
export type CreateOrganizationResponse = z.infer<typeof ZCreateOrganizationResponse>;

export const ZCreateSpreadsheetWithStoredTokenRequest = z.lazy(() => z.object({
  spreadsheet_name: z.string(),
}));
export type CreateSpreadsheetWithStoredTokenRequest = z.infer<typeof ZCreateSpreadsheetWithStoredTokenRequest>;

export const ZCredits = z.lazy(() => z.object({
  credits: z.number(),
}));
export type Credits = z.infer<typeof ZCredits>;

export const ZCreditsDataPoint = z.lazy(() => z.object({
  date: z.string(),
  credits: z.number(),
}));
export type CreditsDataPoint = z.infer<typeof ZCreditsDataPoint>;

export const ZCreditsTimeSeries = z.lazy(() => z.object({
  data: z.array(ZCreditsDataPoint),
}));
export type CreditsTimeSeries = z.infer<typeof ZCreditsTimeSeries>;

export const ZCustomDomain = z.lazy(() => z.object({
  id: z.string(),
  domain: z.string(),
  status: z.union([z.enum(["active", "pending", "active_redeploying", "moved", "pending_deletion", "deleted", "pending_blocked", "pending_migration", "pending_provisioned", "test_pending", "test_active", "test_active_apex", "test_blocked", "test_failed", "provisioned", "blocked"]), z.null()]),
  default: z.boolean().optional(),
}));
export type CustomDomain = z.infer<typeof ZCustomDomain>;

export const ZDBFile = z.lazy(() => z.object({
  object: z.string().optional(),
  id: z.string(),
  filename: z.string(),
}));
export type DBFile = z.infer<typeof ZDBFile>;

export const ZDataPoint = z.lazy(() => z.object({
  date: z.string(),
  value: z.number(),
}));
export type DataPoint = z.infer<typeof ZDataPoint>;

export const ZDetachProcessorRequest = z.lazy(() => z.object({
  new_processor_name: z.string(),
}));
export type DetachProcessorRequest = z.infer<typeof ZDetachProcessorRequest>;

export const ZDisplayMetadata = z.lazy(() => z.object({
  url: z.string(),
  type: z.enum(["image", "pdf", "txt"]),
}));
export type DisplayMetadata = z.infer<typeof ZDisplayMetadata>;

export const ZDistancesResult = z.lazy(() => z.object({
  distances: z.object({}),
  mean_distance: z.number(),
  metric_type: z.enum(["levenshtein", "jaccard", "hamming"]),
}));
export type DistancesResult = z.infer<typeof ZDistancesResult>;

export const ZDocumentBlockParam = z.lazy(() => z.object({
  source: z.union([ZBase64PDFSourceParam, ZPlainTextSourceParam, ZContentBlockSourceParam, ZURLPDFSourceParam]),
  type: z.string(),
  cache_control: z.union([ZCacheControlEphemeralParam, z.null()]).optional(),
  citations: ZCitationsConfigParam.optional(),
  context: z.union([z.string(), z.null()]).optional(),
  title: z.union([z.string(), z.null()]).optional(),
}));
export type DocumentBlockParam = z.infer<typeof ZDocumentBlockParam>;

export const ZDocumentCreateInputRequest = z.lazy(() => z.object({
  document: ZMIMEDataInput,
  modality: z.enum(["text", "image", "native", "image+text"]),
  image_resolution_dpi: z.number().optional(),
  browser_canvas: z.enum(["A3", "A4", "A5"]).optional(),
  json_schema: z.object({}),
}));
export type DocumentCreateInputRequest = z.infer<typeof ZDocumentCreateInputRequest>;

export const ZDocumentCreateMessageRequest = z.lazy(() => z.object({
  document: ZMIMEDataInput,
  modality: z.enum(["text", "image", "native", "image+text"]),
  image_resolution_dpi: z.number().optional(),
  browser_canvas: z.enum(["A3", "A4", "A5"]).optional(),
}));
export type DocumentCreateMessageRequest = z.infer<typeof ZDocumentCreateMessageRequest>;

export const ZDocumentItem = z.lazy(() => z.object({
  mime_data: ZMIMEDataInput,
  annotation: z.object({}).optional(),
  annotation_metadata: z.union([ZPredictionMetadata, z.null()]).optional(),
}));
export type DocumentItem = z.infer<typeof ZDocumentItem>;

export const ZDocumentMessage = z.lazy(() => z.object({
  id: z.string(),
  object: z.string().optional(),
  messages: z.array(ZChatCompletionRetabMessageOutput),
  created: z.number(),
  modality: z.enum(["text", "image", "native", "image+text"]),
  token_count: ZTokenCount,
}));
export type DocumentMessage = z.infer<typeof ZDocumentMessage>;

export const ZDocumentStatus = z.lazy(() => z.object({
  document_id: z.string(),
  filename: z.string(),
  needs_update: z.boolean(),
  has_prediction: z.boolean(),
  prediction_updated_at: z.union([DateOrISO, z.null()]),
  iteration_updated_at: DateOrISO,
}));
export type DocumentStatus = z.infer<typeof ZDocumentStatus>;

export const ZDocumentTransformRequest = z.lazy(() => z.object({
  document: ZMIMEDataInput,
}));
export type DocumentTransformRequest = z.infer<typeof ZDocumentTransformRequest>;

export const ZDocumentTransformResponse = z.lazy(() => z.object({
  document: ZRetabTypesMimeMIMEData,
}));
export type DocumentTransformResponse = z.infer<typeof ZDocumentTransformResponse>;

export const ZDocumentUploadRequest = z.lazy(() => z.object({
  document: ZMIMEDataInput,
}));
export type DocumentUploadRequest = z.infer<typeof ZDocumentUploadRequest>;

export const ZDuplicateEvaluationRequest = z.lazy(() => z.object({
  project_id: z.union([z.string(), z.null()]).optional(),
  name: z.union([z.string(), z.null()]).optional(),
}));
export type DuplicateEvaluationRequest = z.infer<typeof ZDuplicateEvaluationRequest>;

export const ZEasyInputMessage = z.lazy(() => z.object({
  content: z.union([z.string(), z.array(z.union([ZResponseInputText, ZResponseInputImage, ZResponseInputFile]))]),
  role: z.enum(["user", "assistant", "system", "developer"]),
  type: z.union([z.string(), z.null()]).optional(),
}));
export type EasyInputMessage = z.infer<typeof ZEasyInputMessage>;

export const ZEasyInputMessageParam = z.lazy(() => z.object({
  content: z.union([z.string(), z.array(z.union([ZResponseInputTextParam, ZResponseInputImageParam, ZResponseInputFileParam]))]),
  role: z.enum(["user", "assistant", "system", "developer"]),
  type: z.string().optional(),
}));
export type EasyInputMessageParam = z.infer<typeof ZEasyInputMessageParam>;

export const ZEmailAddressData = z.lazy(() => z.object({
  email: z.string(),
  display_name: z.union([z.string(), z.null()]).optional(),
}));
export type EmailAddressData = z.infer<typeof ZEmailAddressData>;

export const ZEmailConversionRequest = z.lazy(() => z.object({
  bytes: z.string(),
}));
export type EmailConversionRequest = z.infer<typeof ZEmailConversionRequest>;

export const ZEmailDataInput = z.lazy(() => z.object({
  id: z.string(),
  tree_id: z.string(),
  subject: z.union([z.string(), z.null()]).optional(),
  body_plain: z.union([z.string(), z.null()]).optional(),
  body_html: z.union([z.string(), z.null()]).optional(),
  sender: ZEmailAddressData,
  recipients_to: z.array(ZEmailAddressData),
  recipients_cc: z.array(ZEmailAddressData).optional(),
  recipients_bcc: z.array(ZEmailAddressData).optional(),
  sent_at: DateOrISO,
  received_at: z.union([DateOrISO, z.null()]).optional(),
  in_reply_to: z.union([z.string(), z.null()]).optional(),
  references: z.array(z.string()).optional(),
  headers: z.object({}).optional(),
  url: z.union([z.string(), z.null()]).optional(),
  attachments: z.array(ZAttachmentMIMEDataInput).optional(),
}));
export type EmailDataInput = z.infer<typeof ZEmailDataInput>;

export const ZEmailDataOutput = z.lazy(() => z.object({
  id: z.string(),
  tree_id: z.string(),
  subject: z.union([z.string(), z.null()]).optional(),
  body_plain: z.union([z.string(), z.null()]).optional(),
  body_html: z.union([z.string(), z.null()]).optional(),
  sender: ZEmailAddressData,
  recipients_to: z.array(ZEmailAddressData),
  recipients_cc: z.array(ZEmailAddressData).optional(),
  recipients_bcc: z.array(ZEmailAddressData).optional(),
  sent_at: DateOrISO,
  received_at: z.union([DateOrISO, z.null()]).optional(),
  in_reply_to: z.union([z.string(), z.null()]).optional(),
  references: z.array(z.string()).optional(),
  headers: z.object({}).optional(),
  url: z.union([z.string(), z.null()]).optional(),
  attachments: z.array(ZAttachmentMIMEDataOutput).optional(),
}));
export type EmailDataOutput = z.infer<typeof ZEmailDataOutput>;

export const ZEmailExtractRequest = z.lazy(() => z.object({
  automation_id: z.string(),
  email_data: ZEmailDataInput,
  modality: z.union([z.enum(["text", "image", "native", "image+text"]), z.null()]).optional(),
  image_resolution_dpi: z.union([z.number(), z.null()]).optional(),
  browser_canvas: z.union([z.enum(["A3", "A4", "A5"]), z.null()]).optional(),
  model: z.union([z.string(), z.null()]).optional(),
  json_schema: z.union([z.object({}), z.null()]).optional(),
  temperature: z.union([z.number(), z.null()]).optional(),
  n_consensus: z.union([z.number(), z.null()]).optional(),
  stream: z.boolean().optional(),
  seed: z.union([z.number(), z.null()]).optional(),
  store: z.boolean().optional(),
}));
export type EmailExtractRequest = z.infer<typeof ZEmailExtractRequest>;

export const ZEndpointInput = z.lazy(() => z.object({
  id: z.string().optional(),
  name: z.string(),
  processor_id: z.string(),
  updated_at: DateOrISO.optional(),
  default_language: z.string().optional(),
  webhook_url: z.string(),
  webhook_headers: z.object({}).optional(),
  need_validation: z.boolean().optional(),
}));
export type EndpointInput = z.infer<typeof ZEndpointInput>;

export const ZEndpointOutput = z.lazy(() => z.object({
  id: z.string().optional(),
  name: z.string(),
  processor_id: z.string(),
  updated_at: DateOrISO.optional(),
  default_language: z.string().optional(),
  webhook_url: z.string(),
  webhook_headers: z.object({}).optional(),
  need_validation: z.boolean().optional(),
  object: z.string(),
}));
export type EndpointOutput = z.infer<typeof ZEndpointOutput>;

export const ZEnhanceSchemaConfig = z.lazy(() => z.object({
  allow_reasoning_fields_added: z.boolean().optional(),
  allow_field_description_update: z.boolean().optional(),
  allow_system_prompt_update: z.boolean().optional(),
  allow_field_simple_type_change: z.boolean().optional(),
  allow_field_data_structure_breakdown: z.boolean().optional(),
}));
export type EnhanceSchemaConfig = z.infer<typeof ZEnhanceSchemaConfig>;

export const ZEnhanceSchemaRequest = z.lazy(() => z.object({
  documents: z.array(ZMIMEDataInput),
  ground_truths: z.union([z.array(z.object({})), z.null()]).optional(),
  model: z.string().optional(),
  temperature: z.number().optional(),
  reasoning_effort: z.union([z.enum(["low", "medium", "high"]), z.null()]).optional(),
  modality: z.enum(["text", "image", "native", "image+text"]),
  image_resolution_dpi: z.number().optional(),
  browser_canvas: z.enum(["A3", "A4", "A5"]).optional(),
  stream: z.boolean().optional(),
  tools_config: ZEnhanceSchemaConfig.optional(),
  json_schema: z.object({}),
  instructions: z.union([z.string(), z.null()]).optional(),
  flat_likelihoods: z.union([z.array(z.object({})), z.object({}), z.null()]).optional(),
}));
export type EnhanceSchemaRequest = z.infer<typeof ZEnhanceSchemaRequest>;

export const ZErrorDetail = z.lazy(() => z.object({
  code: z.string(),
  message: z.string(),
  details: z.union([z.object({}), z.null()]).optional(),
}));
export type ErrorDetail = z.infer<typeof ZErrorDetail>;

export const ZEvaluateSchemaRequest = z.lazy(() => z.object({
  documents: z.array(ZMIMEDataInput),
  ground_truths: z.union([z.array(z.object({})), z.null()]).optional(),
  model: z.string().optional(),
  reasoning_effort: z.union([z.enum(["low", "medium", "high"]), z.null()]).optional(),
  modality: z.enum(["text", "image", "native", "image+text"]),
  image_resolution_dpi: z.number().optional(),
  browser_canvas: z.enum(["A3", "A4", "A5"]).optional(),
  n_consensus: z.number().optional(),
  json_schema: z.object({}),
}));
export type EvaluateSchemaRequest = z.infer<typeof ZEvaluateSchemaRequest>;

export const ZEvaluateSchemaResponse = z.lazy(() => z.object({
  item_metrics: z.array(ZItemMetric),
}));
export type EvaluateSchemaResponse = z.infer<typeof ZEvaluateSchemaResponse>;

export const ZEvaluationDocumentInput = z.lazy(() => z.object({
  mime_data: ZMIMEDataInput,
  annotation: z.object({}).optional(),
  annotation_metadata: z.union([ZPredictionMetadata, z.null()]).optional(),
  id: z.string(),
}));
export type EvaluationDocumentInput = z.infer<typeof ZEvaluationDocumentInput>;

export const ZEvaluationDocumentOutput = z.lazy(() => z.object({
  mime_data: ZRetabTypesMimeMIMEData,
  annotation: z.object({}).optional(),
  annotation_metadata: z.union([ZPredictionMetadata, z.null()]).optional(),
  id: z.string(),
}));
export type EvaluationDocumentOutput = z.infer<typeof ZEvaluationDocumentOutput>;

export const ZEvent = z.lazy(() => z.object({
  object: z.string().optional(),
  id: z.string().optional(),
  event: z.string(),
  created_at: DateOrISO.optional(),
  data: z.object({}),
  metadata: z.union([z.object({}), z.null()]).optional(),
}));
export type Event = z.infer<typeof ZEvent>;

export const ZExportToCsvRequest = z.lazy(() => z.object({
  json_data: z.any(),
  json_schema: z.object({}),
  delimiter: z.string().optional(),
  line_delimiter: z.string().optional(),
  quote: z.string().optional(),
}));
export type ExportToCsvRequest = z.infer<typeof ZExportToCsvRequest>;

export const ZExternalAPIKey = z.lazy(() => z.object({
  provider: z.enum(["OpenAI", "Anthropic", "Gemini", "xAI", "Retab"]),
  is_configured: z.boolean(),
  last_updated: z.union([DateOrISO, z.null()]),
}));
export type ExternalAPIKey = z.infer<typeof ZExternalAPIKey>;

export const ZExternalAPIKeyRequest = z.lazy(() => z.object({
  provider: z.enum(["OpenAI", "Anthropic", "Gemini", "xAI", "Retab"]),
  api_key: z.string(),
}));
export type ExternalAPIKeyRequest = z.infer<typeof ZExternalAPIKeyRequest>;

export const ZExternalRequestLog = z.lazy(() => z.object({
  webhook_url: z.union([z.string(), z.null()]),
  request_body: z.object({}),
  request_headers: z.object({}),
  request_at: DateOrISO,
  response_body: z.object({}),
  response_headers: z.object({}),
  response_at: DateOrISO,
  status_code: z.number(),
  error: z.union([z.string(), z.null()]).optional(),
  duration_ms: z.number(),
}));
export type ExternalRequestLog = z.infer<typeof ZExternalRequestLog>;

export const ZExtraction = z.lazy(() => z.object({
  id: z.string().optional(),
  messages: z.array(ZChatCompletionRetabMessageOutput).optional(),
  messages_gcs: z.string(),
  file_gcs_paths: z.array(z.string()),
  file_ids: z.array(z.string()),
  file_gcs: z.string().optional(),
  file_id: z.string().optional(),
  status: z.enum(["success", "failed"]),
  completion: z.union([ZRetabParsedChatCompletionOutput, ZChatCompletionOutput]),
  json_schema: z.any(),
  model: z.string(),
  temperature: z.number().optional(),
  source: ZExtractionSource,
  image_resolution_dpi: z.number().optional(),
  browser_canvas: z.enum(["A3", "A4", "A5"]).optional(),
  modality: z.enum(["text", "image", "native", "image+text"]).optional(),
  reasoning_effort: z.union([z.enum(["low", "medium", "high"]), z.null()]).optional(),
  n_consensus: z.number().optional(),
  timings: z.array(ZExtractionTimingStep).optional(),
  schema_id: z.string(),
  schema_data_id: z.string(),
  created_at: DateOrISO.optional(),
  request_at: z.union([DateOrISO, z.null()]).optional(),
  organization_id: z.string(),
  validation_state: z.union([z.enum(["pending", "validated", "invalid"]), z.null()]).optional(),
  billed: z.boolean().optional(),
  api_cost: z.union([ZAmount, z.null()]),
  cost_breakdown: z.union([ZCostBreakdown, z.null()]),
}));
export type Extraction = z.infer<typeof ZExtraction>;

export const ZExtractionCount = z.lazy(() => z.object({
  total: z.number(),
}));
export type ExtractionCount = z.infer<typeof ZExtractionCount>;

export const ZExtractionSource = z.lazy(() => z.object({
  type: z.enum(["api", "annotation", "processor", "automation", "automation.link", "automation.mailbox", "automation.cron", "automation.outlook", "automation.endpoint", "schema.extract"]),
  id: z.union([z.string(), z.null()]).optional(),
}));
export type ExtractionSource = z.infer<typeof ZExtractionSource>;

export const ZExtractionTimingStep = z.lazy(() => z.object({
  name: z.union([z.string(), z.enum(["initialization", "prepare_messages", "yield_first_token", "completion"])]),
  duration: z.number(),
  notes: z.union([z.string(), z.null()]).optional(),
}));
export type ExtractionTimingStep = z.infer<typeof ZExtractionTimingStep>;

export const ZFetchParams = z.lazy(() => z.object({
  endpoint: z.string(),
  headers: z.object({}),
  name: z.string(),
}));
export type FetchParams = z.infer<typeof ZFetchParams>;

export const ZFieldLocation = z.lazy(() => z.object({
  label: z.string(),
  value: z.string(),
  quote: z.string(),
  file_id: z.union([z.string(), z.null()]).optional(),
  page: z.union([z.number(), z.null()]).optional(),
  bbox_normalized: z.union([z.tuple([z.number(), z.number(), z.number(), z.number()]), z.null()]).optional(),
  score: z.union([z.number(), z.null()]).optional(),
  match_level: z.union([z.enum(["token", "line", "block"]), z.null()]).optional(),
}));
export type FieldLocation = z.infer<typeof ZFieldLocation>;

export const ZFieldLocationsResult = z.lazy(() => z.object({
  choices: z.array(z.object({})),
}));
export type FieldLocationsResult = z.infer<typeof ZFieldLocationsResult>;

export const ZFile = z.lazy(() => z.object({
  file: ZFileFile,
  type: z.string(),
}));
export type File = z.infer<typeof ZFile>;

export const ZFileFile = z.lazy(() => z.object({
  file_data: z.string().optional(),
  file_id: z.string().optional(),
  filename: z.string().optional(),
}));
export type FileFile = z.infer<typeof ZFileFile>;

export const ZFileLink = z.lazy(() => z.object({
  download_url: z.string(),
  expires_in: z.string(),
  filename: z.string(),
}));
export type FileLink = z.infer<typeof ZFileLink>;

export const ZFileScoreIndex = z.lazy(() => z.object({
  file_id: z.string(),
  created_at: DateOrISO,
  file_embedding: z.array(z.number()),
  llm_output: z.object({}),
  hil_output: z.object({}),
  schema_id: z.string(),
  levenshtein_similarity: z.object({}),
  jaccard_similarity: z.object({}),
  hamming_similarity: z.object({}),
  schema_data_id: z.string(),
  organization_id: z.string(),
}));
export type FileScoreIndex = z.infer<typeof ZFileScoreIndex>;

export const ZFileSearchTool = z.lazy(() => z.object({
  type: z.string(),
  vector_store_ids: z.array(z.string()),
  filters: z.union([ZComparisonFilter, ZCompoundFilter, z.null()]).optional(),
  max_num_results: z.union([z.number(), z.null()]).optional(),
  ranking_options: z.union([ZRankingOptions, z.null()]).optional(),
}));
export type FileSearchTool = z.infer<typeof ZFileSearchTool>;

export const ZFinetunedModel = z.lazy(() => z.object({
  object: z.string().optional(),
  organization_id: z.string(),
  model: z.string(),
  schema_id: z.string(),
  schema_data_id: z.string(),
  finetuning_props: ZInferenceSettings,
  evaluation_id: z.union([z.string(), z.null()]).optional(),
  created_at: DateOrISO.optional(),
}));
export type FinetunedModel = z.infer<typeof ZFinetunedModel>;

export const ZFreightDocumentAnalysisResponse = z.lazy(() => z.object({
  analyses: z.array(ZFreightEmailAnalysisAPIResponse),
  has_more: z.boolean(),
}));
export type FreightDocumentAnalysisResponse = z.infer<typeof ZFreightDocumentAnalysisResponse>;

export const ZFreightEmailAnalysisAPIResponse = z.lazy(() => z.object({
  id: z.string(),
  is_demo: z.boolean().optional(),
  extraction_source: z.enum(["mailbox", "plugin"]),
  extraction_status: z.enum(["success", "pending", "failed", "sent_to_tms", "need_review"]),
  user: z.union([ZMainServerServicesCustomBertfakeRoutesUser, z.null()]).optional(),
  request_at: DateOrISO,
  extraction_type: z.enum(["RoadBookingConfirmation", "RoadTransportOrder", "AirBookingConfirmation", "RoadBookingConfirmationBert", "RoadBookingConfirmationGroussard", "RoadBookingConfirmationJourdan", "RoadBookingConfirmationMGE", "RoadBookingConfirmationSuus", "RoadBookingConfirmationThevenon", "RoadQuoteRequest", "RoadPickupMazet", "RoadCMR"]),
  extraction: z.any(),
  uncertainties: z.any(),
  mappings: z.object({}),
  action_type: z.enum(["Creation", "Modification", "Deletion"]),
  documents: z.array(z.union([ZMainServerServicesCustomBertCubemimedataBaseMIMEData, ZMainServerServicesCustomBertCubemimedataMIMEData])),
  email_data: ZBaseEmailData,
}));
export type FreightEmailAnalysisAPIResponse = z.infer<typeof ZFreightEmailAnalysisAPIResponse>;

export const ZFunctionOutput = z.lazy(() => z.object({
  arguments: z.string(),
  name: z.string(),
}));
export type FunctionOutput = z.infer<typeof ZFunctionOutput>;

export const ZFunctionCallOutput = z.lazy(() => z.object({
  arguments: z.string(),
  name: z.string(),
}));
export type FunctionCallOutput = z.infer<typeof ZFunctionCallOutput>;

export const ZFunctionTool = z.lazy(() => z.object({
  name: z.string(),
  parameters: z.union([z.object({}), z.null()]).optional(),
  strict: z.union([z.boolean(), z.null()]).optional(),
  type: z.string(),
  description: z.union([z.string(), z.null()]).optional(),
}));
export type FunctionTool = z.infer<typeof ZFunctionTool>;

export const ZGenerateSchemaRequest = z.lazy(() => z.object({
  documents: z.array(ZMIMEDataInput),
  model: z.string().optional(),
  temperature: z.number().optional(),
  reasoning_effort: z.union([z.enum(["low", "medium", "high"]), z.null()]).optional(),
  modality: z.enum(["text", "image", "native", "image+text"]),
  instructions: z.union([z.string(), z.null()]).optional(),
  image_resolution_dpi: z.number().optional(),
  browser_canvas: z.enum(["A3", "A4", "A5"]).optional(),
  stream: z.boolean().optional(),
}));
export type GenerateSchemaRequest = z.infer<typeof ZGenerateSchemaRequest>;

export const ZGenerateSystemPromptRequest = z.lazy(() => z.object({
  documents: z.array(ZMIMEDataInput),
  model: z.string().optional(),
  temperature: z.number().optional(),
  reasoning_effort: z.union([z.enum(["low", "medium", "high"]), z.null()]).optional(),
  modality: z.enum(["text", "image", "native", "image+text"]),
  instructions: z.union([z.string(), z.null()]).optional(),
  image_resolution_dpi: z.number().optional(),
  browser_canvas: z.enum(["A3", "A4", "A5"]).optional(),
  stream: z.boolean().optional(),
  json_schema: z.object({}),
}));
export type GenerateSystemPromptRequest = z.infer<typeof ZGenerateSystemPromptRequest>;

export const ZGoogleSpreadsheet = z.lazy(() => z.object({
  id: z.string(),
  name: z.string(),
  url: z.union([z.string(), z.null()]).optional(),
}));
export type GoogleSpreadsheet = z.infer<typeof ZGoogleSpreadsheet>;

export const ZGoogleWorksheet = z.lazy(() => z.object({
  id: z.number(),
  title: z.string(),
}));
export type GoogleWorksheet = z.infer<typeof ZGoogleWorksheet>;

export const ZHTTPValidationError = z.lazy(() => z.object({
  detail: z.array(ZValidationError).optional(),
}));
export type HTTPValidationError = z.infer<typeof ZHTTPValidationError>;

export const ZIdentity = z.lazy(() => z.object({
  user_id: z.string(),
  organization_id: z.union([z.string(), z.null()]).optional(),
  tier: z.number().optional(),
  auth_method: z.enum(["api_key", "bearer_token", "master_key", "outlook_auth"]).optional(),
}));
export type Identity = z.infer<typeof ZIdentity>;

export const ZImageBlockParam = z.lazy(() => z.object({
  source: z.union([ZBase64ImageSourceParam, ZURLImageSourceParam]),
  type: z.string(),
  cache_control: z.union([ZCacheControlEphemeralParam, z.null()]).optional(),
}));
export type ImageBlockParam = z.infer<typeof ZImageBlockParam>;

export const ZImageGeneration = z.lazy(() => z.object({
  type: z.string(),
  background: z.union([z.enum(["transparent", "opaque", "auto"]), z.null()]).optional(),
  input_image_mask: z.union([ZImageGenerationInputImageMask, z.null()]).optional(),
  model: z.union([z.string(), z.null()]).optional(),
  moderation: z.union([z.enum(["auto", "low"]), z.null()]).optional(),
  output_compression: z.union([z.number(), z.null()]).optional(),
  output_format: z.union([z.enum(["png", "webp", "jpeg"]), z.null()]).optional(),
  partial_images: z.union([z.number(), z.null()]).optional(),
  quality: z.union([z.enum(["low", "medium", "high", "auto"]), z.null()]).optional(),
  size: z.union([z.enum(["1024x1024", "1024x1536", "1536x1024", "auto"]), z.null()]).optional(),
}));
export type ImageGeneration = z.infer<typeof ZImageGeneration>;

export const ZImageGenerationInputImageMask = z.lazy(() => z.object({
  file_id: z.union([z.string(), z.null()]).optional(),
  image_url: z.union([z.string(), z.null()]).optional(),
}));
export type ImageGenerationInputImageMask = z.infer<typeof ZImageGenerationInputImageMask>;

export const ZImageURL = z.lazy(() => z.object({
  url: z.string(),
  detail: z.enum(["auto", "low", "high"]).optional(),
}));
export type ImageURL = z.infer<typeof ZImageURL>;

export const ZImportAnnotationsCsvResponse = z.lazy(() => z.object({
  success: z.boolean(),
  evaluation_id: z.string(),
  file_data: z.object({}),
  total_files: z.number(),
  message: z.string(),
}));
export type ImportAnnotationsCsvResponse = z.infer<typeof ZImportAnnotationsCsvResponse>;

export const ZIncompleteDetails = z.lazy(() => z.object({
  reason: z.union([z.enum(["max_output_tokens", "content_filter"]), z.null()]).optional(),
}));
export type IncompleteDetails = z.infer<typeof ZIncompleteDetails>;

export const ZInferenceSettings = z.lazy(() => z.object({
  model: z.string().optional(),
  temperature: z.number().optional(),
  modality: z.enum(["text", "image", "native", "image+text"]).optional(),
  reasoning_effort: z.union([z.enum(["low", "medium", "high"]), z.null()]).optional(),
  image_resolution_dpi: z.number().optional(),
  browser_canvas: z.enum(["A3", "A4", "A5"]).optional(),
  n_consensus: z.number().optional(),
}));
export type InferenceSettings = z.infer<typeof ZInferenceSettings>;

export const ZInputAudio = z.lazy(() => z.object({
  data: z.string(),
  format: z.enum(["wav", "mp3"]),
}));
export type InputAudio = z.infer<typeof ZInputAudio>;

export const ZInputTokensDetails = z.lazy(() => z.object({
  cached_tokens: z.number(),
}));
export type InputTokensDetails = z.infer<typeof ZInputTokensDetails>;

export const ZItemMetric = z.lazy(() => z.object({
  id: z.string(),
  name: z.string(),
  similarity: z.number(),
  similarities: z.object({}),
  flat_similarities: z.object({}),
  aligned_similarity: z.number(),
  aligned_similarities: z.object({}),
  aligned_flat_similarities: z.object({}),
}));
export type ItemMetric = z.infer<typeof ZItemMetric>;

export const ZIterationDocumentStatusResponse = z.lazy(() => z.object({
  iteration_id: z.string(),
  documents: z.array(ZDocumentStatus),
  total_documents: z.number(),
  documents_needing_update: z.number(),
  documents_up_to_date: z.number(),
}));
export type IterationDocumentStatusResponse = z.infer<typeof ZIterationDocumentStatusResponse>;

export const ZJSONSchema = z.lazy(() => z.object({
  name: z.string(),
  description: z.string().optional(),
  schema: z.object({}).optional(),
  strict: z.union([z.boolean(), z.null()]).optional(),
}));
export type JSONSchema = z.infer<typeof ZJSONSchema>;

export const ZKeyValidationResponse = z.lazy(() => z.object({
  is_valid: z.boolean(),
  message: z.string(),
}));
export type KeyValidationResponse = z.infer<typeof ZKeyValidationResponse>;

export const ZLLMAnnotateDocumentRequest = z.lazy(() => z.object({
  stream: z.boolean().optional(),
}));
export type LLMAnnotateDocumentRequest = z.infer<typeof ZLLMAnnotateDocumentRequest>;

export const ZLinkInput = z.lazy(() => z.object({
  id: z.string().optional(),
  name: z.string(),
  processor_id: z.string(),
  updated_at: DateOrISO.optional(),
  default_language: z.string().optional(),
  webhook_url: z.string(),
  webhook_headers: z.object({}).optional(),
  need_validation: z.boolean().optional(),
  password: z.union([z.string(), z.null()]).optional(),
}));
export type LinkInput = z.infer<typeof ZLinkInput>;

export const ZLinkOutput = z.lazy(() => z.object({
  id: z.string().optional(),
  name: z.string(),
  processor_id: z.string(),
  updated_at: DateOrISO.optional(),
  default_language: z.string().optional(),
  webhook_url: z.string(),
  webhook_headers: z.object({}).optional(),
  need_validation: z.boolean().optional(),
  password: z.union([z.string(), z.null()]).optional(),
  object: z.string(),
}));
export type LinkOutput = z.infer<typeof ZLinkOutput>;

export const ZListAutomations = z.lazy(() => z.object({
  data: z.array(ZAutomationConfig),
  list_metadata: ZListMetadata,
}));
export type ListAutomations = z.infer<typeof ZListAutomations>;

export const ZListDomainsResponse = z.lazy(() => z.object({
  domains: z.array(ZCustomDomain),
}));
export type ListDomainsResponse = z.infer<typeof ZListDomainsResponse>;

export const ZListEndpoints = z.lazy(() => z.object({
  data: z.array(ZEndpointOutput),
  list_metadata: ZListMetadata,
}));
export type ListEndpoints = z.infer<typeof ZListEndpoints>;

export const ZListEvaluationDocumentsResponse = z.lazy(() => z.object({
  data: z.array(ZEvaluationDocumentOutput),
}));
export type ListEvaluationDocumentsResponse = z.infer<typeof ZListEvaluationDocumentsResponse>;

export const ZListFiles = z.lazy(() => z.object({
  data: z.array(ZStoredDBFile),
  list_metadata: ZListMetadata,
}));
export type ListFiles = z.infer<typeof ZListFiles>;

export const ZListFinetunedModels = z.lazy(() => z.object({
  data: z.array(ZFinetunedModel),
  list_metadata: ZListMetadata,
}));
export type ListFinetunedModels = z.infer<typeof ZListFinetunedModels>;

export const ZListLinks = z.lazy(() => z.object({
  data: z.array(ZLinkOutput),
  list_metadata: ZListMetadata,
}));
export type ListLinks = z.infer<typeof ZListLinks>;

export const ZListLogs = z.lazy(() => z.object({
  data: z.array(ZAutomationLog),
  list_metadata: ZListMetadata,
}));
export type ListLogs = z.infer<typeof ZListLogs>;

export const ZListMetadata = z.lazy(() => z.object({
  before: z.union([z.string(), z.null()]),
  after: z.union([z.string(), z.null()]),
}));
export type ListMetadata = z.infer<typeof ZListMetadata>;

export const ZListTemplates = z.lazy(() => z.object({
  data: z.array(ZTemplateSchema),
  list_metadata: ZListMetadata,
}));
export type ListTemplates = z.infer<typeof ZListTemplates>;

export const ZLocalShell = z.lazy(() => z.object({
  type: z.string(),
}));
export type LocalShell = z.infer<typeof ZLocalShell>;

export const ZLogExtractionRequest = z.lazy(() => z.object({
  messages: z.union([z.array(ZChatCompletionRetabMessageInput), z.null()]).optional(),
  openai_messages: z.union([z.array(z.union([ZChatCompletionDeveloperMessageParam, ZChatCompletionSystemMessageParam, ZChatCompletionUserMessageParam, ZChatCompletionAssistantMessageParam, ZChatCompletionToolMessageParam, ZChatCompletionFunctionMessageParam])), z.null()]).optional(),
  openai_responses_input: z.union([z.array(z.union([ZEasyInputMessageParam, ZOpenaiTypesResponsesResponseInputParamMessage, ZResponseOutputMessageParam, ZResponseFileSearchToolCallParam, ZResponseComputerToolCallParam, ZOpenaiTypesResponsesResponseInputParamComputerCallOutput, ZResponseFunctionWebSearchParam, ZResponseFunctionToolCallParam, ZOpenaiTypesResponsesResponseInputParamFunctionCallOutput, ZResponseReasoningItemParam, ZOpenaiTypesResponsesResponseInputParamImageGenerationCall, ZResponseCodeInterpreterToolCallParam, ZOpenaiTypesResponsesResponseInputParamLocalShellCall, ZOpenaiTypesResponsesResponseInputParamLocalShellCallOutput, ZOpenaiTypesResponsesResponseInputParamMcpListTools, ZOpenaiTypesResponsesResponseInputParamMcpApprovalRequest, ZOpenaiTypesResponsesResponseInputParamMcpApprovalResponse, ZOpenaiTypesResponsesResponseInputParamMcpCall, ZOpenaiTypesResponsesResponseInputParamItemReference])), z.null()]).optional(),
  anthropic_messages: z.union([z.array(ZMessageParam), z.null()]).optional(),
  anthropic_system_prompt: z.union([z.string(), z.null()]).optional(),
  document: ZMIMEDataInput.optional(),
  completion: z.union([z.object({}), ZRetabParsedChatCompletionInput, ZAnthropicTypesMessageMessage, ZParsedChatCompletion, ZChatCompletionInput, z.null()]).optional(),
  openai_responses_output: z.union([ZResponse, z.null()]).optional(),
  json_schema: z.object({}),
  model: z.string(),
  temperature: z.number(),
}));
export type LogExtractionRequest = z.infer<typeof ZLogExtractionRequest>;

export const ZLogExtractionResponse = z.lazy(() => z.object({
  extraction_id: z.union([z.string(), z.null()]).optional(),
  status: z.enum(["success", "error"]),
  error_message: z.union([z.string(), z.null()]).optional(),
}));
export type LogExtractionResponse = z.infer<typeof ZLogExtractionResponse>;

export const ZMIMEDataInput = z.lazy(() => z.object({
  filename: z.string(),
  url: z.string(),
}));
export type MIMEDataInput = z.infer<typeof ZMIMEDataInput>;

export const ZMailboxInput = z.lazy(() => z.object({
  id: z.string().optional(),
  name: z.string(),
  processor_id: z.string(),
  updated_at: DateOrISO.optional(),
  default_language: z.string().optional(),
  webhook_url: z.string(),
  webhook_headers: z.object({}).optional(),
  need_validation: z.boolean().optional(),
  email: z.string(),
  authorized_domains: z.array(z.string()).optional(),
  authorized_emails: z.array(z.string()).optional(),
}));
export type MailboxInput = z.infer<typeof ZMailboxInput>;

export const ZMailboxOutput = z.lazy(() => z.object({
  id: z.string().optional(),
  name: z.string(),
  processor_id: z.string(),
  updated_at: DateOrISO.optional(),
  default_language: z.string().optional(),
  webhook_url: z.string(),
  webhook_headers: z.object({}).optional(),
  need_validation: z.boolean().optional(),
  email: z.string(),
  authorized_domains: z.array(z.string()).optional(),
  authorized_emails: z.array(z.string()).optional(),
  object: z.string(),
}));
export type MailboxOutput = z.infer<typeof ZMailboxOutput>;

export const ZMappingObject = z.lazy(() => z.object({
  internal_code: z.union([z.string(), z.null()]),
  extracted_object: z.any().optional(),
  mapped_object: z.any().optional(),
}));
export type MappingObject = z.infer<typeof ZMappingObject>;

export const ZMatchParams = z.lazy(() => z.object({
  endpoint: z.string(),
  headers: z.object({}),
  path: z.string(),
}));
export type MatchParams = z.infer<typeof ZMatchParams>;

export const ZMatchResultModel = z.lazy(() => z.object({
  record: z.object({}),
  similarity: z.number(),
}));
export type MatchResultModel = z.infer<typeof ZMatchResultModel>;

export const ZMatrix = z.lazy(() => z.object({
  rows: z.number(),
  cols: z.number(),
  type_: z.number(),
  data: z.string(),
}));
export type Matrix = z.infer<typeof ZMatrix>;

export const ZMcp = z.lazy(() => z.object({
  server_label: z.string(),
  server_url: z.string(),
  type: z.string(),
  allowed_tools: z.union([z.array(z.string()), ZMcpAllowedToolsMcpAllowedToolsFilter, z.null()]).optional(),
  headers: z.union([z.object({}), z.null()]).optional(),
  require_approval: z.union([ZMcpRequireApprovalMcpToolApprovalFilter, z.enum(["always", "never"]), z.null()]).optional(),
}));
export type Mcp = z.infer<typeof ZMcp>;

export const ZMcpAllowedToolsMcpAllowedToolsFilter = z.lazy(() => z.object({
  tool_names: z.union([z.array(z.string()), z.null()]).optional(),
}));
export type McpAllowedToolsMcpAllowedToolsFilter = z.infer<typeof ZMcpAllowedToolsMcpAllowedToolsFilter>;

export const ZMcpRequireApprovalMcpToolApprovalFilter = z.lazy(() => z.object({
  always: z.union([ZMcpRequireApprovalMcpToolApprovalFilterAlways, z.null()]).optional(),
  never: z.union([ZMcpRequireApprovalMcpToolApprovalFilterNever, z.null()]).optional(),
}));
export type McpRequireApprovalMcpToolApprovalFilter = z.infer<typeof ZMcpRequireApprovalMcpToolApprovalFilter>;

export const ZMcpRequireApprovalMcpToolApprovalFilterAlways = z.lazy(() => z.object({
  tool_names: z.union([z.array(z.string()), z.null()]).optional(),
}));
export type McpRequireApprovalMcpToolApprovalFilterAlways = z.infer<typeof ZMcpRequireApprovalMcpToolApprovalFilterAlways>;

export const ZMcpRequireApprovalMcpToolApprovalFilterNever = z.lazy(() => z.object({
  tool_names: z.union([z.array(z.string()), z.null()]).optional(),
}));
export type McpRequireApprovalMcpToolApprovalFilterNever = z.infer<typeof ZMcpRequireApprovalMcpToolApprovalFilterNever>;

export const ZMessageParam = z.lazy(() => z.object({
  content: z.union([z.string(), z.array(z.union([ZTextBlockParam, ZImageBlockParam, ZDocumentBlockParam, ZThinkingBlockParam, ZRedactedThinkingBlockParam, ZToolUseBlockParam, ZToolResultBlockParam, ZServerToolUseBlockParam, ZWebSearchToolResultBlockParam, ZTextBlock, ZThinkingBlock, ZRedactedThinkingBlock, ZToolUseBlock, ZServerToolUseBlock, ZWebSearchToolResultBlock]))]),
  role: z.enum(["user", "assistant"]),
}));
export type MessageParam = z.infer<typeof ZMessageParam>;

export const ZMetricResult = z.lazy(() => z.object({
  item_metrics: z.array(ZItemMetric),
  mean_similarity: z.number(),
  aligned_mean_similarity: z.number(),
  metric_type: z.enum(["levenshtein", "jaccard", "hamming"]),
}));
export type MetricResult = z.infer<typeof ZMetricResult>;

export const ZModel = z.lazy(() => z.object({
  id: z.string(),
  created: z.number(),
  object: z.string(),
  owned_by: z.string(),
}));
export type Model = z.infer<typeof ZModel>;

export const ZModelCapabilities = z.lazy(() => z.object({
  modalities: z.array(z.enum(["text", "audio", "image"])),
  endpoints: z.array(z.enum(["chat_completions", "responses", "assistants", "batch", "fine_tuning", "embeddings", "speech_generation", "translation", "completions_legacy", "image_generation", "transcription", "moderation", "realtime"])),
  features: z.array(z.enum(["streaming", "function_calling", "structured_outputs", "distillation", "fine_tuning", "predicted_outputs", "schema_generation"])),
}));
export type ModelCapabilities = z.infer<typeof ZModelCapabilities>;

export const ZModelCard = z.lazy(() => z.object({
  model: z.union([z.enum(["gpt-4o", "gpt-4o-mini", "chatgpt-4o-latest", "gpt-4.1", "gpt-4.1-mini", "gpt-4.1-mini-2025-04-14", "gpt-4.1-2025-04-14", "gpt-4.1-nano", "gpt-4.1-nano-2025-04-14", "gpt-4o-2024-11-20", "gpt-4o-2024-08-06", "gpt-4o-2024-05-13", "gpt-4o-mini-2024-07-18", "o1", "o1-2024-12-17", "o3", "o3-2025-04-16", "o4-mini", "o4-mini-2025-04-16", "gpt-4o-audio-preview-2024-12-17", "gpt-4o-audio-preview-2024-10-01", "gpt-4o-realtime-preview-2024-12-17", "gpt-4o-realtime-preview-2024-10-01", "gpt-4o-mini-audio-preview-2024-12-17", "gpt-4o-mini-realtime-preview-2024-12-17", "claude-3-5-sonnet-latest", "claude-3-5-sonnet-20241022", "claude-3-opus-20240229", "claude-3-sonnet-20240229", "claude-3-haiku-20240307", "grok-3", "grok-3-mini", "gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.5-pro-preview-06-05", "gemini-2.5-pro-preview-05-06", "gemini-2.5-pro-preview-03-25", "gemini-2.5-flash-preview-05-20", "gemini-2.5-flash-preview-04-17", "gemini-2.5-flash-lite-preview-06-17", "gemini-2.5-pro-exp-03-25", "gemini-2.0-flash-lite", "gemini-2.0-flash", "auto-large", "auto-small", "auto-micro", "human"]), z.string()]),
  pricing: ZPricing,
  capabilities: ZModelCapabilities,
  temperature_support: z.boolean().optional(),
  reasoning_effort_support: z.boolean().optional(),
  permissions: ZModelCardPermissions.optional(),
  is_finetuned: z.boolean(),
  model_credit_usage_per_page: z.number(),
}));
export type ModelCard = z.infer<typeof ZModelCard>;

export const ZModelCardPermissions = z.lazy(() => z.object({
  show_in_free_picker: z.boolean().optional(),
  show_in_paid_picker: z.boolean().optional(),
}));
export type ModelCardPermissions = z.infer<typeof ZModelCardPermissions>;

export const ZModelCardsResponse = z.lazy(() => z.object({
  data: z.array(ZModelCard),
  object: z.string().optional(),
}));
export type ModelCardsResponse = z.infer<typeof ZModelCardsResponse>;

export const ZModelsResponse = z.lazy(() => z.object({
  data: z.array(ZModel),
  object: z.string().optional(),
}));
export type ModelsResponse = z.infer<typeof ZModelsResponse>;

export const ZMonthlyUsageResponseContent = z.lazy(() => z.object({
  credits_count: z.number(),
}));
export type MonthlyUsageResponseContent = z.infer<typeof ZMonthlyUsageResponseContent>;

export const ZMultipleUploadResponse = z.lazy(() => z.object({
  files: z.array(ZDBFile),
}));
export type MultipleUploadResponse = z.infer<typeof ZMultipleUploadResponse>;

export const ZOCRInput = z.lazy(() => z.object({
  pages: z.array(ZPageInput),
}));
export type OCRInput = z.infer<typeof ZOCRInput>;

export const ZOCRMetadata = z.lazy(() => z.object({
  result: z.union([ZRetabTypesMimeOCROutput, z.null()]).optional(),
  file_gcs: z.union([z.string(), z.null()]).optional(),
  file_page_count: z.union([z.number(), z.null()]).optional(),
}));
export type OCRMetadata = z.infer<typeof ZOCRMetadata>;

export const ZOpenAIRateLimits = z.lazy(() => z.object({
  limit_requests: z.number().optional(),
  limit_tokens: z.number().optional(),
  remaining_requests: z.number().optional(),
  remaining_tokens: z.number().optional(),
  reset_requests: z.string().optional(),
  reset_tokens: z.string().optional(),
}));
export type OpenAIRateLimits = z.infer<typeof ZOpenAIRateLimits>;

export const ZOpenAITierResponse = z.lazy(() => z.object({
  rate_limits: ZOpenAIRateLimits,
}));
export type OpenAITierResponse = z.infer<typeof ZOpenAITierResponse>;

export const ZOptimizedDocumentMetrics = z.lazy(() => z.object({
  document_id: z.string(),
  filename: z.string(),
  true_positives: z.array(z.object({})),
  true_negatives: z.array(z.object({})),
  false_positives: z.array(z.object({})),
  false_negatives: z.array(z.object({})),
  mismatched_values: z.array(z.object({})),
  field_similarities: z.object({}),
}));
export type OptimizedDocumentMetrics = z.infer<typeof ZOptimizedDocumentMetrics>;

export const ZOptimizedIterationMetrics = z.lazy(() => z.object({
  overall_metrics: ZOptimizedOverallMetrics,
  document_metrics: z.array(ZOptimizedDocumentMetrics),
}));
export type OptimizedIterationMetrics = z.infer<typeof ZOptimizedIterationMetrics>;

export const ZOptimizedOverallMetrics = z.lazy(() => z.object({
  accuracy: z.number(),
  similarity: z.number(),
  total_error_rate: z.number(),
  true_positive_rate: z.number(),
  true_negative_rate: z.number(),
  false_positive_rate: z.number(),
  false_negative_rate: z.number(),
  mismatched_value_rate: z.number(),
  accuracy_per_field: z.object({}),
  similarity_per_field: z.object({}),
  total_documents: z.number(),
  total_fields_compared: z.number(),
}));
export type OptimizedOverallMetrics = z.infer<typeof ZOptimizedOverallMetrics>;

export const ZOrganization = z.lazy(() => z.object({
  id: z.string(),
  object: z.string(),
  name: z.string(),
  domains: z.array(ZOrganizationDomain),
  created_at: z.string(),
  updated_at: z.string(),
  allow_profiles_outside_organization: z.boolean(),
  stripe_customer_id: z.union([z.string(), z.null()]).optional(),
  external_id: z.union([z.string(), z.null()]).optional(),
  metadata: z.object({}).optional(),
}));
export type Organization = z.infer<typeof ZOrganization>;

export const ZOrganizationDomain = z.lazy(() => z.object({
  id: z.string(),
  organization_id: z.string(),
  object: z.string(),
  domain: z.string(),
  state: z.union([z.enum(["failed", "pending", "legacy_verified", "verified"]), z.null()]).optional(),
  verification_strategy: z.union([z.enum(["manual", "dns"]), z.null()]).optional(),
  verification_token: z.union([z.string(), z.null()]).optional(),
}));
export type OrganizationDomain = z.infer<typeof ZOrganizationDomain>;

export const ZOutlookInput = z.lazy(() => z.object({
  id: z.string().optional(),
  name: z.string(),
  processor_id: z.string(),
  updated_at: DateOrISO.optional(),
  default_language: z.string().optional(),
  webhook_url: z.string(),
  webhook_headers: z.object({}).optional(),
  need_validation: z.boolean().optional(),
  authorized_domains: z.array(z.string()).optional(),
  authorized_emails: z.array(z.string()).optional(),
  layout_schema: z.union([z.object({}), z.null()]).optional(),
  match_params: z.array(ZMatchParams).optional(),
  fetch_params: z.array(ZFetchParams).optional(),
}));
export type OutlookInput = z.infer<typeof ZOutlookInput>;

export const ZOutlookOutput = z.lazy(() => z.object({
  id: z.string().optional(),
  name: z.string(),
  processor_id: z.string(),
  updated_at: DateOrISO.optional(),
  default_language: z.string().optional(),
  webhook_url: z.string(),
  webhook_headers: z.object({}).optional(),
  need_validation: z.boolean().optional(),
  authorized_domains: z.array(z.string()).optional(),
  authorized_emails: z.array(z.string()).optional(),
  layout_schema: z.union([z.object({}), z.null()]).optional(),
  match_params: z.array(ZMatchParams).optional(),
  fetch_params: z.array(ZFetchParams).optional(),
  object: z.string(),
}));
export type OutlookOutput = z.infer<typeof ZOutlookOutput>;

export const ZOutlookSubmitRequest = z.lazy(() => z.object({
  email_data: ZEmailDataInput,
  completion: ZRetabParsedChatCompletionInput,
  user_email: z.string(),
  metadata: z.object({}),
  store: z.boolean().optional(),
}));
export type OutlookSubmitRequest = z.infer<typeof ZOutlookSubmitRequest>;

export const ZOutputTokensDetails = z.lazy(() => z.object({
  reasoning_tokens: z.number(),
}));
export type OutputTokensDetails = z.infer<typeof ZOutputTokensDetails>;

export const ZPageInput = z.lazy(() => z.object({
  page_number: z.number(),
  width: z.number(),
  height: z.number(),
  unit: z.string().optional(),
  blocks: z.array(ZTextBox),
  lines: z.array(ZTextBox),
  tokens: z.array(ZTextBox),
  transforms: z.array(ZMatrix).optional(),
}));
export type PageInput = z.infer<typeof ZPageInput>;

export const ZPaginatedList = z.lazy(() => z.object({
  data: z.array(z.any()),
  list_metadata: ZListMetadata,
}));
export type PaginatedList = z.infer<typeof ZPaginatedList>;

export const ZParseRequest = z.lazy(() => z.object({
  document: ZMIMEDataInput,
  model: z.enum(["gpt-4o", "gpt-4o-mini", "chatgpt-4o-latest", "gpt-4.1", "gpt-4.1-mini", "gpt-4.1-mini-2025-04-14", "gpt-4.1-2025-04-14", "gpt-4.1-nano", "gpt-4.1-nano-2025-04-14", "gpt-4o-2024-11-20", "gpt-4o-2024-08-06", "gpt-4o-2024-05-13", "gpt-4o-mini-2024-07-18", "o1", "o1-2024-12-17", "o3", "o3-2025-04-16", "o4-mini", "o4-mini-2025-04-16", "gpt-4o-audio-preview-2024-12-17", "gpt-4o-audio-preview-2024-10-01", "gpt-4o-realtime-preview-2024-12-17", "gpt-4o-realtime-preview-2024-10-01", "gpt-4o-mini-audio-preview-2024-12-17", "gpt-4o-mini-realtime-preview-2024-12-17", "claude-3-5-sonnet-latest", "claude-3-5-sonnet-20241022", "claude-3-opus-20240229", "claude-3-sonnet-20240229", "claude-3-haiku-20240307", "grok-3", "grok-3-mini", "gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.5-pro-preview-06-05", "gemini-2.5-pro-preview-05-06", "gemini-2.5-pro-preview-03-25", "gemini-2.5-flash-preview-05-20", "gemini-2.5-flash-preview-04-17", "gemini-2.5-flash-lite-preview-06-17", "gemini-2.5-pro-exp-03-25", "gemini-2.0-flash-lite", "gemini-2.0-flash", "auto-large", "auto-small", "auto-micro", "human"]).optional(),
  table_parsing_format: z.enum(["markdown", "yaml", "html", "json"]).optional(),
  image_resolution_dpi: z.number().optional(),
  browser_canvas: z.enum(["A3", "A4", "A5"]).optional(),
}));
export type ParseRequest = z.infer<typeof ZParseRequest>;

export const ZParseResult = z.lazy(() => z.object({
  document: ZRetabTypesMimeBaseMIMEData,
  usage: ZRetabUsage,
  pages: z.array(z.string()),
  text: z.string(),
}));
export type ParseResult = z.infer<typeof ZParseResult>;

export const ZParsedChatCompletion = z.lazy(() => z.object({
  id: z.string(),
  choices: z.array(ZParsedChoice),
  created: z.number(),
  model: z.string(),
  object: z.string(),
  service_tier: z.union([z.enum(["auto", "default", "flex", "scale", "priority"]), z.null()]).optional(),
  system_fingerprint: z.union([z.string(), z.null()]).optional(),
  usage: z.union([ZCompletionUsage, z.null()]).optional(),
}));
export type ParsedChatCompletion = z.infer<typeof ZParsedChatCompletion>;

export const ZParsedChatCompletionMessageInput = z.lazy(() => z.object({
  content: z.union([z.string(), z.null()]).optional(),
  refusal: z.union([z.string(), z.null()]).optional(),
  role: z.string(),
  annotations: z.union([z.array(ZAnnotationInput), z.null()]).optional(),
  audio: z.union([ZChatCompletionAudio, z.null()]).optional(),
  function_call: z.union([ZOpenaiTypesChatChatCompletionMessageFunctionCall, z.null()]).optional(),
  tool_calls: z.union([z.array(ZParsedFunctionToolCall), z.null()]).optional(),
  parsed: z.union([z.any(), z.null()]).optional(),
}));
export type ParsedChatCompletionMessageInput = z.infer<typeof ZParsedChatCompletionMessageInput>;

export const ZParsedChatCompletionMessageOutput = z.lazy(() => z.object({
  content: z.union([z.string(), z.null()]).optional(),
  refusal: z.union([z.string(), z.null()]).optional(),
  role: z.string(),
  annotations: z.union([z.array(ZAnnotationOutput), z.null()]).optional(),
  audio: z.union([ZChatCompletionAudio, z.null()]).optional(),
  function_call: z.union([ZFunctionCallOutput, z.null()]).optional(),
  tool_calls: z.union([z.array(ZParsedFunctionToolCall), z.null()]).optional(),
  parsed: z.union([z.any(), z.null()]).optional(),
}));
export type ParsedChatCompletionMessageOutput = z.infer<typeof ZParsedChatCompletionMessageOutput>;

export const ZParsedChoice = z.lazy(() => z.object({
  finish_reason: z.enum(["stop", "length", "tool_calls", "content_filter", "function_call"]),
  index: z.number(),
  logprobs: z.union([ZChoiceLogprobsInput, z.null()]).optional(),
  message: ZParsedChatCompletionMessageInput,
}));
export type ParsedChoice = z.infer<typeof ZParsedChoice>;

export const ZParsedFunction = z.lazy(() => z.object({
  arguments: z.string(),
  name: z.string(),
  parsed_arguments: z.union([z.any(), z.null()]).optional(),
}));
export type ParsedFunction = z.infer<typeof ZParsedFunction>;

export const ZParsedFunctionToolCall = z.lazy(() => z.object({
  id: z.string(),
  function: ZParsedFunction,
  type: z.string(),
}));
export type ParsedFunctionToolCall = z.infer<typeof ZParsedFunctionToolCall>;

export const ZPartialSchema = z.lazy(() => z.object({
  object: z.string().optional(),
  created_at: DateOrISO.optional(),
  json_schema: z.object({}).optional(),
  strict: z.boolean().optional(),
}));
export type PartialSchema = z.infer<typeof ZPartialSchema>;

export const ZPatchEvaluationDocumentRequest = z.lazy(() => z.object({
  annotation: z.union([z.object({}), z.null()]).optional(),
  annotation_metadata: z.union([ZPredictionMetadata, z.null()]).optional(),
}));
export type PatchEvaluationDocumentRequest = z.infer<typeof ZPatchEvaluationDocumentRequest>;

export const ZPatchIterationDocumentPredictionRequest = z.lazy(() => z.object({
  metadata: ZPredictionMetadata,
}));
export type PatchIterationDocumentPredictionRequest = z.infer<typeof ZPatchIterationDocumentPredictionRequest>;

export const ZPatchIterationRequest = z.lazy(() => z.object({
  inference_settings: z.union([ZInferenceSettings, z.null()]).optional(),
  json_schema: z.union([z.object({}), z.null()]).optional(),
}));
export type PatchIterationRequest = z.infer<typeof ZPatchIterationRequest>;

export const ZPerformOCROnlyRequest = z.lazy(() => z.object({
  extraction_id: z.string(),
}));
export type PerformOCROnlyRequest = z.infer<typeof ZPerformOCROnlyRequest>;

export const ZPerformOCROnlyResponse = z.lazy(() => z.object({
  status: z.enum(["success", "error"]),
  message: z.string(),
  extraction_id: z.string(),
  ocr_file_id: z.string(),
  ocr_file_url: z.string(),
  ocr_result: z.union([ZRetabTypesMimeOCROutput, z.null()]).optional(),
}));
export type PerformOCROnlyResponse = z.infer<typeof ZPerformOCROnlyResponse>;

export const ZPerformOCRRequest = z.lazy(() => z.object({
  extraction_id: z.string(),
}));
export type PerformOCRRequest = z.infer<typeof ZPerformOCRRequest>;

export const ZPerformOCRResponse = z.lazy(() => z.object({
  status: z.enum(["success", "error"]),
  message: z.string(),
  extraction_id: z.string(),
  ocr_file_id: z.string(),
  ocr_file_url: z.string(),
  field_locations_result: z.union([ZFieldLocationsResult, z.null()]).optional(),
}));
export type PerformOCRResponse = z.infer<typeof ZPerformOCRResponse>;

export const ZPlainTextSourceParam = z.lazy(() => z.object({
  data: z.string(),
  media_type: z.string(),
  type: z.string(),
}));
export type PlainTextSourceParam = z.infer<typeof ZPlainTextSourceParam>;

export const ZPoint = z.lazy(() => z.object({
  x: z.number(),
  y: z.number(),
}));
export type Point = z.infer<typeof ZPoint>;

export const ZPredictionMetadata = z.lazy(() => z.object({
  extraction_id: z.union([z.string(), z.null()]).optional(),
  likelihoods: z.union([z.object({}), z.null()]).optional(),
  field_locations: z.union([z.object({}), z.null()]).optional(),
  agentic_field_locations: z.union([z.object({}), z.null()]).optional(),
  consensus_details: z.union([z.array(z.object({})), z.null()]).optional(),
  api_cost: z.union([ZAmount, z.null()]).optional(),
}));
export type PredictionMetadata = z.infer<typeof ZPredictionMetadata>;

export const ZPreprocessingLogResponse = z.lazy(() => z.object({
  id: z.string(),
  credits_count: z.number(),
  page_count: z.union([z.number(), z.null()]).optional(),
  filename: z.string(),
  operation: z.string(),
}));
export type PreprocessingLogResponse = z.infer<typeof ZPreprocessingLogResponse>;

export const ZPricing = z.lazy(() => z.object({
  text: ZTokenPrice,
  audio: z.union([ZTokenPrice, z.null()]).optional(),
  ft_price_hike: z.number().optional(),
}));
export type Pricing = z.infer<typeof ZPricing>;

export const ZProcessIterationDocument = z.lazy(() => z.object({
  stream: z.boolean().optional(),
}));
export type ProcessIterationDocument = z.infer<typeof ZProcessIterationDocument>;

export const ZProcessIterationRequest = z.lazy(() => z.object({
  document_ids: z.union([z.array(z.string()), z.null()]).optional(),
  only_outdated: z.boolean().optional(),
}));
export type ProcessIterationRequest = z.infer<typeof ZProcessIterationRequest>;

export const ZProcessorConfig = z.lazy(() => z.object({
  object: z.string().optional(),
  id: z.string().optional(),
  updated_at: DateOrISO.optional(),
  name: z.string(),
  modality: z.enum(["text", "image", "native", "image+text"]),
  image_resolution_dpi: z.number().optional(),
  browser_canvas: z.enum(["A3", "A4", "A5"]).optional(),
  model: z.string(),
  json_schema: z.object({}),
  temperature: z.number().optional(),
  reasoning_effort: z.union([z.enum(["low", "medium", "high"]), z.null()]).optional(),
  n_consensus: z.number().optional(),
}));
export type ProcessorConfig = z.infer<typeof ZProcessorConfig>;

export const ZProject = z.lazy(() => z.object({
  id: z.string().optional(),
  name: z.string(),
  updated_at: DateOrISO.optional(),
}));
export type Project = z.infer<typeof ZProject>;

export const ZPromptTokensDetails = z.lazy(() => z.object({
  audio_tokens: z.union([z.number(), z.null()]).optional(),
  cached_tokens: z.union([z.number(), z.null()]).optional(),
}));
export type PromptTokensDetails = z.infer<typeof ZPromptTokensDetails>;

export const ZRankingOptions = z.lazy(() => z.object({
  ranker: z.union([z.enum(["auto", "default-2024-11-15"]), z.null()]).optional(),
  score_threshold: z.union([z.number(), z.null()]).optional(),
}));
export type RankingOptions = z.infer<typeof ZRankingOptions>;

export const ZReconciliationResponse = z.lazy(() => z.object({
  consensus_dict: z.object({}),
  likelihoods: z.object({}),
}));
export type ReconciliationResponse = z.infer<typeof ZReconciliationResponse>;

export const ZRedactedThinkingBlock = z.lazy(() => z.object({
  data: z.string(),
  type: z.string(),
}));
export type RedactedThinkingBlock = z.infer<typeof ZRedactedThinkingBlock>;

export const ZRedactedThinkingBlockParam = z.lazy(() => z.object({
  data: z.string(),
  type: z.string(),
}));
export type RedactedThinkingBlockParam = z.infer<typeof ZRedactedThinkingBlockParam>;

export const ZResponse = z.lazy(() => z.object({
  id: z.string(),
  created_at: z.number(),
  error: z.union([ZResponseError, z.null()]).optional(),
  incomplete_details: z.union([ZIncompleteDetails, z.null()]).optional(),
  instructions: z.union([z.string(), z.array(z.union([ZEasyInputMessage, ZOpenaiTypesResponsesResponseInputItemMessage, ZResponseOutputMessage, ZResponseFileSearchToolCall, ZResponseComputerToolCall, ZOpenaiTypesResponsesResponseInputItemComputerCallOutput, ZResponseFunctionWebSearch, ZResponseFunctionToolCall, ZOpenaiTypesResponsesResponseInputItemFunctionCallOutput, ZResponseReasoningItem, ZOpenaiTypesResponsesResponseInputItemImageGenerationCall, ZResponseCodeInterpreterToolCall, ZOpenaiTypesResponsesResponseInputItemLocalShellCall, ZOpenaiTypesResponsesResponseInputItemLocalShellCallOutput, ZOpenaiTypesResponsesResponseInputItemMcpListTools, ZOpenaiTypesResponsesResponseInputItemMcpApprovalRequest, ZOpenaiTypesResponsesResponseInputItemMcpApprovalResponse, ZOpenaiTypesResponsesResponseInputItemMcpCall, ZOpenaiTypesResponsesResponseInputItemItemReference])), z.null()]).optional(),
  metadata: z.union([z.object({}), z.null()]).optional(),
  model: z.union([z.string(), z.enum(["gpt-4.1", "gpt-4.1-mini", "gpt-4.1-nano", "gpt-4.1-2025-04-14", "gpt-4.1-mini-2025-04-14", "gpt-4.1-nano-2025-04-14", "o4-mini", "o4-mini-2025-04-16", "o3", "o3-2025-04-16", "o3-mini", "o3-mini-2025-01-31", "o1", "o1-2024-12-17", "o1-preview", "o1-preview-2024-09-12", "o1-mini", "o1-mini-2024-09-12", "gpt-4o", "gpt-4o-2024-11-20", "gpt-4o-2024-08-06", "gpt-4o-2024-05-13", "gpt-4o-audio-preview", "gpt-4o-audio-preview-2024-10-01", "gpt-4o-audio-preview-2024-12-17", "gpt-4o-audio-preview-2025-06-03", "gpt-4o-mini-audio-preview", "gpt-4o-mini-audio-preview-2024-12-17", "gpt-4o-search-preview", "gpt-4o-mini-search-preview", "gpt-4o-search-preview-2025-03-11", "gpt-4o-mini-search-preview-2025-03-11", "chatgpt-4o-latest", "codex-mini-latest", "gpt-4o-mini", "gpt-4o-mini-2024-07-18", "gpt-4-turbo", "gpt-4-turbo-2024-04-09", "gpt-4-0125-preview", "gpt-4-turbo-preview", "gpt-4-1106-preview", "gpt-4-vision-preview", "gpt-4", "gpt-4-0314", "gpt-4-0613", "gpt-4-32k", "gpt-4-32k-0314", "gpt-4-32k-0613", "gpt-3.5-turbo", "gpt-3.5-turbo-16k", "gpt-3.5-turbo-0301", "gpt-3.5-turbo-0613", "gpt-3.5-turbo-1106", "gpt-3.5-turbo-0125", "gpt-3.5-turbo-16k-0613"]), z.enum(["o1-pro", "o1-pro-2025-03-19", "o3-pro", "o3-pro-2025-06-10", "o3-deep-research", "o3-deep-research-2025-06-26", "o4-mini-deep-research", "o4-mini-deep-research-2025-06-26", "computer-use-preview", "computer-use-preview-2025-03-11"])]),
  object: z.string(),
  output: z.array(z.union([ZResponseOutputMessage, ZResponseFileSearchToolCall, ZResponseFunctionToolCall, ZResponseFunctionWebSearch, ZResponseComputerToolCall, ZResponseReasoningItem, ZOpenaiTypesResponsesResponseOutputItemImageGenerationCall, ZResponseCodeInterpreterToolCall, ZOpenaiTypesResponsesResponseOutputItemLocalShellCall, ZOpenaiTypesResponsesResponseOutputItemMcpCall, ZOpenaiTypesResponsesResponseOutputItemMcpListTools, ZOpenaiTypesResponsesResponseOutputItemMcpApprovalRequest])),
  parallel_tool_calls: z.boolean(),
  temperature: z.union([z.number(), z.null()]).optional(),
  tool_choice: z.union([z.enum(["none", "auto", "required"]), ZToolChoiceTypes, ZToolChoiceFunction, ZToolChoiceMcp]),
  tools: z.array(z.union([ZFunctionTool, ZFileSearchTool, ZWebSearchTool, ZComputerTool, ZMcp, ZCodeInterpreter, ZImageGeneration, ZLocalShell])),
  top_p: z.union([z.number(), z.null()]).optional(),
  background: z.union([z.boolean(), z.null()]).optional(),
  max_output_tokens: z.union([z.number(), z.null()]).optional(),
  max_tool_calls: z.union([z.number(), z.null()]).optional(),
  previous_response_id: z.union([z.string(), z.null()]).optional(),
  prompt: z.union([ZResponsePrompt, z.null()]).optional(),
  reasoning: z.union([ZOpenaiTypesSharedReasoningReasoning, z.null()]).optional(),
  service_tier: z.union([z.enum(["auto", "default", "flex", "scale", "priority"]), z.null()]).optional(),
  status: z.union([z.enum(["completed", "failed", "in_progress", "cancelled", "queued", "incomplete"]), z.null()]).optional(),
  text: z.union([ZResponseTextConfig, z.null()]).optional(),
  top_logprobs: z.union([z.number(), z.null()]).optional(),
  truncation: z.union([z.enum(["auto", "disabled"]), z.null()]).optional(),
  usage: z.union([ZResponseUsage, z.null()]).optional(),
  user: z.union([z.string(), z.null()]).optional(),
}));
export type Response = z.infer<typeof ZResponse>;

export const ZResponseCodeInterpreterToolCall = z.lazy(() => z.object({
  id: z.string(),
  code: z.union([z.string(), z.null()]).optional(),
  container_id: z.string(),
  outputs: z.union([z.array(z.union([ZOpenaiTypesResponsesResponseCodeInterpreterToolCallOutputLogs, ZOpenaiTypesResponsesResponseCodeInterpreterToolCallOutputImage])), z.null()]).optional(),
  status: z.enum(["in_progress", "completed", "incomplete", "interpreting", "failed"]),
  type: z.string(),
}));
export type ResponseCodeInterpreterToolCall = z.infer<typeof ZResponseCodeInterpreterToolCall>;

export const ZResponseCodeInterpreterToolCallParam = z.lazy(() => z.object({
  id: z.string(),
  code: z.union([z.string(), z.null()]),
  container_id: z.string(),
  outputs: z.union([z.array(z.union([ZOpenaiTypesResponsesResponseCodeInterpreterToolCallParamOutputLogs, ZOpenaiTypesResponsesResponseCodeInterpreterToolCallParamOutputImage])), z.null()]),
  status: z.enum(["in_progress", "completed", "incomplete", "interpreting", "failed"]),
  type: z.string(),
}));
export type ResponseCodeInterpreterToolCallParam = z.infer<typeof ZResponseCodeInterpreterToolCallParam>;

export const ZResponseComputerToolCall = z.lazy(() => z.object({
  id: z.string(),
  action: z.union([ZOpenaiTypesResponsesResponseComputerToolCallActionClick, ZOpenaiTypesResponsesResponseComputerToolCallActionDoubleClick, ZOpenaiTypesResponsesResponseComputerToolCallActionDrag, ZOpenaiTypesResponsesResponseComputerToolCallActionKeypress, ZOpenaiTypesResponsesResponseComputerToolCallActionMove, ZOpenaiTypesResponsesResponseComputerToolCallActionScreenshot, ZOpenaiTypesResponsesResponseComputerToolCallActionScroll, ZOpenaiTypesResponsesResponseComputerToolCallActionType, ZOpenaiTypesResponsesResponseComputerToolCallActionWait]),
  call_id: z.string(),
  pending_safety_checks: z.array(ZOpenaiTypesResponsesResponseComputerToolCallPendingSafetyCheck),
  status: z.enum(["in_progress", "completed", "incomplete"]),
  type: z.string(),
}));
export type ResponseComputerToolCall = z.infer<typeof ZResponseComputerToolCall>;

export const ZResponseComputerToolCallOutputScreenshot = z.lazy(() => z.object({
  type: z.string(),
  file_id: z.union([z.string(), z.null()]).optional(),
  image_url: z.union([z.string(), z.null()]).optional(),
}));
export type ResponseComputerToolCallOutputScreenshot = z.infer<typeof ZResponseComputerToolCallOutputScreenshot>;

export const ZResponseComputerToolCallOutputScreenshotParam = z.lazy(() => z.object({
  type: z.string(),
  file_id: z.string().optional(),
  image_url: z.string().optional(),
}));
export type ResponseComputerToolCallOutputScreenshotParam = z.infer<typeof ZResponseComputerToolCallOutputScreenshotParam>;

export const ZResponseComputerToolCallParam = z.lazy(() => z.object({
  id: z.string(),
  action: z.union([ZOpenaiTypesResponsesResponseComputerToolCallParamActionClick, ZOpenaiTypesResponsesResponseComputerToolCallParamActionDoubleClick, ZOpenaiTypesResponsesResponseComputerToolCallParamActionDrag, ZOpenaiTypesResponsesResponseComputerToolCallParamActionKeypress, ZOpenaiTypesResponsesResponseComputerToolCallParamActionMove, ZOpenaiTypesResponsesResponseComputerToolCallParamActionScreenshot, ZOpenaiTypesResponsesResponseComputerToolCallParamActionScroll, ZOpenaiTypesResponsesResponseComputerToolCallParamActionType, ZOpenaiTypesResponsesResponseComputerToolCallParamActionWait]),
  call_id: z.string(),
  pending_safety_checks: z.array(ZOpenaiTypesResponsesResponseComputerToolCallParamPendingSafetyCheck),
  status: z.enum(["in_progress", "completed", "incomplete"]),
  type: z.string(),
}));
export type ResponseComputerToolCallParam = z.infer<typeof ZResponseComputerToolCallParam>;

export const ZResponseError = z.lazy(() => z.object({
  code: z.enum(["server_error", "rate_limit_exceeded", "invalid_prompt", "vector_store_timeout", "invalid_image", "invalid_image_format", "invalid_base64_image", "invalid_image_url", "image_too_large", "image_too_small", "image_parse_error", "image_content_policy_violation", "invalid_image_mode", "image_file_too_large", "unsupported_image_media_type", "empty_image_file", "failed_to_download_image", "image_file_not_found"]),
  message: z.string(),
}));
export type ResponseError = z.infer<typeof ZResponseError>;

export const ZResponseFileSearchToolCall = z.lazy(() => z.object({
  id: z.string(),
  queries: z.array(z.string()),
  status: z.enum(["in_progress", "searching", "completed", "incomplete", "failed"]),
  type: z.string(),
  results: z.union([z.array(ZOpenaiTypesResponsesResponseFileSearchToolCallResult), z.null()]).optional(),
}));
export type ResponseFileSearchToolCall = z.infer<typeof ZResponseFileSearchToolCall>;

export const ZResponseFileSearchToolCallParam = z.lazy(() => z.object({
  id: z.string(),
  queries: z.array(z.string()),
  status: z.enum(["in_progress", "searching", "completed", "incomplete", "failed"]),
  type: z.string(),
  results: z.union([z.array(ZOpenaiTypesResponsesResponseFileSearchToolCallParamResult), z.null()]).optional(),
}));
export type ResponseFileSearchToolCallParam = z.infer<typeof ZResponseFileSearchToolCallParam>;

export const ZResponseFormatJSONSchema = z.lazy(() => z.object({
  json_schema: ZJSONSchema,
  type: z.string(),
}));
export type ResponseFormatJSONSchema = z.infer<typeof ZResponseFormatJSONSchema>;

export const ZResponseFormatTextJSONSchemaConfig = z.lazy(() => z.object({
  name: z.string(),
  schema: z.object({}),
  type: z.string(),
  description: z.union([z.string(), z.null()]).optional(),
  strict: z.union([z.boolean(), z.null()]).optional(),
}));
export type ResponseFormatTextJSONSchemaConfig = z.infer<typeof ZResponseFormatTextJSONSchemaConfig>;

export const ZResponseFormatTextJSONSchemaConfigParam = z.lazy(() => z.object({
  name: z.string(),
  schema: z.object({}),
  type: z.string(),
  description: z.string().optional(),
  strict: z.union([z.boolean(), z.null()]).optional(),
}));
export type ResponseFormatTextJSONSchemaConfigParam = z.infer<typeof ZResponseFormatTextJSONSchemaConfigParam>;

export const ZResponseFunctionToolCall = z.lazy(() => z.object({
  arguments: z.string(),
  call_id: z.string(),
  name: z.string(),
  type: z.string(),
  id: z.union([z.string(), z.null()]).optional(),
  status: z.union([z.enum(["in_progress", "completed", "incomplete"]), z.null()]).optional(),
}));
export type ResponseFunctionToolCall = z.infer<typeof ZResponseFunctionToolCall>;

export const ZResponseFunctionToolCallParam = z.lazy(() => z.object({
  arguments: z.string(),
  call_id: z.string(),
  name: z.string(),
  type: z.string(),
  id: z.string().optional(),
  status: z.enum(["in_progress", "completed", "incomplete"]).optional(),
}));
export type ResponseFunctionToolCallParam = z.infer<typeof ZResponseFunctionToolCallParam>;

export const ZResponseFunctionWebSearch = z.lazy(() => z.object({
  id: z.string(),
  action: z.union([ZOpenaiTypesResponsesResponseFunctionWebSearchActionSearch, ZOpenaiTypesResponsesResponseFunctionWebSearchActionOpenPage, ZOpenaiTypesResponsesResponseFunctionWebSearchActionFind]),
  status: z.enum(["in_progress", "searching", "completed", "failed"]),
  type: z.string(),
}));
export type ResponseFunctionWebSearch = z.infer<typeof ZResponseFunctionWebSearch>;

export const ZResponseFunctionWebSearchParam = z.lazy(() => z.object({
  id: z.string(),
  action: z.union([ZOpenaiTypesResponsesResponseFunctionWebSearchParamActionSearch, ZOpenaiTypesResponsesResponseFunctionWebSearchParamActionOpenPage, ZOpenaiTypesResponsesResponseFunctionWebSearchParamActionFind]),
  status: z.enum(["in_progress", "searching", "completed", "failed"]),
  type: z.string(),
}));
export type ResponseFunctionWebSearchParam = z.infer<typeof ZResponseFunctionWebSearchParam>;

export const ZResponseInputFile = z.lazy(() => z.object({
  type: z.string(),
  file_data: z.union([z.string(), z.null()]).optional(),
  file_id: z.union([z.string(), z.null()]).optional(),
  filename: z.union([z.string(), z.null()]).optional(),
}));
export type ResponseInputFile = z.infer<typeof ZResponseInputFile>;

export const ZResponseInputFileParam = z.lazy(() => z.object({
  type: z.string(),
  file_data: z.string().optional(),
  file_id: z.union([z.string(), z.null()]).optional(),
  filename: z.string().optional(),
}));
export type ResponseInputFileParam = z.infer<typeof ZResponseInputFileParam>;

export const ZResponseInputImage = z.lazy(() => z.object({
  detail: z.enum(["low", "high", "auto"]),
  type: z.string(),
  file_id: z.union([z.string(), z.null()]).optional(),
  image_url: z.union([z.string(), z.null()]).optional(),
}));
export type ResponseInputImage = z.infer<typeof ZResponseInputImage>;

export const ZResponseInputImageParam = z.lazy(() => z.object({
  detail: z.enum(["low", "high", "auto"]),
  type: z.string(),
  file_id: z.union([z.string(), z.null()]).optional(),
  image_url: z.union([z.string(), z.null()]).optional(),
}));
export type ResponseInputImageParam = z.infer<typeof ZResponseInputImageParam>;

export const ZResponseInputText = z.lazy(() => z.object({
  text: z.string(),
  type: z.string(),
}));
export type ResponseInputText = z.infer<typeof ZResponseInputText>;

export const ZResponseInputTextParam = z.lazy(() => z.object({
  text: z.string(),
  type: z.string(),
}));
export type ResponseInputTextParam = z.infer<typeof ZResponseInputTextParam>;

export const ZResponseOutputMessage = z.lazy(() => z.object({
  id: z.string(),
  content: z.array(z.union([ZResponseOutputText, ZResponseOutputRefusal])),
  role: z.string(),
  status: z.enum(["in_progress", "completed", "incomplete"]),
  type: z.string(),
}));
export type ResponseOutputMessage = z.infer<typeof ZResponseOutputMessage>;

export const ZResponseOutputMessageParam = z.lazy(() => z.object({
  id: z.string(),
  content: z.array(z.union([ZResponseOutputTextParam, ZResponseOutputRefusalParam])),
  role: z.string(),
  status: z.enum(["in_progress", "completed", "incomplete"]),
  type: z.string(),
}));
export type ResponseOutputMessageParam = z.infer<typeof ZResponseOutputMessageParam>;

export const ZResponseOutputRefusal = z.lazy(() => z.object({
  refusal: z.string(),
  type: z.string(),
}));
export type ResponseOutputRefusal = z.infer<typeof ZResponseOutputRefusal>;

export const ZResponseOutputRefusalParam = z.lazy(() => z.object({
  refusal: z.string(),
  type: z.string(),
}));
export type ResponseOutputRefusalParam = z.infer<typeof ZResponseOutputRefusalParam>;

export const ZResponseOutputText = z.lazy(() => z.object({
  annotations: z.array(z.union([ZOpenaiTypesResponsesResponseOutputTextAnnotationFileCitation, ZOpenaiTypesResponsesResponseOutputTextAnnotationURLCitation, ZOpenaiTypesResponsesResponseOutputTextAnnotationContainerFileCitation, ZOpenaiTypesResponsesResponseOutputTextAnnotationFilePath])),
  text: z.string(),
  type: z.string(),
  logprobs: z.union([z.array(ZOpenaiTypesResponsesResponseOutputTextLogprob), z.null()]).optional(),
}));
export type ResponseOutputText = z.infer<typeof ZResponseOutputText>;

export const ZResponseOutputTextParam = z.lazy(() => z.object({
  annotations: z.array(z.union([ZOpenaiTypesResponsesResponseOutputTextParamAnnotationFileCitation, ZOpenaiTypesResponsesResponseOutputTextParamAnnotationURLCitation, ZOpenaiTypesResponsesResponseOutputTextParamAnnotationContainerFileCitation, ZOpenaiTypesResponsesResponseOutputTextParamAnnotationFilePath])),
  text: z.string(),
  type: z.string(),
  logprobs: z.array(ZOpenaiTypesResponsesResponseOutputTextParamLogprob).optional(),
}));
export type ResponseOutputTextParam = z.infer<typeof ZResponseOutputTextParam>;

export const ZResponsePrompt = z.lazy(() => z.object({
  id: z.string(),
  variables: z.union([z.object({}), z.null()]).optional(),
  version: z.union([z.string(), z.null()]).optional(),
}));
export type ResponsePrompt = z.infer<typeof ZResponsePrompt>;

export const ZResponseReasoningItem = z.lazy(() => z.object({
  id: z.string(),
  summary: z.array(ZOpenaiTypesResponsesResponseReasoningItemSummary),
  type: z.string(),
  encrypted_content: z.union([z.string(), z.null()]).optional(),
  status: z.union([z.enum(["in_progress", "completed", "incomplete"]), z.null()]).optional(),
}));
export type ResponseReasoningItem = z.infer<typeof ZResponseReasoningItem>;

export const ZResponseReasoningItemParam = z.lazy(() => z.object({
  id: z.string(),
  summary: z.array(ZOpenaiTypesResponsesResponseReasoningItemParamSummary),
  type: z.string(),
  encrypted_content: z.union([z.string(), z.null()]).optional(),
  status: z.enum(["in_progress", "completed", "incomplete"]).optional(),
}));
export type ResponseReasoningItemParam = z.infer<typeof ZResponseReasoningItemParam>;

export const ZResponseTextConfig = z.lazy(() => z.object({
  format: z.union([ZOpenaiTypesSharedResponseFormatTextResponseFormatText, ZResponseFormatTextJSONSchemaConfig, ZOpenaiTypesSharedResponseFormatJsonObjectResponseFormatJSONObject, z.null()]).optional(),
}));
export type ResponseTextConfig = z.infer<typeof ZResponseTextConfig>;

export const ZResponseTextConfigParam = z.lazy(() => z.object({
  format: z.union([ZOpenaiTypesSharedParamsResponseFormatTextResponseFormatText, ZResponseFormatTextJSONSchemaConfigParam, ZOpenaiTypesSharedParamsResponseFormatJsonObjectResponseFormatJSONObject]).optional(),
}));
export type ResponseTextConfigParam = z.infer<typeof ZResponseTextConfigParam>;

export const ZResponseUsage = z.lazy(() => z.object({
  input_tokens: z.number(),
  input_tokens_details: ZInputTokensDetails,
  output_tokens: z.number(),
  output_tokens_details: ZOutputTokensDetails,
  total_tokens: z.number(),
}));
export type ResponseUsage = z.infer<typeof ZResponseUsage>;

export const ZRetabChatCompletionsRequest = z.lazy(() => z.object({
  model: z.string(),
  messages: z.array(ZChatCompletionRetabMessageInput),
  response_format: ZResponseFormatJSONSchema,
  temperature: z.number().optional(),
  reasoning_effort: z.union([z.enum(["low", "medium", "high"]), z.null()]).optional(),
  stream: z.boolean().optional(),
  seed: z.union([z.number(), z.null()]).optional(),
  n_consensus: z.number().optional(),
}));
export type RetabChatCompletionsRequest = z.infer<typeof ZRetabChatCompletionsRequest>;

export const ZRetabChatResponseCreateRequest = z.lazy(() => z.object({
  input: z.union([z.string(), z.array(z.union([ZEasyInputMessageParam, ZOpenaiTypesResponsesResponseInputParamMessage, ZResponseOutputMessageParam, ZResponseFileSearchToolCallParam, ZResponseComputerToolCallParam, ZOpenaiTypesResponsesResponseInputParamComputerCallOutput, ZResponseFunctionWebSearchParam, ZResponseFunctionToolCallParam, ZOpenaiTypesResponsesResponseInputParamFunctionCallOutput, ZResponseReasoningItemParam, ZOpenaiTypesResponsesResponseInputParamImageGenerationCall, ZResponseCodeInterpreterToolCallParam, ZOpenaiTypesResponsesResponseInputParamLocalShellCall, ZOpenaiTypesResponsesResponseInputParamLocalShellCallOutput, ZOpenaiTypesResponsesResponseInputParamMcpListTools, ZOpenaiTypesResponsesResponseInputParamMcpApprovalRequest, ZOpenaiTypesResponsesResponseInputParamMcpApprovalResponse, ZOpenaiTypesResponsesResponseInputParamMcpCall, ZOpenaiTypesResponsesResponseInputParamItemReference]))]),
  instructions: z.union([z.string(), z.null()]).optional(),
  model: z.string(),
  temperature: z.union([z.number(), z.null()]).optional(),
  reasoning: z.union([ZOpenaiTypesSharedParamsReasoningReasoning, z.null()]).optional(),
  stream: z.union([z.boolean(), z.null()]).optional(),
  seed: z.union([z.number(), z.null()]).optional(),
  text: ZResponseTextConfigParam.optional(),
  n_consensus: z.number().optional(),
}));
export type RetabChatResponseCreateRequest = z.infer<typeof ZRetabChatResponseCreateRequest>;

export const ZRetabParsedChatCompletionInput = z.lazy(() => z.object({
  id: z.string(),
  choices: z.array(ZRetabParsedChoiceInput),
  created: z.number(),
  model: z.string(),
  object: z.string(),
  service_tier: z.union([z.enum(["auto", "default", "flex", "scale", "priority"]), z.null()]).optional(),
  system_fingerprint: z.union([z.string(), z.null()]).optional(),
  usage: z.union([ZCompletionUsage, z.null()]).optional(),
  extraction_id: z.union([z.string(), z.null()]).optional(),
  likelihoods: z.union([z.object({}), z.null()]).optional(),
  schema_validation_error: z.union([ZErrorDetail, z.null()]).optional(),
  request_at: z.union([DateOrISO, z.null()]).optional(),
  first_token_at: z.union([DateOrISO, z.null()]).optional(),
  last_token_at: z.union([DateOrISO, z.null()]).optional(),
}));
export type RetabParsedChatCompletionInput = z.infer<typeof ZRetabParsedChatCompletionInput>;

export const ZRetabParsedChatCompletionOutput = z.lazy(() => z.object({
  id: z.string(),
  choices: z.array(ZRetabParsedChoiceOutput),
  created: z.number(),
  model: z.string(),
  object: z.string(),
  service_tier: z.union([z.enum(["auto", "default", "flex", "scale", "priority"]), z.null()]).optional(),
  system_fingerprint: z.union([z.string(), z.null()]).optional(),
  usage: z.union([ZCompletionUsage, z.null()]).optional(),
  extraction_id: z.union([z.string(), z.null()]).optional(),
  likelihoods: z.union([z.object({}), z.null()]).optional(),
  schema_validation_error: z.union([ZErrorDetail, z.null()]).optional(),
  request_at: z.union([DateOrISO, z.null()]).optional(),
  first_token_at: z.union([DateOrISO, z.null()]).optional(),
  last_token_at: z.union([DateOrISO, z.null()]).optional(),
  api_cost: z.union([ZAmount, z.null()]),
}));
export type RetabParsedChatCompletionOutput = z.infer<typeof ZRetabParsedChatCompletionOutput>;

export const ZRetabParsedChatCompletionChunk = z.lazy(() => z.object({
  id: z.string(),
  choices: z.array(ZRetabParsedChoiceChunk),
  created: z.number(),
  model: z.string(),
  object: z.string(),
  service_tier: z.union([z.enum(["auto", "default", "flex", "scale", "priority"]), z.null()]).optional(),
  system_fingerprint: z.union([z.string(), z.null()]).optional(),
  usage: z.union([ZCompletionUsage, z.null()]).optional(),
  streaming_error: z.union([ZErrorDetail, z.null()]).optional(),
  extraction_id: z.union([z.string(), z.null()]).optional(),
  schema_validation_error: z.union([ZErrorDetail, z.null()]).optional(),
  request_at: z.union([DateOrISO, z.null()]).optional(),
  first_token_at: z.union([DateOrISO, z.null()]).optional(),
  last_token_at: z.union([DateOrISO, z.null()]).optional(),
  api_cost: z.union([ZAmount, z.null()]),
  cost_breakdown: z.union([ZCostBreakdown, z.null()]),
}));
export type RetabParsedChatCompletionChunk = z.infer<typeof ZRetabParsedChatCompletionChunk>;

export const ZRetabParsedChoiceInput = z.lazy(() => z.object({
  finish_reason: z.union([z.enum(["stop", "length", "tool_calls", "content_filter", "function_call"]), z.null()]).optional(),
  index: z.number(),
  logprobs: z.union([ZChoiceLogprobsInput, z.null()]).optional(),
  message: ZParsedChatCompletionMessageInput,
  field_locations: z.union([z.object({}), z.null()]).optional(),
  key_mapping: z.union([z.object({}), z.null()]).optional(),
}));
export type RetabParsedChoiceInput = z.infer<typeof ZRetabParsedChoiceInput>;

export const ZRetabParsedChoiceOutput = z.lazy(() => z.object({
  finish_reason: z.union([z.enum(["stop", "length", "tool_calls", "content_filter", "function_call"]), z.null()]).optional(),
  index: z.number(),
  logprobs: z.union([ZChoiceLogprobsOutput, z.null()]).optional(),
  message: ZParsedChatCompletionMessageOutput,
  field_locations: z.union([z.object({}), z.null()]).optional(),
  key_mapping: z.union([z.object({}), z.null()]).optional(),
}));
export type RetabParsedChoiceOutput = z.infer<typeof ZRetabParsedChoiceOutput>;

export const ZRetabParsedChoiceChunk = z.lazy(() => z.object({
  delta: ZRetabParsedChoiceDeltaChunk,
  finish_reason: z.union([z.enum(["stop", "length", "tool_calls", "content_filter", "function_call"]), z.null()]).optional(),
  index: z.number(),
  logprobs: z.union([ZChoiceLogprobsOutput, z.null()]).optional(),
}));
export type RetabParsedChoiceChunk = z.infer<typeof ZRetabParsedChoiceChunk>;

export const ZRetabParsedChoiceDeltaChunk = z.lazy(() => z.object({
  content: z.union([z.string(), z.null()]).optional(),
  function_call: z.union([ZChoiceDeltaFunctionCall, z.null()]).optional(),
  refusal: z.union([z.string(), z.null()]).optional(),
  role: z.union([z.enum(["developer", "system", "user", "assistant", "tool"]), z.null()]).optional(),
  tool_calls: z.union([z.array(ZChoiceDeltaToolCall), z.null()]).optional(),
  flat_likelihoods: z.object({}).optional(),
  flat_parsed: z.object({}).optional(),
  flat_deleted_keys: z.array(z.string()).optional(),
  field_locations: z.union([z.object({}), z.null()]).optional(),
  is_valid_json: z.boolean().optional(),
  key_mapping: z.union([z.object({}), z.null()]).optional(),
}));
export type RetabParsedChoiceDeltaChunk = z.infer<typeof ZRetabParsedChoiceDeltaChunk>;

export const ZRetabUsage = z.lazy(() => z.object({
  page_count: z.number(),
  credits: z.number(),
}));
export type RetabUsage = z.infer<typeof ZRetabUsage>;

export const ZReviewExtractionRequest = z.lazy(() => z.object({
  extraction: z.union([z.object({}), z.null()]).optional(),
}));
export type ReviewExtractionRequest = z.infer<typeof ZReviewExtractionRequest>;

export const ZSchema = z.lazy(() => z.object({
  object: z.string().optional(),
  created_at: DateOrISO.optional(),
  json_schema: z.object({}).optional(),
  strict: z.boolean().optional(),
  data_id: z.string(),
  id: z.string(),
}));
export type Schema = z.infer<typeof ZSchema>;

export const ZServerToolUsage = z.lazy(() => z.object({
  web_search_requests: z.number(),
}));
export type ServerToolUsage = z.infer<typeof ZServerToolUsage>;

export const ZServerToolUseBlock = z.lazy(() => z.object({
  id: z.string(),
  input: z.any(),
  name: z.string(),
  type: z.string(),
}));
export type ServerToolUseBlock = z.infer<typeof ZServerToolUseBlock>;

export const ZServerToolUseBlockParam = z.lazy(() => z.object({
  id: z.string(),
  input: z.any(),
  name: z.string(),
  type: z.string(),
  cache_control: z.union([ZCacheControlEphemeralParam, z.null()]).optional(),
}));
export type ServerToolUseBlockParam = z.infer<typeof ZServerToolUseBlockParam>;

export const ZSpreadsheetDetails = z.lazy(() => z.object({
  accessible: z.boolean(),
  spreadsheet_id: z.string(),
  spreadsheet_name: z.string(),
  worksheets: z.array(ZGoogleWorksheet),
  spreadsheet_url: z.string(),
  error: z.union([z.string(), z.null()]).optional(),
}));
export type SpreadsheetDetails = z.infer<typeof ZSpreadsheetDetails>;

export const ZStoredDBFile = z.lazy(() => z.object({
  object: z.string().optional(),
  id: z.string(),
  filename: z.string(),
  organization_id: z.string(),
  created_at: DateOrISO.optional(),
  page_count: z.union([z.number(), z.null()]).optional(),
  ocr: z.union([ZOCRMetadata, z.null()]).optional(),
}));
export type StoredDBFile = z.infer<typeof ZStoredDBFile>;

export const ZStoredProcessor = z.lazy(() => z.object({
  object: z.string().optional(),
  id: z.string().optional(),
  updated_at: DateOrISO.optional(),
  name: z.string(),
  modality: z.enum(["text", "image", "native", "image+text"]),
  image_resolution_dpi: z.number().optional(),
  browser_canvas: z.enum(["A3", "A4", "A5"]).optional(),
  model: z.string(),
  json_schema: z.object({}),
  temperature: z.number().optional(),
  reasoning_effort: z.union([z.enum(["low", "medium", "high"]), z.null()]).optional(),
  n_consensus: z.number().optional(),
  organization_id: z.string(),
  schema_data_id: z.string(),
  schema_id: z.string(),
}));
export type StoredProcessor = z.infer<typeof ZStoredProcessor>;

export const ZSubscriptionStatus = z.lazy(() => z.object({
  status: z.enum(["active", "canceled", "incomplete", "incomplete_expired", "past_due", "paused", "trialing", "unpaid"]).optional(),
  plan: z.string().optional(),
  current_period_end: z.number().optional(),
  plan_name: z.string().optional(),
  cancel_at_period_end: z.boolean().optional(),
}));
export type SubscriptionStatus = z.infer<typeof ZSubscriptionStatus>;

export const ZTemplateSchema = z.lazy(() => z.object({
  id: z.string().optional(),
  name: z.string(),
  object: z.string().optional(),
  updated_at: DateOrISO.optional(),
  json_schema: z.object({}).optional(),
  python_code: z.union([z.string(), z.null()]).optional(),
  sample_document_filename: z.union([z.string(), z.null()]).optional(),
  schema_data_id: z.string(),
  schema_id: z.string(),
}));
export type TemplateSchema = z.infer<typeof ZTemplateSchema>;

export const ZTextBlock = z.lazy(() => z.object({
  citations: z.union([z.array(z.union([ZCitationCharLocation, ZCitationPageLocation, ZCitationContentBlockLocation, ZCitationsWebSearchResultLocation])), z.null()]).optional(),
  text: z.string(),
  type: z.string(),
}));
export type TextBlock = z.infer<typeof ZTextBlock>;

export const ZTextBlockParam = z.lazy(() => z.object({
  text: z.string(),
  type: z.string(),
  cache_control: z.union([ZCacheControlEphemeralParam, z.null()]).optional(),
  citations: z.union([z.array(z.union([ZCitationCharLocationParam, ZCitationPageLocationParam, ZCitationContentBlockLocationParam, ZCitationWebSearchResultLocationParam])), z.null()]).optional(),
}));
export type TextBlockParam = z.infer<typeof ZTextBlockParam>;

export const ZTextBox = z.lazy(() => z.object({
  width: z.number(),
  height: z.number(),
  center: ZPoint,
  vertices: z.tuple([ZPoint, ZPoint, ZPoint, ZPoint]),
  text: z.string(),
}));
export type TextBox = z.infer<typeof ZTextBox>;

export const ZThinkingBlock = z.lazy(() => z.object({
  signature: z.string(),
  thinking: z.string(),
  type: z.string(),
}));
export type ThinkingBlock = z.infer<typeof ZThinkingBlock>;

export const ZThinkingBlockParam = z.lazy(() => z.object({
  signature: z.string(),
  thinking: z.string(),
  type: z.string(),
}));
export type ThinkingBlockParam = z.infer<typeof ZThinkingBlockParam>;

export const ZTimeRange = z.lazy(() => z.enum(["day", "week", "month", "three_months"]));
export type TimeRange = z.infer<typeof ZTimeRange>;

export const ZTokenCount = z.lazy(() => z.object({
  total_tokens: z.number().optional(),
  developer_tokens: z.number().optional(),
  user_tokens: z.number().optional(),
}));
export type TokenCount = z.infer<typeof ZTokenCount>;

export const ZTokenCounts = z.lazy(() => z.object({
  prompt_regular_text: z.number(),
  prompt_cached_text: z.number(),
  prompt_audio: z.number(),
  completion_regular_text: z.number(),
  completion_audio: z.number(),
  total_tokens: z.number(),
}));
export type TokenCounts = z.infer<typeof ZTokenCounts>;

export const ZTokenPrice = z.lazy(() => z.object({
  prompt: z.number(),
  completion: z.number(),
  cached_discount: z.number().optional(),
}));
export type TokenPrice = z.infer<typeof ZTokenPrice>;

export const ZToolChoiceFunction = z.lazy(() => z.object({
  name: z.string(),
  type: z.string(),
}));
export type ToolChoiceFunction = z.infer<typeof ZToolChoiceFunction>;

export const ZToolChoiceMcp = z.lazy(() => z.object({
  server_label: z.string(),
  type: z.string(),
  name: z.union([z.string(), z.null()]).optional(),
}));
export type ToolChoiceMcp = z.infer<typeof ZToolChoiceMcp>;

export const ZToolChoiceTypes = z.lazy(() => z.object({
  type: z.enum(["file_search", "web_search_preview", "computer_use_preview", "web_search_preview_2025_03_11", "image_generation", "code_interpreter"]),
}));
export type ToolChoiceTypes = z.infer<typeof ZToolChoiceTypes>;

export const ZToolResultBlockParam = z.lazy(() => z.object({
  tool_use_id: z.string(),
  type: z.string(),
  cache_control: z.union([ZCacheControlEphemeralParam, z.null()]).optional(),
  content: z.union([z.string(), z.array(z.union([ZTextBlockParam, ZImageBlockParam]))]).optional(),
  is_error: z.boolean().optional(),
}));
export type ToolResultBlockParam = z.infer<typeof ZToolResultBlockParam>;

export const ZToolUseBlock = z.lazy(() => z.object({
  id: z.string(),
  input: z.any(),
  name: z.string(),
  type: z.string(),
}));
export type ToolUseBlock = z.infer<typeof ZToolUseBlock>;

export const ZToolUseBlockParam = z.lazy(() => z.object({
  id: z.string(),
  input: z.any(),
  name: z.string(),
  type: z.string(),
  cache_control: z.union([ZCacheControlEphemeralParam, z.null()]).optional(),
}));
export type ToolUseBlockParam = z.infer<typeof ZToolUseBlockParam>;

export const ZTopLogprob = z.lazy(() => z.object({
  token: z.string(),
  bytes: z.union([z.array(z.number()), z.null()]).optional(),
  logprob: z.number(),
}));
export type TopLogprob = z.infer<typeof ZTopLogprob>;

export const ZURLImageSourceParam = z.lazy(() => z.object({
  type: z.string(),
  url: z.string(),
}));
export type URLImageSourceParam = z.infer<typeof ZURLImageSourceParam>;

export const ZURLPDFSourceParam = z.lazy(() => z.object({
  type: z.string(),
  url: z.string(),
}));
export type URLPDFSourceParam = z.infer<typeof ZURLPDFSourceParam>;

export const ZUpdateEmailDataRequest = z.lazy(() => z.object({
  email_data: ZEmailDataInput,
  additional_documents: z.union([z.array(ZAttachmentMIMEDataInput), z.null()]).optional(),
}));
export type UpdateEmailDataRequest = z.infer<typeof ZUpdateEmailDataRequest>;

export const ZUpdateEndpointRequest = z.lazy(() => z.object({
  name: z.union([z.string(), z.null()]).optional(),
  default_language: z.union([z.string(), z.null()]).optional(),
  webhook_url: z.union([z.string(), z.null()]).optional(),
  webhook_headers: z.union([z.object({}), z.null()]).optional(),
  need_validation: z.union([z.boolean(), z.null()]).optional(),
}));
export type UpdateEndpointRequest = z.infer<typeof ZUpdateEndpointRequest>;

export const ZUpdateLinkRequest = z.lazy(() => z.object({
  name: z.union([z.string(), z.null()]).optional(),
  default_language: z.union([z.string(), z.null()]).optional(),
  webhook_url: z.union([z.string(), z.null()]).optional(),
  webhook_headers: z.union([z.object({}), z.null()]).optional(),
  need_validation: z.union([z.boolean(), z.null()]).optional(),
  password: z.union([z.string(), z.null()]).optional(),
}));
export type UpdateLinkRequest = z.infer<typeof ZUpdateLinkRequest>;

export const ZUpdateMailboxRequest = z.lazy(() => z.object({
  name: z.union([z.string(), z.null()]).optional(),
  default_language: z.union([z.string(), z.null()]).optional(),
  webhook_url: z.union([z.string(), z.null()]).optional(),
  webhook_headers: z.union([z.object({}), z.null()]).optional(),
  need_validation: z.union([z.boolean(), z.null()]).optional(),
  authorized_domains: z.union([z.array(z.string()), z.null()]).optional(),
  authorized_emails: z.union([z.array(z.string()), z.null()]).optional(),
}));
export type UpdateMailboxRequest = z.infer<typeof ZUpdateMailboxRequest>;

export const ZUpdateOutlookRequest = z.lazy(() => z.object({
  name: z.union([z.string(), z.null()]).optional(),
  default_language: z.union([z.string(), z.null()]).optional(),
  webhook_url: z.union([z.string(), z.null()]).optional(),
  webhook_headers: z.union([z.object({}), z.null()]).optional(),
  need_validation: z.union([z.boolean(), z.null()]).optional(),
  authorized_domains: z.union([z.array(z.string()), z.null()]).optional(),
  authorized_emails: z.union([z.array(z.string()), z.null()]).optional(),
  match_params: z.union([z.array(ZMatchParams), z.null()]).optional(),
  fetch_params: z.union([z.array(ZFetchParams), z.null()]).optional(),
  layout_schema: z.union([z.object({}), z.null()]).optional(),
}));
export type UpdateOutlookRequest = z.infer<typeof ZUpdateOutlookRequest>;

export const ZUpdateProcessorRequest = z.lazy(() => z.object({
  name: z.union([z.string(), z.null()]).optional(),
  modality: z.union([z.enum(["text", "image", "native", "image+text"]), z.null()]).optional(),
  image_resolution_dpi: z.union([z.number(), z.null()]).optional(),
  browser_canvas: z.union([z.enum(["A3", "A4", "A5"]), z.null()]).optional(),
  model: z.union([z.string(), z.null()]).optional(),
  json_schema: z.union([z.object({}), z.null()]).optional(),
  temperature: z.union([z.number(), z.null()]).optional(),
  reasoning_effort: z.union([z.enum(["low", "medium", "high"]), z.null()]).optional(),
  n_consensus: z.union([z.number(), z.null()]).optional(),
}));
export type UpdateProcessorRequest = z.infer<typeof ZUpdateProcessorRequest>;

export const ZUpdateTemplateRequest = z.lazy(() => z.object({
  id: z.string(),
  name: z.union([z.string(), z.null()]).optional(),
  json_schema: z.union([z.object({}), z.null()]).optional(),
  python_code: z.union([z.string(), z.null()]).optional(),
  sample_document: z.union([ZMIMEDataInput, z.null()]).optional(),
}));
export type UpdateTemplateRequest = z.infer<typeof ZUpdateTemplateRequest>;

export const ZUsage = z.lazy(() => z.object({
  cache_creation_input_tokens: z.union([z.number(), z.null()]).optional(),
  cache_read_input_tokens: z.union([z.number(), z.null()]).optional(),
  input_tokens: z.number(),
  output_tokens: z.number(),
  server_tool_use: z.union([ZServerToolUsage, z.null()]).optional(),
  service_tier: z.union([z.enum(["standard", "priority", "batch"]), z.null()]).optional(),
}));
export type Usage = z.infer<typeof ZUsage>;

export const ZUsageTimeSeries = z.lazy(() => z.object({
  data: z.array(ZDataPoint),
}));
export type UsageTimeSeries = z.infer<typeof ZUsageTimeSeries>;

export const ZUserLocation = z.lazy(() => z.object({
  type: z.string(),
  city: z.union([z.string(), z.null()]).optional(),
  country: z.union([z.string(), z.null()]).optional(),
  region: z.union([z.string(), z.null()]).optional(),
  timezone: z.union([z.string(), z.null()]).optional(),
}));
export type UserLocation = z.infer<typeof ZUserLocation>;

export const ZUserParameters = z.lazy(() => z.object({
  language: z.enum(["fr", "en", "de", "es", "it", "nl", "pt", "pl"]),
  agency_id: z.string(),
  organization_id: z.string(),
  phone_number: z.union([z.string(), z.null()]).optional(),
}));
export type UserParameters = z.infer<typeof ZUserParameters>;

export const ZValidationError = z.lazy(() => z.object({
  loc: z.array(z.union([z.string(), z.number()])),
  msg: z.string(),
  type: z.string(),
}));
export type ValidationError = z.infer<typeof ZValidationError>;

export const ZVectorSearchRequest = z.lazy(() => z.object({
  query_vector: z.array(z.number()),
  organization_id: z.string(),
  schema_id: z.string(),
  limit: z.number().optional(),
  num_candidates: z.number().optional(),
  similarity_metric: z.enum(["cosine", "euclidean", "dotProduct"]).optional(),
}));
export type VectorSearchRequest = z.infer<typeof ZVectorSearchRequest>;

export const ZVectorSearchResponse = z.lazy(() => z.object({
  results: z.array(ZVectorSearchResult),
  average_search_score: z.number(),
}));
export type VectorSearchResponse = z.infer<typeof ZVectorSearchResponse>;

export const ZVectorSearchResult = z.lazy(() => z.object({
  file_id: z.string(),
  search_score: z.number(),
  llm_output: z.object({}),
  hil_output: z.object({}),
  schema_id: z.string(),
  levenshtein_similarity: z.object({}),
  created_at: DateOrISO,
}));
export type VectorSearchResult = z.infer<typeof ZVectorSearchResult>;

export const ZWebSearchResultBlock = z.lazy(() => z.object({
  encrypted_content: z.string(),
  page_age: z.union([z.string(), z.null()]).optional(),
  title: z.string(),
  type: z.string(),
  url: z.string(),
}));
export type WebSearchResultBlock = z.infer<typeof ZWebSearchResultBlock>;

export const ZWebSearchResultBlockParam = z.lazy(() => z.object({
  encrypted_content: z.string(),
  title: z.string(),
  type: z.string(),
  url: z.string(),
  page_age: z.union([z.string(), z.null()]).optional(),
}));
export type WebSearchResultBlockParam = z.infer<typeof ZWebSearchResultBlockParam>;

export const ZWebSearchTool = z.lazy(() => z.object({
  type: z.enum(["web_search_preview", "web_search_preview_2025_03_11"]),
  search_context_size: z.union([z.enum(["low", "medium", "high"]), z.null()]).optional(),
  user_location: z.union([ZUserLocation, z.null()]).optional(),
}));
export type WebSearchTool = z.infer<typeof ZWebSearchTool>;

export const ZWebSearchToolRequestErrorParam = z.lazy(() => z.object({
  error_code: z.enum(["invalid_tool_input", "unavailable", "max_uses_exceeded", "too_many_requests", "query_too_long"]),
  type: z.string(),
}));
export type WebSearchToolRequestErrorParam = z.infer<typeof ZWebSearchToolRequestErrorParam>;

export const ZWebSearchToolResultBlock = z.lazy(() => z.object({
  content: z.union([ZWebSearchToolResultError, z.array(ZWebSearchResultBlock)]),
  tool_use_id: z.string(),
  type: z.string(),
}));
export type WebSearchToolResultBlock = z.infer<typeof ZWebSearchToolResultBlock>;

export const ZWebSearchToolResultBlockParam = z.lazy(() => z.object({
  content: z.union([z.array(ZWebSearchResultBlockParam), ZWebSearchToolRequestErrorParam]),
  tool_use_id: z.string(),
  type: z.string(),
  cache_control: z.union([ZCacheControlEphemeralParam, z.null()]).optional(),
}));
export type WebSearchToolResultBlockParam = z.infer<typeof ZWebSearchToolResultBlockParam>;

export const ZWebSearchToolResultError = z.lazy(() => z.object({
  error_code: z.enum(["invalid_tool_input", "unavailable", "max_uses_exceeded", "too_many_requests", "query_too_long"]),
  type: z.string(),
}));
export type WebSearchToolResultError = z.infer<typeof ZWebSearchToolResultError>;

export const ZWebhookRequest = z.lazy(() => z.object({
  completion: ZRetabParsedChatCompletionInput,
  user: z.union([z.string(), z.null()]).optional(),
  file_payload: ZMIMEDataInput,
  metadata: z.union([z.object({}), z.null()]).optional(),
}));
export type WebhookRequest = z.infer<typeof ZWebhookRequest>;

export const ZWebhookSignature = z.lazy(() => z.object({
  organization_id: z.string(),
  updated_at: DateOrISO,
  signature: z.string(),
}));
export type WebhookSignature = z.infer<typeof ZWebhookSignature>;

export const ZDocumentExtractRequest = z.lazy(() => z.object({
  document: ZMIMEDataInput.optional(),
  documents: z.array(ZMIMEDataInput),
  modality: z.enum(["text", "image", "native", "image+text"]),
  image_resolution_dpi: z.number().optional(),
  browser_canvas: z.enum(["A3", "A4", "A5"]).optional(),
  model: z.string(),
  json_schema: z.object({}),
  temperature: z.number().optional(),
  reasoning_effort: z.union([z.enum(["low", "medium", "high"]), z.null()]).optional(),
  n_consensus: z.number().optional(),
  stream: z.boolean().optional(),
  seed: z.union([z.number(), z.null()]).optional(),
  store: z.boolean().optional(),
  need_validation: z.boolean().optional(),
  test_exception: z.union([z.enum(["before_handle_extraction", "within_extraction_parse_or_stream", "after_handle_extraction", "within_process_document_stream_generator"]), z.null()]).optional(),
}));
export type DocumentExtractRequest = z.infer<typeof ZDocumentExtractRequest>;

export const ZAnthropicTypesMessageMessage = z.lazy(() => z.object({
  id: z.string(),
  content: z.array(z.union([ZTextBlock, ZThinkingBlock, ZRedactedThinkingBlock, ZToolUseBlock, ZServerToolUseBlock, ZWebSearchToolResultBlock])),
  model: z.union([z.enum(["claude-3-7-sonnet-latest", "claude-3-7-sonnet-20250219", "claude-3-5-haiku-latest", "claude-3-5-haiku-20241022", "claude-sonnet-4-20250514", "claude-sonnet-4-0", "claude-4-sonnet-20250514", "claude-3-5-sonnet-latest", "claude-3-5-sonnet-20241022", "claude-3-5-sonnet-20240620", "claude-opus-4-0", "claude-opus-4-20250514", "claude-4-opus-20250514", "claude-3-opus-latest", "claude-3-opus-20240229", "claude-3-sonnet-20240229", "claude-3-haiku-20240307", "claude-2.1", "claude-2.0"]), z.string()]),
  role: z.string(),
  stop_reason: z.union([z.enum(["end_turn", "max_tokens", "stop_sequence", "tool_use", "pause_turn", "refusal"]), z.null()]).optional(),
  stop_sequence: z.union([z.string(), z.null()]).optional(),
  type: z.string(),
  usage: ZUsage,
}));
export type AnthropicTypesMessageMessage = z.infer<typeof ZAnthropicTypesMessageMessage>;

export const ZMainServerServicesCustomBertCubemimedataAttachmentMetadata = z.lazy(() => z.object({
  is_inline: z.boolean().optional(),
  inline_cid: z.union([z.string(), z.null()]).optional(),
  url: z.union([z.string(), z.null()]).optional(),
  display_metadata: z.union([ZDisplayMetadata, z.null()]).optional(),
  ocr: z.union([ZMainServerServicesCustomBertCubemimedataOCR, z.null()]).optional(),
  source: z.union([z.string(), z.null()]).optional(),
}));
export type MainServerServicesCustomBertCubemimedataAttachmentMetadata = z.infer<typeof ZMainServerServicesCustomBertCubemimedataAttachmentMetadata>;

export const ZMainServerServicesCustomBertCubemimedataBaseMIMEData = z.lazy(() => z.object({
  id: z.string(),
  name: z.string(),
  size: z.number(),
  mime_type: z.string(),
  metadata: ZMainServerServicesCustomBertCubemimedataAttachmentMetadata,
}));
export type MainServerServicesCustomBertCubemimedataBaseMIMEData = z.infer<typeof ZMainServerServicesCustomBertCubemimedataBaseMIMEData>;

export const ZMainServerServicesCustomBertCubemimedataMIMEData = z.lazy(() => z.object({
  id: z.string(),
  name: z.string(),
  size: z.number(),
  mime_type: z.string(),
  metadata: ZMainServerServicesCustomBertCubemimedataAttachmentMetadata,
  content: z.string(),
}));
export type MainServerServicesCustomBertCubemimedataMIMEData = z.infer<typeof ZMainServerServicesCustomBertCubemimedataMIMEData>;

export const ZMainServerServicesCustomBertCubemimedataOCR = z.lazy(() => z.object({
  pages: z.array(ZMainServerServicesCustomBertCubemimedataPage),
}));
export type MainServerServicesCustomBertCubemimedataOCR = z.infer<typeof ZMainServerServicesCustomBertCubemimedataOCR>;

export const ZMainServerServicesCustomBertCubemimedataPage = z.lazy(() => z.object({
  page_number: z.number(),
  width: z.number(),
  height: z.number(),
  blocks: z.array(ZTextBox),
  lines: z.array(ZTextBox),
}));
export type MainServerServicesCustomBertCubemimedataPage = z.infer<typeof ZMainServerServicesCustomBertCubemimedataPage>;

export const ZMainServerServicesCustomBertfakeRoutesUser = z.lazy(() => z.object({
  object: z.string(),
  id: z.string(),
  email: z.string(),
  first_name: z.union([z.string(), z.null()]).optional(),
  last_name: z.union([z.string(), z.null()]).optional(),
  email_verified: z.boolean(),
  profile_picture_url: z.union([z.string(), z.null()]).optional(),
  last_sign_in_at: z.union([z.string(), z.null()]).optional(),
  created_at: z.string(),
  updated_at: z.string(),
  external_id: z.union([z.string(), z.null()]).optional(),
  metadata: z.object({}).optional(),
  parameters: ZUserParameters,
}));
export type MainServerServicesCustomBertfakeRoutesUser = z.infer<typeof ZMainServerServicesCustomBertfakeRoutesUser>;

export const ZMainServerServicesInternalBlogModelsUser = z.lazy(() => z.object({
  id: z.enum(["louis-de-benoist", "sacha-ichbiah", "victor-plaisance"]),
  name: z.string(),
  avatarUrl: z.union([z.string(), z.null()]).optional(),
  bio: z.union([z.string(), z.null()]).optional(),
  organization_id: z.union([z.string(), z.null()]).optional(),
}));
export type MainServerServicesInternalBlogModelsUser = z.infer<typeof ZMainServerServicesInternalBlogModelsUser>;

export const ZMainServerServicesV1EvalsDistancesRoutesIterationMetricsFromEvaluationRequest = z.lazy(() => z.object({
  evaluation: ZRetabTypesEvalsEvaluationInput,
  iteration_id: z.string(),
}));
export type MainServerServicesV1EvalsDistancesRoutesIterationMetricsFromEvaluationRequest = z.infer<typeof ZMainServerServicesV1EvalsDistancesRoutesIterationMetricsFromEvaluationRequest>;

export const ZMainServerServicesV1EvalsIoRoutesExportToCsvResponse = z.lazy(() => z.object({
  csv_data: z.string(),
  rows: z.number(),
  columns: z.number(),
}));
export type MainServerServicesV1EvalsIoRoutesExportToCsvResponse = z.infer<typeof ZMainServerServicesV1EvalsIoRoutesExportToCsvResponse>;

export const ZMainServerServicesV1EvalsIterationsRoutesListIterationsResponse = z.lazy(() => z.object({
  data: z.array(ZRetabTypesEvalsIterationOutput),
}));
export type MainServerServicesV1EvalsIterationsRoutesListIterationsResponse = z.infer<typeof ZMainServerServicesV1EvalsIterationsRoutesListIterationsResponse>;

export const ZMainServerServicesV1EvalsRoutesListEvaluations = z.lazy(() => z.object({
  data: z.array(ZRetabTypesEvalsEvaluationOutput),
  list_metadata: ZListMetadata,
}));
export type MainServerServicesV1EvalsRoutesListEvaluations = z.infer<typeof ZMainServerServicesV1EvalsRoutesListEvaluations>;

export const ZMainServerServicesV1EvalsRoutesPatchEvaluationRequest = z.lazy(() => z.object({
  name: z.union([z.string(), z.null()]).optional(),
  project_id: z.union([z.string(), z.null()]).optional(),
  json_schema: z.union([z.object({}), z.null()]).optional(),
}));
export type MainServerServicesV1EvalsRoutesPatchEvaluationRequest = z.infer<typeof ZMainServerServicesV1EvalsRoutesPatchEvaluationRequest>;

export const ZMainServerServicesV1EvaluationsDistancesRoutesIterationMetricsFromEvaluationRequest = z.lazy(() => z.object({
  evaluation: ZRetabTypesEvaluationsModelEvaluationInput,
  iteration_id: z.string(),
}));
export type MainServerServicesV1EvaluationsDistancesRoutesIterationMetricsFromEvaluationRequest = z.infer<typeof ZMainServerServicesV1EvaluationsDistancesRoutesIterationMetricsFromEvaluationRequest>;

export const ZMainServerServicesV1EvaluationsIoRoutesExportToCsvResponse = z.lazy(() => z.object({
  csv_data: z.string(),
  rows: z.number(),
  columns: z.number(),
}));
export type MainServerServicesV1EvaluationsIoRoutesExportToCsvResponse = z.infer<typeof ZMainServerServicesV1EvaluationsIoRoutesExportToCsvResponse>;

export const ZMainServerServicesV1EvaluationsIterationsRoutesListIterationsResponse = z.lazy(() => z.object({
  data: z.array(ZRetabTypesEvaluationsIterationsIterationOutput),
}));
export type MainServerServicesV1EvaluationsIterationsRoutesListIterationsResponse = z.infer<typeof ZMainServerServicesV1EvaluationsIterationsRoutesListIterationsResponse>;

export const ZMainServerServicesV1EvaluationsRoutesListEvaluations = z.lazy(() => z.object({
  data: z.array(ZRetabTypesEvaluationsModelEvaluationOutput),
  list_metadata: ZListMetadata,
}));
export type MainServerServicesV1EvaluationsRoutesListEvaluations = z.infer<typeof ZMainServerServicesV1EvaluationsRoutesListEvaluations>;

export const ZMainServerServicesV1IntegrationsGoogleSheetsRoutesExportToCsvResponse = z.lazy(() => z.object({
  csv_data: z.string(),
  rows: z.number(),
  columns: z.number(),
}));
export type MainServerServicesV1IntegrationsGoogleSheetsRoutesExportToCsvResponse = z.infer<typeof ZMainServerServicesV1IntegrationsGoogleSheetsRoutesExportToCsvResponse>;

export const ZMainServerServicesV1SchemasDefaultTemplatesRoutesCreateTemplateRequest = z.lazy(() => z.object({
  id: z.string().optional(),
  name: z.string(),
  json_schema: z.object({}),
  python_code: z.union([z.string(), z.null()]).optional(),
  sample_document: z.union([ZMIMEDataInput, z.null()]).optional(),
}));
export type MainServerServicesV1SchemasDefaultTemplatesRoutesCreateTemplateRequest = z.infer<typeof ZMainServerServicesV1SchemasDefaultTemplatesRoutesCreateTemplateRequest>;

export const ZMainServerServicesV1SchemasTemplatesRoutesCreateTemplateRequest = z.lazy(() => z.object({
  name: z.string(),
  json_schema: z.object({}),
  python_code: z.union([z.string(), z.null()]).optional(),
  sample_document: z.union([ZMIMEDataInput, z.null()]).optional(),
}));
export type MainServerServicesV1SchemasTemplatesRoutesCreateTemplateRequest = z.infer<typeof ZMainServerServicesV1SchemasTemplatesRoutesCreateTemplateRequest>;

export const ZOpenaiTypesChatChatCompletionAssistantMessageParamFunctionCall = z.lazy(() => z.object({
  arguments: z.string(),
  name: z.string(),
}));
export type OpenaiTypesChatChatCompletionAssistantMessageParamFunctionCall = z.infer<typeof ZOpenaiTypesChatChatCompletionAssistantMessageParamFunctionCall>;

export const ZOpenaiTypesChatChatCompletionMessageAnnotationURLCitation = z.lazy(() => z.object({
  end_index: z.number(),
  start_index: z.number(),
  title: z.string(),
  url: z.string(),
}));
export type OpenaiTypesChatChatCompletionMessageAnnotationURLCitation = z.infer<typeof ZOpenaiTypesChatChatCompletionMessageAnnotationURLCitation>;

export const ZOpenaiTypesChatChatCompletionMessageFunctionCall = z.lazy(() => z.object({
  arguments: z.string(),
  name: z.string(),
}));
export type OpenaiTypesChatChatCompletionMessageFunctionCall = z.infer<typeof ZOpenaiTypesChatChatCompletionMessageFunctionCall>;

export const ZOpenaiTypesChatChatCompletionMessageToolCallFunction = z.lazy(() => z.object({
  arguments: z.string(),
  name: z.string(),
}));
export type OpenaiTypesChatChatCompletionMessageToolCallFunction = z.infer<typeof ZOpenaiTypesChatChatCompletionMessageToolCallFunction>;

export const ZOpenaiTypesChatChatCompletionMessageToolCallParamFunction = z.lazy(() => z.object({
  arguments: z.string(),
  name: z.string(),
}));
export type OpenaiTypesChatChatCompletionMessageToolCallParamFunction = z.infer<typeof ZOpenaiTypesChatChatCompletionMessageToolCallParamFunction>;

export const ZOpenaiTypesResponsesResponseCodeInterpreterToolCallOutputImage = z.lazy(() => z.object({
  type: z.string(),
  url: z.string(),
}));
export type OpenaiTypesResponsesResponseCodeInterpreterToolCallOutputImage = z.infer<typeof ZOpenaiTypesResponsesResponseCodeInterpreterToolCallOutputImage>;

export const ZOpenaiTypesResponsesResponseCodeInterpreterToolCallOutputLogs = z.lazy(() => z.object({
  logs: z.string(),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseCodeInterpreterToolCallOutputLogs = z.infer<typeof ZOpenaiTypesResponsesResponseCodeInterpreterToolCallOutputLogs>;

export const ZOpenaiTypesResponsesResponseCodeInterpreterToolCallParamOutputImage = z.lazy(() => z.object({
  type: z.string(),
  url: z.string(),
}));
export type OpenaiTypesResponsesResponseCodeInterpreterToolCallParamOutputImage = z.infer<typeof ZOpenaiTypesResponsesResponseCodeInterpreterToolCallParamOutputImage>;

export const ZOpenaiTypesResponsesResponseCodeInterpreterToolCallParamOutputLogs = z.lazy(() => z.object({
  logs: z.string(),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseCodeInterpreterToolCallParamOutputLogs = z.infer<typeof ZOpenaiTypesResponsesResponseCodeInterpreterToolCallParamOutputLogs>;

export const ZOpenaiTypesResponsesResponseComputerToolCallActionClick = z.lazy(() => z.object({
  button: z.enum(["left", "right", "wheel", "back", "forward"]),
  type: z.string(),
  x: z.number(),
  y: z.number(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallActionClick = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallActionClick>;

export const ZOpenaiTypesResponsesResponseComputerToolCallActionDoubleClick = z.lazy(() => z.object({
  type: z.string(),
  x: z.number(),
  y: z.number(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallActionDoubleClick = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallActionDoubleClick>;

export const ZOpenaiTypesResponsesResponseComputerToolCallActionDrag = z.lazy(() => z.object({
  path: z.array(ZOpenaiTypesResponsesResponseComputerToolCallActionDragPath),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallActionDrag = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallActionDrag>;

export const ZOpenaiTypesResponsesResponseComputerToolCallActionDragPath = z.lazy(() => z.object({
  x: z.number(),
  y: z.number(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallActionDragPath = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallActionDragPath>;

export const ZOpenaiTypesResponsesResponseComputerToolCallActionKeypress = z.lazy(() => z.object({
  keys: z.array(z.string()),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallActionKeypress = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallActionKeypress>;

export const ZOpenaiTypesResponsesResponseComputerToolCallActionMove = z.lazy(() => z.object({
  type: z.string(),
  x: z.number(),
  y: z.number(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallActionMove = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallActionMove>;

export const ZOpenaiTypesResponsesResponseComputerToolCallActionScreenshot = z.lazy(() => z.object({
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallActionScreenshot = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallActionScreenshot>;

export const ZOpenaiTypesResponsesResponseComputerToolCallActionScroll = z.lazy(() => z.object({
  scroll_x: z.number(),
  scroll_y: z.number(),
  type: z.string(),
  x: z.number(),
  y: z.number(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallActionScroll = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallActionScroll>;

export const ZOpenaiTypesResponsesResponseComputerToolCallActionType = z.lazy(() => z.object({
  text: z.string(),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallActionType = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallActionType>;

export const ZOpenaiTypesResponsesResponseComputerToolCallActionWait = z.lazy(() => z.object({
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallActionWait = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallActionWait>;

export const ZOpenaiTypesResponsesResponseComputerToolCallPendingSafetyCheck = z.lazy(() => z.object({
  id: z.string(),
  code: z.string(),
  message: z.string(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallPendingSafetyCheck = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallPendingSafetyCheck>;

export const ZOpenaiTypesResponsesResponseComputerToolCallParamActionClick = z.lazy(() => z.object({
  button: z.enum(["left", "right", "wheel", "back", "forward"]),
  type: z.string(),
  x: z.number(),
  y: z.number(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallParamActionClick = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallParamActionClick>;

export const ZOpenaiTypesResponsesResponseComputerToolCallParamActionDoubleClick = z.lazy(() => z.object({
  type: z.string(),
  x: z.number(),
  y: z.number(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallParamActionDoubleClick = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallParamActionDoubleClick>;

export const ZOpenaiTypesResponsesResponseComputerToolCallParamActionDrag = z.lazy(() => z.object({
  path: z.array(ZOpenaiTypesResponsesResponseComputerToolCallParamActionDragPath),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallParamActionDrag = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallParamActionDrag>;

export const ZOpenaiTypesResponsesResponseComputerToolCallParamActionDragPath = z.lazy(() => z.object({
  x: z.number(),
  y: z.number(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallParamActionDragPath = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallParamActionDragPath>;

export const ZOpenaiTypesResponsesResponseComputerToolCallParamActionKeypress = z.lazy(() => z.object({
  keys: z.array(z.string()),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallParamActionKeypress = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallParamActionKeypress>;

export const ZOpenaiTypesResponsesResponseComputerToolCallParamActionMove = z.lazy(() => z.object({
  type: z.string(),
  x: z.number(),
  y: z.number(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallParamActionMove = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallParamActionMove>;

export const ZOpenaiTypesResponsesResponseComputerToolCallParamActionScreenshot = z.lazy(() => z.object({
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallParamActionScreenshot = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallParamActionScreenshot>;

export const ZOpenaiTypesResponsesResponseComputerToolCallParamActionScroll = z.lazy(() => z.object({
  scroll_x: z.number(),
  scroll_y: z.number(),
  type: z.string(),
  x: z.number(),
  y: z.number(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallParamActionScroll = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallParamActionScroll>;

export const ZOpenaiTypesResponsesResponseComputerToolCallParamActionType = z.lazy(() => z.object({
  text: z.string(),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallParamActionType = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallParamActionType>;

export const ZOpenaiTypesResponsesResponseComputerToolCallParamActionWait = z.lazy(() => z.object({
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallParamActionWait = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallParamActionWait>;

export const ZOpenaiTypesResponsesResponseComputerToolCallParamPendingSafetyCheck = z.lazy(() => z.object({
  id: z.string(),
  code: z.string(),
  message: z.string(),
}));
export type OpenaiTypesResponsesResponseComputerToolCallParamPendingSafetyCheck = z.infer<typeof ZOpenaiTypesResponsesResponseComputerToolCallParamPendingSafetyCheck>;

export const ZOpenaiTypesResponsesResponseFileSearchToolCallResult = z.lazy(() => z.object({
  attributes: z.union([z.object({}), z.null()]).optional(),
  file_id: z.union([z.string(), z.null()]).optional(),
  filename: z.union([z.string(), z.null()]).optional(),
  score: z.union([z.number(), z.null()]).optional(),
  text: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseFileSearchToolCallResult = z.infer<typeof ZOpenaiTypesResponsesResponseFileSearchToolCallResult>;

export const ZOpenaiTypesResponsesResponseFileSearchToolCallParamResult = z.lazy(() => z.object({
  attributes: z.union([z.object({}), z.null()]).optional(),
  file_id: z.string().optional(),
  filename: z.string().optional(),
  score: z.number().optional(),
  text: z.string().optional(),
}));
export type OpenaiTypesResponsesResponseFileSearchToolCallParamResult = z.infer<typeof ZOpenaiTypesResponsesResponseFileSearchToolCallParamResult>;

export const ZOpenaiTypesResponsesResponseFunctionWebSearchActionFind = z.lazy(() => z.object({
  pattern: z.string(),
  type: z.string(),
  url: z.string(),
}));
export type OpenaiTypesResponsesResponseFunctionWebSearchActionFind = z.infer<typeof ZOpenaiTypesResponsesResponseFunctionWebSearchActionFind>;

export const ZOpenaiTypesResponsesResponseFunctionWebSearchActionOpenPage = z.lazy(() => z.object({
  type: z.string(),
  url: z.string(),
}));
export type OpenaiTypesResponsesResponseFunctionWebSearchActionOpenPage = z.infer<typeof ZOpenaiTypesResponsesResponseFunctionWebSearchActionOpenPage>;

export const ZOpenaiTypesResponsesResponseFunctionWebSearchActionSearch = z.lazy(() => z.object({
  query: z.string(),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseFunctionWebSearchActionSearch = z.infer<typeof ZOpenaiTypesResponsesResponseFunctionWebSearchActionSearch>;

export const ZOpenaiTypesResponsesResponseFunctionWebSearchParamActionFind = z.lazy(() => z.object({
  pattern: z.string(),
  type: z.string(),
  url: z.string(),
}));
export type OpenaiTypesResponsesResponseFunctionWebSearchParamActionFind = z.infer<typeof ZOpenaiTypesResponsesResponseFunctionWebSearchParamActionFind>;

export const ZOpenaiTypesResponsesResponseFunctionWebSearchParamActionOpenPage = z.lazy(() => z.object({
  type: z.string(),
  url: z.string(),
}));
export type OpenaiTypesResponsesResponseFunctionWebSearchParamActionOpenPage = z.infer<typeof ZOpenaiTypesResponsesResponseFunctionWebSearchParamActionOpenPage>;

export const ZOpenaiTypesResponsesResponseFunctionWebSearchParamActionSearch = z.lazy(() => z.object({
  query: z.string(),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseFunctionWebSearchParamActionSearch = z.infer<typeof ZOpenaiTypesResponsesResponseFunctionWebSearchParamActionSearch>;

export const ZOpenaiTypesResponsesResponseInputItemComputerCallOutput = z.lazy(() => z.object({
  call_id: z.string(),
  output: ZResponseComputerToolCallOutputScreenshot,
  type: z.string(),
  id: z.union([z.string(), z.null()]).optional(),
  acknowledged_safety_checks: z.union([z.array(ZOpenaiTypesResponsesResponseInputItemComputerCallOutputAcknowledgedSafetyCheck), z.null()]).optional(),
  status: z.union([z.enum(["in_progress", "completed", "incomplete"]), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputItemComputerCallOutput = z.infer<typeof ZOpenaiTypesResponsesResponseInputItemComputerCallOutput>;

export const ZOpenaiTypesResponsesResponseInputItemComputerCallOutputAcknowledgedSafetyCheck = z.lazy(() => z.object({
  id: z.string(),
  code: z.union([z.string(), z.null()]).optional(),
  message: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputItemComputerCallOutputAcknowledgedSafetyCheck = z.infer<typeof ZOpenaiTypesResponsesResponseInputItemComputerCallOutputAcknowledgedSafetyCheck>;

export const ZOpenaiTypesResponsesResponseInputItemFunctionCallOutput = z.lazy(() => z.object({
  call_id: z.string(),
  output: z.string(),
  type: z.string(),
  id: z.union([z.string(), z.null()]).optional(),
  status: z.union([z.enum(["in_progress", "completed", "incomplete"]), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputItemFunctionCallOutput = z.infer<typeof ZOpenaiTypesResponsesResponseInputItemFunctionCallOutput>;

export const ZOpenaiTypesResponsesResponseInputItemImageGenerationCall = z.lazy(() => z.object({
  id: z.string(),
  result: z.union([z.string(), z.null()]).optional(),
  status: z.enum(["in_progress", "completed", "generating", "failed"]),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseInputItemImageGenerationCall = z.infer<typeof ZOpenaiTypesResponsesResponseInputItemImageGenerationCall>;

export const ZOpenaiTypesResponsesResponseInputItemItemReference = z.lazy(() => z.object({
  id: z.string(),
  type: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputItemItemReference = z.infer<typeof ZOpenaiTypesResponsesResponseInputItemItemReference>;

export const ZOpenaiTypesResponsesResponseInputItemLocalShellCall = z.lazy(() => z.object({
  id: z.string(),
  action: ZOpenaiTypesResponsesResponseInputItemLocalShellCallAction,
  call_id: z.string(),
  status: z.enum(["in_progress", "completed", "incomplete"]),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseInputItemLocalShellCall = z.infer<typeof ZOpenaiTypesResponsesResponseInputItemLocalShellCall>;

export const ZOpenaiTypesResponsesResponseInputItemLocalShellCallAction = z.lazy(() => z.object({
  command: z.array(z.string()),
  env: z.object({}),
  type: z.string(),
  timeout_ms: z.union([z.number(), z.null()]).optional(),
  user: z.union([z.string(), z.null()]).optional(),
  working_directory: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputItemLocalShellCallAction = z.infer<typeof ZOpenaiTypesResponsesResponseInputItemLocalShellCallAction>;

export const ZOpenaiTypesResponsesResponseInputItemLocalShellCallOutput = z.lazy(() => z.object({
  id: z.string(),
  output: z.string(),
  type: z.string(),
  status: z.union([z.enum(["in_progress", "completed", "incomplete"]), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputItemLocalShellCallOutput = z.infer<typeof ZOpenaiTypesResponsesResponseInputItemLocalShellCallOutput>;

export const ZOpenaiTypesResponsesResponseInputItemMcpApprovalRequest = z.lazy(() => z.object({
  id: z.string(),
  arguments: z.string(),
  name: z.string(),
  server_label: z.string(),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseInputItemMcpApprovalRequest = z.infer<typeof ZOpenaiTypesResponsesResponseInputItemMcpApprovalRequest>;

export const ZOpenaiTypesResponsesResponseInputItemMcpApprovalResponse = z.lazy(() => z.object({
  approval_request_id: z.string(),
  approve: z.boolean(),
  type: z.string(),
  id: z.union([z.string(), z.null()]).optional(),
  reason: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputItemMcpApprovalResponse = z.infer<typeof ZOpenaiTypesResponsesResponseInputItemMcpApprovalResponse>;

export const ZOpenaiTypesResponsesResponseInputItemMcpCall = z.lazy(() => z.object({
  id: z.string(),
  arguments: z.string(),
  name: z.string(),
  server_label: z.string(),
  type: z.string(),
  error: z.union([z.string(), z.null()]).optional(),
  output: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputItemMcpCall = z.infer<typeof ZOpenaiTypesResponsesResponseInputItemMcpCall>;

export const ZOpenaiTypesResponsesResponseInputItemMcpListTools = z.lazy(() => z.object({
  id: z.string(),
  server_label: z.string(),
  tools: z.array(ZOpenaiTypesResponsesResponseInputItemMcpListToolsTool),
  type: z.string(),
  error: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputItemMcpListTools = z.infer<typeof ZOpenaiTypesResponsesResponseInputItemMcpListTools>;

export const ZOpenaiTypesResponsesResponseInputItemMcpListToolsTool = z.lazy(() => z.object({
  input_schema: z.any(),
  name: z.string(),
  annotations: z.union([z.any(), z.null()]).optional(),
  description: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputItemMcpListToolsTool = z.infer<typeof ZOpenaiTypesResponsesResponseInputItemMcpListToolsTool>;

export const ZOpenaiTypesResponsesResponseInputItemMessage = z.lazy(() => z.object({
  content: z.array(z.union([ZResponseInputText, ZResponseInputImage, ZResponseInputFile])),
  role: z.enum(["user", "system", "developer"]),
  status: z.union([z.enum(["in_progress", "completed", "incomplete"]), z.null()]).optional(),
  type: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputItemMessage = z.infer<typeof ZOpenaiTypesResponsesResponseInputItemMessage>;

export const ZOpenaiTypesResponsesResponseInputParamComputerCallOutput = z.lazy(() => z.object({
  call_id: z.string(),
  output: ZResponseComputerToolCallOutputScreenshotParam,
  type: z.string(),
  id: z.union([z.string(), z.null()]).optional(),
  acknowledged_safety_checks: z.union([z.array(ZOpenaiTypesResponsesResponseInputParamComputerCallOutputAcknowledgedSafetyCheck), z.null()]).optional(),
  status: z.union([z.enum(["in_progress", "completed", "incomplete"]), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputParamComputerCallOutput = z.infer<typeof ZOpenaiTypesResponsesResponseInputParamComputerCallOutput>;

export const ZOpenaiTypesResponsesResponseInputParamComputerCallOutputAcknowledgedSafetyCheck = z.lazy(() => z.object({
  id: z.string(),
  code: z.union([z.string(), z.null()]).optional(),
  message: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputParamComputerCallOutputAcknowledgedSafetyCheck = z.infer<typeof ZOpenaiTypesResponsesResponseInputParamComputerCallOutputAcknowledgedSafetyCheck>;

export const ZOpenaiTypesResponsesResponseInputParamFunctionCallOutput = z.lazy(() => z.object({
  call_id: z.string(),
  output: z.string(),
  type: z.string(),
  id: z.union([z.string(), z.null()]).optional(),
  status: z.union([z.enum(["in_progress", "completed", "incomplete"]), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputParamFunctionCallOutput = z.infer<typeof ZOpenaiTypesResponsesResponseInputParamFunctionCallOutput>;

export const ZOpenaiTypesResponsesResponseInputParamImageGenerationCall = z.lazy(() => z.object({
  id: z.string(),
  result: z.union([z.string(), z.null()]),
  status: z.enum(["in_progress", "completed", "generating", "failed"]),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseInputParamImageGenerationCall = z.infer<typeof ZOpenaiTypesResponsesResponseInputParamImageGenerationCall>;

export const ZOpenaiTypesResponsesResponseInputParamItemReference = z.lazy(() => z.object({
  id: z.string(),
  type: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputParamItemReference = z.infer<typeof ZOpenaiTypesResponsesResponseInputParamItemReference>;

export const ZOpenaiTypesResponsesResponseInputParamLocalShellCall = z.lazy(() => z.object({
  id: z.string(),
  action: ZOpenaiTypesResponsesResponseInputParamLocalShellCallAction,
  call_id: z.string(),
  status: z.enum(["in_progress", "completed", "incomplete"]),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseInputParamLocalShellCall = z.infer<typeof ZOpenaiTypesResponsesResponseInputParamLocalShellCall>;

export const ZOpenaiTypesResponsesResponseInputParamLocalShellCallAction = z.lazy(() => z.object({
  command: z.array(z.string()),
  env: z.object({}),
  type: z.string(),
  timeout_ms: z.union([z.number(), z.null()]).optional(),
  user: z.union([z.string(), z.null()]).optional(),
  working_directory: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputParamLocalShellCallAction = z.infer<typeof ZOpenaiTypesResponsesResponseInputParamLocalShellCallAction>;

export const ZOpenaiTypesResponsesResponseInputParamLocalShellCallOutput = z.lazy(() => z.object({
  id: z.string(),
  output: z.string(),
  type: z.string(),
  status: z.union([z.enum(["in_progress", "completed", "incomplete"]), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputParamLocalShellCallOutput = z.infer<typeof ZOpenaiTypesResponsesResponseInputParamLocalShellCallOutput>;

export const ZOpenaiTypesResponsesResponseInputParamMcpApprovalRequest = z.lazy(() => z.object({
  id: z.string(),
  arguments: z.string(),
  name: z.string(),
  server_label: z.string(),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseInputParamMcpApprovalRequest = z.infer<typeof ZOpenaiTypesResponsesResponseInputParamMcpApprovalRequest>;

export const ZOpenaiTypesResponsesResponseInputParamMcpApprovalResponse = z.lazy(() => z.object({
  approval_request_id: z.string(),
  approve: z.boolean(),
  type: z.string(),
  id: z.union([z.string(), z.null()]).optional(),
  reason: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputParamMcpApprovalResponse = z.infer<typeof ZOpenaiTypesResponsesResponseInputParamMcpApprovalResponse>;

export const ZOpenaiTypesResponsesResponseInputParamMcpCall = z.lazy(() => z.object({
  id: z.string(),
  arguments: z.string(),
  name: z.string(),
  server_label: z.string(),
  type: z.string(),
  error: z.union([z.string(), z.null()]).optional(),
  output: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputParamMcpCall = z.infer<typeof ZOpenaiTypesResponsesResponseInputParamMcpCall>;

export const ZOpenaiTypesResponsesResponseInputParamMcpListTools = z.lazy(() => z.object({
  id: z.string(),
  server_label: z.string(),
  tools: z.array(ZOpenaiTypesResponsesResponseInputParamMcpListToolsTool),
  type: z.string(),
  error: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputParamMcpListTools = z.infer<typeof ZOpenaiTypesResponsesResponseInputParamMcpListTools>;

export const ZOpenaiTypesResponsesResponseInputParamMcpListToolsTool = z.lazy(() => z.object({
  input_schema: z.any(),
  name: z.string(),
  annotations: z.union([z.any(), z.null()]).optional(),
  description: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseInputParamMcpListToolsTool = z.infer<typeof ZOpenaiTypesResponsesResponseInputParamMcpListToolsTool>;

export const ZOpenaiTypesResponsesResponseInputParamMessage = z.lazy(() => z.object({
  content: z.array(z.union([ZResponseInputTextParam, ZResponseInputImageParam, ZResponseInputFileParam])),
  role: z.enum(["user", "system", "developer"]),
  status: z.enum(["in_progress", "completed", "incomplete"]).optional(),
  type: z.string().optional(),
}));
export type OpenaiTypesResponsesResponseInputParamMessage = z.infer<typeof ZOpenaiTypesResponsesResponseInputParamMessage>;

export const ZOpenaiTypesResponsesResponseOutputItemImageGenerationCall = z.lazy(() => z.object({
  id: z.string(),
  result: z.union([z.string(), z.null()]).optional(),
  status: z.enum(["in_progress", "completed", "generating", "failed"]),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseOutputItemImageGenerationCall = z.infer<typeof ZOpenaiTypesResponsesResponseOutputItemImageGenerationCall>;

export const ZOpenaiTypesResponsesResponseOutputItemLocalShellCall = z.lazy(() => z.object({
  id: z.string(),
  action: ZOpenaiTypesResponsesResponseOutputItemLocalShellCallAction,
  call_id: z.string(),
  status: z.enum(["in_progress", "completed", "incomplete"]),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseOutputItemLocalShellCall = z.infer<typeof ZOpenaiTypesResponsesResponseOutputItemLocalShellCall>;

export const ZOpenaiTypesResponsesResponseOutputItemLocalShellCallAction = z.lazy(() => z.object({
  command: z.array(z.string()),
  env: z.object({}),
  type: z.string(),
  timeout_ms: z.union([z.number(), z.null()]).optional(),
  user: z.union([z.string(), z.null()]).optional(),
  working_directory: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseOutputItemLocalShellCallAction = z.infer<typeof ZOpenaiTypesResponsesResponseOutputItemLocalShellCallAction>;

export const ZOpenaiTypesResponsesResponseOutputItemMcpApprovalRequest = z.lazy(() => z.object({
  id: z.string(),
  arguments: z.string(),
  name: z.string(),
  server_label: z.string(),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseOutputItemMcpApprovalRequest = z.infer<typeof ZOpenaiTypesResponsesResponseOutputItemMcpApprovalRequest>;

export const ZOpenaiTypesResponsesResponseOutputItemMcpCall = z.lazy(() => z.object({
  id: z.string(),
  arguments: z.string(),
  name: z.string(),
  server_label: z.string(),
  type: z.string(),
  error: z.union([z.string(), z.null()]).optional(),
  output: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseOutputItemMcpCall = z.infer<typeof ZOpenaiTypesResponsesResponseOutputItemMcpCall>;

export const ZOpenaiTypesResponsesResponseOutputItemMcpListTools = z.lazy(() => z.object({
  id: z.string(),
  server_label: z.string(),
  tools: z.array(ZOpenaiTypesResponsesResponseOutputItemMcpListToolsTool),
  type: z.string(),
  error: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseOutputItemMcpListTools = z.infer<typeof ZOpenaiTypesResponsesResponseOutputItemMcpListTools>;

export const ZOpenaiTypesResponsesResponseOutputItemMcpListToolsTool = z.lazy(() => z.object({
  input_schema: z.any(),
  name: z.string(),
  annotations: z.union([z.any(), z.null()]).optional(),
  description: z.union([z.string(), z.null()]).optional(),
}));
export type OpenaiTypesResponsesResponseOutputItemMcpListToolsTool = z.infer<typeof ZOpenaiTypesResponsesResponseOutputItemMcpListToolsTool>;

export const ZOpenaiTypesResponsesResponseOutputTextAnnotationContainerFileCitation = z.lazy(() => z.object({
  container_id: z.string(),
  end_index: z.number(),
  file_id: z.string(),
  filename: z.string(),
  start_index: z.number(),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseOutputTextAnnotationContainerFileCitation = z.infer<typeof ZOpenaiTypesResponsesResponseOutputTextAnnotationContainerFileCitation>;

export const ZOpenaiTypesResponsesResponseOutputTextAnnotationFileCitation = z.lazy(() => z.object({
  file_id: z.string(),
  filename: z.string(),
  index: z.number(),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseOutputTextAnnotationFileCitation = z.infer<typeof ZOpenaiTypesResponsesResponseOutputTextAnnotationFileCitation>;

export const ZOpenaiTypesResponsesResponseOutputTextAnnotationFilePath = z.lazy(() => z.object({
  file_id: z.string(),
  index: z.number(),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseOutputTextAnnotationFilePath = z.infer<typeof ZOpenaiTypesResponsesResponseOutputTextAnnotationFilePath>;

export const ZOpenaiTypesResponsesResponseOutputTextAnnotationURLCitation = z.lazy(() => z.object({
  end_index: z.number(),
  start_index: z.number(),
  title: z.string(),
  type: z.string(),
  url: z.string(),
}));
export type OpenaiTypesResponsesResponseOutputTextAnnotationURLCitation = z.infer<typeof ZOpenaiTypesResponsesResponseOutputTextAnnotationURLCitation>;

export const ZOpenaiTypesResponsesResponseOutputTextLogprob = z.lazy(() => z.object({
  token: z.string(),
  bytes: z.array(z.number()),
  logprob: z.number(),
  top_logprobs: z.array(ZOpenaiTypesResponsesResponseOutputTextLogprobTopLogprob),
}));
export type OpenaiTypesResponsesResponseOutputTextLogprob = z.infer<typeof ZOpenaiTypesResponsesResponseOutputTextLogprob>;

export const ZOpenaiTypesResponsesResponseOutputTextLogprobTopLogprob = z.lazy(() => z.object({
  token: z.string(),
  bytes: z.array(z.number()),
  logprob: z.number(),
}));
export type OpenaiTypesResponsesResponseOutputTextLogprobTopLogprob = z.infer<typeof ZOpenaiTypesResponsesResponseOutputTextLogprobTopLogprob>;

export const ZOpenaiTypesResponsesResponseOutputTextParamAnnotationContainerFileCitation = z.lazy(() => z.object({
  container_id: z.string(),
  end_index: z.number(),
  file_id: z.string(),
  filename: z.string(),
  start_index: z.number(),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseOutputTextParamAnnotationContainerFileCitation = z.infer<typeof ZOpenaiTypesResponsesResponseOutputTextParamAnnotationContainerFileCitation>;

export const ZOpenaiTypesResponsesResponseOutputTextParamAnnotationFileCitation = z.lazy(() => z.object({
  file_id: z.string(),
  filename: z.string(),
  index: z.number(),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseOutputTextParamAnnotationFileCitation = z.infer<typeof ZOpenaiTypesResponsesResponseOutputTextParamAnnotationFileCitation>;

export const ZOpenaiTypesResponsesResponseOutputTextParamAnnotationFilePath = z.lazy(() => z.object({
  file_id: z.string(),
  index: z.number(),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseOutputTextParamAnnotationFilePath = z.infer<typeof ZOpenaiTypesResponsesResponseOutputTextParamAnnotationFilePath>;

export const ZOpenaiTypesResponsesResponseOutputTextParamAnnotationURLCitation = z.lazy(() => z.object({
  end_index: z.number(),
  start_index: z.number(),
  title: z.string(),
  type: z.string(),
  url: z.string(),
}));
export type OpenaiTypesResponsesResponseOutputTextParamAnnotationURLCitation = z.infer<typeof ZOpenaiTypesResponsesResponseOutputTextParamAnnotationURLCitation>;

export const ZOpenaiTypesResponsesResponseOutputTextParamLogprob = z.lazy(() => z.object({
  token: z.string(),
  bytes: z.array(z.number()),
  logprob: z.number(),
  top_logprobs: z.array(ZOpenaiTypesResponsesResponseOutputTextParamLogprobTopLogprob),
}));
export type OpenaiTypesResponsesResponseOutputTextParamLogprob = z.infer<typeof ZOpenaiTypesResponsesResponseOutputTextParamLogprob>;

export const ZOpenaiTypesResponsesResponseOutputTextParamLogprobTopLogprob = z.lazy(() => z.object({
  token: z.string(),
  bytes: z.array(z.number()),
  logprob: z.number(),
}));
export type OpenaiTypesResponsesResponseOutputTextParamLogprobTopLogprob = z.infer<typeof ZOpenaiTypesResponsesResponseOutputTextParamLogprobTopLogprob>;

export const ZOpenaiTypesResponsesResponseReasoningItemSummary = z.lazy(() => z.object({
  text: z.string(),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseReasoningItemSummary = z.infer<typeof ZOpenaiTypesResponsesResponseReasoningItemSummary>;

export const ZOpenaiTypesResponsesResponseReasoningItemParamSummary = z.lazy(() => z.object({
  text: z.string(),
  type: z.string(),
}));
export type OpenaiTypesResponsesResponseReasoningItemParamSummary = z.infer<typeof ZOpenaiTypesResponsesResponseReasoningItemParamSummary>;

export const ZOpenaiTypesSharedReasoningReasoning = z.lazy(() => z.object({
  effort: z.union([z.enum(["low", "medium", "high"]), z.null()]).optional(),
  generate_summary: z.union([z.enum(["auto", "concise", "detailed"]), z.null()]).optional(),
  summary: z.union([z.enum(["auto", "concise", "detailed"]), z.null()]).optional(),
}));
export type OpenaiTypesSharedReasoningReasoning = z.infer<typeof ZOpenaiTypesSharedReasoningReasoning>;

export const ZOpenaiTypesSharedResponseFormatJsonObjectResponseFormatJSONObject = z.lazy(() => z.object({
  type: z.string(),
}));
export type OpenaiTypesSharedResponseFormatJsonObjectResponseFormatJSONObject = z.infer<typeof ZOpenaiTypesSharedResponseFormatJsonObjectResponseFormatJSONObject>;

export const ZOpenaiTypesSharedResponseFormatTextResponseFormatText = z.lazy(() => z.object({
  type: z.string(),
}));
export type OpenaiTypesSharedResponseFormatTextResponseFormatText = z.infer<typeof ZOpenaiTypesSharedResponseFormatTextResponseFormatText>;

export const ZOpenaiTypesSharedParamsReasoningReasoning = z.lazy(() => z.object({
  effort: z.union([z.enum(["low", "medium", "high"]), z.null()]).optional(),
  generate_summary: z.union([z.enum(["auto", "concise", "detailed"]), z.null()]).optional(),
  summary: z.union([z.enum(["auto", "concise", "detailed"]), z.null()]).optional(),
}));
export type OpenaiTypesSharedParamsReasoningReasoning = z.infer<typeof ZOpenaiTypesSharedParamsReasoningReasoning>;

export const ZOpenaiTypesSharedParamsResponseFormatJsonObjectResponseFormatJSONObject = z.lazy(() => z.object({
  type: z.string(),
}));
export type OpenaiTypesSharedParamsResponseFormatJsonObjectResponseFormatJSONObject = z.infer<typeof ZOpenaiTypesSharedParamsResponseFormatJsonObjectResponseFormatJSONObject>;

export const ZOpenaiTypesSharedParamsResponseFormatTextResponseFormatText = z.lazy(() => z.object({
  type: z.string(),
}));
export type OpenaiTypesSharedParamsResponseFormatTextResponseFormatText = z.infer<typeof ZOpenaiTypesSharedParamsResponseFormatTextResponseFormatText>;

export const ZRetabTypesEvalsCreateIterationRequest = z.lazy(() => z.object({
  inference_settings: ZInferenceSettings,
  json_schema: z.union([z.object({}), z.null()]).optional(),
}));
export type RetabTypesEvalsCreateIterationRequest = z.infer<typeof ZRetabTypesEvalsCreateIterationRequest>;

export const ZRetabTypesEvalsEvaluationInput = z.lazy(() => z.object({
  id: z.string().optional(),
  updated_at: DateOrISO.optional(),
  name: z.string(),
  old_documents: z.union([z.array(ZEvaluationDocumentInput), z.null()]).optional(),
  documents: z.array(ZEvaluationDocumentInput),
  iterations: z.array(ZRetabTypesEvalsIterationInput),
  json_schema: z.object({}),
  project_id: z.string().optional(),
  default_inference_settings: z.union([ZInferenceSettings, z.null()]).optional(),
}));
export type RetabTypesEvalsEvaluationInput = z.infer<typeof ZRetabTypesEvalsEvaluationInput>;

export const ZRetabTypesEvalsEvaluationOutput = z.lazy(() => z.object({
  id: z.string().optional(),
  updated_at: DateOrISO.optional(),
  name: z.string(),
  old_documents: z.union([z.array(ZEvaluationDocumentOutput), z.null()]).optional(),
  documents: z.array(ZEvaluationDocumentOutput),
  iterations: z.array(ZRetabTypesEvalsIterationOutput),
  json_schema: z.object({}),
  project_id: z.string().optional(),
  default_inference_settings: z.union([ZInferenceSettings, z.null()]).optional(),
  schema_data_id: z.string(),
  schema_id: z.string(),
}));
export type RetabTypesEvalsEvaluationOutput = z.infer<typeof ZRetabTypesEvalsEvaluationOutput>;

export const ZRetabTypesEvalsIterationInput = z.lazy(() => z.object({
  id: z.string().optional(),
  inference_settings: ZInferenceSettings,
  json_schema: z.object({}),
  predictions: z.array(ZRetabTypesEvalsPredictionDataInput).optional(),
  metric_results: z.union([ZMetricResult, z.null()]).optional(),
}));
export type RetabTypesEvalsIterationInput = z.infer<typeof ZRetabTypesEvalsIterationInput>;

export const ZRetabTypesEvalsIterationOutput = z.lazy(() => z.object({
  id: z.string().optional(),
  inference_settings: ZInferenceSettings,
  json_schema: z.object({}),
  predictions: z.array(ZRetabTypesEvalsPredictionDataOutput).optional(),
  metric_results: z.union([ZMetricResult, z.null()]).optional(),
  schema_data_id: z.string(),
  schema_id: z.string(),
}));
export type RetabTypesEvalsIterationOutput = z.infer<typeof ZRetabTypesEvalsIterationOutput>;

export const ZRetabTypesEvalsPredictionDataInput = z.lazy(() => z.object({
  prediction: z.object({}).optional(),
  metadata: z.union([ZPredictionMetadata, z.null()]).optional(),
}));
export type RetabTypesEvalsPredictionDataInput = z.infer<typeof ZRetabTypesEvalsPredictionDataInput>;

export const ZRetabTypesEvalsPredictionDataOutput = z.lazy(() => z.object({
  prediction: z.object({}).optional(),
  metadata: z.union([ZPredictionMetadata, z.null()]).optional(),
}));
export type RetabTypesEvalsPredictionDataOutput = z.infer<typeof ZRetabTypesEvalsPredictionDataOutput>;

export const ZRetabTypesEvaluationsIterationsCreateIterationRequest = z.lazy(() => z.object({
  inference_settings: ZInferenceSettings,
  json_schema: z.union([z.object({}), z.null()]).optional(),
  from_iteration_id: z.union([z.string(), z.null()]).optional(),
}));
export type RetabTypesEvaluationsIterationsCreateIterationRequest = z.infer<typeof ZRetabTypesEvaluationsIterationsCreateIterationRequest>;

export const ZRetabTypesEvaluationsIterationsIterationInput = z.lazy(() => z.object({
  id: z.string().optional(),
  updated_at: DateOrISO.optional(),
  inference_settings: ZInferenceSettings,
  json_schema: z.object({}),
  predictions: z.object({}).optional(),
  metric_results: z.union([ZMetricResult, z.null()]).optional(),
}));
export type RetabTypesEvaluationsIterationsIterationInput = z.infer<typeof ZRetabTypesEvaluationsIterationsIterationInput>;

export const ZRetabTypesEvaluationsIterationsIterationOutput = z.lazy(() => z.object({
  id: z.string().optional(),
  updated_at: DateOrISO.optional(),
  inference_settings: ZInferenceSettings,
  json_schema: z.object({}),
  predictions: z.object({}).optional(),
  metric_results: z.union([ZMetricResult, z.null()]).optional(),
  schema_data_id: z.string(),
  schema_id: z.string(),
}));
export type RetabTypesEvaluationsIterationsIterationOutput = z.infer<typeof ZRetabTypesEvaluationsIterationsIterationOutput>;

export const ZRetabTypesEvaluationsModelEvaluationInput = z.lazy(() => z.object({
  id: z.string().optional(),
  updated_at: DateOrISO.optional(),
  name: z.string(),
  documents: z.array(ZEvaluationDocumentInput).optional(),
  iterations: z.array(ZRetabTypesEvaluationsIterationsIterationInput).optional(),
  json_schema: z.object({}),
  project_id: z.string().optional(),
  default_inference_settings: ZInferenceSettings.optional(),
}));
export type RetabTypesEvaluationsModelEvaluationInput = z.infer<typeof ZRetabTypesEvaluationsModelEvaluationInput>;

export const ZRetabTypesEvaluationsModelEvaluationOutput = z.lazy(() => z.object({
  id: z.string().optional(),
  updated_at: DateOrISO.optional(),
  name: z.string(),
  documents: z.array(ZEvaluationDocumentOutput).optional(),
  iterations: z.array(ZRetabTypesEvaluationsIterationsIterationOutput).optional(),
  json_schema: z.object({}),
  project_id: z.string().optional(),
  default_inference_settings: ZInferenceSettings.optional(),
  schema_data_id: z.string(),
  schema_id: z.string(),
}));
export type RetabTypesEvaluationsModelEvaluationOutput = z.infer<typeof ZRetabTypesEvaluationsModelEvaluationOutput>;

export const ZRetabTypesEvaluationsModelPatchEvaluationRequest = z.lazy(() => z.object({
  name: z.union([z.string(), z.null()]).optional(),
  json_schema: z.union([z.object({}), z.null()]).optional(),
  project_id: z.union([z.string(), z.null()]).optional(),
  default_inference_settings: z.union([ZInferenceSettings, z.null()]).optional(),
}));
export type RetabTypesEvaluationsModelPatchEvaluationRequest = z.infer<typeof ZRetabTypesEvaluationsModelPatchEvaluationRequest>;

export const ZRetabTypesMimeAttachmentMetadata = z.lazy(() => z.object({
  is_inline: z.boolean().optional(),
  inline_cid: z.union([z.string(), z.null()]).optional(),
  source: z.union([z.string(), z.null()]).optional(),
}));
export type RetabTypesMimeAttachmentMetadata = z.infer<typeof ZRetabTypesMimeAttachmentMetadata>;

export const ZRetabTypesMimeBaseMIMEData = z.lazy(() => z.object({
  filename: z.string(),
  url: z.string(),
}));
export type RetabTypesMimeBaseMIMEData = z.infer<typeof ZRetabTypesMimeBaseMIMEData>;

export const ZRetabTypesMimeMIMEData = z.lazy(() => z.object({
  filename: z.string(),
  url: z.string(),
}));
export type RetabTypesMimeMIMEData = z.infer<typeof ZRetabTypesMimeMIMEData>;

export const ZRetabTypesMimeOCROutput = z.lazy(() => z.object({
  pages: z.array(ZRetabTypesMimePageOutput),
}));
export type RetabTypesMimeOCROutput = z.infer<typeof ZRetabTypesMimeOCROutput>;

export const ZRetabTypesMimePageOutput = z.lazy(() => z.object({
  page_number: z.number(),
  width: z.number(),
  height: z.number(),
  unit: z.string().optional(),
  blocks: z.array(ZTextBox),
  lines: z.array(ZTextBox),
  tokens: z.array(ZTextBox),
  transforms: z.array(ZMatrix).optional(),
}));
export type RetabTypesMimePageOutput = z.infer<typeof ZRetabTypesMimePageOutput>;

export const ZRetabTypesPredictionsPredictionDataInput = z.lazy(() => z.object({
  prediction: z.object({}).optional(),
  metadata: z.union([ZPredictionMetadata, z.null()]).optional(),
  updated_at: z.union([DateOrISO, z.null()]).optional(),
}));
export type RetabTypesPredictionsPredictionDataInput = z.infer<typeof ZRetabTypesPredictionsPredictionDataInput>;

export const ZRetabTypesPredictionsPredictionDataOutput = z.lazy(() => z.object({
  prediction: z.object({}).optional(),
  metadata: z.union([ZPredictionMetadata, z.null()]).optional(),
  updated_at: z.union([DateOrISO, z.null()]).optional(),
}));
export type RetabTypesPredictionsPredictionDataOutput = z.infer<typeof ZRetabTypesPredictionsPredictionDataOutput>;

