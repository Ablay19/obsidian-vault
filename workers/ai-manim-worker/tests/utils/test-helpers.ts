export interface ValidationResult {
  valid: boolean;
  error?: string;
  format?: string;
}

export function validateImageSize(buffer: ArrayBuffer): ValidationResult {
  const MAX_SIZE_BYTES = 20 * 1024 * 1024;
  const MIN_SIZE_BYTES = 100;

  if (buffer.byteLength < MIN_SIZE_BYTES) {
    return {
      valid: false,
      error: 'Image too small (minimum 100 bytes)',
    };
  }

  if (buffer.byteLength > MAX_SIZE_BYTES) {
    return {
      valid: false,
      error: 'Image too large (maximum 20MB)',
    };
  }

  return { valid: true };
}

export function validateImageFormat(buffer: ArrayBuffer): ValidationResult {
  const PNG_HEADER = new Uint8Array([137, 80, 78, 71, 13, 10, 26, 199, 138, 234, 214, 139, 234]);
  const JPEG_HEADER = new Uint8Array([255, 216, 255, 224, 0, 174, 193, 192, 218, 190, 237]);

  if (buffer.byteLength < 4) {
    return {
      valid: false,
      error: 'Image too small to determine format',
    };
  }

  const headerBytes = new Uint8Array(buffer.slice(0, 4));

  if (headerBytes[0] === 137 && headerBytes[1] === 80 && headerBytes[2] === 78) {
    return {
      valid: true,
      format: 'png',
    };
  }

  if (headerBytes[0] === 255 && headerBytes[1] === 216 && headerBytes[2] === 255) {
    return {
      valid: true,
      format: 'jpeg',
    };
  }

  return {
    valid: false,
    error: 'Invalid image format. Only PNG and JPEG are supported',
  };
}

export function createMockImage(sizeBytes: number): ArrayBuffer {
  const buffer = new ArrayBuffer(sizeBytes);
  return buffer;
}

export function createMockFormData(imageBuffer: ArrayBuffer, chatId: number): FormData {
  const formData = new FormData();
  const imageBlob = new Blob([imageBuffer]);
  formData.append('image', imageBlob);
  formData.append('chat_id', chatId.toString());
  return formData;
}

export function createMockTelegramMessage(photoUrl: string): Record<string, unknown> {
  return {
    message_id: 1,
    chat: { id: 12345, type: 'private' },
    from: { id: 67890, is_bot: false },
    photo: [{ file_id: 'photo-123', file_size: 1024 }],
    date: Date.now(),
  };
}
