import { BrowserContext, Browser } from 'playwright';
import { promises as fs } from 'fs';
import path from 'path';

// Simple interfaces (not from protobuf)
interface LoginCredentials {
  login: string;
  password: string;
}
import { logger } from '../utils/logger';
import { LibrusUrls } from '../utils/LibrusUrls';
import { ProxyConfig } from '../utils/ProxyConfig';

export class SessionManager {
  private sessionsDir = './sessions';
  private browser: Browser;

  constructor(browser: Browser) {
    this.browser = browser;
    this.ensureSessionsDir();
  }

  /**
   * Creates browser context with proxy configuration if available
   */
  private async createContextWithProxy(options: any = {}): Promise<BrowserContext> {
    const proxyConfig = ProxyConfig.getProxyConfig();

    const contextOptions = {
      viewport: { width: 2000, height: 2000 },
      userAgent: LibrusUrls.USER_AGENT,
      ...options
    };

    if (proxyConfig) {
      contextOptions.proxy = proxyConfig;
      logger.info('Creating browser context with proxy', {
        server: proxyConfig.server,
        hasAuth: !!(proxyConfig.username && proxyConfig.password)
      });
    } else {
      logger.debug('Creating browser context without proxy');
    }

    return await this.browser.newContext(contextOptions);
  }

  private async ensureSessionsDir(): Promise<void> {
    try {
      await fs.mkdir(this.sessionsDir, { recursive: true });
    } catch (error) {
      logger.error('Failed to create sessions directory', { error });
    }
  }

  private getSessionFilePath(userId: string): string {
    return path.join(this.sessionsDir, `${userId}.json`);
  }

  /**
   * Creates an ephemeral session that should be closed after use
   * Session state is restored from file if available and valid
   */
  async getEphemeralSession(
    credentials: LoginCredentials
  ): Promise<BrowserContext> {
    const userId = credentials.login;
    logger.debug('Creating ephemeral session', { userId });

    // Try to restore from file
    const sessionFile = this.getSessionFilePath(userId);
    try {
      const sessionData = await fs.readFile(sessionFile, 'utf-8');
      const storageState = JSON.parse(sessionData);

      logger.debug('Attempting to restore session from file', { userId });
      const context = await this.createContextWithProxy({ storageState });

      if (await this.validateSession(context)) {
        logger.info('Successfully restored ephemeral session from file', { userId });
        return context;
      } else {
        logger.info('Restored session is invalid, creating new one', { userId });
        await context.close();
        await fs.unlink(sessionFile).catch(() => {}); // Ignore errors
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.debug('Failed to restore session from file', { userId, error: errorMessage });
    }

    // Create new session
    logger.info('Creating new ephemeral session', { userId });
    return await this.createEphemeralSession(credentials);
  }

  private async createEphemeralSession(
    credentials: LoginCredentials
  ): Promise<BrowserContext> {
    const userId = credentials.login;

    // Create context with viewport size like in Go scraper
    const context = await this.createContextWithProxy();

    try {
      await this.performLogin(context, credentials);

      // Save session state immediately after login
      await this.saveSessionState(context, userId);

      logger.info('Successfully created ephemeral session', { userId });
      return context;
    } catch (error) {
      await context.close();
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Failed to create ephemeral session', { userId, error: errorMessage });
      throw error;
    }
  }

  /**
   * Saves session state and closes the context
   * Should be called after each operation to free memory
   */
  async saveAndCloseSession(context: BrowserContext, userId: string): Promise<void> {
    try {
      await this.saveSessionState(context, userId);
      logger.debug('Saved session state to file', { userId });
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Failed to save session state', { userId, error: errorMessage });
    } finally {
      // Always close context to free memory
      await context.close();
      logger.debug('Closed ephemeral session', { userId });
    }
  }

  private async saveSessionState(context: BrowserContext, userId: string): Promise<void> {
    const storageState = await context.storageState();
    const sessionFile = this.getSessionFilePath(userId);
    await fs.writeFile(sessionFile, JSON.stringify(storageState, null, 2));
  }

  private async performLogin(
    context: BrowserContext, 
    credentials: LoginCredentials
  ): Promise<void> {
    logger.debug('Performing login', { login: credentials.login });
    
    const page = await context.newPage();
    
    try {
      // Navigate to login page
      await page.goto(LibrusUrls.PORTAL_URL);
      logger.trace('Navigated to portal page');

      // Wait for page to load completely
      await page.waitForLoadState('networkidle');
      logger.debug('Page loaded completely');

      // Accept cookies - try multiple selectors
      try {
        // Try different cookie consent selectors
        const cookieSelectors = [
          'button:has-text("Akceptuję i przechodzę do")',
          'button:has-text("Akceptuję")',
          '[data-testid="cookie-accept"]',
          '.cookie-accept',
          '#cookie-accept'
        ];

        let cookiesAccepted = false;
        for (const selector of cookieSelectors) {
          try {
            await page.locator(selector).click({ timeout: 3000 });
            logger.debug('Accepted cookies with selector:', selector);
            cookiesAccepted = true;
            break;
          } catch (e) {
            // Try next selector
          }
        }

        if (!cookiesAccepted) {
          logger.debug('No cookies dialog found with any selector');
        }
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error';
        logger.debug('Error handling cookies:', errorMessage);
      }

      // Wait for LIBRUS Synergia button to appear and click it with retry logic
      logger.debug('Looking for LIBRUS Synergia button...');
      await page.getByRole('link', { name: ' LIBRUS Synergia' }).waitFor({ timeout: 10000 });

      // Retry logic for LIBRUS Synergia click and Zaloguj appearance
      let zalogujFound = false;
      for (let attempt = 1; attempt <= 3; attempt++) {
        logger.debug(`Clicking LIBRUS Synergia (attempt ${attempt}/3)...`);
        await page.getByRole('link', { name: ' LIBRUS Synergia' }).first().click();

        try {
          await page.getByRole('link', { name: ' Zaloguj' }).waitFor({ timeout: 1000 });
          logger.debug('Zaloguj button appeared after click');
          zalogujFound = true;
          break;
        } catch (error) {
          logger.debug(`Zaloguj not found after attempt ${attempt}, waiting 500ms...`);
          if (attempt < 3) {
            await page.waitForTimeout(500);
          }
        }
      }

      if (!zalogujFound) {
        throw new Error('Zaloguj button did not appear after 3 attempts to click LIBRUS Synergia');
      }

      // Click on Zaloguj
      logger.debug('Clicking Zaloguj button...');
      await page.getByRole('link', { name: ' Zaloguj' }).click();
      logger.debug('Clicked Zaloguj');

      // Wait for login iframe to appear and load
      await page.locator('#caLoginIframe').waitFor({ timeout: 10000 });
      logger.trace('Login iframe appeared');

      const iframe = page.locator('#caLoginIframe').contentFrame();

      // Wait for login form fields to be available in iframe
      await iframe.getByRole('textbox', { name: 'Login' }).waitFor({ timeout: 10000 });
      await iframe.getByRole('textbox', { name: 'Login' }).fill(credentials.login);
      logger.trace('Filled login');

      // Wait for password field and fill it
      await iframe.getByRole('textbox', { name: 'Hasło' }).waitFor({ timeout: 5000 });
      await iframe.getByRole('textbox', { name: 'Hasło' }).fill(credentials.password);
      logger.trace('Filled password');

      // Wait for login button and click it
      await iframe.getByRole('button', { name: 'Zaloguj' }).waitFor({ timeout: 5000 });
      await iframe.getByRole('button', { name: 'Zaloguj' }).click();
      logger.trace('Clicked login button');

      // Check for authentication error first
      try {
        const errorAlert = iframe.locator('div[role="alert"]');
        await errorAlert.waitFor({ timeout: 3000 });
        const errorText = await errorAlert.textContent();

        if (errorText && errorText.includes('Nieprawidłowy login i/lub hasło')) {
          logger.error('Authentication failed - invalid credentials', { login: credentials.login });
          throw new Error('Invalid login credentials');
        }
      } catch (error) {
        // If it's our authentication error, re-throw it
        if (error instanceof Error && error.message === 'Invalid login credentials') {
          throw error;
        }
        // Otherwise it's a timeout - no error alert found, continue with normal flow
        logger.trace('No authentication error detected');
      }

      // Wait for navigation to main page (any synergia page indicates success)
      await page.waitForURL(LibrusUrls.getLoginSuccessPattern(), { timeout: LibrusUrls.LOGIN_TIMEOUT });
      await page.waitForLoadState('networkidle'); // Wait for page to fully load
      logger.debug('Successfully logged in');

    } finally {
      await page.close();
    }
  }

  private async validateSession(context: BrowserContext): Promise<boolean> {
    logger.trace('Validating session');

    const page = await context.newPage();

    try {
      // Try to access a protected page
      await page.goto(LibrusUrls.MESSAGES_URL, { timeout: LibrusUrls.VALIDATION_TIMEOUT });

      // Check if we're redirected to login page
      const currentUrl = page.url();
      const isValid = LibrusUrls.isLoggedInUrl(currentUrl);

      logger.trace('Session validation result', { isValid, currentUrl });
      return isValid;

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.debug('Session validation failed', { error: errorMessage });
      return false;
    } finally {
      await page.close();
    }
  }

  /**
   * Removes session file for a user (logout)
   */
  async removeSession(login: string): Promise<void> {
    const userId = login;
    logger.debug('Removing session', { userId });

    // Remove session file
    const sessionFile = this.getSessionFilePath(userId);
    try {
      await fs.unlink(sessionFile);
      logger.debug('Removed session file', { userId });
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.debug('Failed to remove session file', { userId, error: errorMessage });
    }
  }

  /**
   * Cleanup method for graceful shutdown
   * Since we don't keep active sessions in memory, this just cleans up session files if needed
   */
  async cleanup(): Promise<void> {
    logger.info('SessionManager cleanup - no active sessions to close');
    // In ephemeral mode, we don't have active sessions to close
    // Session files remain for future use
  }
}
