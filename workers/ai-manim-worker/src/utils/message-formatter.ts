import type { Platform } from '../types';

export interface MessageFormatOptions {
  includePlatform?: boolean;
  includeTimestamp?: boolean;
  maxLength?: number;
}

export class MessageFormatter {
  static formatConfirmation(platform: Platform, jobId: string, content: string, options: MessageFormatOptions = {}): string {
    const platformIcon = this.getPlatformIcon(platform);
    const preview = content.length > 50 ? content.substring(0, 50) + '...' : content;

    let message = `${platformIcon} Processing your request...\n\n`;
    message += `Job ID: ${jobId}\n`;
    message += `Content: "${preview}"\n\n`;
    message += `This may take 2-5 minutes.`;

    if (options.includePlatform) {
      message += `\n\nPlatform: ${this.getPlatformDisplayName(platform)}`;
    }

    if (options.includeTimestamp) {
      message += `\n\nTime: ${new Date().toLocaleString()}`;
    }

    return this.truncateMessage(message, options.maxLength);
  }

  static formatCodeConfirmation(platform: Platform, jobId: string, options: MessageFormatOptions = {}): string {
    const platformIcon = this.getPlatformIcon(platform);

    let message = `${platformIcon} Processing your Manim code...\n\n`;
    message += `Job ID: ${jobId}\n\n`;
    message += `This may take 2-5 minutes.`;

    if (options.includePlatform) {
      message += `\n\nPlatform: ${this.getPlatformDisplayName(platform)}`;
    }

    if (options.includeTimestamp) {
      message += `\n\nTime: ${new Date().toLocaleString()}`;
    }

    return this.truncateMessage(message, options.maxLength);
  }

  static formatVideoReady(platform: Platform, jobId: string, videoUrl: string, options: MessageFormatOptions = {}): string {
    const platformIcon = this.getPlatformIcon(platform);

    let message = `${platformIcon} ✅ Your video is ready!\n\n`;
    message += `🎬 ${videoUrl}\n\n`;
    message += `Job ID: ${jobId}`;

    if (options.includePlatform) {
      message += `\nPlatform: ${this.getPlatformDisplayName(platform)}`;
    }

    if (options.includeTimestamp) {
      message += `\nCompleted: ${new Date().toLocaleString()}`;
    }

    return this.truncateMessage(message, options.maxLength);
  }

  static formatError(platform: Platform, error: string, options: MessageFormatOptions = {}): string {
    const platformIcon = this.getPlatformIcon(platform);

    let message = `${platformIcon} ❌ Error: ${error}`;

    if (options.includePlatform) {
      message += `\n\nPlatform: ${this.getPlatformDisplayName(platform)}`;
    }

    if (options.includeTimestamp) {
      message += `\nTime: ${new Date().toLocaleString()}`;
    }

    return this.truncateMessage(message, options.maxLength);
  }

  static formatTimeout(platform: Platform, jobId: string, options: MessageFormatOptions = {}): string {
    const platformIcon = this.getPlatformIcon(platform);

    let message = `${platformIcon} ⏰ Video generation is taking longer than expected.\n\n`;
    message += `Job ID: ${jobId}\n\n`;
    message += `You can check the status later on the web dashboard.`;

    if (options.includePlatform) {
      message += `\n\nPlatform: ${this.getPlatformDisplayName(platform)}`;
    }

    return this.truncateMessage(message, options.maxLength);
  }

  private static getPlatformIcon(platform: Platform): string {
    switch (platform) {
      case 'telegram':
        return '📱';
      case 'whatsapp':
        return '💬';
      case 'web':
        return '🌐';
      default:
        return '❓';
    }
  }

  private static getPlatformDisplayName(platform: Platform): string {
    switch (platform) {
      case 'telegram':
        return 'Telegram';
      case 'whatsapp':
        return 'WhatsApp';
      case 'web':
        return 'Web Dashboard';
      default:
        return 'Unknown';
    }
  }

  private static truncateMessage(message: string, maxLength?: number): string {
    if (!maxLength || message.length <= maxLength) {
      return message;
    }

    return message.substring(0, maxLength - 3) + '...';
  }
}

// Convenience functions for common formatting
export function formatConfirmation(platform: Platform, jobId: string, content: string): string {
  return MessageFormatter.formatConfirmation(platform, jobId, content);
}

export function formatVideoReady(platform: Platform, jobId: string, videoUrl: string): string {
  return MessageFormatter.formatVideoReady(platform, jobId, videoUrl);
}

export function formatError(platform: Platform, error: string): string {
  return MessageFormatter.formatError(platform, error);
}