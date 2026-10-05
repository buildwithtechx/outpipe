import { completeShareUrl, encryptShare } from '@outpipe/share-crypto';
import { useState } from 'react';
import { createShareApi } from '../lib/api';

export function ShareComposer() {
  const [content, setContent] = useState('');
  const [ttl, setTtl] = useState(86400);
  const [maxViews, setMaxViews] = useState(1);
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [shareUrl, setShareUrl] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!content.trim()) {
      setError('Please enter a secret to share.');
      return;
    }
    if (password && password.length < 8) {
      setError('Optional password must be at least 8 characters long.');
      return;
    }

    try {
      setLoading(true);
      setError(null);
      const encrypted = await encryptShare(
        { type: 'text', text: content },
        password.length > 0 ? password : undefined,
      );

      const response = await createShareApi({
        ciphertext: encrypted.ciphertext,
        iv: encrypted.iv,
        keyVerifier: encrypted.verifier,
        passwordSalt: encrypted.passwordSalt,
        passwordVerifier: encrypted.passwordVerifier,
        contentFormat: 'text',
        ttlSeconds: ttl,
        maxViews: maxViews,
      });

      const url = completeShareUrl(
        window.location.origin,
        response.id,
        encrypted.key,
      );
      setShareUrl(url);
    } catch (err: unknown) {
      setError(
        err instanceof Error ? err.message : 'Failed to create share link.',
      );
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = async () => {
    if (!shareUrl) return;
    try {
      await navigator.clipboard.writeText(shareUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2500);
    } catch {
      setError('Failed to copy to clipboard.');
    }
  };

  const handleReset = () => {
    setContent('');
    setPassword('');
    setShareUrl(null);
    setError(null);
  };

  if (shareUrl) {
    return (
      <div className="card">
        <h2 className="card-title">Secret Link Ready</h2>
        <p className="card-subtitle">
          Share this link with your recipient. The decryption key exists only in
          the hash fragment.
        </p>

        <div className="alert-banner">
          <svg
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            role="img"
          >
            <title>Information</title>
            <circle cx="12" cy="12" r="10" />
            <line x1="12" y1="8" x2="12" y2="12" />
            <line x1="12" y1="16" x2="12.01" y2="16" />
          </svg>
          <div>
            The secret key is stored in the <strong>#hash</strong> portion of
            the URL. Our servers never receive or store the key.
          </div>
        </div>

        <div className="form-group">
          <label htmlFor="share-result-url" className="form-label">
            Shareable Link
          </label>
          <input
            id="share-result-url"
            type="text"
            readOnly
            value={shareUrl}
            className="form-input"
            onClick={(e) => (e.target as HTMLInputElement).select()}
          />
        </div>

        <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1.5rem' }}>
          <button
            type="button"
            className="btn-primary"
            onClick={handleCopy}
            style={{ flex: 1 }}
          >
            {copied ? 'Copied to Clipboard!' : 'Copy Link'}
          </button>
          <button type="button" className="btn-secondary" onClick={handleReset}>
            New Secret
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="card">
      <h1 className="card-title">Share a Secret</h1>
      <p className="card-subtitle">
        Encrypt sensitive information in your browser before sending. Only the
        link recipient can decrypt it.
      </p>

      {error && (
        <div className="alert-banner alert-danger">
          <span>{error}</span>
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="secret-content" className="form-label">
            Secret Content
          </label>
          <textarea
            id="secret-content"
            className="form-textarea"
            placeholder="Paste password, private key, API token, or sensitive notes..."
            value={content}
            onChange={(e) => setContent(e.target.value)}
            disabled={loading}
            required
          />
        </div>

        <div className="grid-2">
          <div className="form-group">
            <label htmlFor="secret-ttl" className="form-label">
              Lifetime
            </label>
            <select
              id="secret-ttl"
              className="form-select"
              value={ttl}
              onChange={(e) => setTtl(Number(e.target.value))}
              disabled={loading}
            >
              <option value={3600}>1 hour</option>
              <option value={86400}>24 hours (1 day)</option>
              <option value={604800}>7 days</option>
            </select>
          </div>

          <div className="form-group">
            <label htmlFor="secret-views" className="form-label">
              View Limit
            </label>
            <select
              id="secret-views"
              className="form-select"
              value={maxViews}
              onChange={(e) => setMaxViews(Number(e.target.value))}
              disabled={loading}
            >
              <option value={1}>1 view (burn on read)</option>
              <option value={5}>5 views</option>
              <option value={10}>10 views</option>
              <option value={50}>50 views</option>
            </select>
          </div>
        </div>

        <div className="form-group">
          <label htmlFor="secret-password" className="form-label">
            Optional Passphrase (minimum 8 characters)
          </label>
          <input
            id="secret-password"
            type="password"
            className="form-input"
            placeholder="Additional password required to decrypt..."
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            disabled={loading}
            autoComplete="new-password"
          />
        </div>

        <div style={{ marginTop: '1.75rem' }}>
          <button
            type="submit"
            className="btn-primary"
            disabled={loading || !content.trim()}
          >
            {loading
              ? 'Encrypting & Generating Link...'
              : 'Create Encrypted Link'}
          </button>
        </div>
      </form>
    </div>
  );
}
