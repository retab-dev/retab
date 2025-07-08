import {
  convertToOpenAIFormat,
  convertFromOpenAIFormat,
  separateMessages,
  stringifyMessages,
  ChatMessage,
  ChatContentPart,
  ParsedChatCompletion
} from '../src/utils/responses.js';

describe('Responses Utilities Tests', () => {
  describe('convertToOpenAIFormat', () => {
    it('should convert simple text messages correctly', () => {
      const messages: ChatMessage[] = [
        { role: 'system', content: 'You are a helpful assistant.' },
        { role: 'user', content: 'Hello, how are you?' },
        { role: 'assistant', content: 'I am doing well, thank you!' }
      ];

      const result = convertToOpenAIFormat(messages);

      expect(result).toEqual([
        { role: 'system', content: 'You are a helpful assistant.' },
        { role: 'user', content: 'Hello, how are you?' },
        { role: 'assistant', content: 'I am doing well, thank you!' }
      ]);
    });

    it('should convert messages with different roles', () => {
      const messages: ChatMessage[] = [
        { role: 'developer', content: 'System prompt for development.' },
        { role: 'user', content: 'User question.' },
        { role: 'assistant', content: 'Assistant response.' }
      ];

      const result = convertToOpenAIFormat(messages);

      expect(result).toHaveLength(3);
      expect(result[0].role).toBe('developer');
      expect(result[1].role).toBe('user');
      expect(result[2].role).toBe('assistant');
    });

    it('should handle complex content with parts', () => {
      const contentParts: ChatContentPart[] = [
        { type: 'text', text: 'Here is an image:' },
        { 
          type: 'image_url', 
          image_url: { 
            url: 'data:image/jpeg;base64,abc123',
            detail: 'high'
          }
        }
      ];

      const messages: ChatMessage[] = [
        { role: 'user', content: contentParts }
      ];

      const result = convertToOpenAIFormat(messages);

      expect(result).toEqual([
        { role: 'user', content: contentParts }
      ]);
    });

    it('should handle empty messages array', () => {
      const result = convertToOpenAIFormat([]);
      expect(result).toEqual([]);
    });

    it('should handle messages with audio content', () => {
      const audioContent: ChatContentPart[] = [
        { type: 'text', text: 'Here is audio:' },
        { 
          type: 'input_audio',
          input_audio: {
            data: 'base64-audio-data',
            format: 'wav'
          }
        }
      ];

      const messages: ChatMessage[] = [
        { role: 'user', content: audioContent }
      ];

      const result = convertToOpenAIFormat(messages);

      expect(result[0].content).toEqual(audioContent);
    });

    it('should preserve all message properties', () => {
      const messages: ChatMessage[] = [
        { role: 'system', content: 'System message' },
        { role: 'user', content: '' }, // Empty content
        { role: 'assistant', content: 'Response with\nnewlines\nand\ttabs' }
      ];

      const result = convertToOpenAIFormat(messages);

      expect(result).toHaveLength(3);
      expect(result[1].content).toBe('');
      expect(result[2].content).toContain('\n');
      expect(result[2].content).toContain('\t');
    });
  });

  describe('convertFromOpenAIFormat', () => {
    it('should convert simple OpenAI format messages correctly', () => {
      const openAIMessages = [
        { role: 'system', content: 'You are a helpful assistant.' },
        { role: 'user', content: 'Hello, how are you?' },
        { role: 'assistant', content: 'I am doing well, thank you!' }
      ];

      const result = convertFromOpenAIFormat(openAIMessages);

      expect(result).toEqual([
        { role: 'system', content: 'You are a helpful assistant.' },
        { role: 'user', content: 'Hello, how are you?' },
        { role: 'assistant', content: 'I am doing well, thank you!' }
      ]);
    });

    it('should handle messages with additional properties', () => {
      const openAIMessages = [
        { 
          role: 'user', 
          content: 'Hello',
          name: 'John',
          function_call: { name: 'test', arguments: '{}' }
        },
        { 
          role: 'assistant', 
          content: 'Hi there!',
          function_call: null
        }
      ];

      const result = convertFromOpenAIFormat(openAIMessages);

      expect(result).toHaveLength(2);
      expect(result[0].role).toBe('user');
      expect(result[0].content).toBe('Hello');
      expect(result[1].role).toBe('assistant');
      expect(result[1].content).toBe('Hi there!');
    });

    it('should handle empty OpenAI messages array', () => {
      const result = convertFromOpenAIFormat([]);
      expect(result).toEqual([]);
    });

    it('should preserve complex content structures', () => {
      const complexContent = [
        { type: 'text', text: 'Complex message' },
        { type: 'image_url', image_url: { url: 'http://example.com/image.jpg' } }
      ];

      const openAIMessages = [
        { role: 'user', content: complexContent }
      ];

      const result = convertFromOpenAIFormat(openAIMessages);

      expect(result[0].content).toEqual(complexContent);
    });

    it('should handle null and undefined content gracefully', () => {
      const openAIMessages = [
        { role: 'user', content: null },
        { role: 'assistant', content: undefined },
        { role: 'system' } // No content property
      ];

      const result = convertFromOpenAIFormat(openAIMessages);

      expect(result).toHaveLength(3);
      expect(result[0].content).toBeNull();
      expect(result[1].content).toBeUndefined();
      expect(result[2].content).toBeUndefined();
    });
  });

  describe('separateMessages', () => {
    it('should separate messages by role correctly', () => {
      const messages: ChatMessage[] = [
        { role: 'system', content: 'System prompt' },
        { role: 'user', content: 'First user message' },
        { role: 'assistant', content: 'First assistant response' },
        { role: 'user', content: 'Second user message' },
        { role: 'assistant', content: 'Second assistant response' }
      ];

      const result = separateMessages(messages);

      expect(result.systemMessage).toEqual({ role: 'system', content: 'System prompt' });
      expect(result.userMessages).toHaveLength(2);
      expect(result.assistantMessages).toHaveLength(2);
      expect(result.userMessages[0].content).toBe('First user message');
      expect(result.userMessages[1].content).toBe('Second user message');
      expect(result.assistantMessages[0].content).toBe('First assistant response');
      expect(result.assistantMessages[1].content).toBe('Second assistant response');
    });

    it('should handle developer role as system message', () => {
      const messages: ChatMessage[] = [
        { role: 'developer', content: 'Developer instructions' },
        { role: 'user', content: 'User message' }
      ];

      const result = separateMessages(messages);

      expect(result.systemMessage).toEqual({ role: 'developer', content: 'Developer instructions' });
      expect(result.userMessages).toHaveLength(1);
      expect(result.assistantMessages).toHaveLength(0);
    });

    it('should handle messages with only user messages', () => {
      const messages: ChatMessage[] = [
        { role: 'user', content: 'First message' },
        { role: 'user', content: 'Second message' }
      ];

      const result = separateMessages(messages);

      expect(result.systemMessage).toBeUndefined();
      expect(result.userMessages).toHaveLength(2);
      expect(result.assistantMessages).toHaveLength(0);
    });

    it('should handle messages with only assistant messages', () => {
      const messages: ChatMessage[] = [
        { role: 'assistant', content: 'First response' },
        { role: 'assistant', content: 'Second response' }
      ];

      const result = separateMessages(messages);

      expect(result.systemMessage).toBeUndefined();
      expect(result.userMessages).toHaveLength(0);
      expect(result.assistantMessages).toHaveLength(2);
    });

    it('should handle empty messages array', () => {
      const result = separateMessages([]);

      expect(result.systemMessage).toBeUndefined();
      expect(result.userMessages).toHaveLength(0);
      expect(result.assistantMessages).toHaveLength(0);
    });

    it('should handle multiple system/developer messages (last one wins)', () => {
      const messages: ChatMessage[] = [
        { role: 'system', content: 'First system message' },
        { role: 'developer', content: 'Developer message' },
        { role: 'system', content: 'Second system message' },
        { role: 'user', content: 'User message' }
      ];

      const result = separateMessages(messages);

      // The last system/developer message should be kept
      expect(result.systemMessage).toEqual({ role: 'system', content: 'Second system message' });
      expect(result.userMessages).toHaveLength(1);
    });

    it('should handle complex content in separated messages', () => {
      const complexContent: ChatContentPart[] = [
        { type: 'text', text: 'Text part' },
        { type: 'image_url', image_url: { url: 'http://example.com/image.jpg', detail: 'high' } }
      ];

      const messages: ChatMessage[] = [
        { role: 'user', content: complexContent },
        { role: 'assistant', content: 'Simple response' }
      ];

      const result = separateMessages(messages);

      expect(result.userMessages[0].content).toEqual(complexContent);
      expect(result.assistantMessages[0].content).toBe('Simple response');
    });
  });

  describe('stringifyMessages', () => {
    it('should stringify simple text messages', () => {
      const messages: ChatMessage[] = [
        { role: 'user', content: 'Hello' },
        { role: 'assistant', content: 'Hi there!' }
      ];

      const result = stringifyMessages(messages);
      const parsed = JSON.parse(result);

      expect(Array.isArray(parsed)).toBe(true);
      expect(parsed).toHaveLength(2);
      expect(parsed[0]).toEqual({ role: 'user', content: 'Hello' });
      expect(parsed[1]).toEqual({ role: 'assistant', content: 'Hi there!' });
    });

    it('should truncate long text content to default length', () => {
      const longText = 'a'.repeat(200);
      const messages: ChatMessage[] = [
        { role: 'user', content: longText }
      ];

      const result = stringifyMessages(messages);
      const parsed = JSON.parse(result);

      expect(parsed[0].content).toHaveLength(103); // 100 chars + '...'
      expect(parsed[0].content.endsWith('...')).toBe(true);
      expect(parsed[0].content.startsWith('aaaa')).toBe(true);
    });

    it('should respect custom maxLength parameter', () => {
      const longText = 'hello world this is a long message';
      const messages: ChatMessage[] = [
        { role: 'user', content: longText }
      ];

      const result = stringifyMessages(messages, 10);
      const parsed = JSON.parse(result);

      expect(parsed[0].content).toHaveLength(13); // 10 chars + '...'
      expect(parsed[0].content).toBe('hello worl...');
    });

    it('should not truncate short content', () => {
      const shortText = 'Hi';
      const messages: ChatMessage[] = [
        { role: 'user', content: shortText }
      ];

      const result = stringifyMessages(messages, 100);
      const parsed = JSON.parse(result);

      expect(parsed[0].content).toBe('Hi');
    });

    it('should handle complex content with text parts', () => {
      const complexContent: ChatContentPart[] = [
        { type: 'text', text: 'This is a very long text that should be truncated because it exceeds the limit' },
        { type: 'image_url', image_url: { url: 'http://example.com/image.jpg', detail: 'high' } }
      ];

      const messages: ChatMessage[] = [
        { role: 'user', content: complexContent }
      ];

      const result = stringifyMessages(messages, 20);
      const parsed = JSON.parse(result);

      expect(parsed[0].content[0].text).toHaveLength(23); // 20 chars + '...'
      expect(parsed[0].content[0].text.endsWith('...')).toBe(true);
      expect(parsed[0].content[1].type).toBe('image_url');
    });

    it('should handle complex content with image URLs', () => {
      const veryLongUrl = 'http://example.com/' + 'a'.repeat(200) + '/image.jpg';
      const complexContent: ChatContentPart[] = [
        { type: 'text', text: 'Short text' },
        { type: 'image_url', image_url: { url: veryLongUrl, detail: 'low' } }
      ];

      const messages: ChatMessage[] = [
        { role: 'user', content: complexContent }
      ];

      const result = stringifyMessages(messages, 30);
      const parsed = JSON.parse(result);

      expect(parsed[0].content[0].text).toBe('Short text');
      expect(parsed[0].content[1].image_url.url).toHaveLength(33); // 30 chars + '...'
      expect(parsed[0].content[1].image_url.url.endsWith('...')).toBe(true);
      expect(parsed[0].content[1].image_url.detail).toBe('low');
    });

    it('should handle audio content parts', () => {
      const complexContent: ChatContentPart[] = [
        { type: 'text', text: 'Audio message' },
        { 
          type: 'input_audio',
          input_audio: {
            data: 'base64-audio-data-that-is-very-long',
            format: 'wav'
          }
        }
      ];

      const messages: ChatMessage[] = [
        { role: 'user', content: complexContent }
      ];

      const result = stringifyMessages(messages);
      const parsed = JSON.parse(result);

      expect(parsed[0].content[0].text).toBe('Audio message');
      expect(parsed[0].content[1].type).toBe('input_audio');
      expect(parsed[0].content[1].input_audio).toBeDefined();
    });

    it('should handle empty messages array', () => {
      const result = stringifyMessages([]);
      const parsed = JSON.parse(result);

      expect(Array.isArray(parsed)).toBe(true);
      expect(parsed).toHaveLength(0);
    });

    it('should handle messages with empty content', () => {
      const messages: ChatMessage[] = [
        { role: 'user', content: '' },
        { role: 'assistant', content: [] }
      ];

      const result = stringifyMessages(messages);
      const parsed = JSON.parse(result);

      expect(parsed[0].content).toBe('');
      expect(parsed[1].content).toEqual([]);
    });

    it('should produce valid JSON with proper formatting', () => {
      const messages: ChatMessage[] = [
        { role: 'user', content: 'Test message' }
      ];

      const result = stringifyMessages(messages);

      expect(() => JSON.parse(result)).not.toThrow();
      expect(result).toContain('{\n');
      expect(result).toContain('  ');
    });

    it('should handle edge case with zero maxLength', () => {
      const messages: ChatMessage[] = [
        { role: 'user', content: 'Hello' }
      ];

      const result = stringifyMessages(messages, 0);
      const parsed = JSON.parse(result);

      expect(parsed[0].content).toBe('...');
    });

    it('should handle special characters in content', () => {
      const messages: ChatMessage[] = [
        { role: 'user', content: 'Special chars: "quotes" \'apostrophes\' \n newlines \t tabs' }
      ];

      const result = stringifyMessages(messages);
      const parsed = JSON.parse(result);

      expect(parsed[0].content).toContain('"quotes"');
      expect(parsed[0].content).toContain("'apostrophes'");
      expect(parsed[0].content).toContain('\n');
      expect(parsed[0].content).toContain('\t');
    });
  });

  describe('Type Definitions', () => {
    it('should define ChatMessage interface correctly', () => {
      const textMessage: ChatMessage = {
        role: 'user',
        content: 'Simple text message'
      };

      const complexMessage: ChatMessage = {
        role: 'user',
        content: [
          { type: 'text', text: 'Text part' },
          { type: 'image_url', image_url: { url: 'http://example.com/image.jpg' } }
        ]
      };

      expect(textMessage.role).toBe('user');
      expect(typeof textMessage.content).toBe('string');
      expect(complexMessage.role).toBe('user');
      expect(Array.isArray(complexMessage.content)).toBe(true);
    });

    it('should define ChatContentPart interface correctly', () => {
      const textPart: ChatContentPart = {
        type: 'text',
        text: 'Text content'
      };

      const imagePart: ChatContentPart = {
        type: 'image_url',
        image_url: {
          url: 'http://example.com/image.jpg',
          detail: 'high'
        }
      };

      const audioPart: ChatContentPart = {
        type: 'input_audio',
        input_audio: {
          data: 'base64-data',
          format: 'wav'
        }
      };

      expect(textPart.type).toBe('text');
      expect(textPart.text).toBe('Text content');
      expect(imagePart.type).toBe('image_url');
      expect(imagePart.image_url?.url).toBe('http://example.com/image.jpg');
      expect(audioPart.type).toBe('input_audio');
      expect(audioPart.input_audio?.format).toBe('wav');
    });

    it('should define ParsedChatCompletion interface correctly', () => {
      const completion: ParsedChatCompletion = {
        id: 'chatcmpl-123',
        object: 'chat.completion',
        created: Date.now(),
        model: 'gpt-4o',
        choices: [
          {
            index: 0,
            message: {
              role: 'assistant',
              content: 'Response content'
            },
            finish_reason: 'stop'
          }
        ],
        usage: {
          prompt_tokens: 10,
          completion_tokens: 5,
          total_tokens: 15
        },
        likelihoods: {
          token_log_probs: []
        }
      };

      expect(completion.object).toBe('chat.completion');
      expect(completion.choices).toHaveLength(1);
      expect(completion.choices[0].message.role).toBe('assistant');
      expect(completion.usage?.total_tokens).toBe(15);
    });
  });

  describe('Integration Tests', () => {
    it('should work together for complete message processing workflow', () => {
      const originalMessages: ChatMessage[] = [
        { role: 'system', content: 'You are helpful' },
        { role: 'user', content: 'Hello' },
        { role: 'assistant', content: 'Hi there!' }
      ];

      // Convert to OpenAI format
      const openAIFormat = convertToOpenAIFormat(originalMessages);
      
      // Convert back from OpenAI format
      const convertedBack = convertFromOpenAIFormat(openAIFormat);
      
      // Separate messages
      const separated = separateMessages(convertedBack);
      
      // Stringify messages
      const stringified = stringifyMessages(convertedBack);

      expect(convertedBack).toEqual(originalMessages);
      expect(separated.systemMessage?.content).toBe('You are helpful');
      expect(separated.userMessages).toHaveLength(1);
      expect(separated.assistantMessages).toHaveLength(1);
      expect(() => JSON.parse(stringified)).not.toThrow();
    });

    it('should handle complex workflow with multimodal content', () => {
      const complexMessages: ChatMessage[] = [
        {
          role: 'user',
          content: [
            { type: 'text', text: 'Analyze this image' },
            { type: 'image_url', image_url: { url: 'data:image/jpeg;base64,abc123', detail: 'high' } }
          ]
        },
        {
          role: 'assistant',
          content: 'I can see the image you provided.'
        }
      ];

      const openAIFormat = convertToOpenAIFormat(complexMessages);
      const convertedBack = convertFromOpenAIFormat(openAIFormat);
      const separated = separateMessages(convertedBack);
      const stringified = stringifyMessages(convertedBack, 50);

      expect(convertedBack).toEqual(complexMessages);
      expect(separated.userMessages[0].content).toEqual(complexMessages[0].content);
      
      const parsed = JSON.parse(stringified);
      expect(parsed[0].content[0].text).toBe('Analyze this image');
      expect(parsed[0].content[1].image_url.url).toContain('data:image/jpeg');
    });
  });

  describe('Error Handling and Edge Cases', () => {
    it('should handle malformed message objects gracefully', () => {
      const malformedMessages = [
        { role: 'user' }, // No content
        { content: 'No role' }, // No role
        {}, // Empty object
        { role: 'invalid', content: 'Invalid role' }
      ];

      expect(() => convertToOpenAIFormat(malformedMessages as any)).not.toThrow();
      expect(() => convertFromOpenAIFormat(malformedMessages)).not.toThrow();
      expect(() => separateMessages(malformedMessages as any)).not.toThrow();
      expect(() => stringifyMessages(malformedMessages as any)).not.toThrow();
    });

    it('should handle null and undefined inputs', () => {
      expect(() => convertToOpenAIFormat(null as any)).toThrow();
      expect(() => convertToOpenAIFormat(undefined as any)).toThrow();
      expect(() => convertFromOpenAIFormat(null as any)).toThrow();
      expect(() => convertFromOpenAIFormat(undefined as any)).toThrow();
      expect(() => separateMessages(null as any)).toThrow();
      expect(() => separateMessages(undefined as any)).toThrow();
      expect(() => stringifyMessages(null as any)).toThrow();
      expect(() => stringifyMessages(undefined as any)).toThrow();
    });

    it('should handle very large message arrays', () => {
      const largeArray = Array(1000).fill(null).map((_, i) => ({
        role: i % 2 === 0 ? 'user' : 'assistant',
        content: `Message ${i}`
      }));

      expect(() => convertToOpenAIFormat(largeArray as any)).not.toThrow();
      expect(() => convertFromOpenAIFormat(largeArray)).not.toThrow();
      expect(() => separateMessages(largeArray as any)).not.toThrow();
      expect(() => stringifyMessages(largeArray as any, 10)).not.toThrow();
    });

    it('should handle messages with circular references in content', () => {
      const circularContent: any = { type: 'text', text: 'Circular' };
      circularContent.self = circularContent;

      const messages: ChatMessage[] = [
        { role: 'user', content: circularContent }
      ];

      // These should handle the circular reference gracefully
      expect(() => convertToOpenAIFormat(messages)).not.toThrow();
      expect(() => separateMessages(messages)).not.toThrow();
      
      // Stringify might throw due to circular reference in JSON.stringify
      expect(() => stringifyMessages(messages)).toThrow();
    });
  });
});