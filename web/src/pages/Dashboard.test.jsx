import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import Dashboard from './Dashboard';
import * as apiModule from '../api';

const mockNavigate = vi.fn();
vi.mock('react-router-dom', async importOriginal => ({
  ...(await importOriginal()),
  useNavigate: () => mockNavigate,
}));
vi.mock('../api');

const LINKS = [
  { id: '1', short_code: 'abc1234', original_url: 'https://example.com', click_count: 5, created_at: '2026-01-01T00:00:00Z' },
  { id: '2', short_code: 'xyz5678', original_url: 'https://anthropic.com', click_count: 0, created_at: '2026-02-01T00:00:00Z' },
];

function renderDashboard() {
  localStorage.setItem('token', 'tok');
  localStorage.setItem('email', 'user@example.com');
  return render(<MemoryRouter><Dashboard /></MemoryRouter>);
}

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  apiModule.api.getLinks = vi.fn().mockResolvedValue({ ok: true, json: async () => [...LINKS] });
  apiModule.api.createLink = vi.fn();
  apiModule.api.deleteLink = vi.fn();
});

describe('Dashboard', () => {
  it('shows the user email and logout button', async () => {
    renderDashboard();
    expect(screen.getByText('user@example.com')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /log out/i })).toBeInTheDocument();
  });

  it('loads and displays links', async () => {
    renderDashboard();
    await waitFor(() => {
      expect(screen.getByText('/abc1234')).toBeInTheDocument();
      expect(screen.getByText('/xyz5678')).toBeInTheDocument();
    });
    expect(screen.getByText('https://example.com')).toBeInTheDocument();
    expect(screen.getByText('5')).toBeInTheDocument();
  });

  it('shows empty state when there are no links', async () => {
    apiModule.api.getLinks = vi.fn().mockResolvedValue({ ok: true, json: async () => [] });
    renderDashboard();
    await waitFor(() => {
      expect(screen.getByText(/no links yet/i)).toBeInTheDocument();
    });
  });

  it('shows the count badge', async () => {
    renderDashboard();
    await waitFor(() => {
      expect(screen.getByText('2')).toBeInTheDocument();
    });
  });

  it('creates a link and prepends it to the list', async () => {
    const newLink = { id: '3', short_code: 'new1111', original_url: 'https://new.com', click_count: 0, created_at: '2026-03-01T00:00:00Z' };
    apiModule.api.createLink = vi.fn().mockResolvedValue({ ok: true, json: async () => newLink });
    const user = userEvent.setup();
    renderDashboard();

    await waitFor(() => screen.getByText('/abc1234'));

    await user.type(screen.getByPlaceholderText(/https:\/\/example.com/i), 'https://new.com');
    await user.click(screen.getByRole('button', { name: /shorten/i }));

    await waitFor(() => {
      expect(screen.getByText('/new1111')).toBeInTheDocument();
    });
    expect(apiModule.api.createLink).toHaveBeenCalledWith('https://new.com');
    expect(screen.getByText('3')).toBeInTheDocument();
  });

  it('shows validation error when URL is empty', async () => {
    const user = userEvent.setup();
    renderDashboard();
    await waitFor(() => screen.getByText('/abc1234'));

    await user.click(screen.getByRole('button', { name: /shorten/i }));

    expect(screen.getByText(/please enter a url/i)).toBeInTheDocument();
    expect(apiModule.api.createLink).not.toHaveBeenCalled();
  });

  it('removes a link after deletion', async () => {
    apiModule.api.deleteLink = vi.fn().mockResolvedValue({ ok: true });
    const user = userEvent.setup();
    renderDashboard();

    await waitFor(() => screen.getByText('/abc1234'));

    const rows = screen.getAllByRole('row');
    const firstDataRow = rows[1];
    const deleteBtn = within(firstDataRow).getByTitle(/delete/i);
    await user.click(deleteBtn);

    await waitFor(() => {
      expect(screen.queryByText('/abc1234')).not.toBeInTheDocument();
    });
    expect(apiModule.api.deleteLink).toHaveBeenCalledWith('abc1234');
  });

  it('clears token and navigates to login on logout', async () => {
    const user = userEvent.setup();
    renderDashboard();

    await user.click(screen.getByRole('button', { name: /log out/i }));

    expect(localStorage.getItem('token')).toBeNull();
    expect(mockNavigate).toHaveBeenCalledWith('/login', { replace: true });
  });
});
