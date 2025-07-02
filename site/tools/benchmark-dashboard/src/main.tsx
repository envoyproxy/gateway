import React from 'react';
import ReactDOM from 'react-dom/client';
import { ShadowDOMWrapper } from './components/ShadowDOMWrapper';
// Import the compiled CSS as raw text for Shadow DOM injection
import cssText from '../dist/benchmark-dashboard.css?raw';

// Make CSS available to the CSSInjector
(window as any).__BENCHMARK_CSS__ = cssText;

// Safe Hugo integration function that uses Shadow DOM for complete isolation
function initializeBenchmarkDashboardShadow() {
  try {
    // Only initialize if we're in a browser environment
    if (typeof window === 'undefined') {
      return;
    }

    const mountPoints = document.querySelectorAll('[data-react-component="benchmark-dashboard"]');

    if (mountPoints.length === 0) {
      return;
    }

    mountPoints.forEach((element, index) => {
      try {
        const htmlElement = element as HTMLElement;

        // Skip if already initialized
        if (htmlElement.dataset.reactInitialized === 'true') {
          return;
        }

        // Parse configuration from data attributes
        const config = {
          apiBase: htmlElement.dataset.apiBase,
          initialVersion: htmlElement.dataset.version,
          theme: (htmlElement.dataset.theme as 'light' | 'dark') || 'light',
          containerClassName: htmlElement.dataset.containerClass || '',
          containerId: `benchmark-dashboard-${index}`,
          features: {
            header: htmlElement.dataset.showHeader === 'true',
            versionSelector: htmlElement.dataset.showVersionSelector !== 'false',
            summaryCards: htmlElement.dataset.showSummaryCards !== 'false',
            tabs: htmlElement.dataset.tabs?.split(',').map(t => t.trim()) || ['overview', 'latency', 'resources']
          }
        };

        // Create React root and render with Shadow DOM wrapper
        const root = ReactDOM.createRoot(htmlElement);

        root.render(
          <React.StrictMode>
            <ShadowDOMWrapper {...config} />
          </React.StrictMode>
        );

        // Mark as initialized
        htmlElement.dataset.reactInitialized = 'true';

      } catch (error) {
        console.error(`Error initializing benchmark dashboard ${index + 1}:`, error);
      }
    });
  } catch (error) {
    console.error('Error in benchmark dashboard initialization:', error);
  }
}

// Wait for DOM to be ready
function waitForSafeInitialization() {
  // Wait a bit to ensure Hugo's scripts have loaded
  setTimeout(() => {
    initializeBenchmarkDashboardShadow();
  }, 100);
}

// Multiple initialization strategies
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', waitForSafeInitialization);
} else {
  waitForSafeInitialization();
}

// Also try after window load (more defensive)
window.addEventListener('load', () => {
  setTimeout(initializeBenchmarkDashboardShadow, 200);
});

// Export for manual initialization if needed
(window as Window & { initializeBenchmarkDashboardShadow?: () => void }).initializeBenchmarkDashboardShadow = initializeBenchmarkDashboardShadow;

// Also export for module usage
export { initializeBenchmarkDashboardShadow as initializeHugoShadowDashboard };
