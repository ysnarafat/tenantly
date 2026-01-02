import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root',
})
export class DateUtilsService {
  /**
   * Parse UTC timestamp from API to Date object
   * API returns timestamps in UTC (e.g., "2025-11-03T17:24:13.554132")
   */
  parseUTCTimestamp(timestamp: string): Date {
    // Ensure the timestamp is treated as UTC
    if (!timestamp.endsWith('Z') && !timestamp.includes('+')) {
      timestamp = timestamp + 'Z';
    }
    return new Date(timestamp);
  }

  /**
   * Format Date to UTC ISO string for API requests
   */
  toUTCString(date: Date): string {
    return date.toISOString();
  }

  /**
   * Get current UTC timestamp
   */
  nowUTC(): Date {
    return new Date();
  }

  /**
   * Format UTC timestamp to local date string
   */
  formatToLocal(
    timestamp: string,
    format: 'short' | 'medium' | 'long' | 'full' = 'medium'
  ): string {
    const date = this.parseUTCTimestamp(timestamp);

    const optionsMap: Record<string, Intl.DateTimeFormatOptions> = {
      short: { dateStyle: 'short' as const, timeStyle: 'short' as const },
      medium: { dateStyle: 'medium' as const, timeStyle: 'short' as const },
      long: { dateStyle: 'long' as const, timeStyle: 'medium' as const },
      full: { dateStyle: 'full' as const, timeStyle: 'long' as const },
    };

    const options = optionsMap[format];

    return new Intl.DateTimeFormat('default', options).format(date);
  }

  /**
   * Format date only (no time)
   */
  formatDateOnly(timestamp: string): string {
    const date = this.parseUTCTimestamp(timestamp);
    return new Intl.DateTimeFormat('default', { dateStyle: 'medium' }).format(date);
  }

  /**
   * Format time only (no date)
   */
  formatTimeOnly(timestamp: string): string {
    const date = this.parseUTCTimestamp(timestamp);
    return new Intl.DateTimeFormat('default', { timeStyle: 'short' }).format(date);
  }

  /**
   * Get relative time (e.g., "2 hours ago", "in 3 days")
   */
  getRelativeTime(timestamp: string): string {
    const date = this.parseUTCTimestamp(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffSec / 60);
    const diffHour = Math.floor(diffMin / 60);
    const diffDay = Math.floor(diffHour / 24);

    if (diffSec < 60) return 'just now';
    if (diffMin < 60) return `${diffMin} minute${diffMin > 1 ? 's' : ''} ago`;
    if (diffHour < 24) return `${diffHour} hour${diffHour > 1 ? 's' : ''} ago`;
    if (diffDay < 30) return `${diffDay} day${diffDay > 1 ? 's' : ''} ago`;

    return this.formatDateOnly(timestamp);
  }

  /**
   * Check if timestamp is today
   */
  isToday(timestamp: string): boolean {
    const date = this.parseUTCTimestamp(timestamp);
    const today = new Date();
    return date.toDateString() === today.toDateString();
  }

  /**
   * Check if timestamp is in the past
   */
  isPast(timestamp: string): boolean {
    const date = this.parseUTCTimestamp(timestamp);
    return date < new Date();
  }

  /**
   * Check if timestamp is in the future
   */
  isFuture(timestamp: string): boolean {
    const date = this.parseUTCTimestamp(timestamp);
    return date > new Date();
  }

  /**
   * Get days between two timestamps
   */
  daysBetween(start: string, end: string): number {
    const startDate = this.parseUTCTimestamp(start);
    const endDate = this.parseUTCTimestamp(end);
    const diffMs = endDate.getTime() - startDate.getTime();
    return Math.floor(diffMs / (1000 * 60 * 60 * 24));
  }

  /**
   * Add days to a timestamp
   */
  addDays(timestamp: string, days: number): Date {
    const date = this.parseUTCTimestamp(timestamp);
    date.setDate(date.getDate() + days);
    return date;
  }

  /**
   * Format for date input (YYYY-MM-DD)
   */
  toDateInputFormat(date: Date): string {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  }

  /**
   * Parse date input format to Date
   */
  fromDateInputFormat(dateString: string): Date {
    return new Date(dateString + 'T00:00:00Z');
  }
}
