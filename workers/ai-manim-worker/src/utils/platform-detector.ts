import type { Platform } from '../types';

export function detectPlatform(userAgent: string, headers: Record<string, string>): Platform {
  // Check for Telegram-specific headers
  if (headers['x-telegram-bot-api-secret-token']) {
    return 'telegram';
  }

  // Check for WhatsApp-specific headers
  if (headers['x-hub-signature-256']) {
    return 'whatsapp';
  }

  // Check user agent for web platform
  if (userAgent && (
    userAgent.includes('Mozilla') ||
    userAgent.includes('Chrome') ||
    userAgent.includes('Safari') ||
    userAgent.includes('Firefox') ||
    userAgent.includes('Edge')
  )) {
    return 'web';
  }

  // Default to web for unknown platforms
  return 'web';
}

export function getPlatformDisplayName(platform: Platform): string {
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

export function getPlatformIcon(platform: Platform): string {
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