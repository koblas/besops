export interface HeaderEntry {
  name: string;
  value: string;
}

/**
 * Merges bodyContentType into the headers array as a Content-Type entry.
 * If a Content-Type header already exists, its value is updated.
 * Returns a new array (does not mutate the input).
 */
export function mergeContentType(headers: HeaderEntry[], contentType: string): HeaderEntry[] {
  const result = headers.map(h => ({ ...h }));
  const existing = result.find(h => h.name.toLowerCase() === 'content-type');
  if (existing) {
    existing.value = contentType;
  } else {
    result.push({ name: 'Content-Type', value: contentType });
  }
  return result;
}

/**
 * Extracts the Content-Type value from a headers array.
 * Returns undefined if no Content-Type header is present.
 */
export function extractContentType(headers: HeaderEntry[]): string | undefined {
  const entry = headers.find(h => h.name.toLowerCase() === 'content-type');
  return entry?.value;
}
