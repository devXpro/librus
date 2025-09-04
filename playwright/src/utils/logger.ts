// Logger utility with configurable log levels

export enum LogLevel {
  ERROR = 0,
  WARN = 1,
  INFO = 2,
  DEBUG = 3,
  TRACE = 4
}

class Logger {
  private logLevel: LogLevel;

  constructor() {
    const envLogLevel = process.env.LOG_LEVEL?.toUpperCase();
    this.logLevel = this.parseLogLevel(envLogLevel || 'INFO');
  }

  private parseLogLevel(level: string): LogLevel {
    switch (level) {
      case 'ERROR': return LogLevel.ERROR;
      case 'WARN': return LogLevel.WARN;
      case 'INFO': return LogLevel.INFO;
      case 'DEBUG': return LogLevel.DEBUG;
      case 'TRACE': return LogLevel.TRACE;
      default: return LogLevel.INFO;
    }
  }

  private shouldLog(level: LogLevel): boolean {
    return level <= this.logLevel;
  }

  private formatMessage(level: string, message: string, meta?: any): string {
    const timestamp = new Date().toISOString();
    let metaStr = '';

    if (meta) {
      if (meta.error && meta.error instanceof Error) {
        // Special handling for Error objects
        metaStr = ` ${JSON.stringify({
          ...meta,
          error: {
            name: meta.error.name,
            message: meta.error.message,
            stack: meta.error.stack
          }
        })}`;
      } else {
        metaStr = ` ${JSON.stringify(meta)}`;
      }
    }

    return `[${timestamp}] [${level}] ${message}${metaStr}`;
  }

  error(message: string, meta?: any): void {
    if (this.shouldLog(LogLevel.ERROR)) {
      console.error(this.formatMessage('ERROR', message, meta));
    }
  }

  warn(message: string, meta?: any): void {
    if (this.shouldLog(LogLevel.WARN)) {
      console.warn(this.formatMessage('WARN', message, meta));
    }
  }

  info(message: string, meta?: any): void {
    if (this.shouldLog(LogLevel.INFO)) {
      console.info(this.formatMessage('INFO', message, meta));
    }
  }

  debug(message: string, meta?: any): void {
    if (this.shouldLog(LogLevel.DEBUG)) {
      console.debug(this.formatMessage('DEBUG', message, meta));
    }
  }

  trace(message: string, meta?: any): void {
    if (this.shouldLog(LogLevel.TRACE)) {
      console.log(this.formatMessage('TRACE', message, meta));
    }
  }
}

export const logger = new Logger();
