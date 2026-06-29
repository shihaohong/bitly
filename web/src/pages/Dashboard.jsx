import { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api';

function CopyIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <rect x="9" y="9" width="13" height="13" rx="2" />
      <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
    </svg>
  );
}

function CheckIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="20 6 9 17 4 12" />
    </svg>
  );
}

function TrashIcon() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <polyline points="3 6 5 6 21 6" />
      <path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6" />
      <path d="M10 11v6M14 11v6" />
      <path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2" />
    </svg>
  );
}

function Toast({ message }) {
  return <div className={`toast ${message ? 'show' : ''}`}>{message}</div>;
}

function LinkRow({ link, onDelete }) {
  const [copied, setCopied] = useState(false);
  const shortUrl = `${window.location.origin}/${link.short_code}`;

  function copy() {
    navigator.clipboard.writeText(shortUrl).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  }

  return (
    <tr>
      <td>
        <div className="short-url-cell">
          <a className="short-link" href={shortUrl} target="_blank" rel="noreferrer">
            /{link.short_code}
          </a>
          <button className={`btn-icon ${copied ? 'copied' : ''}`} onClick={copy} title="Copy">
            {copied ? <CheckIcon /> : <CopyIcon />}
          </button>
        </div>
      </td>
      <td>
        <div className="original-url" title={link.original_url}>{link.original_url}</div>
      </td>
      <td className="clicks">{link.click_count}</td>
      <td className="date">{new Date(link.created_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })}</td>
      <td>
        <button className="btn-icon danger" onClick={() => onDelete(link.short_code)} title="Delete">
          <TrashIcon />
        </button>
      </td>
    </tr>
  );
}

export default function Dashboard() {
  const navigate = useNavigate();
  const [links, setLinks] = useState([]);
  const [url, setUrl] = useState('');
  const [createError, setCreateError] = useState('');
  const [creating, setCreating] = useState(false);
  const [toast, setToast] = useState('');
  const toastTimer = useRef(null);

  function showToast(msg) {
    setToast(msg);
    clearTimeout(toastTimer.current);
    toastTimer.current = setTimeout(() => setToast(''), 2200);
  }

  const loadLinks = useCallback(async () => {
    const res = await api.getLinks();
    if (res && res.ok) setLinks(await res.json());
  }, []);

  useEffect(() => { loadLinks(); }, [loadLinks]);

  function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('email');
    navigate('/login', { replace: true });
  }

  async function handleCreate(e) {
    e.preventDefault();
    setCreateError('');
    if (!url.trim()) { setCreateError('Please enter a URL'); return; }
    setCreating(true);
    try {
      const res = await api.createLink(url.trim());
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed to create link');
      setLinks(prev => [data, ...prev]);
      setUrl('');
      showToast('Link created!');
    } catch (err) {
      setCreateError(err.message);
    } finally {
      setCreating(false);
    }
  }

  async function handleDelete(code) {
    const res = await api.deleteLink(code);
    if (res && res.ok) {
      setLinks(prev => prev.filter(l => l.short_code !== code));
      showToast('Link deleted');
    }
  }

  return (
    <>
      <nav>
        <span className="nav-logo">Bitly</span>
        <div className="nav-right">
          <span className="nav-email">{localStorage.getItem('email')}</span>
          <button className="btn-logout" onClick={logout}>Log out</button>
        </div>
      </nav>

      <main>
        <div className="card">
          <div className="card-title">Create new link</div>
          <form onSubmit={handleCreate}>
            <div className="create-row">
              <input
                className="url-input"
                type="url"
                placeholder="https://example.com/your-long-url"
                value={url}
                onChange={e => setUrl(e.target.value)}
                autoComplete="off"
              />
              <button className="btn-create" type="submit" disabled={creating}>
                {creating ? 'Shortening…' : 'Shorten'}
              </button>
            </div>
          </form>
          {createError && <p className="create-error">{createError}</p>}
        </div>

        <div className="section-header">
          <span className="section-title">Your links</span>
          <span className="count-badge">{links.length}</span>
        </div>

        <div className="card card-flush">
          {links.length === 0 ? (
            <div className="empty-state">
              <div className="empty-icon">🔗</div>
              <p>No links yet. Create your first one above!</p>
            </div>
          ) : (
            <table className="links-table">
              <thead>
                <tr>
                  <th>Short URL</th>
                  <th>Original URL</th>
                  <th>Clicks</th>
                  <th>Created</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {links.map(link => (
                  <LinkRow key={link.id} link={link} onDelete={handleDelete} />
                ))}
              </tbody>
            </table>
          )}
        </div>
      </main>

      <Toast message={toast} />
    </>
  );
}
