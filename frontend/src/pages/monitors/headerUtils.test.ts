import { describe, it, expect } from 'vitest';
import { mergeContentType, extractContentType } from './headerUtils';

describe('mergeContentType', () => {
  it('appends Content-Type when no existing header', () => {
    const result = mergeContentType([], 'application/json');
    expect(result).toEqual([{ name: 'Content-Type', value: 'application/json' }]);
  });

  it('appends to existing headers without Content-Type', () => {
    const headers = [{ name: 'Authorization', value: 'Bearer token' }];
    const result = mergeContentType(headers, 'text/plain');
    expect(result).toEqual([
      { name: 'Authorization', value: 'Bearer token' },
      { name: 'Content-Type', value: 'text/plain' },
    ]);
  });

  it('updates existing Content-Type header (exact case)', () => {
    const headers = [
      { name: 'Content-Type', value: 'text/html' },
      { name: 'Accept', value: '*/*' },
    ];
    const result = mergeContentType(headers, 'application/json');
    expect(result).toEqual([
      { name: 'Content-Type', value: 'application/json' },
      { name: 'Accept', value: '*/*' },
    ]);
  });

  it('updates existing Content-Type header (case-insensitive)', () => {
    const headers = [{ name: 'content-type', value: 'old' }];
    const result = mergeContentType(headers, 'new');
    expect(result).toEqual([{ name: 'content-type', value: 'new' }]);
  });

  it('does not mutate the original array', () => {
    const headers = [{ name: 'Content-Type', value: 'old' }];
    const result = mergeContentType(headers, 'new');
    expect(headers[0].value).toBe('old');
    expect(result[0].value).toBe('new');
  });
});

describe('extractContentType', () => {
  it('returns undefined for empty headers', () => {
    expect(extractContentType([])).toBeUndefined();
  });

  it('returns undefined when no Content-Type header exists', () => {
    const headers = [{ name: 'Authorization', value: 'Bearer xyz' }];
    expect(extractContentType(headers)).toBeUndefined();
  });

  it('extracts Content-Type value (exact case)', () => {
    const headers = [
      { name: 'Content-Type', value: 'application/json' },
      { name: 'Accept', value: '*/*' },
    ];
    expect(extractContentType(headers)).toBe('application/json');
  });

  it('extracts Content-Type value (case-insensitive)', () => {
    const headers = [{ name: 'content-type', value: 'text/xml' }];
    expect(extractContentType(headers)).toBe('text/xml');
  });
});
