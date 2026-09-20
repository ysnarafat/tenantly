export function maskNid(value: string | null | undefined): string {
  if (!value || value.trim() === '') return '—';
  if (value.length <= 4) return value;
  return '•'.repeat(value.length - 4) + value.slice(-4);
}

// maskFromLastFour renders a masked NID when only the last four digits are known
// (the full value is no longer sent to the client by default). The exact length
// is unknown, so a fixed-width bullet prefix is used.
export function maskFromLastFour(lastFour: string | null | undefined): string {
  if (!lastFour || lastFour.trim() === '') return '—';
  return '••••••' + lastFour;
}

export function maskPhone(value: string | null | undefined): string {
  if (!value || value.trim() === '') return '—';
  const str = value.trim();
  if (str.length <= 7) return '•'.repeat(Math.max(0, str.length - 4)) + str.slice(-4);
  return str.slice(0, 3) + '•'.repeat(str.length - 7) + str.slice(-4);
}

export const RESTRICTED_LABEL = '[Restricted]';
