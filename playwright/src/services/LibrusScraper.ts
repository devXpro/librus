import { BrowserContext, Page } from 'playwright';
import { SessionManager } from './SessionManager';
import { SessionExecutor } from './SessionExecutor';
import { AttachmentDownloader } from './AttachmentDownloader';
import {
  Message,
  MessageType
} from '../generated/librus_scraper';
import { logger } from '../utils/logger';
import { PageNavigator } from '../utils/PageNavigator';
import { MessageParser } from '../parsers/MessageParser';
import { NewsParser } from '../parsers/NewsParser';
import { LibrusUrls } from '../utils/LibrusUrls';

// Simple interface for credentials (not from protobuf)
interface LoginCredentials {
  login: string;
  password: string;
}

export class LibrusScraper {
  private sessionExecutor: SessionExecutor;

  constructor(sessionManager: SessionManager) {
    this.sessionExecutor = new SessionExecutor(sessionManager);
  }

  async getMessages(
    credentials: LoginCredentials
  ): Promise<Message[]> {
    logger.info('Getting messages', { login: credentials.login });

    return await this.sessionExecutor.executeWithPage(
      credentials,
      async (page: Page) => {
        // Navigate to messages page
        await PageNavigator.navigateToMessages(page);

        // Get message links
        const messageLinks = await MessageParser.extractMessageLinks(page);
        logger.info('Found message links', { login: credentials.login, count: messageLinks.length });

        // Process each message individually
        const messages: Message[] = [];
        for (const link of messageLinks) {
          try {
            const fullUrl = LibrusUrls.getFullUrl(link);

            // Skip javascript links
            if (LibrusUrls.isJavaScriptLink(fullUrl)) {
              continue;
            }

            logger.debug('Processing message', { link: fullUrl });

            // Navigate to message page and scrape it
            await PageNavigator.navigateToMessage(page, fullUrl);
            const message = await this.scrapeSingleMessageWithAttachments(page, fullUrl);

            if (message) {
              messages.push(message);
              logger.trace('Successfully processed message', { id: message.id, title: message.title });
            }

          } catch (error) {
            const errorMessage = error instanceof Error ? error.message : 'Unknown error';
            logger.warn('Failed to process message', { link, error: errorMessage });
            // Continue with other messages
          }
        }

        logger.info('Successfully processed all messages', { login: credentials.login, count: messages.length });
        return messages;
      },
      'getMessages'
    );
  }

  async getNews(
    credentials: LoginCredentials
  ): Promise<Message[]> {
    logger.info('Getting news', { login: credentials.login });

    return await this.sessionExecutor.executeWithPage(
      credentials,
      async (page: Page) => {
        // Navigate to news page
        await PageNavigator.navigateToNews(page);

        // Parse news from page
        const news = await NewsParser.parseNewsList(page);
        logger.info('Successfully scraped news', { login: credentials.login, count: news.length });

        return news;
      },
      'getNews'
    );
  }

  async getAllUpdates(
    credentials: LoginCredentials
  ): Promise<{ messages: Message[], news: Message[] }> {
    logger.info('Getting all updates', { login: credentials.login });

    const results = await this.sessionExecutor.executeParallel(
      credentials,
      [
        {
          operation: async (context: BrowserContext) => {
            const page = await context.newPage();
            try {
              // Use the same logic as GetMessages - get links and parse each message
              await PageNavigator.navigateToMessages(page);
              const messageLinks = await MessageParser.extractMessageLinks(page);

              const messages: Message[] = [];
              for (const link of messageLinks) {
                try {
                  const fullUrl = LibrusUrls.getFullUrl(link);
                  if (LibrusUrls.isJavaScriptLink(fullUrl)) {
                    continue;
                  }

                  await PageNavigator.navigateToMessage(page, fullUrl);
                  const message = await this.scrapeSingleMessageWithAttachments(page, fullUrl);

                  if (message) {
                    messages.push(message);
                  }
                } catch (error) {
                  const errorMessage = error instanceof Error ? error.message : 'Unknown error';
                  logger.warn('Failed to process message in GetAllUpdates', { link, error: errorMessage });
                }
              }

              return messages;
            } finally {
              await page.close();
            }
          },
          name: 'getMessages'
        },
        {
          operation: async (context: BrowserContext) => {
            const page = await context.newPage();
            try {
              await PageNavigator.navigateToNews(page);
              return await NewsParser.parseNewsList(page);
            } finally {
              await page.close();
            }
          },
          name: 'getNews'
        }
      ],
      'getAllUpdates'
    );

    const messages = results[0] || [];
    const news = results[1] || [];

    logger.info('Successfully got all updates', {
      login: credentials.login,
      messagesCount: messages.length,
      newsCount: news.length
    });

    return { messages, news };
  }

  async getSingleMessage(
    credentials: LoginCredentials,
    messageUrl: string
  ): Promise<Message | null> {
    logger.info('Getting single message', { login: credentials.login, messageUrl });

    try {
      return await this.sessionExecutor.executeWithPage(
        credentials,
        async (page: Page) => {
          await PageNavigator.navigateToMessage(page, messageUrl);
          const message = await this.scrapeSingleMessageWithAttachments(page, messageUrl);
          logger.info('Successfully scraped single message', { login: credentials.login, messageId: message?.id });
          return message;
        },
        'getSingleMessage'
      );
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Failed to get single message', { login: credentials.login, messageUrl, error: errorMessage });
      return null;
    }
  }

  async answerMessage(
    credentials: LoginCredentials,
    messageUrl: string,
    answerText: string
  ): Promise<void> {
    logger.info('Answering message', { login: credentials.login, messageUrl });

    await this.sessionExecutor.executeWithRetry(
      credentials,
      async (page: Page) => {
        await PageNavigator.navigateToMessage(page, messageUrl);

        // Wait for reply button to appear and click it
        await page.waitForSelector(LibrusUrls.REPLY_BUTTON_SELECTOR);
        await page.locator(LibrusUrls.REPLY_BUTTON_SELECTOR).click();
        logger.debug('Clicked reply button');

        // Wait for textarea to appear and be ready
        await page.waitForSelector(LibrusUrls.MESSAGE_TEXT_AREA_SELECTOR);

        // Get current text and prepend our answer (like in Go version)
        const textArea = page.locator(LibrusUrls.MESSAGE_TEXT_AREA_SELECTOR);

        // Focus on textarea and move cursor to the beginning
        await textArea.focus();
        await page.keyboard.press('Control+Home'); // Move cursor to start

        // Type our answer at the beginning
        await page.keyboard.type(answerText);
        logger.debug('Filled answer text');

        // Verify text was added at the beginning
        const textareaValue = await textArea.inputValue();
        if (!textareaValue.startsWith(answerText)) {
          throw new Error('Failed to add answer text at the beginning');
        }
        logger.debug('Verified answer text was added at the beginning');

        // Wait for send button to be ready and click it
        await page.waitForSelector(LibrusUrls.SEND_BUTTON_SELECTOR);
        await page.locator(LibrusUrls.SEND_BUTTON_SELECTOR).click();
        logger.debug('Clicked send button');

        // Wait for success message "Wysłano wiadomość" to appear
        await page.waitForSelector('text=Wysłano wiadomość', { timeout: 10000 });
        logger.debug('Success message appeared - message sent successfully');

        logger.info('Successfully answered message', { login: credentials.login, messageUrl });
      },
      'answerMessage',
      2, // maxRetries
      2000 // retryDelay
    );
  }

  /**
   * Scrapes a single message with attachments
   * Combines message parsing and attachment downloading
   */
  private async scrapeSingleMessageWithAttachments(page: Page, messageUrl: string): Promise<Message | null> {
    // Parse basic message data
    const message = await MessageParser.parseSingleMessage(page, messageUrl);
    if (!message) {
      return null;
    }

    // Download attachments if present
    try {
      const attachmentsDir = await AttachmentDownloader.downloadAttachments(page);
      message.attachmentsDir = attachmentsDir;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.warn('Failed to download attachments', { messageUrl, error: errorMessage });
      // Continue without attachments
    }

    return message;
  }

  async validateLogin(credentials: LoginCredentials): Promise<void> {
    logger.info('Validating login credentials', { login: credentials.login });

    // Use SessionExecutor to perform login and validation
    // This will create ephemeral session, perform login, validate it, and clean up
    await this.sessionExecutor.executeWithPage(
      credentials,
      async (page: Page) => {
        // Just navigate to a protected page to ensure session is valid
        await PageNavigator.navigateToMessages(page);
        logger.debug('Successfully navigated to messages page - login valid');

        // Return void - success is indicated by no exception
        return;
      },
      'validateLogin'
    );

    logger.info('Login validation successful', { login: credentials.login });
  }
}
