/**
 * Proxy configuration utility
 * Handles HTTP proxy settings from environment variables
 */

export interface ProxySettings {
  server: string;
  username?: string;
  password?: string;
}

export class ProxyConfig {
  /**
   * Gets proxy configuration from environment variables
   * Returns null if proxy is not configured
   */
  static getProxyConfig(): ProxySettings | null {
    const host = process.env.HTTP_PROXY_HOST;
    const port = process.env.HTTP_PROXY_PORT;
    
    // If host and port are not set, no proxy
    if (!host || !port) {
      return null;
    }

    const username = process.env.HTTP_PROXY_USER;
    const password = process.env.HTTP_PROXY_PASSWORD;

    const server = `http://${host}:${port}`;

    // Return proxy config with optional auth
    const config: ProxySettings = { server };
    
    if (username && password) {
      config.username = username;
      config.password = password;
    }

    return config;
  }

  /**
   * Checks if proxy is configured
   */
  static isProxyEnabled(): boolean {
    return this.getProxyConfig() !== null;
  }

  /**
   * Gets proxy server URL for logging purposes (without credentials)
   */
  static getProxyServerUrl(): string | null {
    const config = this.getProxyConfig();
    return config ? config.server : null;
  }
}
