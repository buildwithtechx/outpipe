import {
  decryptShare,
  shareNeedsPassword,
  sharePasswordVerifier,
  shareVerifier,
} from '@outpipe/share-crypto';
import { useEffect, useState } from 'react';
import {
  getShareMetaApi,
  revealShareApi,
  type ShareMetaResponse,
} from '../lib/api';

interface ShareRevealProps {
  id?: string;
}

export function ShareReveal({ id: propId }: ShareRevealProps) {
  const [meta, setMeta] = useState<ShareMetaResponse | null>(null);
  const [keyText, setKeyText] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(true);
  const [revealing, setRevealing] = useState(false);
  const [revealedText, setRevealedText] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [resolvedId, setResolvedId] = useState(propId || '');

  useEffect(() => {
    const hash = window.location.hash.replace(/^#/, '');
    if (!hash) {
      setError('Invalid link: decryption key is missing from URL.');
      setLoading(false);
      return;
    }
    setKeyText(hash);

    const shareId =
      propId ||
      new URLSearchParams(window.location.search).get('id') ||
      window.location.pathname.split('/').filter(Boolean).pop() ||
      '';

    if (!shareId || shareId === 'reveal' || shareId === 's') {
      setError('Secret ID is missing from link.');
      setLoading(false);
      return;
    }

    setResolvedId(shareId);

    getShareMetaApi(shareId)
      .then((data) => {
        setMeta(data);
        setLoading(false);
      })
      .catch((err: unknown) => {
        setError(
          err instanceof Error ? err.message : 'Secret not found or expired.',
        );
        setLoading(false);
      });
  }, [propId]);

  const handleReveal = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!keyText) return;

    try {
      setRevealing(true);
      setError(null);

      const verifier = await shareVerifier(keyText);
      let passwordVerifier: string | undefined;

      const requiresPass = meta?.needsPassword || shareNeedsPassword(keyText);
      if (requiresPass) {
        if (!password) {
          setError('Please enter the password to unlock this secret.');
          setRevealing(false);
          return;
        }
        passwordVerifier = await sharePasswordVerifier(keyText, password);
      }

      const response = await revealShareApi(
        resolvedId,
        verifier,
        passwordVerifier,
      );
      const decrypted = await decryptShare(
        { ciphertext: response.ciphertext, iv: response.iv },
        keyText,
        password || undefined,
      );

      if (decrypted.type === 'text') {
        setRevealedText(decrypted.text);
      } else {
        setRevealedText(JSON.stringify(decrypted.entries, null, 2));
      }
    } catch (err: unknown) {
      setError(
        err instanceof Error ? err.message : 'Failed to decrypt secret.',
      );
    } finally {
      setRevealing(false);
    }
  };

  const handleCopy = async () => {
    if (!revealedText) return;
    try {
      await navigator.clipboard.writeText(revealedText);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      setError('Failed to copy to clipboard.');
    }
  };

  if (loading) {
    return (
      <div
        className="card"
        style={{ textAlign: 'center', padding: '3rem 1.5rem' }}
      >
        <p style={{ color: 'var(--text-secondary)' }}>
          Loading secret metadata...
        </p>
      </div>
    );
  }

  if (error && !revealedText) {
    return (
      <div className="card">
        <h2 className="card-title">Secret Unavailable</h2>
        <div
          className="alert-banner alert-danger"
          style={{ marginTop: '1rem' }}
        >
          <span>{error}</span>
        </div>
        <div style={{ marginTop: '1.5rem' }}>
          <a
            href="/"
            className="btn-secondary"
            style={{ display: 'inline-flex' }}
          >
            Create a New Secret
          </a>
        </div>
      </div>
    );
  }

  if (revealedText !== null) {
    return (
      <div className="card">
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            marginBottom: '0.75rem',
          }}
        >
          <h2 className="card-title" style={{ margin: 0 }}>
            Decrypted Secret
          </h2>
          <span className="badge-tag">Decrypted in Browser</span>
        </div>
        <p className="card-subtitle">
          Make sure to copy or save this now. You may not be able to view it
          again.
        </p>

        <div className="code-box">{revealedText}</div>

        <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1.5rem' }}>
          <button
            type="button"
            className="btn-primary"
            onClick={handleCopy}
            style={{ flex: 1 }}
          >
            {copied ? 'Copied to Clipboard!' : 'Copy Secret'}
          </button>
          <a
            href="/"
            className="btn-secondary"
            style={{ textDecoration: 'none' }}
          >
            Share Another
          </a>
        </div>
      </div>
    );
  }

  const requiresPassword =
    meta?.needsPassword || (keyText && shareNeedsPassword(keyText));

  return (
    <div className="card">
      <h1 className="card-title">Reveal Secret</h1>
      <p className="card-subtitle">
        You have received an end-to-end encrypted secret.
      </p>

      {meta && (
        <div className="alert-banner">
          <div>
            {meta.maxViews === 1 ? (
              <strong>
                Warning: This secret will burn immediately after being viewed.
              </strong>
            ) : (
              <span>
                Views remaining: <strong>{meta.viewsRemaining}</strong> of{' '}
                {meta.maxViews}
              </span>
            )}
          </div>
        </div>
      )}

      {error && (
        <div className="alert-banner alert-danger">
          <span>{error}</span>
        </div>
      )}

      <form onSubmit={handleReveal}>
        {requiresPassword && (
          <div className="form-group">
            <label htmlFor="reveal-password" className="form-label">
              Passphrase Required
            </label>
            <input
              id="reveal-password"
              type="password"
              className="form-input"
              placeholder="Enter the sender's passphrase..."
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={revealing}
              required
            />
          </div>
        )}

        <div style={{ marginTop: '1.5rem' }}>
          <button type="submit" className="btn-primary" disabled={revealing}>
            {revealing ? 'Decrypting Secret...' : 'Reveal Secret'}
          </button>
        </div>
      </form>
    </div>
  );
}
