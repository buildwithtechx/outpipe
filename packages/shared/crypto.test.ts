import { describe, expect, it } from 'vitest';
import {
  completeShareUrl,
  decryptShare,
  encryptShare,
  shareNeedsPassword,
  sharePasswordVerifier,
  shareVerifier,
} from './crypto';

describe('shared secret cryptography', () => {
  it('encrypts a bundle without exposing its values in stored fields', async () => {
    const content = {
      type: 'bundle' as const,
      entries: [{ key: 'API_KEY', value: 'super-secret' }],
    };
    const encrypted = await encryptShare(content);
    expect(encrypted.key.length).toBe(43);
    expect(encrypted.verifier).toBe(await shareVerifier(encrypted.key));
    expect(
      JSON.stringify({
        ciphertext: encrypted.ciphertext,
        iv: encrypted.iv,
        verifier: encrypted.verifier,
      }),
    ).not.toContain('super-secret');
    expect(await decryptShare(encrypted, encrypted.key)).toEqual(content);
    expect(
      completeShareUrl(
        'https://secrets.outpipe.dev',
        'a'.repeat(22),
        encrypted.key,
      ),
    ).toMatch(/#.+$/);
    await expect(
      decryptShare(encrypted, (await encryptShare(content)).key),
    ).rejects.toThrow();
  });

  it('encrypted shares are snapshots, not references to later source edits', async () => {
    const entry = { key: 'TOKEN', value: 'first' };
    const source = {
      type: 'bundle' as const,
      entries: [entry],
    };
    const encrypted = await encryptShare(source);
    entry.value = 'second';
    expect(await decryptShare(encrypted, encrypted.key)).toEqual({
      type: 'bundle',
      entries: [{ key: 'TOKEN', value: 'first' }],
    });
  });

  it('a password-protected share requires the fragment and password', async () => {
    const content = { type: 'text' as const, text: 'private note' };
    const encrypted = await encryptShare(content, 'a separate password');
    expect(shareNeedsPassword(encrypted.key)).toBe(true);
    expect(await shareVerifier(encrypted.key)).toBe(encrypted.verifier);
    expect(
      await sharePasswordVerifier(encrypted.key, 'a separate password'),
    ).toBe(encrypted.passwordVerifier);
    expect(
      await sharePasswordVerifier(encrypted.key, 'wrong password'),
    ).not.toBe(encrypted.passwordVerifier);
    expect(
      await decryptShare(encrypted, encrypted.key, 'a separate password'),
    ).toEqual(content);
    await expect(decryptShare(encrypted, encrypted.key)).rejects.toThrow(
      /password/i,
    );
    await expect(
      decryptShare(encrypted, encrypted.key, 'wrong password'),
    ).rejects.toThrow();
    expect(
      JSON.stringify({
        ciphertext: encrypted.ciphertext,
        salt: encrypted.passwordSalt,
        proof: encrypted.passwordVerifier,
      }),
    ).not.toContain('private note');
    expect(JSON.stringify(encrypted)).not.toContain('a separate password');
    await expect(encryptShare(content, 'short')).rejects.toThrow(/password/i);
  });

  it('rejects too many named entries and oversized plaintext before encryption', async () => {
    await expect(
      encryptShare({
        type: 'bundle',
        entries: Array.from({ length: 51 }, (_, index) => ({
          key: `K${index}`,
          value: 'v',
        })),
      }),
    ).rejects.toThrow();
    await expect(
      encryptShare({ type: 'text', text: 'x'.repeat(256 * 1024 + 1) }),
    ).rejects.toThrow();
  });
});
