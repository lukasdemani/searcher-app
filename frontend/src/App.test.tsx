import React from 'react';
import { render, screen } from '@testing-library/react';
import App from './App';

describe('App Component', () => {
  test('renders Website Analyzer title', () => {
    render(<App />);
    const titleElement = screen.getByText(/Website Analyzer/i);
    expect(titleElement).toBeInTheDocument();
  });

  test('renders setup complete message', () => {
    render(<App />);
    const setupMessage = screen.getByText(/Setup Complete!/i);
    expect(setupMessage).toBeInTheDocument();
  });

  test('renders all three status cards', () => {
    render(<App />);
    
    const frontendCard = screen.getByText(/Frontend ✅/i);
    const backendCard = screen.getByText(/Backend 🔧/i);
    const databaseCard = screen.getByText(/Database ✅/i);
    
    expect(frontendCard).toBeInTheDocument();
    expect(backendCard).toBeInTheDocument();
    expect(databaseCard).toBeInTheDocument();
  });

  test('has proper responsive grid layout', () => {
    render(<App />);
    const gridContainer = screen.getByText(/Frontend ✅/i).closest('.grid');
    
    expect(gridContainer).toHaveClass('grid-cols-1');
    expect(gridContainer).toHaveClass('md:grid-cols-3');
  });

  test('applies correct styling classes', () => {
    const { container } = render(<App />);
    const mainDiv = container.firstChild as HTMLElement;
    
    expect(mainDiv).toHaveClass('min-h-screen');
    expect(mainDiv).toHaveClass('bg-gray-50');
  });
}); 