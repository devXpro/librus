import { BrowserContext, Browser } from 'playwright';
import { promises as fs } from 'fs';
import path from 'path';
// Simple interfaces (not from protobuf)
interface SessionInfo {
  userId: string;
  context: import('playwright').BrowserContext;
  lastUsed: Date;
  isValid: boolean;
}

interface LoginCredentials {
  login: string;
  password: string;
}
import { logger } from '../utils/logger';

export class SessionManager {
  private activeSessions = new Map<string, BrowserContext>();
  private sessionsDir = './sessions';
  private browser: Browser;

  constructor(browser: Browser) {
    this.browser = browser;
    this.ensureSessionsDir();
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

  async getOrCreateSession(
    credentials: LoginCredentials
  ): Promise<BrowserContext> {
    const userId = credentials.login; // Use login as unique identifier
    logger.debug('Getting or creating session', { userId });

    // Check if we have an active session
    if (this.activeSessions.has(userId)) {
      const context = this.activeSessions.get(userId)!;
      
      // Validate session is still working
      if (await this.validateSession(context)) {
        logger.debug('Using existing active session', { userId });
        return context;
      } else {
        logger.info('Active session is invalid, removing', { userId });
        await context.close();
        this.activeSessions.delete(userId);
      }
    }

    // Try to restore from file
    const sessionFile = this.getSessionFilePath(userId);
    try {
      const sessionData = await fs.readFile(sessionFile, 'utf-8');
      const storageState = JSON.parse(sessionData);
      
      logger.debug('Attempting to restore session from file', { userId });
      const context = await this.browser.newContext({ storageState });
      
      if (await this.validateSession(context)) {
        logger.info('Successfully restored session from file', { userId });
        this.activeSessions.set(userId, context);
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
    logger.info('Creating new session', { userId });
    return await this.createNewSession(credentials);
  }

  private async createNewSession(
    credentials: LoginCredentials
  ): Promise<BrowserContext> {
    const userId = credentials.login;
    const context = await this.browser.newContext();
    
    try {
      await this.performLogin(context, credentials);
      
      // Save session state
      const storageState = await context.storageState();
      const sessionFile = this.getSessionFilePath(userId);
      await fs.writeFile(sessionFile, JSON.stringify(storageState, null, 2));
      
      this.activeSessions.set(userId, context);
      logger.info('Successfully created and saved new session', { userId });
      
      return context;
    } catch (error) {
      await context.close();
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('Failed to create new session', { userId, error: errorMessage });
      throw error;
    }
  }

  private async performLogin(
    context: BrowserContext, 
    credentials: LoginCredentials
  ): Promise<void> {
    logger.debug('Performing login', { login: credentials.login });
    
    const page = await context.newPage();
    
    try {
      // Navigate to login page
      await page.goto('https://portal.librus.pl/rodzina');
      logger.trace('Navigated to portal page');

      // Accept cookies
      try {
        await page.getByRole('button', { name: 'Akceptuję i przechodzę do' }).click({ timeout: 5000 });
        logger.trace('Accepted cookies');
      } catch (error) {
        logger.debug('No cookies dialog found or already accepted');
      }

      // Wait for LIBRUS Synergia button to appear and click it
      await page.getByRole('link', { name: ' LIBRUS Synergia' }).waitFor({ timeout: 2000 });
      await page.getByRole('link', { name: ' LIBRUS Synergia' }).click();
      logger.trace('Clicked LIBRUS Synergia');

      // Wait for dropdown menu to appear and click on Zaloguj
      await page.getByRole('link', { name: ' Zaloguj' }).waitFor({ timeout: 5000 });
      await page.getByRole('link', { name: ' Zaloguj' }).click();
      logger.trace('Clicked Zaloguj');

      // Wait for login form in iframe
      const iframe = page.locator('#caLoginIframe').contentFrame();
      
      // Fill login
      await iframe.getByRole('textbox', { name: 'Login' }).fill(credentials.login);
      logger.trace('Filled login');

      // Fill password
      await iframe.getByRole('textbox', { name: 'Hasło' }).fill(credentials.password);
      logger.trace('Filled password');

      // Click login button
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
      await page.waitForURL('**/synergia.librus.pl/**', { timeout: 30000 });
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
      await page.goto('https://synergia.librus.pl/wiadomosci', { timeout: 10000 });

      // Check if we're redirected to login page
      const currentUrl = page.url();
      const isValid = !currentUrl.includes('/loguj');

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

  async closeSession(login: string): Promise<void> {
    const userId = login;
    logger.debug('Closing session', { userId });
    
    const context = this.activeSessions.get(userId);
    if (context) {
      await context.close();
      this.activeSessions.delete(userId);
    }

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

  async closeAllSessions(): Promise<void> {
    logger.info('Closing all sessions');
    
    for (const [userId, context] of this.activeSessions) {
      try {
        await context.close();
        logger.debug('Closed session', { userId });
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error';
        logger.error('Failed to close session', { userId, error: errorMessage });
      }
    }
    
    this.activeSessions.clear();
  }
}
