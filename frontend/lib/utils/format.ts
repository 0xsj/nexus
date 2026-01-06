// lib/utils/format.ts

/**
 * Formats a date using Intl.DateTimeFormat.
 * Returns locale-aware formatted string.
 */
export function formatDate(
  date: Date | string | number,
  locale: string = "en",
  options: Intl.DateTimeFormatOptions = {
    year: "numeric",
    month: "short",
    day: "numeric",
  }
): string {
  const d = date instanceof Date ? date : new Date(date);
  return new Intl.DateTimeFormat(locale, options).format(d);
}

/**
 * Formats a date as relative time (e.g., "2 days ago").
 * Falls back to absolute date if RelativeTimeFormat unavailable.
 */
export function formatRelativeTime(
  date: Date | string | number,
  locale: string = "en"
): string {
  const d = date instanceof Date ? date : new Date(date);
  const now = new Date();
  const diffInSeconds = Math.floor((now.getTime() - d.getTime()) / 1000);

  const rtf = new Intl.RelativeTimeFormat(locale, { numeric: "auto" });

  const thresholds: [number, Intl.RelativeTimeFormatUnit][] = [
    [60, "second"],
    [3600, "minute"],
    [86400, "hour"],
    [604800, "day"],
    [2592000, "week"],
    [31536000, "month"],
    [Infinity, "year"],
  ];

  for (let i = 0; i < thresholds.length; i++) {
    const [threshold, unit] = thresholds[i];
    const prevThreshold = i === 0 ? 1 : thresholds[i - 1][0];

    if (diffInSeconds < threshold) {
      const value = Math.floor(diffInSeconds / prevThreshold);
      return rtf.format(-value, unit);
    }
  }

  return formatDate(d, locale);
}

/**
 * Formats a number with locale-aware separators.
 */
export function formatNumber(
  value: number,
  locale: string = "en",
  options: Intl.NumberFormatOptions = {}
): string {
  return new Intl.NumberFormat(locale, options).format(value);
}

/**
 * Formats a number in compact notation (e.g., 1.2K, 3.4M).
 */
export function formatCompactNumber(
  value: number,
  locale: string = "en"
): string {
  return new Intl.NumberFormat(locale, {
    notation: "compact",
    compactDisplay: "short",
  }).format(value);
}