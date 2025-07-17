import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import StatusBadge from './StatusBadge';

describe('StatusBadge Component', () => {
  test('renders queued status correctly', () => {
    render(<StatusBadge status="queued" />);
    expect(screen.getByText('Queued')).toBeInTheDocument();
    expect(screen.getByText('Queued').parentElement).toHaveClass('bg-gray-100', 'text-gray-800');
  });

  test('renders processing status correctly', () => {
    render(<StatusBadge status="processing" />);
    expect(screen.getByText('Processing')).toBeInTheDocument();
    expect(screen.getByText('Processing').parentElement).toHaveClass('bg-blue-100', 'text-blue-800');
  });

  test('renders completed status correctly', () => {
    render(<StatusBadge status="completed" />);
    expect(screen.getByText('Completed')).toBeInTheDocument();
    expect(screen.getByText('Completed').parentElement).toHaveClass('bg-green-100', 'text-green-800');
  });

  test('renders error status correctly', () => {
    render(<StatusBadge status="error" />);
    expect(screen.getByText('Error')).toBeInTheDocument();
    expect(screen.getByText('Error').parentElement).toHaveClass('bg-red-100', 'text-red-800');
  });

  test('shows spinner icon for processing status', () => {
    render(<StatusBadge status="processing" />);
    const container = screen.getByText('Processing').parentElement;
    const spinner = container?.querySelector('.animate-spin');
    expect(spinner).toBeInTheDocument();
  });

  test('applies correct base classes', () => {
    render(<StatusBadge status="completed" />);
    const badge = screen.getByText('Completed').parentElement;
    expect(badge).toHaveClass(
      'inline-flex',
      'items-center',
      'rounded-full',
      'font-medium'
    );
  });
}); 
