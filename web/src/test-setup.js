import '@testing-library/jest-dom';

// jsdom's localStorage stub is incomplete - replace it with a full implementation
const store = {};
const localStorageMock = {
  getItem:    (k)    => store[k] ?? null,
  setItem:    (k, v) => { store[k] = String(v); },
  removeItem: (k)    => { delete store[k]; },
  clear:      ()     => { Object.keys(store).forEach(k => delete store[k]); },
  get length()       { return Object.keys(store).length; },
};
Object.defineProperty(globalThis, 'localStorage', { value: localStorageMock, writable: true });
