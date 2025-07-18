import fs from 'fs';
import { Readable } from 'stream';
import {
    fileTypeFromBuffer,
    fileTypeFromFile,
    FileTypeResult
} from 'file-type';
import { MIMEData } from '@/types';

function streamToBuffer(stream: Readable): Promise<Buffer> {
    return new Promise((resolve, reject) => {
        const chunks: Buffer[] = [];
        stream.on('data', (chunk) => chunks.push(chunk));
        stream.on('end', () => resolve(Buffer.concat(chunks)));
        stream.on('error', reject);
    });
}

export function mimeToBlob(mime: MIMEData): Blob {
    const splitDataURI = mime.url.split(',')
    const byteString = splitDataURI[0].indexOf('base64') >= 0 ? atob(splitDataURI[1]) : decodeURI(splitDataURI[1])
    const mimeString = splitDataURI[0].split(':')[1].split(';')[0]

    const ia = new Uint8Array(byteString.length)
    for (let i = 0; i < byteString.length; i++)
        ia[i] = byteString.charCodeAt(i)

    return new Blob([ia], { type: mimeString })
  }

export async function inferFileInfo(input: Buffer | string | Readable): Promise<MIMEData> {
    let buffer: Buffer;
    let filePath: string | null = null;
    let mime: string | null = null;

    if (Buffer.isBuffer(input)) {
        buffer = input;
    } else if (typeof input === 'string') {
        if (await fs.promises.stat(input).then(stat => stat.isFile()).catch(() => false)) {
            filePath = input;
            buffer = await fs.promises.readFile(filePath);
        } else if (input.startsWith('data:')) {
            mime = input.split(/[,;]/)[0].split(':')[1];
            const base64 = input.split(',')[1];
            buffer = Buffer.from(base64, 'base64');
        } else {
            buffer = Buffer.from(input, 'base64');
        }
    } else if (input instanceof Readable) {
        buffer = await streamToBuffer(input);
    } else {
        throw new Error('Unsupported input type');
    }

    let fileType: FileTypeResult | undefined;

    if (filePath) {
        fileType = await fileTypeFromFile(filePath);
    } else {
        fileType = await fileTypeFromBuffer(buffer);
    }

    if (!fileType) {
        throw new Error('Unable to determine file type');
    }

    const { ext, mime: inferredMime } = fileType;
    if (mime === null) mime = inferredMime;
    const base64Data = buffer.toString('base64');

    return {
        filename: `file.${ext}`,
        url: `data:${mime};base64,${base64Data}`,
    };
}

