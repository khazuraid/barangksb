import React, { Component, ErrorInfo, ReactNode } from 'react';
import { createRoot } from 'react-dom/client';
import App from './App.tsx';
import './index.css';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Unhandled app error:', error, errorInfo);
  }

  public render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen bg-slate-950 text-white flex items-center justify-center p-6 font-sans">
          <div className="bg-slate-900 border border-slate-800 rounded-3xl p-8 max-w-md w-full text-center shadow-2xl">
            <div className="w-12 h-12 rounded-2xl bg-rose-500/10 border border-rose-500/20 text-rose-400 flex items-center justify-center mx-auto mb-4 font-bold text-xl">
              !
            </div>
            <h2 className="text-xl font-bold tracking-tight mb-2">Terjadi Kesalahan Tampilan</h2>
            <p className="text-xs text-slate-400 mb-6 leading-relaxed">
              Sistem telah mengisolasi error agar data Anda tetap aman. Silakan muat ulang halaman.
            </p>
            <button
              onClick={() => {
                this.setState({ hasError: false, error: null });
                window.location.reload();
              }}
              className="w-full py-2.5 px-4 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold rounded-xl transition-all cursor-pointer"
            >
              Muat Ulang Aplikasi
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

// Global safe error catching for cross-origin or deferred scripts
if (typeof window !== 'undefined') {
  const originalOnError = window.onerror;
  window.onerror = function (message, source, lineno, colno, error) {
    if (
      message === 'Script error.' ||
      !message ||
      String(message).toLowerCase().includes('script error')
    ) {
      console.warn('Caught and silenced external script error.');
      return true; // Prevents the error from firing as unhandled
    }
    if (typeof originalOnError === 'function') {
      return (originalOnError as any)(message, source, lineno, colno, error);
    }
    return false;
  };

  window.addEventListener(
    'error',
    (event) => {
      const msg = event?.message ? String(event.message).toLowerCase() : '';
      if (!msg || msg === 'script error.' || msg.includes('script error')) {
        event.preventDefault();
        event.stopImmediatePropagation();
      }
    },
    true
  );

  window.addEventListener('unhandledrejection', (event) => {
    const reasonMsg = event?.reason?.message ? String(event.reason.message).toLowerCase() : String(event?.reason || '').toLowerCase();
    if (
      !reasonMsg ||
      reasonMsg.includes('script error') ||
      reasonMsg.includes('canceled') ||
      reasonMsg.includes('aborted')
    ) {
      event.preventDefault();
      event.stopImmediatePropagation();
    }
  });
}

createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <ErrorBoundary>
      <App />
    </ErrorBoundary>
  </React.StrictMode>
);
