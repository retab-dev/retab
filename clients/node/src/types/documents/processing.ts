/**
 * Specialized document processing types
 * Equivalent to Python's types/documents/ modules
 */

// Document orientation correction
export interface OrientationCorrection {
  original_orientation: number; // degrees
  corrected_orientation: number; // degrees
  rotation_applied: number; // degrees to rotate
  confidence_score: number; // 0-1
  detection_method: 'text_analysis' | 'image_analysis' | 'metadata' | 'manual';
  processing_time_ms: number;
}

export interface OrientationCorrectionRequest {
  image_data: string; // base64 encoded
  detection_methods?: Array<'text_analysis' | 'image_analysis' | 'metadata'>;
  auto_correct?: boolean;
  return_corrected_image?: boolean;
  confidence_threshold?: number; // 0-1, minimum confidence to apply correction
}

export interface OrientationCorrectionResponse {
  correction: OrientationCorrection;
  corrected_image_data?: string; // base64 encoded if requested
  original_dimensions: {
    width: number;
    height: number;
  };
  corrected_dimensions: {
    width: number;
    height: number;
  };
}

// Message creation utilities
export interface DocumentMessage {
  role: 'system' | 'user' | 'assistant';
  content: string | Array<{
    type: 'text' | 'image_url';
    text?: string;
    image_url?: {
      url: string;
      detail?: 'low' | 'high' | 'auto';
    };
  }>;
  name?: string;
  function_call?: {
    name: string;
    arguments: string;
  };
  tool_calls?: Array<{
    id: string;
    type: 'function';
    function: {
      name: string;
      arguments: string;
    };
  }>;
}

export interface MessageCreationConfig {
  include_system_prompt: boolean;
  system_prompt_template?: string;
  user_prompt_template?: string;
  image_detail_level: 'low' | 'high' | 'auto';
  max_image_size?: {
    width: number;
    height: number;
  };
  image_format: 'jpeg' | 'png' | 'webp';
  image_quality?: number; // 1-100 for JPEG
  include_metadata: boolean;
  metadata_fields?: string[];
}

export interface CreateMessagesRequest {
  documents: Array<{
    content: string | Buffer;
    mime_type: string;
    filename?: string;
    metadata?: Record<string, any>;
  }>;
  schema?: Record<string, any>;
  instructions?: string;
  config: MessageCreationConfig;
}

export interface CreateMessagesResponse {
  messages: DocumentMessage[];
  token_count_estimate: {
    text_tokens: number;
    image_tokens: number;
    total_tokens: number;
  };
  processing_stats: {
    documents_processed: number;
    images_processed: number;
    text_extracted_chars: number;
    processing_time_ms: number;
  };
}

// Document parsing types
export interface DocumentParseConfig {
  extract_text: boolean;
  extract_images: boolean;
  extract_tables: boolean;
  extract_metadata: boolean;
  ocr_enabled: boolean;
  ocr_language?: string;
  image_dpi?: number;
  preserve_layout: boolean;
  merge_text_blocks: boolean;
  detect_columns: boolean;
  extract_headers_footers: boolean;
  page_range?: {
    start: number;
    end: number;
  };
}

export interface ParsedDocument {
  document_id: string;
  filename: string;
  mime_type: string;
  total_pages: number;
  processing_time_ms: number;
  metadata: {
    title?: string;
    author?: string;
    subject?: string;
    creator?: string;
    producer?: string;
    creation_date?: Date;
    modification_date?: Date;
    keywords?: string[];
    language?: string;
    page_size?: {
      width: number;
      height: number;
    };
  };
  content: {
    full_text: string;
    pages: Array<ParsedPage>;
    images: Array<ExtractedImage>;
    tables: Array<ExtractedTable>;
  };
  quality_metrics: {
    text_confidence: number; // 0-1
    ocr_confidence?: number; // 0-1
    layout_confidence: number; // 0-1
    extraction_completeness: number; // 0-1
  };
}

export interface ParsedPage {
  page_number: number;
  text: string;
  layout: {
    bounding_box: BoundingBox;
    text_blocks: Array<TextBlock>;
    columns?: Array<ColumnLayout>;
  };
  images: string[]; // image IDs on this page
  tables: string[]; // table IDs on this page
  headers: string[];
  footers: string[];
  annotations?: Array<PageAnnotation>;
}

export interface TextBlock {
  id: string;
  text: string;
  bounding_box: BoundingBox;
  font_info?: {
    family: string;
    size: number;
    weight: 'normal' | 'bold';
    style: 'normal' | 'italic';
  };
  confidence: number; // 0-1
  block_type: 'paragraph' | 'heading' | 'list_item' | 'caption' | 'footer' | 'header';
  reading_order: number;
}

export interface ColumnLayout {
  column_number: number;
  bounding_box: BoundingBox;
  text_blocks: string[]; // text block IDs in this column
}

export interface ExtractedImage {
  id: string;
  page_number: number;
  bounding_box: BoundingBox;
  image_data: string; // base64 encoded
  format: 'jpeg' | 'png' | 'tiff' | 'gif';
  dimensions: {
    width: number;
    height: number;
  };
  dpi?: number;
  description?: string; // from OCR or image analysis
  alt_text?: string;
}

export interface ExtractedTable {
  id: string;
  page_number: number;
  bounding_box: BoundingBox;
  rows: Array<TableRow>;
  column_headers?: string[];
  caption?: string;
  confidence: number; // 0-1
  table_type: 'data' | 'layout' | 'form';
}

export interface TableRow {
  row_number: number;
  cells: Array<TableCell>;
  is_header: boolean;
}

export interface TableCell {
  column_number: number;
  text: string;
  bounding_box: BoundingBox;
  colspan: number;
  rowspan: number;
  cell_type: 'data' | 'header' | 'empty';
  confidence: number; // 0-1
}

export interface BoundingBox {
  x: number;
  y: number;
  width: number;
  height: number;
}

export interface PageAnnotation {
  id: string;
  type: 'highlight' | 'note' | 'redaction' | 'stamp';
  bounding_box: BoundingBox;
  content?: string;
  author?: string;
  created_at?: Date;
  color?: string;
}

// Document quality assessment
export interface DocumentQualityAssessment {
  overall_score: number; // 0-1
  readability_score: number; // 0-1
  image_quality_score: number; // 0-1
  text_extraction_quality: number; // 0-1
  layout_preservation_score: number; // 0-1
  issues: Array<QualityIssue>;
  recommendations: string[];
  processing_suitable: boolean;
}

export interface QualityIssue {
  type: 'low_resolution' | 'skewed_text' | 'poor_contrast' | 'corrupted_content' | 'unsupported_format';
  severity: 'low' | 'medium' | 'high' | 'critical';
  description: string;
  affected_pages?: number[];
  suggested_fix?: string;
}

// No default export for type-only modules