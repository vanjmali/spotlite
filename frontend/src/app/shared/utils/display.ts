export function initialsFromName(name: string, maxChars: number = 2): string {
  const trimmed = name.trim();
  if (!trimmed) {
    return '?';
  }

  const parts = trimmed.split(/\s+/).filter(Boolean);
  if (parts.length === 1) {
    return parts[0].slice(0, maxChars).toUpperCase();
  }

  return parts
    .slice(0, maxChars)
    .map((part) => part.charAt(0).toUpperCase())
    .join('');
}

export function colorFromName(name: string): string {
  const hash = hashFromName(name);
  if (hash === null) {
    return 'linear-gradient(140deg, hsl(280 52% 44%), hsl(210 56% 34%))';
  }

  const hue = hash % 360;
  const secondHue = (hue + 38) % 360;
  return `linear-gradient(140deg, hsl(${hue} 62% 44%), hsl(${secondHue} 58% 34%))`;
}

export function iconFromName(name: string): string {
  const icons = [
    'music_note',
    'album',
    'graphic_eq',
    'equalizer',
    'queue_music',
    'library_music',
    'headphones',
    'audiotrack',
  ];

  const hash = hashFromName(name);
  if (hash === null) {
    return icons[0];
  }

  return icons[hash % icons.length];
}

export function angleFromName(name: string): number {
  const hash = hashFromName(name);
  if (hash === null) {
    return 18;
  }

  return (hash % 44) - 22;
}

function hashFromName(name: string): number | null {
  const input = name.trim().toLowerCase();
  if (!input) {
    return null;
  }

  let hash = 0;
  for (let i = 0; i < input.length; i += 1) {
    hash = (hash << 5) - hash + input.charCodeAt(i);
    hash |= 0;
  }

  return Math.abs(hash);
}
