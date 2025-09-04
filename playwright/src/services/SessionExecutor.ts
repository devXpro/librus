import { BrowserContext, Page } from 'playwright';
import { SessionManager } from './SessionManager';
import { logger } from '../utils/logger';

/**
 * Wrapper class that handles the session management pattern
 * Eliminates duplication of session creation, management, and cleanup across all methods
 */

// Simple interface for credentials (not from protobuf)
interface LoginCredentials {
  login: string;
  password: string;
}

// Generic function type for operations that need a page
type PageOperation<T> = (page: Page) => Promise<T>;

// Generic function type for operations that need a context
type ContextOperation<T> = (context: BrowserContext) => Promise<T>;

export class SessionExecutor {
  private sessionManager: SessionManager;

  constructor(sessionManager: SessionManager) {
    this.sessionManager = sessionManager;
  }

  /**
   * Executes an operation with a single page in an ephemeral session
   * Handles session creation, page management, and cleanup automatically
   */
  async executeWithPage<T>(
    credentials: LoginCredentials,
    operation: PageOperation<T>,
    operationName: string = 'unknown'
  ): Promise<T> {
    logger.debug('Executing operation with page', { 
      login: credentials.login, 
      operation: operationName 
    });

    const context = await this.sessionManager.getEphemeralSession(credentials);

    try {
      const page = await context.newPage();
      try {
        const result = await operation(page);
        logger.debug('Operation completed successfully', { 
          login: credentials.login, 
          operation: operationName 
        });
        return result;
      } finally {
        await page.close();
      }
    } finally {
      await this.sessionManager.saveAndCloseSession(context, credentials.login);
    }
  }



  /**
   * Executes multiple operations in parallel using the same context
   * More efficient for operations that can run concurrently
   */
  async executeParallel<T>(
    credentials: LoginCredentials,
    operations: Array<{
      operation: ContextOperation<T>;
      name: string;
    }>,
    operationName: string = 'parallel operations'
  ): Promise<T[]> {
    logger.debug('Executing parallel operations', { 
      login: credentials.login, 
      operation: operationName,
      count: operations.length
    });

    const context = await this.sessionManager.getEphemeralSession(credentials);

    try {
      const promises = operations.map(async ({ operation, name }) => {
        try {
          logger.debug('Starting parallel operation', { name });
          const result = await operation(context);
          logger.debug('Parallel operation completed', { name });
          return result;
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Unknown error';
          logger.error('Parallel operation failed', { name, error: errorMessage });
          throw error;
        }
      });

      const results = await Promise.all(promises);
      logger.debug('All parallel operations completed', { 
        login: credentials.login, 
        operation: operationName,
        count: results.length
      });
      return results;
    } finally {
      await this.sessionManager.saveAndCloseSession(context, credentials.login);
    }
  }

  /**
   * Executes an operation with error handling and retry logic
   * Useful for operations that might fail due to temporary issues
   */
  async executeWithRetry<T>(
    credentials: LoginCredentials,
    operation: PageOperation<T>,
    operationName: string = 'unknown',
    maxRetries: number = 3,
    retryDelay: number = 1000
  ): Promise<T> {
    let lastError: Error | null = null;

    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        logger.debug('Executing operation with retry', { 
          login: credentials.login, 
          operation: operationName,
          attempt,
          maxRetries
        });

        return await this.executeWithPage(credentials, operation, operationName);
      } catch (error) {
        lastError = error instanceof Error ? error : new Error('Unknown error');
        logger.warn('Operation attempt failed', { 
          login: credentials.login,
          operation: operationName,
          attempt,
          maxRetries,
          error: lastError.message
        });

        if (attempt < maxRetries) {
          const delay = retryDelay * attempt; // Exponential backoff
          logger.debug('Waiting before retry', { delay });
          await new Promise(resolve => setTimeout(resolve, delay));
        }
      }
    }

    throw new Error(`Operation failed after ${maxRetries} attempts: ${lastError?.message}`);
  }
}
