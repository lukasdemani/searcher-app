import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import '@testing-library/jest-dom';
import StatusBadge from './StatusBadge';

describe('StatusBadge Component', () => {
  it('renders queued status correctly', () => {
    render(<StatusBadge status="queued" />);
    expect(screen.getByText('Queued')).toBeInTheDocument();
  });

  it('renders processing status correctly', () => {
    render(<StatusBadge status="processing" />);
    expect(screen.getByText('Processing')).toBeInTheDocument();
  });

  it('renders completed status correctly', () => {
    render(<StatusBadge status="completed" />);
    expect(screen.getByText('Completed')).toBeInTheDocument();
  });

  it('renders error status correctly', () => {
    render(<StatusBadge status="error" />);
    expect(screen.getByText('Error')).toBeInTheDocument();
  });

  it('shows spinner icon for processing status', () => {
    render(<StatusBadge status="processing" />);
    const container = screen.getByText('Processing').parentElement;
    const spinner = container?.querySelector('.animate-spin');
    expect(spinner).toBeInTheDocument();
  });

  it('applies correct base classes', () => {
    render(<StatusBadge status="completed" />);
    const badge = screen.getByText('Completed');
    expect(badge).toHaveClass('px-2.5', 'py-0.5', 'text-xs');
  });
});