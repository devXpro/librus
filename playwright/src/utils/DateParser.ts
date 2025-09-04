/**
 * Utility class for parsing different date formats used in Librus
 * Handles Polish date formats and converts them to Unix timestamps
 */

export class DateParser {
  /**
   * Parses simple date format: YYYY-MM-DD
   * Used in message lists and news
   */
  static parseSimpleDate(dateStr: string): number {
    if (!dateStr?.trim()) {
      return Date.now();
    }

    const match = dateStr.match(/(\d{4})-(\d{2})-(\d{2})/);
    if (match && match.length >= 4) {
      const year = match[1];
      const month = match[2];
      const day = match[3];

      if (year && month && day) {
        const date = new Date(parseInt(year), parseInt(month) - 1, parseInt(day));
        return date.getTime();
      }
    }

    return Date.now();
  }

  /**
   * Parses full Librus date format: "YYYY-MM-DD HH:mm:ss"
   * Used in detailed message views
   */
  static parseLibrusDateTime(dateStr: string): number {
    if (!dateStr?.trim()) {
      return Date.now();
    }

    const match = dateStr.match(/(\d{4})-(\d{2})-(\d{2})\s+(\d{2}):(\d{2}):(\d{2})/);
    if (match && match.length >= 7) {
      const [, year, month, day, hour, minute, second] = match;

      if (year && month && day && hour && minute && second) {
        const date = new Date(
          parseInt(year),
          parseInt(month) - 1, // JavaScript months are 0-based
          parseInt(day),
          parseInt(hour),
          parseInt(minute),
          parseInt(second)
        );
        return date.getTime();
      }
    }

    // Fallback to simple date parsing
    return DateParser.parseSimpleDate(dateStr);
  }

  /**
   * Parses any date format by trying different parsers
   * Automatically detects the format and uses appropriate parser
   */
  static parseAnyDate(dateStr: string): number {
    if (!dateStr?.trim()) {
      return Date.now();
    }

    // Try full datetime format first
    if (dateStr.includes(':')) {
      return DateParser.parseLibrusDateTime(dateStr);
    }

    // Fall back to simple date format
    return DateParser.parseSimpleDate(dateStr);
  }
}
