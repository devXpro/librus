import { createServer } from 'nice-grpc';
import { ServerReflectionService, ServerReflection } from 'nice-grpc-server-reflection';
import { Status, ServerError } from 'nice-grpc-common';
import * as fs from 'fs';
import * as path from 'path';
import { chromium, Browser } from 'playwright';
import { SessionManager } from './services/SessionManager';
import { LibrusScraper } from './services/LibrusScraper';
import { logger } from './utils/logger';
import { LibrusUrls } from './utils/LibrusUrls';
import {
  LibrusScraperDefinition,
  LibrusScraperServiceImplementation,
  GetMessagesRequest,
  GetMessagesResponse,
  GetNewsRequest,
  GetNewsResponse,
  GetAllUpdatesRequest,
  GetAllUpdatesResponse,
  GetSingleMessageRequest,
  GetSingleMessageResponse,
  AnswerMessageRequest,
  AnswerMessageResponse,
  HealthCheckRequest,
  HealthCheckResponse
} from './generated/librus_scraper';
import type { CallContext } from 'nice-grpc-common';
import type { DeepPartial } from './generated/librus_scraper';
import dotenv from 'dotenv';

// Load environment variables
dotenv.config({ path: '../.env' });

class LibrusScraperService {
  private browser: Browser | null = null;
  private sessionManager: SessionManager | null = null;
  private scraper: LibrusScraper | null = null;

  private isAuthenticationError(errorMessage: string): boolean {
    return errorMessage.includes('authentication') ||
           errorMessage.includes('login') ||
           errorMessage.includes('Invalid login credentials');
  }

  async initialize(): Promise<void> {
    logger.info('Initializing Librus Scraper Service');

    try {
      // Launch browser
      const isDevelopment = process.env.DEVELOPMENT === 'true';
      const headless = !isDevelopment; // If DEVELOPMENT=true, run with GUI

      logger.info('Browser configuration', {
        isDevelopment,
        headless,
        mode: isDevelopment ? 'GUI (development)' : 'headless (production)'
      });

      this.browser = await chromium.launch({
        headless,
        args: isDevelopment ? [
          '--disable-web-security'
        ] : [
          '--disable-web-security',
          '--disable-gpu',
          '--no-sandbox',
          `--user-agent=${LibrusUrls.USER_AGENT}`
        ]
      });

      this.sessionManager = new SessionManager(this.browser);
      this.scraper = new LibrusScraper(this.sessionManager);

      logger.info('Service initialized successfully');
    } catch (error) {
      logger.error('Failed to initialize service', { error });
      throw error;
    }
  }

  async shutdown(): Promise<void> {
    logger.info('Shutting down service');

    if (this.sessionManager) {
      await this.sessionManager.cleanup();
    }

    if (this.browser) {
      await this.browser.close();
    }

    logger.info('Service shutdown complete');
  }

  // gRPC method implementations with proper error handling
  async getMessages(request: GetMessagesRequest, context: CallContext): Promise<DeepPartial<GetMessagesResponse>> {
    const { login, password } = request;

    logger.info('gRPC: GetMessages called', { login });

    if (!this.scraper) {
      throw new ServerError(Status.INTERNAL, 'Service not initialized');
    }

    try {
      const messages = await this.scraper.getMessages({ login, password });

      return {
        messages
      };

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('gRPC: GetMessages failed', { login, error: errorMessage });

      // Determine appropriate gRPC status code
      if (this.isAuthenticationError(errorMessage)) {
        throw new ServerError(Status.UNAUTHENTICATED, 'Authentication failed');
      }

      throw new ServerError(Status.INTERNAL, errorMessage);
    }
  }

  async getNews(request: GetNewsRequest, context: CallContext): Promise<DeepPartial<GetNewsResponse>> {
    const { login, password } = request;

    logger.info('gRPC: GetNews called', { login });

    if (!this.scraper) {
      throw new ServerError(Status.INTERNAL, 'Service not initialized');
    }

    try {
      const news = await this.scraper.getNews({ login, password });

      return {
        news
      };

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('gRPC: GetNews failed', { login, error: errorMessage });

      if (this.isAuthenticationError(errorMessage)) {
        throw new ServerError(Status.UNAUTHENTICATED, 'Authentication failed');
      }

      throw new ServerError(Status.INTERNAL, errorMessage);
    }
  }

  async getAllUpdates(request: GetAllUpdatesRequest, context: CallContext): Promise<DeepPartial<GetAllUpdatesResponse>> {
    const { login, password } = request;

    logger.info('gRPC: GetAllUpdates called', { login });

    if (!this.scraper) {
      throw new ServerError(Status.INTERNAL, 'Service not initialized');
    }

    try {
      const { messages, news } = await this.scraper.getAllUpdates({ login, password });

      return {
        messages,
        news
      };

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('gRPC: GetAllUpdates failed', { login, error: errorMessage });

      if (this.isAuthenticationError(errorMessage)) {
        throw new ServerError(Status.UNAUTHENTICATED, 'Authentication failed');
      }

      throw new ServerError(Status.INTERNAL, errorMessage);
    }
  }

  async getSingleMessage(request: GetSingleMessageRequest, context: CallContext): Promise<DeepPartial<GetSingleMessageResponse>> {
    const { login, password, messageUrl } = request;

    logger.info('gRPC: GetSingleMessage called', { login, messageUrl });

    if (!this.scraper) {
      throw new ServerError(Status.INTERNAL, 'Service not initialized');
    }

    try {
      const message = await this.scraper.getSingleMessage({ login, password }, messageUrl);

      if (!message) {
        throw new ServerError(Status.NOT_FOUND, 'Message not found');
      }

      return {
        message
      };

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('gRPC: GetSingleMessage failed', { login, messageUrl, error: errorMessage });

      if (errorMessage.includes('not found')) {
        throw new ServerError(Status.NOT_FOUND, 'Message not found');
      }

      if (this.isAuthenticationError(errorMessage)) {
        throw new ServerError(Status.UNAUTHENTICATED, 'Authentication failed');
      }

      throw new ServerError(Status.INTERNAL, errorMessage);
    }
  }

  async answerMessage(request: AnswerMessageRequest, context: CallContext): Promise<DeepPartial<AnswerMessageResponse>> {
    const { login, password, messageUrl, answerText } = request;

    logger.info('gRPC: AnswerMessage called', { login, messageUrl });

    if (!this.scraper) {
      throw new ServerError(Status.INTERNAL, 'Service not initialized');
    }

    try {
      await this.scraper.answerMessage({ login, password }, messageUrl, answerText);

      // Empty response indicates success
      return {};

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error('gRPC: AnswerMessage failed', { login, messageUrl, error: errorMessage });

      if (this.isAuthenticationError(errorMessage)) {
        throw new ServerError(Status.UNAUTHENTICATED, 'Authentication failed');
      }

      throw new ServerError(Status.INTERNAL, errorMessage);
    }
  }

  async healthCheck(request: HealthCheckRequest, context: CallContext): Promise<DeepPartial<HealthCheckResponse>> {
    logger.debug('gRPC: HealthCheck called');

    const healthy = this.browser !== null && this.scraper !== null;

    return {
      healthy,
      status: healthy ? 'Service is running' : 'Service not initialized'
    };
  }
}

async function startServer(): Promise<void> {
  const service = new LibrusScraperService();

  try {
    await service.initialize();

    const server = createServer();
    server.add(LibrusScraperDefinition, service as any);

    // Add reflection for grpcurl support
    try {
      const protosetPath = path.join(__dirname, '../protoset.bin');
      if (fs.existsSync(protosetPath)) {
        server.add(
          ServerReflectionService,
          ServerReflection(
            fs.readFileSync(protosetPath),
            [LibrusScraperDefinition.fullName],
          ),
        );
        logger.info('Server reflection enabled');
      } else {
        logger.warn('Protoset file not found, reflection disabled');
      }
    } catch (error) {
      logger.warn('Failed to enable reflection', { error });
    }

    const port = process.env.GRPC_PORT || '50051';
    const bindAddress = `0.0.0.0:${port}`;

    await server.listen(bindAddress);
    logger.info(`nice-grpc server started on ${bindAddress}`);

    // Graceful shutdown
    const shutdown = async () => {
      logger.info('Shutting down gracefully');
      server.shutdown();
      await service.shutdown();
      process.exit(0);
    };

    process.on('SIGINT', shutdown);
    process.on('SIGTERM', shutdown);

  } catch (error) {
    logger.error('Failed to start server', { error });
    process.exit(1);
  }
}

if (require.main === module) {
  startServer();
}
