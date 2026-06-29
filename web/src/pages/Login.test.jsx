import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import Login from './Login';
import * as apiModule from '../api';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async importOriginal => ({
  ...(await importOriginal()),
  useNavigate: () => mockNavigate,
}));
vi.mock('../api');

function renderLogin(search = '') {
  return render(
    <MemoryRouter initialEntries={[`/login${search}`]}>
      <Login />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
});

describe('Login', () => {
  it('renders the form', () => {
    renderLogin();
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
  });

  it('shows success banner when ?registered=1 is present', () => {
    renderLogin('?registered=1');
    expect(screen.getByText(/account created/i)).toBeInTheDocument();
  });

  it('redirects to dashboard if already logged in', () => {
    localStorage.setItem('token', 'existing-token');
    renderLogin();
    expect(mockNavigate).toHaveBeenCalledWith('/dashboard', { replace: true });
  });

  it('shows error on failed login', async () => {
    apiModule.api.login = vi.fn().mockResolvedValue({
      ok: false,
      json: async () => ({ error: 'invalid credentials' }),
    });
    const user = userEvent.setup();
    renderLogin();

    await user.type(screen.getByLabelText(/email/i), 'test@example.com');
    await user.type(screen.getByLabelText(/password/i), 'wrongpassword');
    await user.click(screen.getByRole('button', { name: /sign in/i }));

    await waitFor(() => {
      expect(screen.getByText('invalid credentials')).toBeInTheDocument();
    });
  });

  it('stores token and navigates on successful login', async () => {
    apiModule.api.login = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ token: 'jwt-token-abc' }),
    });
    const user = userEvent.setup();
    renderLogin();

    await user.type(screen.getByLabelText(/email/i), 'test@example.com');
    await user.type(screen.getByLabelText(/password/i), 'password123');
    await user.click(screen.getByRole('button', { name: /sign in/i }));

    await waitFor(() => {
      expect(localStorage.getItem('token')).toBe('jwt-token-abc');
      expect(localStorage.getItem('email')).toBe('test@example.com');
      expect(mockNavigate).toHaveBeenCalledWith('/dashboard', { replace: true });
    });
  });

  it('disables the button while submitting', async () => {
    let resolve;
    apiModule.api.login = vi.fn().mockReturnValue(new Promise(r => { resolve = r; }));
    const user = userEvent.setup();
    renderLogin();

    await user.type(screen.getByLabelText(/email/i), 'test@example.com');
    await user.type(screen.getByLabelText(/password/i), 'password123');
    await user.click(screen.getByRole('button', { name: /sign in/i }));

    expect(screen.getByRole('button', { name: /signing in/i })).toBeDisabled();
    resolve({ ok: true, json: async () => ({ token: 'tok' }) });
  });
});
