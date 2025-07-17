import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import '@testing-library/jest-dom';
import { MemoryRouter } from 'react-router-dom';
import { Layout } from './components/layout/Layout';

describe('App Layout', () => {
  it('renders layout without crashing', () => {
    render(
      <MemoryRouter>
        <Layout><div>Test Content</div></Layout>
      </MemoryRouter>
    );
    expect(screen.getByRole('navigation')).toBeInTheDocument();
  });

  it('renders header navigation', () => {
    render(
      <MemoryRouter>
        <Layout><div>Test Content</div></Layout>
      </MemoryRouter>
    );
    expect(screen.getByText('header.title')).toBeInTheDocument();
    expect(screen.getByText('header.dashboard')).toBeInTheDocument();
  });

  it('renders language selector', () => {
    render(
      <MemoryRouter>
        <Layout><div>Test Content</div></Layout>
      </MemoryRouter>
    );
    expect(screen.getByText('EN')).toBeInTheDocument();
    expect(screen.getByText('English')).toBeInTheDocument();
  });

  it('has proper navigation structure', () => {
    render(
      <MemoryRouter>
        <Layout><div>Test Content</div></Layout>
      </MemoryRouter>
    );
    const nav = screen.getByRole('navigation');
    expect(nav).toHaveClass('bg-white', 'shadow-sm');
  });

  it('renders main content area', () => {
    render(
      <MemoryRouter>
        <Layout><div>Test Content</div></Layout>
      </MemoryRouter>
    );
    const main = screen.getByRole('main');
    expect(main).toHaveClass('flex-1');
  });
});