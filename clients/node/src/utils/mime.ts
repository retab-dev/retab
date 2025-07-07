import fs from 'fs';
import path from 'path';
import { MIMEData, MIMEDataSchema } from '../types/mime.js';

export function prepareMimeDocument(document: string | Buffer | MIMEData): MIMEData {
  if (typeof document === 'string') {
    // If it's a file path, read the file
    if (fs.existsSync(document)) {
      const filename = path.basename(document);
      const content = fs.readFileSync(document);
      const base64Content = content.toString('base64');
      const mimeType = getMimeType(document);
      return {
        id: 'doc_' + Date.now(),
        extension: path.extname(filename).slice(1) || 'unknown',
        content: base64Content,
        mime_type: mimeType,
        unique_filename: filename,
        size: content.length,
        filename,
        url: `data:${mimeType};base64,${base64Content}`,
      };
    }
    // Otherwise treat as text content
    const base64Content = Buffer.from(document).toString('base64');
    return {
      id: 'doc_' + Date.now(),
      extension: 'txt',
      content: base64Content,
      mime_type: 'text/plain',
      unique_filename: 'content.txt',
      size: Buffer.byteLength(document),
      filename: 'content.txt',
      url: `data:text/plain;base64,${base64Content}`,
    };
  } else if (Buffer.isBuffer(document)) {
    const base64Content = document.toString('base64');
    return {
      id: 'doc_' + Date.now(),
      extension: 'bin',
      content: base64Content,
      mime_type: 'application/octet-stream',
      unique_filename: 'content.bin',
      size: document.length,
      filename: 'content.bin',
      url: `data:application/octet-stream;base64,${base64Content}`,
    };
  }
  return document as MIMEData;
}

export function prepareMimeDocumentList(documents: Array<string | Buffer | MIMEData>): MIMEData[] {
  return documents.map(doc => prepareMimeDocument(doc));
}

export function getMimeType(filePath: string): string {
  const ext = path.extname(filePath).toLowerCase();
  const mimeTypes: Record<string, string> = {
    '.pdf': 'application/pdf',
    '.jpg': 'image/jpeg',
    '.jpeg': 'image/jpeg',
    '.png': 'image/png',
    '.gif': 'image/gif',
    '.bmp': 'image/bmp',
    '.tiff': 'image/tiff',
    '.webp': 'image/webp',
    '.txt': 'text/plain',
    '.json': 'application/json',
    '.html': 'text/html',
    '.htm': 'text/html',
    '.xml': 'application/xml',
    '.csv': 'text/csv',
    '.tsv': 'text/tab-separated-values',
    '.md': 'text/markdown',
    '.log': 'text/plain',
    '.yaml': 'application/x-yaml',
    '.yml': 'application/x-yaml',
    '.rtf': 'application/rtf',
    '.ini': 'text/plain',
    '.conf': 'text/plain',
    '.cfg': 'text/plain',
    '.nfo': 'text/plain',
    '.srt': 'text/plain',
    '.sql': 'text/plain',
    '.sh': 'text/x-shellscript',
    '.bat': 'text/plain',
    '.ps1': 'text/plain',
    '.js': 'text/javascript',
    '.jsx': 'text/javascript',
    '.ts': 'text/typescript',
    '.tsx': 'text/typescript',
    '.py': 'text/x-python',
    '.java': 'text/x-java-source',
    '.c': 'text/x-c',
    '.cpp': 'text/x-c++',
    '.cs': 'text/x-csharp',
    '.rb': 'text/x-ruby',
    '.php': 'text/x-php',
    '.swift': 'text/x-swift',
    '.kt': 'text/x-kotlin',
    '.go': 'text/x-go',
    '.rs': 'text/x-rust',
    '.pl': 'text/x-perl',
    '.r': 'text/x-r',
    '.m': 'text/x-objc',
    '.scala': 'text/x-scala',
    '.xls': 'application/vnd.ms-excel',
    '.xlsx': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    '.ods': 'application/vnd.oasis.opendocument.spreadsheet',
    '.doc': 'application/msword',
    '.docx': 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    '.odt': 'application/vnd.oasis.opendocument.text',
    '.ppt': 'application/vnd.ms-powerpoint',
    '.pptx': 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
    '.odp': 'application/vnd.oasis.opendocument.presentation',
    '.mhtml': 'message/rfc822',
    '.eml': 'message/rfc822',
    '.msg': 'application/vnd.ms-outlook',
    '.mp3': 'audio/mpeg',
    '.mp4': 'video/mp4',
    '.mpeg': 'video/mpeg',
    '.mpga': 'audio/mpeg',
    '.m4a': 'audio/mp4',
    '.wav': 'audio/wav',
    '.webm': 'video/webm',
  };
  return mimeTypes[ext] || 'application/octet-stream';
}

export function validateMimeData(data: any): MIMEData {
  return MIMEDataSchema.parse(data);
}