export function maskNid(value: string | null | undefined): string {
  if (!value || value.trim() === '') return '—';
  if (value.length <= 4) return value;
  return '•'.repeat(value.length - 4) + value.slice(-4);
}

export function maskPhone(value: string | null | undefined): string {
  if (!value || value.trim() === '') return '—';
  const str = value.trim();
  if (str.length <= 7) return '•'.repeat(Math.max(0, str.length - 4)) + str.slice(-4);
  return str.slice(0, 3) + '•'.repeat(str.length - 7) + str.slice(-4);
}

export const RESTRICTED_LABEL = '[Restricted]';
