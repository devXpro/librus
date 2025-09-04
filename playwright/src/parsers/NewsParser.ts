import { Page } from 'playwright';
import { Message, MessageType } from '../generated/librus_scraper';
import { logger } from '../utils/logger';
import { LibrusUrls } from '../utils/LibrusUrls';
import { DateParser } from '../utils/DateParser';
import { IdGenerator } from '../utils/IdGenerator';

/**
 * Parser class responsible for extracting news/announcements data from Librus pages
 * Handles parsing of news content from the announcements page
 */

export class NewsParser {
  /**
   * Parses all news items from the news/announcements page
   * Returns array of Message objects with type MESSAGE_TYPE_NOTIFICATION
   */
  static async parseNewsList(page: Page): Promise<Message[]> {
    logger.debug('Parsing news list from page');

    try {
      const news: Message[] = [];

      // Find all tables on the news page - each table represents one news item
      const newsTables = await page.locator(LibrusUrls.NEWS_TABLE_SELECTOR).all();

      for (const table of newsTables) {
        try {
          const newsItem = await NewsParser.parseSingleNewsTable(table);
          if (newsItem && NewsParser.validateNewsItem(newsItem)) {
            news.push(newsItem);
            logger.trace('Parsed news item', { 
              id: newsItem.id, 
              title: newsItem.title, 
              author: newsItem.author 
            });
          }
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Unknown error';
          logger.warn('Failed to parse news table', { error: errorMessage });
          // Continue with other news items
        }
      }

      logger.debug('Parsed news list', { count: news.length });
      return news;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Failed to parse news list', { error: errorMessage });
      throw new Error(`News list parsing failed: ${errorMessage}`);
    }
  }

  /**
   * Parses a single news table and extracts news item data
   * Returns Message object or null if parsing fails
   */
  private static async parseSingleNewsTable(table: any): Promise<Message | null> {
    try {
      // Get the title from thead td[colspan='2']
      const title = await table.locator(LibrusUrls.NEWS_TITLE_SELECTOR).textContent();
      if (!title?.trim()) {
        return null; // Skip tables without titles
      }

      // Get author from th:contains('Dodał') + next cell
      const authorCell = table.locator(LibrusUrls.NEWS_AUTHOR_SELECTOR);
      const author = await authorCell.textContent() || '';

      // Get date from th:contains('Data publikacji') + next cell
      const dateCell = table.locator(LibrusUrls.NEWS_DATE_SELECTOR);
      const dateStr = await dateCell.textContent() || '';

      // Get content from th:contains('Treść') + next cell
      const contentCell = table.locator(LibrusUrls.NEWS_CONTENT_SELECTOR);
      const content = await contentCell.textContent() || '';

      // Generate ID based on title + content + date (like in Go version)
      const newsId = IdGenerator.generateNewsId(
        title.trim(), 
        content.trim(), 
        dateStr.trim()
      );

      const newsItem: Message = {
        id: newsId,
        type: MessageType.MESSAGE_TYPE_NOTIFICATION,
        link: LibrusUrls.NEWS_URL,
        author: author.trim(),
        title: title.trim(),
        content: content.trim(),
        dateTimestamp: DateParser.parseAnyDate(dateStr.trim()),
        attachmentsDir: '' // News usually don't have attachments
      };

      return newsItem;

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.warn('Failed to parse single news table', { error: errorMessage });
      return null;
    }
  }

  /**
   * Validates if a news item has required fields
   */
  static validateNewsItem(newsItem: Message): boolean {
    const hasRequiredFields = !!(
      newsItem.id &&
      newsItem.title &&
      newsItem.type === MessageType.MESSAGE_TYPE_NOTIFICATION
    );

    if (!hasRequiredFields) {
      logger.warn('News item validation failed', { 
        id: newsItem.id,
        title: newsItem.title,
        type: newsItem.type
      });
    }

    return hasRequiredFields;
  }
}
