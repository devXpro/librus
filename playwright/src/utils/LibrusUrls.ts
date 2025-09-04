/**
 * Constants for Librus URLs and selectors
 * Centralized configuration to avoid hardcoded values throughout the codebase
 */

export class LibrusUrls {
  // Base URLs
  static readonly BASE_URL = 'https://synergia.librus.pl';
  static readonly PORTAL_URL = 'https://portal.librus.pl/rodzina';
  static readonly MESSAGES_URL = `${LibrusUrls.BASE_URL}/wiadomosci`;
  static readonly NEWS_URL = `${LibrusUrls.BASE_URL}/ogloszenia`;

  // Selectors for messages page
  static readonly MESSAGE_TABLE_SELECTOR = 'table.decorated';
  static readonly MESSAGE_LINKS_SELECTOR = 'table.decorated td a';
  static readonly UNREAD_MESSAGE_LINKS_SELECTOR = 'table.decorated td[style*="font-weight: bold"] a';

  // Selectors for news page
  static readonly NEWS_TABLE_SELECTOR = 'table';
  static readonly NEWS_TITLE_SELECTOR = 'thead td[colspan="2"]';
  static readonly NEWS_AUTHOR_SELECTOR = 'th:has-text("Dodał") + td';
  static readonly NEWS_DATE_SELECTOR = 'th:has-text("Data publikacji") + td';
  static readonly NEWS_CONTENT_SELECTOR = 'th:has-text("Treść") + td';

  // Selectors for single message page
  static readonly MESSAGE_DETAILS_TABLE_INDEX = 6;
  static readonly MESSAGE_AUTHOR_SELECTOR = 'tr:has(td b:text("Nadawca")) td';
  static readonly MESSAGE_TITLE_SELECTOR = 'tr:has(td b:text("Temat")) td';
  static readonly MESSAGE_DATE_SELECTOR = 'tr:has(td b:text("Wysłano")) td';
  static readonly GENERIC_TABLE_SELECTOR = 'table';

  // Selectors for attachments
  static readonly DOWNLOAD_BUTTON_SELECTOR = 'img[src="/assets/img/homework_files_icons/download.png"]';

  // Selectors for message reply
  static readonly REPLY_BUTTON_SELECTOR = 'input[type="button"][value="Odpowiedz"]';
  static readonly MESSAGE_TEXT_AREA_SELECTOR = '#tresc_wiadomosci';
  static readonly SEND_BUTTON_SELECTOR = 'input[type="submit"][name="wyslij"]';

  // Selectors for login process
  static readonly COOKIES_ACCEPT_BUTTON = 'button:has-text("Akceptuję i przechodzę do")';
  static readonly LIBRUS_SYNERGIA_LINK = 'link:has-text(" LIBRUS Synergia")';
  static readonly LOGIN_LINK = 'link:has-text(" Zaloguj")';
  static readonly LOGIN_BUTTON = 'button:has-text("Zaloguj")';
  static readonly USERNAME_FIELD = 'textbox:has-text("Login")';
  static readonly PASSWORD_FIELD = 'textbox:has-text("Hasło")';
  static readonly ERROR_ALERT_SELECTOR = 'div[role="alert"]';

  // Timeouts
  static readonly DEFAULT_TIMEOUT = 10000;
  static readonly NETWORK_IDLE_TIMEOUT = 30000;
  static readonly REPLY_TIMEOUT = 2000;
  static readonly LOGIN_TIMEOUT = 30000;
  static readonly VALIDATION_TIMEOUT = 10000;

  // User agent
  static readonly USER_AGENT = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.140 Safari/537.36 Edge/17.17134';

  /**
   * Constructs full URL from relative path
   */
  static getFullUrl(relativePath: string): string {
    return new URL(relativePath, LibrusUrls.BASE_URL).toString();
  }

  /**
   * Checks if URL is a JavaScript link that should be skipped
   */
  static isJavaScriptLink(url: string): boolean {
    return url.includes('javascript');
  }

  /**
   * Validates if URL belongs to Librus domain
   */
  static isLibrusUrl(url: string): boolean {
    try {
      const parsedUrl = new URL(url);
      return parsedUrl.hostname.includes('librus.pl');
    } catch {
      return false;
    }
  }

  /**
   * Checks if URL indicates successful login (synergia domain)
   */
  static isLoggedInUrl(url: string): boolean {
    return url.includes('synergia.librus.pl') && !url.includes('/loguj');
  }

  /**
   * Gets URL pattern for login validation
   */
  static getLoginSuccessPattern(): string {
    return '**/synergia.librus.pl/**';
  }
}
