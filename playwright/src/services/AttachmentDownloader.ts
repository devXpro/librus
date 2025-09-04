import { Page } from 'playwright';
import * as fs from 'fs';
import * as path from 'path';
import axios from 'axios';
import { logger } from '../utils/logger';
import { LibrusUrls } from '../utils/LibrusUrls';
import { IdGenerator } from '../utils/IdGenerator';

/**
 * Service class responsible for downloading message attachments
 * Handles attachment detection, URL extraction, and file downloading
 */

export class AttachmentDownloader {
  private static readonly ATTACHMENTS_BASE_DIR = './attachments';
  private static readonly MAX_RETRIES = 3;
  private static readonly RETRY_DELAY_MS = 1000;

  /**
   * Downloads all attachments from a message page
   * Returns UUID of the directory where attachments were saved, or empty string if no attachments
   */
  static async downloadAttachments(page: Page): Promise<string> {
    try {
      logger.debug('Looking for download buttons...');

      // Find download buttons
      const downloadButtons = await page.locator(LibrusUrls.DOWNLOAD_BUTTON_SELECTOR).all();

      if (downloadButtons.length === 0) {
        logger.debug('No attachments found on this page');
        return '';
      }

      logger.debug(`Found ${downloadButtons.length} attachment(s)`);

      // Create UUID directory for attachments
      const uuid = IdGenerator.generateAttachmentDirId();
      const attachmentsDir = path.join(AttachmentDownloader.ATTACHMENTS_BASE_DIR, uuid);

      // Create directory
      await AttachmentDownloader.ensureDirectoryExists(attachmentsDir);
      logger.debug(`Created attachments directory: ${attachmentsDir}`);

      // Get cookies from page context for HTTP requests
      const cookies = await page.context().cookies();

      // Process each attachment
      for (let i = 0; i < downloadButtons.length; i++) {
        try {
          logger.debug(`Processing attachment ${i + 1} of ${downloadButtons.length}`);
          await AttachmentDownloader.downloadSingleAttachment(
            downloadButtons[i], 
            cookies, 
            attachmentsDir, 
            i + 1
          );
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

  /**
   * Downloads a single attachment
   */
  private static async downloadSingleAttachment(
    downloadButton: any,
    cookies: any[],
    targetDir: string,
    index: number
  ): Promise<void> {
    // Get onclick attribute value
    const onclickValue = await downloadButton.getAttribute('onclick');
    if (!onclickValue) {
      throw new Error(`No onclick attribute for attachment ${index}`);
    }

    // Extract URL from onclick
    const relativeURL = AttachmentDownloader.extractURLFromOnclick(onclickValue);
    if (!relativeURL) {
      throw new Error(`Failed to extract URL from onclick for attachment ${index}`);
    }

    // Construct full URL
    const fullURL = LibrusUrls.BASE_URL + relativeURL;
    logger.debug(`Extracted URL: ${fullURL}`);

    // Download with retry logic
    await AttachmentDownloader.downloadFileWithRetry(
      fullURL, 
      cookies, 
      targetDir, 
      index
    );
  }

  /**
   * Downloads file with retry logic
   */
  private static async downloadFileWithRetry(
    url: string,
    cookies: any[],
    targetDir: string,
    index: number
  ): Promise<void> {
    let lastError: Error | null = null;

    for (let attempt = 1; attempt <= AttachmentDownloader.MAX_RETRIES; attempt++) {
      try {
        logger.debug(`Download attempt ${attempt}/${AttachmentDownloader.MAX_RETRIES}`, { url });
        await AttachmentDownloader.downloadFile(url, cookies, targetDir, index);
        logger.debug(`Successfully downloaded attachment ${index}`);
        return;
      } catch (error) {
        lastError = error instanceof Error ? error : new Error('Unknown error');
        logger.warn(`Download attempt ${attempt} failed`, { 
          url, 
          attempt, 
          error: lastError.message 
        });

        if (attempt < AttachmentDownloader.MAX_RETRIES) {
          const delay = AttachmentDownloader.RETRY_DELAY_MS * attempt;
          logger.debug(`Waiting ${delay}ms before retry`);
          await new Promise(resolve => setTimeout(resolve, delay));
        }
      }
    }

    throw new Error(`Download failed after ${AttachmentDownloader.MAX_RETRIES} attempts: ${lastError?.message}`);
  }

  /**
   * Downloads a single file
   */
  private static async downloadFile(
    url: string,
    cookies: any[],
    targetDir: string,
    index: number
  ): Promise<void> {
    // Filter relevant cookies
    const relevantCookies = AttachmentDownloader.filterRelevantCookies(cookies, url);
    const cookieString = relevantCookies
      .map(cookie => `${cookie.name}=${cookie.value}`)
      .join('; ');

    logger.debug(`Using ${relevantCookies.length} relevant cookies out of ${cookies.length} total`);

    // Get redirect URL
    const redirectResponse = await axios.get(url, {
      headers: {
        'Cookie': cookieString,
        'User-Agent': LibrusUrls.USER_AGENT,
        'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8'
      },
      maxRedirects: 0,
      validateStatus: () => true // Accept all status codes
    });

    // Validate redirect response
    AttachmentDownloader.validateRedirectResponse(redirectResponse);

    // Get download URL
    const downloadURL = AttachmentDownloader.getDownloadURL(redirectResponse, url);
    logger.debug(`Download URL: ${downloadURL}`);

    // Download the file
    const fileResponse = await axios.get(downloadURL, {
      headers: {
        'Cookie': cookieString,
        'User-Agent': LibrusUrls.USER_AGENT
      },
      responseType: 'stream'
    });

    // Save file
    const filename = AttachmentDownloader.extractFilename(fileResponse, downloadURL, index);
    const filePath = path.join(targetDir, filename);

    await AttachmentDownloader.saveStreamToFile(fileResponse.data, filePath);
    logger.debug(`Successfully saved attachment to ${filePath}`);
  }

  /**
   * Extracts URL from onclick attribute
   */
  private static extractURLFromOnclick(onclick: string): string | null {
    try {
      const startStr = 'otworz_w_nowym_oknie(';
      if (!onclick.includes(startStr)) {
        return null;
      }

      const parts = onclick.split(startStr);
      if (parts.length < 2) {
        return null;
      }

      const paramsPart = parts[1]?.trim();
      if (!paramsPart) {
        return null;
      }

      // Try different matching patterns
      const patterns = [
        // Pattern 1: Double quotes
        /"([^"]+)"/,
        // Pattern 2: HTML entity quotes
        /&quot;([^&]+)&quot;/,
        // Pattern 3: Single quotes
        /'([^']+)'/
      ];

      for (const pattern of patterns) {
        const match = paramsPart.match(pattern);
        if (match && match[1]) {
          return match[1].replace(/\\\//g, '/');
        }
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

  /**
   * Filters cookies relevant for the request domain
   */
  private static filterRelevantCookies(cookies: any[], url: string): any[] {
    return cookies.filter(cookie => {
      if (url.includes(cookie.domain)) {
        return true;
      }

      try {
        const parsedURL = new URL(url);
        return cookie.domain.includes(parsedURL.hostname);
      } catch {
        return false;
      }
    });
  }

  /**
   * Validates redirect response status
   */
  private static validateRedirectResponse(response: any): void {
    const validRedirectCodes = [301, 302, 307, 308];
    if (!validRedirectCodes.includes(response.status)) {
      throw new Error(`Expected redirect, got status code: ${response.status}`);
    }

    if (!response.headers.location) {
      throw new Error(`No Location header in response. Status: ${response.status}`);
    }
  }

  /**
   * Constructs download URL from redirect response
   */
  private static getDownloadURL(redirectResponse: any, originalUrl: string): string {
    let redirectURL = redirectResponse.headers.location;
    
    if (!redirectURL.startsWith('http')) {
      const baseURL = new URL(originalUrl);
      redirectURL = new URL(redirectURL, baseURL.origin).toString();
    }

    return redirectURL + '/get';
  }

  /**
   * Extracts filename from response headers or generates one
   */
  private static extractFilename(response: any, downloadURL: string, index: number): string {
    const contentDisposition = response.headers['content-disposition'];
    if (contentDisposition && contentDisposition.includes('filename=')) {
      const parts = contentDisposition.split('filename=');
      if (parts.length > 1) {
        const filename = parts[1]?.replace(/['"]/g, '') || '';
        if (filename) {
          return path.basename(filename);
        }
      }
    }

    // Generate filename if not found
    const urlParts = downloadURL.split('/');
    const tokenPart = urlParts[urlParts.length - 2] || 'unknown';
    return `attachment_${index}_${tokenPart}`;
  }

  /**
   * Saves stream to file
   */
  private static async saveStreamToFile(stream: any, filePath: string): Promise<void> {
    return new Promise((resolve, reject) => {
      const writer = fs.createWriteStream(filePath);
      stream.pipe(writer);
      writer.on('finish', resolve);
      writer.on('error', reject);
    });
  }

  /**
   * Ensures directory exists
   */
  private static async ensureDirectoryExists(dirPath: string): Promise<void> {
    try {
      await fs.promises.mkdir(dirPath, { recursive: true });
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      throw new Error(`Failed to create directory ${dirPath}: ${errorMessage}`);
    }
  }
}
