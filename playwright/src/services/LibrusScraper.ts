import { BrowserContext, Page } from 'playwright';
import { SessionManager } from './SessionManager';
import {
  Message,
  MessageType
} from '../generated/librus_scraper';
import { logger } from '../utils/logger';

// Simple interface for credentials (not from protobuf)
interface LoginCredentials {
  login: string;
  password: string;
}

export class LibrusScraper {
  private sessionManager: SessionManager;

  constructor(sessionManager: SessionManager) {
    this.sessionManager = sessionManager;
  }

  async getMessages(
    credentials: LoginCredentials
  ): Promise<Message[]> {
    logger.info('Getting messages', { login: credentials.login });

    const context = await this.sessionManager.getEphemeralSession(credentials);

    try {
      const page = await context.newPage();
      try {
        await page.goto('https://synergia.librus.pl/wiadomosci');
        logger.debug('Navigated to messages page');

        // Wait for page to fully load before scraping
        await page.waitForLoadState('networkidle');
        await page.waitForSelector('table.decorated', { timeout: 10000 });

        // Get message links (like in Go version)
        const messageLinks = await this.getMessageLinks(page);
        logger.info('Found message links', { login: credentials.login, count: messageLinks.length });

        // Process each message individually (like in Go version)
        const messages: Message[] = [];
        for (const link of messageLinks) {
          try {
            const fullUrl = new URL(link, 'https://synergia.librus.pl').toString();

            // Skip javascript links
            if (fullUrl.includes('javascript')) {
              continue;
            }

            logger.debug('Processing message', { link: fullUrl });

            // Navigate to message page and scrape it
            await page.goto(fullUrl);
            const message = await this.scrapeSingleMessage(page, fullUrl);

            if (message) {
              messages.push(message);
              logger.trace('Successfully processed message', { id: message.id, title: message.title });
            }

          } catch (error) {
            const errorMessage = error instanceof Error ? error.message : 'Unknown error';
            logger.warn('Failed to process message', { link, error: errorMessage });
            // Continue with other messages (like in Go version)
          }
        }

        logger.info('Successfully processed all messages', { login: credentials.login, count: messages.length });
        return messages;

      } finally {
        await page.close();
      }
    } finally {
      // Save session state and close context to free memory
      await this.sessionManager.saveAndCloseSession(context, credentials.login);
    }
  }

  async getNews(
    credentials: LoginCredentials
  ): Promise<Message[]> {
    logger.info('Getting news', { login: credentials.login });

    const context = await this.sessionManager.getEphemeralSession(credentials);

    try {
      const page = await context.newPage();
      try {
        await page.goto('https://synergia.librus.pl/ogloszenia');
        logger.debug('Navigated to news page');

        // Wait for page to fully load before scraping
        await page.waitForLoadState('networkidle');
        await page.waitForSelector('table', { timeout: 10000 });

        const news = await this.scrapeNews(page);
        logger.info('Successfully scraped news', { login: credentials.login, count: news.length });

        return news;

      } finally {
        await page.close();
      }
    } finally {
      // Save session state and close context to free memory
      await this.sessionManager.saveAndCloseSession(context, credentials.login);
    }
  }

  async getAllUpdates(
    credentials: LoginCredentials
  ): Promise<{ messages: Message[], news: Message[] }> {
    logger.info('Getting all updates', { login: credentials.login });

    const context = await this.sessionManager.getEphemeralSession(credentials);

    try {
      // Get both messages and news using the same context for efficiency
      const messagesPromise = this.scrapeMessagesWithContext(context);
      const newsPromise = this.scrapeNewsWithContext(context);

      const [messages, news] = await Promise.all([messagesPromise, newsPromise]);

      logger.info('Successfully got all updates', {
        login: credentials.login,
        messagesCount: messages.length,
        newsCount: news.length
      });

      return { messages, news };
    } finally {
      // Save session state and close context to free memory
      await this.sessionManager.saveAndCloseSession(context, credentials.login);
    }
  }

  private async scrapeMessagesWithContext(context: BrowserContext): Promise<Message[]> {
    const page = await context.newPage();
    try {
      await page.goto('https://synergia.librus.pl/wiadomosci');
      await page.waitForLoadState('networkidle');
      await page.waitForSelector('table.decorated', { timeout: 10000 });
      return await this.scrapeMessages(page);
    } finally {
      await page.close();
    }
  }

  private async scrapeNewsWithContext(context: BrowserContext): Promise<Message[]> {
    const page = await context.newPage();
    try {
      await page.goto('https://synergia.librus.pl/ogloszenia');
      await page.waitForLoadState('networkidle');
      await page.waitForSelector('table', { timeout: 10000 });
      return await this.scrapeNews(page);
    } finally {
      await page.close();
    }
  }

  async getSingleMessage(
    credentials: LoginCredentials,
    messageUrl: string
  ): Promise<Message | null> {
    logger.info('Getting single message', { login: credentials.login, messageUrl });

    const context = await this.sessionManager.getEphemeralSession(credentials);

    try {
      const page = await context.newPage();
      try {
        await page.goto(messageUrl);
        logger.debug('Navigated to message page');

        const message = await this.scrapeSingleMessage(page, messageUrl);
        logger.info('Successfully scraped single message', { login: credentials.login, messageId: message?.id });

        return message;

      } finally {
        await page.close();
      }

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Failed to get single message', { login: credentials.login, messageUrl, error: errorMessage });
      return null;
    } finally {
      // Save session state and close context to free memory
      await this.sessionManager.saveAndCloseSession(context, credentials.login);
    }
  }

  async answerMessage(
    credentials: LoginCredentials,
    messageUrl: string,
    answerText: string
  ): Promise<void> {
    logger.info('Answering message', { login: credentials.login, messageUrl });

    const context = await this.sessionManager.getEphemeralSession(credentials);

    try {
      const page = await context.newPage();
      try {
        await page.goto(messageUrl);
        logger.debug('Navigated to message page');

        // Click "Odpowiedz" button
        await page.getByRole('button', { name: 'Odpowiedz' }).click();
        logger.debug('Clicked reply button');

        // Fill the answer text
        const textArea = page.locator('#tresc_wiadomosci');
        await textArea.fill(answerText + '\n\n');
        logger.debug('Filled answer text');

        // Click "Wyślij" button
        await page.getByRole('button', { name: 'Wyślij' }).click();
        logger.debug('Clicked send button');

        // Wait for success (page navigation or success message)
        await page.waitForTimeout(2000);

        logger.info('Successfully answered message', { login: credentials.login, messageUrl });

      } finally {
        await page.close();
      }
    } finally {
      // Save session state and close context to free memory
      await this.sessionManager.saveAndCloseSession(context, credentials.login);
    }
  }

  private async getMessageLinks(page: Page): Promise<string[]> {
    logger.debug('Getting message links from page');

    // Find unread messages (bold text in table) - same selector as Go version
    // const messageRows = await page.locator('table.decorated td[style*="font-weight: bold"] a').all();

    // Temporary: Find ALL messages (both read and unread) for testing
    const messageRows = await page.locator('table.decorated td a').all();

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

    // Remove duplicates (like in Go version)
    const uniqueLinks = [...new Set(links)];
    logger.debug('Extracted message links', { total: links.length, unique: uniqueLinks.length });

    return uniqueLinks;
  }

  private async scrapeMessages(page: Page): Promise<Message[]> {
    logger.debug('Scraping messages from page');

    const messages: Message[] = [];

    // Find unread messages (bold text in table)
    // const messageRows = await page.locator('table.decorated td[style*="font-weight: bold"] a').all();

    // Temporary: Find ALL messages (both read and unread) for testing
    const messageRows = await page.locator('table.decorated td a').all();
    
    for (const messageLink of messageRows) {
      try {
        const href = await messageLink.getAttribute('href');
        const title = await messageLink.textContent();
        
        if (!href || !title) continue;
        
        const fullUrl = new URL(href, 'https://synergia.librus.pl').toString();
        
        // Get additional info from the row
        const row = messageLink.locator('xpath=ancestor::tr');
        const cells = await row.locator('td').all();
        
        let author = '';
        let dateStr = '';
        
        if (cells.length >= 3) {
          author = (await cells[1]?.textContent())?.trim() || '';
          dateStr = (await cells[2]?.textContent())?.trim() || '';
        }
        
        const message: Message = {
          id: this.extractMessageId(fullUrl),
          type: MessageType.MESSAGE_TYPE_MESSAGE,
          link: fullUrl,
          author,
          title: title.trim(),
          content: '', // Will be filled when getting single message
          dateTimestamp: this.parseDate(dateStr),
          attachmentsDir: '' // Will be filled when downloading attachments
        };
        
        messages.push(message);
        logger.trace('Scraped message', { id: message.id, title: message.title });
        
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error';
        logger.warn('Failed to scrape message row', { error: errorMessage });
      }
    }
    
    return messages;
  }

  private async scrapeNews(page: Page): Promise<Message[]> {
    logger.debug('Scraping news from page');

    const news: Message[] = [];

    // Find all tables on the news page - each table represents one news item
    const newsTables = await page.locator('table').all();

    for (const table of newsTables) {
      try {
        // Get the title from thead td[colspan='2'] - like in Go code
        const title = await table.locator('thead td[colspan="2"]').textContent();

        if (!title?.trim()) continue;

        // Get author from th:contains('Dodał') + next cell - like in Go code
        const authorCell = table.locator('th:has-text("Dodał")').locator('+ td');
        const author = await authorCell.textContent() || '';

        // Get date from th:contains('Data publikacji') + next cell - like in Go code
        const dateCell = table.locator('th:has-text("Data publikacji")').locator('+ td');
        const dateStr = await dateCell.textContent() || '';

        // Get content from th:contains('Treść') + next cell - like in Go code
        const contentCell = table.locator('th:has-text("Treść")').locator('+ td');
        const content = await contentCell.textContent() || '';

        // Generate ID like in Go code (title + content + date)
        const newsId = this.generateNewsId(title.trim(), content.trim(), dateStr.trim());
        const newsUrl = 'https://synergia.librus.pl/ogloszenia';

        const newsItem: Message = {
          id: newsId,
          type: MessageType.MESSAGE_TYPE_NOTIFICATION,
          link: newsUrl,
          author: author.trim(),
          title: title.trim(),
          content: content.trim(),
          dateTimestamp: this.parseDate(dateStr.trim()),
          attachmentsDir: '' // News usually don't have attachments
        };

        news.push(newsItem);
        logger.trace('Scraped news item', { id: newsItem.id, title: newsItem.title, author: newsItem.author });

      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error';
        logger.warn('Failed to scrape news table', { error: errorMessage });
      }
    }

    return news;
  }

  private async scrapeSingleMessage(page: Page, messageUrl: string): Promise<Message | null> {
    logger.debug('Scraping single message content');

    try {
      // Wait for page to load
      await page.waitForLoadState('networkidle');
      await page.waitForSelector('table', { timeout: 10000 });

      // Extract author from table 6 (index 6) which contains message details
      const messageTable = page.locator('table').nth(6);
      const authorCell = messageTable.locator('tr:has(td b:text("Nadawca")) td').nth(1);
      const author = await authorCell.textContent() || '';

      // Extract title from the same table
      const titleCell = messageTable.locator('tr:has(td b:text("Temat")) td').nth(1);
      const title = await titleCell.textContent() || '';

      // Extract date from the same table
      const dateCell = messageTable.locator('tr:has(td b:text("Wysłano")) td').nth(1);
      const dateString = await dateCell.textContent() || '';

      // Extract content from the generic block that comes after the message table
      const contentElement = messageTable.locator('+ *');
      const content = await contentElement.textContent() || '';

      // Parse date like in Go version: "2006-01-02 15:04:05" format
      const dateTimestamp = this.parseLibrusDate(dateString.trim());

      // TODO: Download attachments if present (like in Go version)
      const attachmentsDir = '';

      const message: Message = {
        id: this.generateMessageId(messageUrl),
        type: MessageType.MESSAGE_TYPE_MESSAGE,
        link: messageUrl,
        author: author.trim(),
        title: title.trim(),
        content: content.trim(),
        dateTimestamp,
        attachmentsDir
      };

      return message;

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Failed to scrape single message', { messageUrl, error: errorMessage });
      return null;
    }
  }

  private generateMessageId(url: string): string {
    // Generate MD5 hash from URL like in Go version (for MESSAGE_TYPE_MESSAGE)
    const crypto = require('crypto');
    return crypto.createHash('md5').update(url).digest('hex');
  }

  private extractMessageId(url: string): string {
    // Extract ID from URL like /wiadomosci/1/123 or /ogloszenia/123
    const match = url.match(/\/(\w+)\/(?:\d+\/)?(\d+)/);
    return match ? `${match[1]}_${match[2]}` : url;
  }

  private generateNewsId(title: string, content: string, dateStr: string): string {
    // Generate MD5 hash like in Go code: title + content + date
    const crypto = require('crypto');
    const stringToHash = title + content + dateStr;
    return crypto.createHash('md5').update(stringToHash).digest('hex');
  }

  private parseDate(dateStr: string): number {
    if (!dateStr) return Date.now();

    // Parse Polish date format: YYYY-MM-DD
    const match = dateStr.match(/(\d{4})-(\d{2})-(\d{2})/);
    if (match && match.length >= 4) {
      const year = match[1];
      const month = match[2];
      const day = match[3];
      if (year && month && day) {
        const date = new Date(parseInt(year), parseInt(month) - 1, parseInt(day));
        return date.getTime();
      }
    }

    // Fallback to current timestamp
    return Date.now();
  }

  private parseLibrusDate(dateStr: string): number {
    if (!dateStr) return Date.now();

    // Parse Librus date format: "2006-01-02 15:04:05" (like in Go version)
    const match = dateStr.match(/(\d{4})-(\d{2})-(\d{2})\s+(\d{2}):(\d{2}):(\d{2})/);
    if (match && match.length >= 7 && match[1] && match[2] && match[3] && match[4] && match[5] && match[6]) {
      const year = parseInt(match[1]);
      const month = parseInt(match[2]) - 1; // JavaScript months are 0-based
      const day = parseInt(match[3]);
      const hour = parseInt(match[4]);
      const minute = parseInt(match[5]);
      const second = parseInt(match[6]);

      const date = new Date(year, month, day, hour, minute, second);
      return date.getTime();
    }

    // Fallback to parseDate for simpler formats
    return this.parseDate(dateStr);
  }
}
