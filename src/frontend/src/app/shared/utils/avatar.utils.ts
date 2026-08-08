export const AVATAR_COLORS = [
  '#1565c0',
  '#2e7d32',
  '#c62828',
  '#6a1b9a',
  '#0277bd',
  '#e65100',
  '#37474f',
  '#00695c',
];

export function avatarInitials(name: string): string {
  return name.slice(0, 2).toUpperCase() || '?';
}

export function avatarColorFor(name: string): string {
  const idx = (name.charCodeAt(0) || 0) % AVATAR_COLORS.length;
  return AVATAR_COLORS[idx];
}
