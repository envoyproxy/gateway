import React, { useEffect, useRef } from 'react';
import ReactDOM from 'react-dom/client';
import { EmbeddableBenchmarkDashboard } from './EmbeddableBenchmarkDashboard';
import { CSSInjector } from '../utils/cssInjector';
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { TooltipProvider } from "@/components/ui/tooltip";

interface ShadowDOMWrapperProps {
  apiBase?: string;
  initialVersion?: string;
  theme?: 'light' | 'dark';
  containerClassName?: string;
  containerId?: string;
  features?: {
    header?: boolean;
    versionSelector?: boolean;
    summaryCards?: boolean;
    tabs?: string[];
  };
}

export const ShadowDOMWrapper: React.FC<ShadowDOMWrapperProps> = (props) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const shadowRootRef = useRef<ShadowRoot | null>(null);
  const reactRootRef = useRef<any>(null);

  useEffect(() => {
    if (!containerRef.current) return;

    try {
      // Create shadow DOM
      const shadowRoot = containerRef.current.attachShadow({ mode: 'open' });
      shadowRootRef.current = shadowRoot;

      // Create container div in shadow DOM
      const shadowContainer = document.createElement('div');
      shadowContainer.id = 'shadow-root';
      shadowContainer.className = 'benchmark-dashboard';
      shadowRoot.appendChild(shadowContainer);

      // Inject CSS into shadow DOM
      CSSInjector.getAllCSS().then((css) => {
        const style = document.createElement('style');
        style.textContent = css;
        shadowRoot.insertBefore(style, shadowContainer);

        // Create React root and render
        const root = ReactDOM.createRoot(shadowContainer);
        reactRootRef.current = root;

        root.render(
          <React.StrictMode>
            <QueryClientProvider client={new QueryClient()}>
              <TooltipProvider>
                <EmbeddableBenchmarkDashboard {...props} />
              </TooltipProvider>
            </QueryClientProvider>
          </React.StrictMode>
        );
      }).catch((error) => {
        console.error('Failed to inject CSS into Shadow DOM:', error);
      });

    } catch (error) {
      console.error('Failed to create Shadow DOM:', error);
    }

    // Cleanup function
    return () => {
      if (reactRootRef.current) {
        reactRootRef.current.unmount();
      }
    };
  }, []);

  return <div ref={containerRef} style={{ width: '100%', minHeight: '400px' }} />;
};

export default ShadowDOMWrapper;
