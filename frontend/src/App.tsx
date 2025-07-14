import React, { Suspense } from 'react';
import { Toaster } from 'react-hot-toast';
import {
  Navigate,
  Route,
  BrowserRouter as Router,
  Routes,
} from 'react-router-dom';

import ErrorBoundary from './components/ErrorBoundary';
import { Layout } from './components/layout/Layout';
import OfflineIndicator from './components/ui/OfflineIndicator';

const DashboardPage = React.lazy(() => import('./pages/DashboardPage'));
const URLDetailsPage = React.lazy(() => import('./pages/URLDetailsPage'));

const PageLoadingSpinner = () => (
  <div className='min-h-screen bg-gray-50 flex items-center justify-center'>
    <div className='text-center'>
      <div className='animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto'></div>
      <p className='mt-4 text-gray-600'>Loading...</p>
    </div>
  </div>
);

const LazyLoadErrorBoundary: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => (
  <ErrorBoundary>
    <Suspense fallback={<PageLoadingSpinner />}>{children}</Suspense>
  </ErrorBoundary>
);

function App() {
  return (
    <ErrorBoundary>
      <OfflineIndicator />

      <Router>
        <div className='App'>
          <Toaster position='top-right' />

          <Routes>
            <Route
              path='/'
              element={
                <Layout>
                  <LazyLoadErrorBoundary>
                    <DashboardPage />
                  </LazyLoadErrorBoundary>
                </Layout>
              }
            />

            <Route path='/dashboard' element={<Navigate to='/' replace />} />

            <Route
              path='/url/:id'
              element={
                <Layout>
                  <LazyLoadErrorBoundary>
                    <URLDetailsPage />
                  </LazyLoadErrorBoundary>
                </Layout>
              }
            />

            <Route path='*' element={<Navigate to='/' replace />} />
          </Routes>
        </div>
      </Router>
    </ErrorBoundary>
  );
}

export default App;
