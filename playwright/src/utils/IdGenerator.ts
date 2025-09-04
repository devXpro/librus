import * as crypto from 'crypto';

/**
 * Utility class for generating consistent IDs for different types of content
 * Uses MD5 hashing to ensure deterministic IDs
 */

export class IdGenerator {
  /**
   * Generates MD5 hash from input string
   */
  private static generateMD5Hash(input: string): string {
    return crypto.createHash('md5').update(input).digest('hex');
  }

  /**
   * Generates ID for a message based on its URL
   * Used for MESSAGE_TYPE_MESSAGE
   */
  static generateMessageId(messageUrl: string): string {
    return IdGenerator.generateMD5Hash(messageUrl);
  }

  /**
   * Generates ID for news/announcement based on title, content, and date
   * Used for MESSAGE_TYPE_NOTIFICATION
   */
  static generateNewsId(title: string, content: string, dateStr: string): string {
    const combinedString = title + content + dateStr;
    return IdGenerator.generateMD5Hash(combinedString);
  }



  /**
   * Generates unique ID for attachment directory
   * Uses UUID v4 for uniqueness
   */
  static generateAttachmentDirId(): string {
    // Using crypto.randomUUID() which is available in Node.js 14.17.0+
    return crypto.randomUUID();
  }
}
