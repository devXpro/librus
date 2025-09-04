import { Page } from 'playwright';
import { logger } from './logger';
import { LibrusUrls } from './LibrusUrls';

/**
 * Utility class for handling page navigation and waiting logic
 * Eliminates duplication of navigation patterns across the codebase
 */

export class PageNavigator {
  /**
   * Navigates to a URL and waits for page to be ready
   * Standard pattern used throughout the scraper
   */
  static async navigateAndWait(
    page: Page, 
    url: string, 
    selector?: string,
    timeout: number = LibrusUrls.DEFAULT_TIMEOUT
  ): Promise<void> {
    logger.debug('Navigating to page', { url });

    try {
      await page.goto(url);
      await page.waitForLoadState('networkidle');

      if (selector) {
        await page.waitForSelector(selector, { timeout });
        logger.debug('Page loaded and selector found', { url, selector });
      } else {
        logger.debug('Page loaded', { url });
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Failed to navigate to page', { url, selector, error: errorMessage });
      throw new Error(`Navigation failed for ${url}: ${errorMessage}`);
    }
  }

  /**
   * Navigates to messages page and waits for content
   */
  static async navigateToMessages(page: Page): Promise<void> {
    await PageNavigator.navigateWithRetry(
      page,
      LibrusUrls.MESSAGES_URL,
      LibrusUrls.MESSAGE_TABLE_SELECTOR
    );
  }

  /**
   * Navigates to news page and waits for content
   */
  static async navigateToNews(page: Page): Promise<void> {
    await PageNavigator.navigateWithRetry(
      page,
      LibrusUrls.NEWS_URL,
      LibrusUrls.NEWS_TABLE_SELECTOR
    );
  }

  /**
   * Navigates to a specific message URL and waits for content
   */
  static async navigateToMessage(page: Page, messageUrl: string): Promise<void> {
    await PageNavigator.navigateWithRetry(
      page,
      messageUrl,
      'table' // Generic table selector for message content
    );
  }



  /**
   * Retries navigation if it fails
   * Useful for handling temporary network issues
   */
  static async navigateWithRetry(
    page: Page,
    url: string,
    selector?: string,
    maxRetries: number = 3,
    timeout: number = LibrusUrls.DEFAULT_TIMEOUT
  ): Promise<void> {
    let lastError: Error | null = null;

    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        logger.debug('Navigation attempt', { url, attempt, maxRetries });
        await PageNavigator.navigateAndWait(page, url, selector, timeout);
        logger.debug('Navigation successful', { url, attempt });
        return;
      } catch (error) {
        lastError = error instanceof Error ? error : new Error('Unknown error');
        logger.warn('Navigation attempt failed', { 
          url, 
          attempt, 
          maxRetries, 
          error: lastError.message 
        });

        if (attempt < maxRetries) {
          // Wait before retry with exponential backoff
          const delay = Math.pow(2, attempt - 1) * 1000;
          logger.debug('Waiting before retry', { delay });
          await new Promise(resolve => setTimeout(resolve, delay));
        }
      }
    }

    throw new Error(`Navigation failed after ${maxRetries} attempts: ${lastError?.message}`);
  }
}
