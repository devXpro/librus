import { Page, Locator } from 'playwright';
import { Message, MessageType } from '../generated/librus_scraper';
import { logger } from '../utils/logger';
import { LibrusUrls } from '../utils/LibrusUrls';
import { DateParser } from '../utils/DateParser';
import { IdGenerator } from '../utils/IdGenerator';

/**
 * Parser class responsible for extracting message data from Librus pages
 * Handles both message list parsing and single message detail parsing
 */

export class MessageParser {
  /**
   * Extracts message links from the messages list page
   * Returns array of relative URLs to individual messages
   */
  static async extractMessageLinks(page: Page): Promise<string[]> {
    logger.debug('Extracting message links from page');

    try {
      // Find unread messages (bold text in table) - same selector as Go version
      const messageRows = await page.locator(LibrusUrls.UNREAD_MESSAGE_LINKS_SELECTOR).all();

      // Temporary: Find ALL messages (both read and unread) for testing
      // const messageRows = await page.locator(LibrusUrls.MESSAGE_LINKS_SELECTOR).all();

      const links: string[] = [];
      for (const messageLink of messageRows) {
        try {
          const href = await messageLink.getAttribute('href');
          if (href) {
            links.push(href);
          }
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Unknown error';
          logger.warn('Failed to extract link from message row', { error: errorMessage });
        }
      }

      // Remove duplicates
      const uniqueLinks = [...new Set(links)];
      logger.debug('Extracted message links', { 
        total: links.length, 
        unique: uniqueLinks.length 
      });

      return uniqueLinks;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Failed to extract message links', { error: errorMessage });
      throw new Error(`Message link extraction failed: ${errorMessage}`);
    }
  }

  /**
   * Parses detailed message content from a single message page
   * Returns complete Message object with content and metadata
   */
  static async parseSingleMessage(page: Page, messageUrl: string): Promise<Message | null> {
    logger.debug('Parsing single message content', { messageUrl });

    try {
      // Extract message details from the specific table (index 6)
      const messageTable = page.locator('table').nth(LibrusUrls.MESSAGE_DETAILS_TABLE_INDEX);

      // Extract author
      const authorCell = messageTable.locator(LibrusUrls.MESSAGE_AUTHOR_SELECTOR).nth(1);
      const author = await authorCell.textContent() || '';

      // Extract title
      const titleCell = messageTable.locator(LibrusUrls.MESSAGE_TITLE_SELECTOR).nth(1);
      const title = await titleCell.textContent() || '';

      // Extract date
      const dateCell = messageTable.locator(LibrusUrls.MESSAGE_DATE_SELECTOR).nth(1);
      const dateString = await dateCell.textContent() || '';

      // Extract content from the element that comes after the message table
      const contentElement = messageTable.locator('+ *');
      const content = await contentElement.textContent() || '';

      const message: Message = {
        id: IdGenerator.generateMessageId(messageUrl),
        type: MessageType.MESSAGE_TYPE_MESSAGE,
        link: messageUrl,
        author: author.trim(),
        title: title.trim(),
        content: content.trim(),
        dateTimestamp: DateParser.parseLibrusDateTime(dateString.trim()),
        attachmentsDir: '' // Will be filled by AttachmentDownloader
      };

      logger.debug('Parsed single message', { 
        id: message.id, 
        title: message.title,
        author: message.author
      });

      return message;

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Failed to parse single message', { messageUrl, error: errorMessage });
      return null;
    }
  }
}
