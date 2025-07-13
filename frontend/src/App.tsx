import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';

import { DashboardPage } from './pages/DashboardPage';
import { URLDetailsPage } from './pages/URLDetailsPage';

import { Layout } from './components/layout/Layout';
import ErrorBoundary from './components/ErrorBoundary';

function App() {
  return (
    <ErrorBoundary>
      <Router>
        <div className="App">
          <Toaster position="top-right" />
          
          <Routes>
            <Route path="/" element={
              <Layout>
                <ErrorBoundary>
                  <DashboardPage />
                </ErrorBoundary>
              </Layout>
            } />
            
            <Route path="/dashboard" element={<Navigate to="/" replace />} />
            
            <Route path="/url/:id" element={
              <Layout>
                <ErrorBoundary>
                  <URLDetailsPage />
                </ErrorBoundary>
              </Layout>
            } />
            
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </div>
      </Router>
    </ErrorBoundary>
  );
}

export default App; 