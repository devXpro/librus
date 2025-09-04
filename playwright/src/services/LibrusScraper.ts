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

      // Download attachments if present (like in Go version)
      const attachmentsDir = await this.downloadAttachments(page);

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

  private async downloadAttachments(page: Page): Promise<string> {
    try {
      logger.debug('Looking for download buttons...');

      // Find download buttons (same selector as Go version)
      const downloadButtons = await page.locator('img[src="/assets/img/homework_files_icons/download.png"]').all();

      if (downloadButtons.length === 0) {
        logger.debug('No attachments found on this page');
        return '';
      }

      logger.debug(`Found ${downloadButtons.length} attachment(s)`);

      // Create UUID directory for attachments
      const { v4: uuidv4 } = require('uuid');
      const fs = require('fs');
      const path = require('path');

      const uuid = uuidv4();
      const attachmentsDir = path.join('./attachments', uuid);

      // Create directory
      fs.mkdirSync(attachmentsDir, { recursive: true });
      logger.debug(`Created attachments directory: ${attachmentsDir}`);

      // Get cookies from page context for HTTP requests
      const cookies = await page.context().cookies();

      // Process each attachment
      for (let i = 0; i < downloadButtons.length; i++) {
        try {
          logger.debug(`Processing attachment ${i + 1} of ${downloadButtons.length}`);

          // Get onclick attribute value
          const onclickValue = await downloadButtons[i]?.getAttribute('onclick');
          if (!onclickValue) {
            logger.warn(`No onclick attribute for attachment ${i + 1}`);
            continue;
          }

          // Extract URL from onclick (same logic as Go version)
          const relativeURL = this.extractURLFromOnclick(onclickValue);
          if (!relativeURL) {
            logger.warn(`Failed to extract URL from onclick for attachment ${i + 1}`);
            continue;
          }

          // Construct full URL
          const fullURL = 'https://synergia.librus.pl' + relativeURL;
          logger.debug(`Extracted URL: ${fullURL}`);

          // Get redirect URL and download file
          await this.downloadSingleAttachment(fullURL, cookies, attachmentsDir, i + 1);

        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Unknown error';
          logger.warn(`Failed to process attachment ${i + 1}`, { error: errorMessage });
          // Continue with other attachments
        }
      }

      return uuid; // Return UUID directory name

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Failed to download attachments', { error: errorMessage });
      return ''; // Return empty string on error
    }
  }

  private extractURLFromOnclick(onclick: string): string | null {
    try {
      // Looking for otworz_w_nowym_oknie("URL", "o2", 420, 250) - same as Go version
      const startStr = 'otworz_w_nowym_oknie(';
      if (!onclick.includes(startStr)) {
        return null;
      }

      // Split by the function name
      const parts = onclick.split(startStr);
      if (parts.length < 2) {
        return null;
      }

      // Get the part after the function name
      const paramsPart = parts[1]?.trim();
      if (!paramsPart) {
        return null;
      }

      // Try different matching patterns (same as Go version)

      // Pattern 1: Double quotes without HTML entities
      let match = paramsPart.indexOf('"');
      if (match >= 0) {
        const endMatch = paramsPart.indexOf('"', match + 1);
        if (endMatch >= 0) {
          let url = paramsPart.substring(match + 1, endMatch);
          url = url.replace(/\\\//g, '/');
          return url;
        }
      }

      // Pattern 2: HTML entity quotes &quot;
      match = paramsPart.indexOf('&quot;');
      if (match >= 0) {
        const endMatch = paramsPart.indexOf('&quot;', match + 6);
        if (endMatch >= 0) {
          let url = paramsPart.substring(match + 6, endMatch);
          url = url.replace(/\\\//g, '/');
          return url;
        }
      }

      // Pattern 3: Regex for quotes
      const regexMatch = paramsPart.match(/["']([^"']+)["']/);
      if (regexMatch && regexMatch[1]) {
        let url = regexMatch[1];
        url = url.replace(/\\\//g, '/');
        return url;
      }

      // Last resort: Look for URL pattern
      const fallbackMatch = paramsPart.match(/\/wiadomosci\/pobierz_zalacznik\/\d+\/\d+/);
      if (fallbackMatch && fallbackMatch[0]) {
        return fallbackMatch[0];
      }

      return null;

    } catch (error) {
      logger.warn('Error extracting URL from onclick', { onclick, error });
      return null;
    }
  }

  private async downloadSingleAttachment(
    url: string,
    cookies: any[],
    targetDir: string,
    index: number
  ): Promise<void> {
    try {
      const axios = require('axios');
      const fs = require('fs');
      const path = require('path');

      // Filter cookies by domain (same logic as Go version)
      const relevantCookies = cookies.filter(cookie => {
        // Only add cookies that are relevant for the domain we're requesting
        if (url.includes(cookie.domain)) {
          return true;
        }

        // Try to check by hostname
        try {
          const parsedURL = new URL(url);
          if (cookie.domain.includes(parsedURL.hostname)) {
            return true;
          }
        } catch (error) {
          // Ignore URL parsing errors
        }

        return false;
      });

      // Convert filtered cookies to cookie string
      const cookieString = relevantCookies
        .map(cookie => `${cookie.name}=${cookie.value}`)
        .join('; ');

      logger.debug(`Using ${relevantCookies.length} relevant cookies out of ${cookies.length} total`);

      // Get redirect URL (same logic as Go version)
      let redirectResponse;
      try {
        redirectResponse = await axios.get(url, {
          headers: {
            'Cookie': cookieString,
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.140 Safari/537.36 Edge/17.17134',
            'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8'
          },
          maxRedirects: 0,
          validateStatus: () => true // Accept all status codes
        });
      } catch (error: any) {
        // Handle axios errors (like redirect errors)
        if (error.response) {
          redirectResponse = error.response;
        } else {
          throw error;
        }
      }

      // Check if we got a redirect status code (same as Go version)
      if (redirectResponse.status !== 302 && // StatusFound
          redirectResponse.status !== 301 && // StatusMovedPermanently
          redirectResponse.status !== 307 && // StatusTemporaryRedirect
          redirectResponse.status !== 308) { // StatusPermanentRedirect
        logger.warn(`Expected redirect, got status code: ${redirectResponse.status}`);

        // Log response body for debugging
        if (redirectResponse.data && typeof redirectResponse.data === 'string') {
          logger.warn('Response body:', redirectResponse.data.substring(0, 500));
        }

        throw new Error(`Expected redirect, got status code: ${redirectResponse.status}`);
      }

      // Get Location header
      const location = redirectResponse.headers.location;
      if (!location) {
        logger.warn(`No Location header in response. Status: ${redirectResponse.status}, Headers:`, redirectResponse.headers);

        // Log response body for debugging
        if (redirectResponse.data && typeof redirectResponse.data === 'string') {
          logger.warn('Response body:', redirectResponse.data.substring(0, 500));
        }

        throw new Error(`No Location header in response. Status: ${redirectResponse.status}`);
      }

      // Construct absolute URL if needed
      let redirectURL = location;
      if (!location.startsWith('http')) {
        const baseURL = new URL(url);
        redirectURL = new URL(location, baseURL.origin).toString();
      }

      // Add /get to the URL (same as Go version)
      const downloadURL = redirectURL + '/get';
      logger.debug(`Download URL: ${downloadURL}`);

      // Download the file
      const fileResponse = await axios.get(downloadURL, {
        headers: {
          'Cookie': cookieString,
          'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
        },
        responseType: 'stream'
      });

      // Get filename from Content-Disposition header or generate one
      let filename = '';
      const contentDisposition = fileResponse.headers['content-disposition'];
      if (contentDisposition && contentDisposition.includes('filename=')) {
        const parts = contentDisposition.split('filename=');
        if (parts.length > 1) {
          filename = parts[1]?.replace(/['"]/g, '') || '';
        }
      }

      // Generate filename if not found
      if (!filename) {
        const urlParts = downloadURL.split('/');
        const tokenPart = urlParts[urlParts.length - 2] || 'unknown';
        filename = `attachment_${index}_${tokenPart}`;
      }

      // Clean filename
      filename = path.basename(filename);
      const filePath = path.join(targetDir, filename);

      // Save file
      const writer = fs.createWriteStream(filePath);
      fileResponse.data.pipe(writer);

      await new Promise((resolve, reject) => {
        writer.on('finish', resolve);
        writer.on('error', reject);
      });

      logger.debug(`Successfully downloaded attachment ${index} to ${filePath}`);

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.warn(`Failed to download attachment ${index}`, { url, error: errorMessage });
      throw error;
    }
  }
}
