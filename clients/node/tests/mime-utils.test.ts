import fs from 'fs';
import { 
  prepareMimeDocument,
  prepareMimeDocumentList,
  getMimeType,
  validateMimeData
} from '../src/utils/mime.js';

// Mock fs module for testing
jest.mock('fs');
const mockFs = fs as jest.Mocked<typeof fs>;

describe('MIME Utilities Tests', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // Reset fs.existsSync to return false by default (treat strings as text content)
    mockFs.existsSync.mockReturnValue(false);
  });

  describe('getMimeType', () => {
    it('should return correct MIME types for common file extensions', () => {
      const testCases = [
        { file: 'document.pdf', expected: 'application/pdf' },
        { file: 'image.jpg', expected: 'image/jpeg' },
        { file: 'image.jpeg', expected: 'image/jpeg' },
        { file: 'image.png', expected: 'image/png' },
        { file: 'image.gif', expected: 'image/gif' },
        { file: 'image.bmp', expected: 'image/bmp' },
        { file: 'image.tiff', expected: 'image/tiff' },
        { file: 'image.webp', expected: 'image/webp' },
        { file: 'text.txt', expected: 'text/plain' },
        { file: 'data.json', expected: 'application/json' },
        { file: 'page.html', expected: 'text/html' },
        { file: 'page.htm', expected: 'text/html' },
        { file: 'config.xml', expected: 'application/xml' },
        { file: 'data.csv', expected: 'text/csv' },
        { file: 'readme.md', expected: 'text/markdown' }
      ];

      testCases.forEach(({ file, expected }) => {
        const mimeType = getMimeType(file);
        expect(mimeType).toBe(expected);
      });
    });

    it('should return correct MIME types for office documents', () => {
      const officeDocs = [
        { file: 'spreadsheet.xls', expected: 'application/vnd.ms-excel' },
        { file: 'spreadsheet.xlsx', expected: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' },
        { file: 'spreadsheet.ods', expected: 'application/vnd.oasis.opendocument.spreadsheet' },
        { file: 'document.doc', expected: 'application/msword' },
        { file: 'document.docx', expected: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' },
        { file: 'document.odt', expected: 'application/vnd.oasis.opendocument.text' },
        { file: 'presentation.ppt', expected: 'application/vnd.ms-powerpoint' },
        { file: 'presentation.pptx', expected: 'application/vnd.openxmlformats-officedocument.presentationml.presentation' },
        { file: 'presentation.odp', expected: 'application/vnd.oasis.opendocument.presentation' }
      ];

      officeDocs.forEach(({ file, expected }) => {
        const mimeType = getMimeType(file);
        expect(mimeType).toBe(expected);
      });
    });

    it('should return correct MIME types for programming languages', () => {
      const codeFiles = [
        { file: 'script.js', expected: 'text/javascript' },
        { file: 'component.jsx', expected: 'text/javascript' },
        { file: 'module.ts', expected: 'text/typescript' },
        { file: 'component.tsx', expected: 'text/typescript' },
        { file: 'script.py', expected: 'text/x-python' },
        { file: 'Main.java', expected: 'text/x-java-source' },
        { file: 'program.c', expected: 'text/x-c' },
        { file: 'program.cpp', expected: 'text/x-c++' },
        { file: 'Program.cs', expected: 'text/x-csharp' },
        { file: 'script.rb', expected: 'text/x-ruby' },
        { file: 'index.php', expected: 'text/x-php' },
        { file: 'App.swift', expected: 'text/x-swift' },
        { file: 'MainActivity.kt', expected: 'text/x-kotlin' },
        { file: 'main.go', expected: 'text/x-go' },
        { file: 'main.rs', expected: 'text/x-rust' }
      ];

      codeFiles.forEach(({ file, expected }) => {
        const mimeType = getMimeType(file);
        expect(mimeType).toBe(expected);
      });
    });

    it('should return correct MIME types for media files', () => {
      const mediaFiles = [
        { file: 'audio.mp3', expected: 'audio/mpeg' },
        { file: 'video.mp4', expected: 'video/mp4' },
        { file: 'video.mpeg', expected: 'video/mpeg' },
        { file: 'audio.mpga', expected: 'audio/mpeg' },
        { file: 'audio.m4a', expected: 'audio/mp4' },
        { file: 'audio.wav', expected: 'audio/wav' },
        { file: 'video.webm', expected: 'video/webm' }
      ];

      mediaFiles.forEach(({ file, expected }) => {
        const mimeType = getMimeType(file);
        expect(mimeType).toBe(expected);
      });
    });

    it('should handle case insensitive extensions', () => {
      const caseVariations = [
        'image.PNG',
        'document.PDF',
        'script.JS',
        'data.JSON',
        'style.CSS'
      ];

      caseVariations.forEach(file => {
        const mimeType = getMimeType(file);
        expect(typeof mimeType).toBe('string');
        expect(mimeType.length).toBeGreaterThan(0);
      });
    });

    it('should return default MIME type for unknown extensions', () => {
      const unknownFiles = [
        'file.unknown',
        'file.xyz',
        'file.custom',
        'file.',
        'file'
      ];

      unknownFiles.forEach(file => {
        const mimeType = getMimeType(file);
        expect(mimeType).toBe('application/octet-stream');
      });
    });

    it('should handle files with multiple dots', () => {
      const complexFiles = [
        'archive.tar.gz',
        'config.local.json',
        'backup.2024.sql',
        'file.name.with.dots.txt'
      ];

      complexFiles.forEach(file => {
        const mimeType = getMimeType(file);
        expect(typeof mimeType).toBe('string');
        expect(mimeType.length).toBeGreaterThan(0);
      });
    });

    it('should handle empty and special paths', () => {
      const specialPaths = ['', '.', '..', '.hidden', '.gitignore'];

      specialPaths.forEach(filePath => {
        const mimeType = getMimeType(filePath);
        expect(mimeType).toBe('application/octet-stream');
      });
    });
  });

  describe('prepareMimeDocument', () => {
    describe('String inputs (text content)', () => {
      it('should handle plain text content', () => {
        const textContent = 'Hello, world!';
        
        const result = prepareMimeDocument(textContent);
        
        expect(result.extension).toBe('txt');
        expect(result.mime_type).toBe('text/plain');
        expect(result.filename).toBe('content.txt');
        expect(result.unique_filename).toBe('content.txt');
        expect(result.content).toBe(Buffer.from(textContent).toString('base64'));
        expect(result.size).toBe(Buffer.byteLength(textContent));
        expect(result.url).toContain('data:text/plain;base64,');
        expect(result.id).toContain('doc_');
      });

      it('should handle empty string content', () => {
        const result = prepareMimeDocument('');
        
        expect(result.extension).toBe('txt');
        expect(result.mime_type).toBe('text/plain');
        expect(result.content).toBe(Buffer.from('').toString('base64'));
        expect(result.size).toBe(0);
      });

      it('should handle unicode text content', () => {
        const unicodeText = 'Hello 世界! 🌍 éñüíóá';
        
        const result = prepareMimeDocument(unicodeText);
        
        expect(result.mime_type).toBe('text/plain');
        expect(result.content).toBe(Buffer.from(unicodeText).toString('base64'));
        expect(result.size).toBe(Buffer.byteLength(unicodeText));
      });

      it('should handle very long text content', () => {
        const longText = 'A'.repeat(100000);
        
        const result = prepareMimeDocument(longText);
        
        expect(result.mime_type).toBe('text/plain');
        expect(result.size).toBe(100000);
        expect(result.content.length).toBeGreaterThan(0);
      });
    });

    describe('String inputs (file paths)', () => {
      it('should handle existing file paths', () => {
        const filePath = '/path/to/document.pdf';
        const fileContent = Buffer.from('PDF content');
        
        mockFs.existsSync.mockReturnValue(true);
        mockFs.readFileSync.mockReturnValue(fileContent);
        
        const result = prepareMimeDocument(filePath);
        
        expect(mockFs.existsSync).toHaveBeenCalledWith(filePath);
        expect(mockFs.readFileSync).toHaveBeenCalledWith(filePath);
        expect(result.extension).toBe('pdf');
        expect(result.mime_type).toBe('application/pdf');
        expect(result.filename).toBe('document.pdf');
        expect(result.unique_filename).toBe('document.pdf');
        expect(result.content).toBe(fileContent.toString('base64'));
        expect(result.size).toBe(fileContent.length);
        expect(result.url).toContain('data:application/pdf;base64,');
      });

      it('should handle files with different extensions', () => {
        const testFiles = [
          { path: '/path/image.png', mime: 'image/png', ext: 'png' },
          { path: '/path/data.json', mime: 'application/json', ext: 'json' },
          { path: '/path/script.js', mime: 'text/javascript', ext: 'js' },
          { path: '/path/style.css', mime: 'application/octet-stream', ext: 'css' }
        ];

        testFiles.forEach(({ path: filePath, mime, ext }) => {
          const fileContent = Buffer.from('test content');
          
          mockFs.existsSync.mockReturnValue(true);
          mockFs.readFileSync.mockReturnValue(fileContent);
          
          const result = prepareMimeDocument(filePath);
          
          expect(result.extension).toBe(ext);
          expect(result.mime_type).toBe(mime);
        });
      });

      it('should handle files without extensions', () => {
        const filePath = '/path/to/README';
        const fileContent = Buffer.from('README content');
        
        mockFs.existsSync.mockReturnValue(true);
        mockFs.readFileSync.mockReturnValue(fileContent);
        
        const result = prepareMimeDocument(filePath);
        
        expect(result.extension).toBe('unknown');
        expect(result.filename).toBe('README');
        expect(result.unique_filename).toBe('README');
      });

      it('should handle non-existent files as text content', () => {
        const filePath = '/non/existent/file.txt';
        
        mockFs.existsSync.mockReturnValue(false);
        
        const result = prepareMimeDocument(filePath);
        
        expect(result.mime_type).toBe('text/plain');
        expect(result.extension).toBe('txt');
        expect(result.content).toBe(Buffer.from(filePath).toString('base64'));
      });

      it('should handle relative file paths', () => {
        const filePath = './document.pdf';
        const fileContent = Buffer.from('PDF content');
        
        mockFs.existsSync.mockReturnValue(true);
        mockFs.readFileSync.mockReturnValue(fileContent);
        
        const result = prepareMimeDocument(filePath);
        
        expect(result.filename).toBe('document.pdf');
        expect(result.mime_type).toBe('application/pdf');
      });
    });

    describe('Buffer inputs', () => {
      it('should handle buffer input', () => {
        const bufferContent = Buffer.from('Binary data content');
        
        const result = prepareMimeDocument(bufferContent);
        
        expect(result.extension).toBe('bin');
        expect(result.mime_type).toBe('application/octet-stream');
        expect(result.filename).toBe('content.bin');
        expect(result.unique_filename).toBe('content.bin');
        expect(result.content).toBe(bufferContent.toString('base64'));
        expect(result.size).toBe(bufferContent.length);
        expect(result.url).toContain('data:application/octet-stream;base64,');
        expect(result.id).toContain('doc_');
      });

      it('should handle empty buffer', () => {
        const emptyBuffer = Buffer.alloc(0);
        
        const result = prepareMimeDocument(emptyBuffer);
        
        expect(result.mime_type).toBe('application/octet-stream');
        expect(result.size).toBe(0);
        expect(result.content).toBe('');
      });

      it('should handle large buffer', () => {
        const largeBuffer = Buffer.alloc(1024 * 1024); // 1MB
        largeBuffer.fill('A');
        
        const result = prepareMimeDocument(largeBuffer);
        
        expect(result.size).toBe(1024 * 1024);
        expect(result.mime_type).toBe('application/octet-stream');
        expect(result.content.length).toBeGreaterThan(0);
      });

      it('should handle binary buffer with special bytes', () => {
        const binaryBuffer = Buffer.from([0x00, 0xFF, 0x7F, 0x80, 0x01]);
        
        const result = prepareMimeDocument(binaryBuffer);
        
        expect(result.mime_type).toBe('application/octet-stream');
        expect(result.size).toBe(5);
        expect(result.content).toBe(binaryBuffer.toString('base64'));
      });
    });

    describe('MIMEData inputs', () => {
      it('should return MIMEData as-is', () => {
        const mimeData = {
          id: 'existing_doc_123',
          extension: 'pdf',
          content: 'base64content',
          mime_type: 'application/pdf',
          unique_filename: 'existing.pdf',
          size: 1024,
          filename: 'existing.pdf',
          url: 'data:application/pdf;base64,base64content'
        };
        
        const result = prepareMimeDocument(mimeData);
        
        expect(result).toEqual(mimeData);
      });

      it('should handle custom MIMEData with extra properties', () => {
        const customMimeData = {
          id: 'custom_doc_456',
          extension: 'json',
          content: 'eyJrZXkiOiJ2YWx1ZSJ9',
          mime_type: 'application/json',
          unique_filename: 'data.json',
          size: 512,
          filename: 'data.json',
          url: 'data:application/json;base64,eyJrZXkiOiJ2YWx1ZSJ9',
          custom_property: 'custom_value'
        };
        
        const result = prepareMimeDocument(customMimeData as any);
        
        expect(result).toEqual(customMimeData);
      });
    });

    describe('ID generation and timestamps', () => {
      it('should generate unique IDs for different calls', () => {
        const result1 = prepareMimeDocument('test1');
        // Add small delay to ensure different timestamps
        jest.useFakeTimers();
        jest.advanceTimersByTime(1);
        const result2 = prepareMimeDocument('test2');
        jest.useRealTimers();
        
        expect(result1.id).not.toBe(result2.id);
        expect(result1.id).toContain('doc_');
        expect(result2.id).toContain('doc_');
      });

      it('should generate IDs with timestamps', () => {
        const beforeTime = Date.now();
        const result = prepareMimeDocument('test');
        const afterTime = Date.now();
        
        const idTimestamp = parseInt(result.id.replace('doc_', ''));
        expect(idTimestamp).toBeGreaterThanOrEqual(beforeTime);
        expect(idTimestamp).toBeLessThanOrEqual(afterTime);
      });
    });
  });

  describe('prepareMimeDocumentList', () => {
    it('should handle empty array', () => {
      const result = prepareMimeDocumentList([]);
      
      expect(result).toEqual([]);
    });

    it('should handle mixed input types', () => {
      const textContent = 'Text content';
      const bufferContent = Buffer.from('Buffer content');
      const mimeData = {
        id: 'existing_doc',
        extension: 'pdf',
        content: 'base64content',
        mime_type: 'application/pdf',
        unique_filename: 'existing.pdf',
        size: 1024,
        filename: 'existing.pdf',
        url: 'data:application/pdf;base64,base64content'
      };
      
      const inputs = [textContent, bufferContent, mimeData];
      const results = prepareMimeDocumentList(inputs);
      
      expect(results).toHaveLength(3);
      expect(results[0].mime_type).toBe('text/plain');
      expect(results[1].mime_type).toBe('application/octet-stream');
      expect(results[2]).toEqual(mimeData);
    });

    it('should handle large arrays', () => {
      const largeArray = Array(100).fill('test content');
      
      const results = prepareMimeDocumentList(largeArray);
      
      expect(results).toHaveLength(100);
      results.forEach(result => {
        expect(result.mime_type).toBe('text/plain');
        expect(result.id).toContain('doc_');
      });
    });

    it('should preserve order of inputs', () => {
      const inputs = ['first', 'second', 'third'];
      
      const results = prepareMimeDocumentList(inputs);
      
      expect(results[0].content).toBe(Buffer.from('first').toString('base64'));
      expect(results[1].content).toBe(Buffer.from('second').toString('base64'));
      expect(results[2].content).toBe(Buffer.from('third').toString('base64'));
    });

    it('should handle file paths in array', () => {
      const filePath = '/path/to/test.pdf';
      const fileContent = Buffer.from('PDF content');
      
      // Mock existsSync to return true only for the specific file path
      mockFs.existsSync.mockImplementation((path) => path === filePath);
      mockFs.readFileSync.mockReturnValue(fileContent);
      
      const results = prepareMimeDocumentList([filePath, 'text content']);
      
      expect(results).toHaveLength(2);
      expect(results[0].mime_type).toBe('application/pdf');
      expect(results[1].mime_type).toBe('text/plain');
    });
  });

  describe('validateMimeData', () => {
    it('should validate correct MIME data', () => {
      const validMimeData = {
        id: 'doc_123',
        extension: 'pdf',
        content: 'base64content',
        mime_type: 'application/pdf',
        unique_filename: 'document.pdf',
        size: 1024,
        filename: 'document.pdf',
        url: 'data:application/pdf;base64,base64content'
      };
      
      const result = validateMimeData(validMimeData);
      
      expect(result).toEqual(validMimeData);
    });

    it('should throw error for invalid MIME data', () => {
      const invalidData = {
        id: 'doc_123',
        // missing required fields
        extension: 'pdf'
      };
      
      expect(() => {
        validateMimeData(invalidData);
      }).toThrow();
    });

    it('should throw error for wrong data types', () => {
      const invalidTypeData = {
        id: 123, // should be string
        extension: 'pdf',
        content: 'base64content',
        mime_type: 'application/pdf',
        unique_filename: 'document.pdf',
        size: '1024', // should be number
        filename: 'document.pdf',
        url: 'data:application/pdf;base64,base64content'
      };
      
      expect(() => {
        validateMimeData(invalidTypeData);
      }).toThrow();
    });

    it('should handle null and undefined gracefully', () => {
      expect(() => {
        validateMimeData(null);
      }).toThrow();
      
      expect(() => {
        validateMimeData(undefined);
      }).toThrow();
    });

    it('should handle empty object', () => {
      expect(() => {
        validateMimeData({});
      }).toThrow();
    });

    it('should validate with additional properties', () => {
      const dataWithExtra = {
        id: 'doc_123',
        extension: 'pdf',
        content: 'base64content',
        mime_type: 'application/pdf',
        unique_filename: 'document.pdf',
        size: 1024,
        filename: 'document.pdf',
        url: 'data:application/pdf;base64,base64content',
        extra_property: 'extra_value'
      };
      
      const result = validateMimeData(dataWithExtra);
      
      // Zod should strip unknown properties by default
      expect(result).not.toHaveProperty('extra_property');
      expect(result.id).toBe('doc_123');
    });
  });

  describe('Integration Tests', () => {
    it('should work together for complete document processing', () => {
      const textDoc = 'Hello world';
      const bufferDoc = Buffer.from('Binary data');
      
      // Process individual documents
      const textResult = prepareMimeDocument(textDoc);
      const bufferResult = prepareMimeDocument(bufferDoc);
      
      // Process as list
      const listResults = prepareMimeDocumentList([textDoc, bufferDoc]);
      
      // Validate results
      const validatedText = validateMimeData(textResult);
      const validatedBuffer = validateMimeData(bufferResult);
      
      expect(validatedText.mime_type).toBe('text/plain');
      expect(validatedBuffer.mime_type).toBe('application/octet-stream');
      expect(listResults).toHaveLength(2);
      
      // Check MIME type consistency
      expect(getMimeType('test.txt')).toBe(validatedText.mime_type);
    });

    it('should handle complex workflow with file operations', () => {
      const filePath = '/complex/path/document.pdf';
      const fileContent = Buffer.from('Complex PDF content');
      
      mockFs.existsSync.mockReturnValue(true);
      mockFs.readFileSync.mockReturnValue(fileContent);
      
      // Get MIME type first
      const mimeType = getMimeType(filePath);
      expect(mimeType).toBe('application/pdf');
      
      // Process document
      const result = prepareMimeDocument(filePath);
      expect(result.mime_type).toBe(mimeType);
      
      // Validate result
      const validated = validateMimeData(result);
      expect(validated.mime_type).toBe('application/pdf');
      
      // Process as part of list
      const listResult = prepareMimeDocumentList([filePath]);
      expect(listResult[0].mime_type).toBe('application/pdf');
    });
  });

  describe('Error Handling and Edge Cases', () => {
    it('should handle file system errors gracefully', () => {
      const filePath = '/path/that/causes/error.pdf';
      
      mockFs.existsSync.mockReturnValue(true);
      mockFs.readFileSync.mockImplementation(() => {
        throw new Error('File system error');
      });
      
      expect(() => {
        prepareMimeDocument(filePath);
      }).toThrow('File system error');
    });

    it('should handle very large file paths', () => {
      const longPath = '/path/' + 'very-long-filename-'.repeat(100) + 'document.pdf';
      
      mockFs.existsSync.mockReturnValue(false);
      
      const result = prepareMimeDocument(longPath);
      
      expect(result.mime_type).toBe('text/plain'); // Treated as text content
      expect(result.size).toBe(Buffer.byteLength(longPath));
    });

    it('should handle special characters in file names', () => {
      const specialPath = '/path/file with spaces & symbols!@#$%.pdf';
      const fileContent = Buffer.from('Content');
      
      mockFs.existsSync.mockReturnValue(true);
      mockFs.readFileSync.mockReturnValue(fileContent);
      
      const result = prepareMimeDocument(specialPath);
      
      expect(result.filename).toBe('file with spaces & symbols!@#$%.pdf');
      expect(result.mime_type).toBe('application/pdf');
    });

    it('should handle files with unusual extensions', () => {
      const unusualFiles = [
        'file.unknown123',
        'file.CAPS',
        'file.mix3d',
        'file.123',
        'file.-ext'
      ];

      unusualFiles.forEach(filename => {
        const mimeType = getMimeType(filename);
        expect(mimeType).toBe('application/octet-stream');
      });
    });

    it('should handle concurrent ID generation', async () => {
      const promises = Array(10).fill(null).map((_, index) => 
        new Promise(resolve => {
          setTimeout(() => {
            resolve(prepareMimeDocument(`test${index}`));
          }, index); // Stagger the calls slightly
        })
      );
      
      const results = await Promise.all(promises);
      const ids = results.map((r: any) => r.id);
      const uniqueIds = new Set(ids);
      
      expect(uniqueIds.size).toBeGreaterThan(1); // Should have some unique IDs
      expect(ids.length).toBe(10);
    });
  });
});