import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import Register from './Register';
import * as apiModule from '../api';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async importOriginal => ({
  ...(await importOriginal()),
  useNavigate: () => mockNavigate,
}));
vi.mock('../api');

function renderRegister() {
  return render(
    <MemoryRouter><Register /></MemoryRouter>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
});

describe('Register', () => {
  it('renders all fields', () => {
    renderRegister();
    expect(screen.getByLabelText(/^email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/^password/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/confirm/i)).toBeInTheDocument();
  });

  it('redirects to dashboard if already logged in', () => {
    localStorage.setItem('token', 'tok');
    renderRegister();
    expect(mockNavigate).toHaveBeenCalledWith('/dashboard', { replace: true });
  });

  it('shows error when passwords do not match', async () => {
    const user = userEvent.setup();
    renderRegister();

    await user.type(screen.getByLabelText(/^email/i), 'a@b.com');
    await user.type(screen.getByLabelText(/^password/i), 'password1');
    await user.type(screen.getByLabelText(/confirm/i), 'password2');
    await user.click(screen.getByRole('button', { name: /create account/i }));

    expect(screen.getByText(/passwords do not match/i)).toBeInTheDocument();
    expect(apiModule.api.register).not.toHaveBeenCalled();
  });

  it('shows API error on failed registration', async () => {
    apiModule.api.register = vi.fn().mockResolvedValue({
      ok: false,
      json: async () => ({ error: 'email already registered' }),
    });
    const user = userEvent.setup();
    renderRegister();

    await user.type(screen.getByLabelText(/^email/i), 'a@b.com');
    await user.type(screen.getByLabelText(/^password/i), 'password123');
    await user.type(screen.getByLabelText(/confirm/i), 'password123');
    await user.click(screen.getByRole('button', { name: /create account/i }));

    await waitFor(() => {
      expect(screen.getByText('email already registered')).toBeInTheDocument();
    });
  });

  it('navigates to login with ?registered=1 on success', async () => {
    apiModule.api.register = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ message: 'account created' }),
    });
    const user = userEvent.setup();
    renderRegister();

    await user.type(screen.getByLabelText(/^email/i), 'new@user.com');
    await user.type(screen.getByLabelText(/^password/i), 'password123');
    await user.type(screen.getByLabelText(/confirm/i), 'password123');
    await user.click(screen.getByRole('button', { name: /create account/i }));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/login?registered=1', { replace: true });
    });
  });
});
