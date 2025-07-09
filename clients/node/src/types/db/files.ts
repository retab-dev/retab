/**
 * Database types for file management and storage
 */

export interface DBFile {
  id: string;
  organization_id: string;
  filename: string;
  original_filename: string;
  mime_type: string;
  size_bytes: number;
  md5_hash: string;
  sha256_hash: string;
  storage_path: string;
  storage_provider: 'gcs' | 's3' | 'local';
  upload_status: 'uploading' | 'completed' | 'failed' | 'processing';
  created_at: Date;
  updated_at: Date;
  uploaded_by?: string;
  metadata?: Record<string, any>;
  tags?: string[];
  is_public: boolean;
  expires_at?: Date;
  download_count: number;
  last_accessed_at?: Date;
  virus_scan_status?: 'pending' | 'clean' | 'infected' | 'failed';
  virus_scan_result?: string;
  content_extracted: boolean;
  content_text?: string;
  content_preview?: string;
  ocr_status?: 'pending' | 'completed' | 'failed' | 'not_needed';
  ocr_confidence?: number;
  document_type?: 'pdf' | 'image' | 'text' | 'spreadsheet' | 'presentation' | 'other';
  page_count?: number;
  image_dimensions?: {
    width: number;
    height: number;
  };
  compression_applied: boolean;
  thumbnail_path?: string;
}

export interface DBFileCreate {
  organization_id: string;
  filename: string;
  original_filename: string;
  mime_type: string;
  size_bytes: number;
  md5_hash: string;
  sha256_hash: string;
  storage_path: string;
  storage_provider: 'gcs' | 's3' | 'local';
  uploaded_by?: string;
  metadata?: Record<string, any>;
  tags?: string[];
  is_public?: boolean;
  expires_at?: Date;
}

export interface DBFileUpdate {
  filename?: string;
  metadata?: Record<string, any>;
  tags?: string[];
  is_public?: boolean;
  expires_at?: Date;
  upload_status?: 'uploading' | 'completed' | 'failed' | 'processing';
  content_text?: string;
  content_preview?: string;
  ocr_status?: 'pending' | 'completed' | 'failed' | 'not_needed';
  ocr_confidence?: number;
  document_type?: 'pdf' | 'image' | 'text' | 'spreadsheet' | 'presentation' | 'other';
  page_count?: number;
  image_dimensions?: {
    width: number;
    height: number;
  };
  virus_scan_status?: 'pending' | 'clean' | 'infected' | 'failed';
  virus_scan_result?: string;
  content_extracted?: boolean;
  compression_applied?: boolean;
  thumbnail_path?: string;
}

export interface DBFileQuery {
  organization_id?: string;
  filename?: string;
  mime_type?: string;
  min_size?: number;
  max_size?: number;
  uploaded_by?: string;
  tags?: string[];
  is_public?: boolean;
  upload_status?: 'uploading' | 'completed' | 'failed' | 'processing';
  created_after?: Date;
  created_before?: Date;
  expires_after?: Date;
  expires_before?: Date;
  document_type?: 'pdf' | 'image' | 'text' | 'spreadsheet' | 'presentation' | 'other';
  has_content_text?: boolean;
  virus_scan_status?: 'pending' | 'clean' | 'infected' | 'failed';
  limit?: number;
  offset?: number;
  order_by?: 'created_at' | 'updated_at' | 'filename' | 'size_bytes' | 'download_count';
  order_direction?: 'asc' | 'desc';
}

export interface FileStorageConfig {
  provider: 'gcs' | 's3' | 'local';
  bucket_name?: string;
  region?: string;
  access_key?: string;
  secret_key?: string;
  endpoint?: string;
  base_path?: string;
  max_file_size_bytes: number;
  allowed_mime_types: string[];
  virus_scanning_enabled: boolean;
  thumbnail_generation_enabled: boolean;
  ocr_enabled: boolean;
  content_extraction_enabled: boolean;
  compression_enabled: boolean;
  encryption_enabled: boolean;
}

export interface FileProcessingJob {
  id: string;
  file_id: string;
  job_type: 'ocr' | 'thumbnail' | 'virus_scan' | 'content_extract' | 'compress';
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  created_at: Date;
  started_at?: Date;
  completed_at?: Date;
  error_message?: string;
  progress_percentage: number;
  result_data?: Record<string, any>;
  retry_count: number;
  max_retries: number;
}

export default {
  DBFile,
  DBFileCreate,
  DBFileUpdate,
  DBFileQuery,
  FileStorageConfig,
  FileProcessingJob,
};