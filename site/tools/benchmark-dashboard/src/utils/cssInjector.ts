// Utility to handle CSS injection into Shadow DOM
export class CSSInjector {
  private static cssCache: Map<string, string> = new Map();

  /**
   * Get all CSS needed for the shadow DOM
   */
  static async getAllCSS(): Promise<string> {
    const cacheKey = 'shadow-dom-css';

    if (this.cssCache.has(cacheKey)) {
      return this.cssCache.get(cacheKey)!;
    }

    try {
      let css = '';

      // Try to use bundled CSS first (this should always work in shadow DOM build)
      const bundledCSS = this.getBundledCSS();

      if (bundledCSS) {
        // Transform the compiled CSS to work in Shadow DOM by converting :root and .dark to :host selectors
        css = this.transformCSSForShadowDOM(bundledCSS);

        // Add Shadow DOM specific styles
        const finalCSS = `
          ${this.getShadowDOMStyles()}
          ${css}
        `;
        this.cssCache.set(cacheKey, finalCSS);
        return finalCSS;
      } else {
        // Fallback CSS if bundled CSS is not available
        const fallbackCSS = this.getFallbackCSS();
        this.cssCache.set(cacheKey, fallbackCSS);
        return fallbackCSS;
      }
    } catch (error) {
      console.error('Error loading CSS for Shadow DOM:', error);
      return this.getFallbackCSS();
    }
  }

  /**
   * Get bundled CSS that was injected at build time
   */
  private static getBundledCSS(): string | null {
    return (window as any).__BENCHMARK_CSS__ || null;
  }

  /**
   * CSS Custom Properties (CSS Variables) for theming
   */
  private static getCSSVariables(): string {
    return `
      /* CSS Custom Properties for theming */
      :host(.benchmark-dashboard) {
        --background: 0 0% 100%;
        --foreground: 222.2 84% 4.9%;
        --card: 0 0% 100%;
        --card-foreground: 222.2 84% 4.9%;
        --popover: 0 0% 100%;
        --popover-foreground: 222.2 84% 4.9%;
        --primary: 222.2 47.4% 11.2%;
        --primary-foreground: 210 40% 98%;
        --secondary: 210 40% 96.1%;
        --secondary-foreground: 222.2 47.4% 11.2%;
        --muted: 210 40% 96.1%;
        --muted-foreground: 215.4 16.3% 46.9%;
        --accent: 210 40% 96.1%;
        --accent-foreground: 222.2 47.4% 11.2%;
        --destructive: 0 84.2% 60.2%;
        --destructive-foreground: 210 40% 98%;
        --border: 214.3 31.8% 91.4%;
        --input: 214.3 31.8% 91.4%;
        --ring: 222.2 84% 4.9%;
        --radius: 0.5rem;
        --sidebar-background: 0 0% 98%;
        --sidebar-foreground: 240 5.3% 26.1%;
        --sidebar-primary: 240 5.9% 10%;
        --sidebar-primary-foreground: 0 0% 98%;
        --sidebar-accent: 240 4.8% 95.9%;
        --sidebar-accent-foreground: 240 5.9% 10%;
        --sidebar-border: 220 13% 91%;
        --sidebar-ring: 217.2 91.2% 59.8%;
      }

      :host(.benchmark-dashboard.dark) {
        --background: 222.2 84% 4.9%;
        --foreground: 210 40% 98%;
        --card: 222.2 84% 4.9%;
        --card-foreground: 210 40% 98%;
        --popover: 222.2 84% 4.9%;
        --popover-foreground: 210 40% 98%;
        --primary: 210 40% 98%;
        --primary-foreground: 222.2 47.4% 11.2%;
        --secondary: 217.2 32.6% 17.5%;
        --secondary-foreground: 210 40% 98%;
        --muted: 217.2 32.6% 17.5%;
        --muted-foreground: 215 20.2% 65.1%;
        --accent: 217.2 32.6% 17.5%;
        --accent-foreground: 210 40% 98%;
        --destructive: 0 62.8% 30.6%;
        --destructive-foreground: 210 40% 98%;
        --border: 217.2 32.6% 17.5%;
        --input: 217.2 32.6% 17.5%;
        --ring: 212.7 26.8% 83.9%;
        --sidebar-background: 240 5.9% 10%;
        --sidebar-foreground: 240 4.8% 95.9%;
        --sidebar-primary: 224.3 76.3% 48%;
        --sidebar-primary-foreground: 0 0% 100%;
        --sidebar-accent: 240 3.7% 15.9%;
        --sidebar-accent-foreground: 240 4.8% 95.9%;
        --sidebar-border: 240 3.7% 15.9%;
        --sidebar-ring: 217.2 91.2% 59.8%;
      }
    `;
  }

  /**
   * Shadow DOM specific styles
   */
  private static getShadowDOMStyles(): string {
    return `
      /* Shadow DOM host styles */
      :host {
        all: initial;
        display: block;
        font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
        line-height: 1.5;
        -webkit-font-smoothing: antialiased;
        -moz-osx-font-smoothing: grayscale;
        color: hsl(var(--foreground));
        background: hsl(var(--background));
      }

      /* Reset all elements */
      * {
        box-sizing: border-box;
      }

      /* Shadow root container */
      #shadow-root {
        width: 100%;
        min-height: 100%;
        position: relative;
      }

      /* Ensure proper isolation */
      :host([hidden]) {
        display: none;
      }
    `;
  }

  /**
   * Transform compiled CSS to work in Shadow DOM
   */
  private static transformCSSForShadowDOM(css: string): string {
    // Transform :root selector to :host for Shadow DOM
    css = css.replace(/:root\s*{([^}]*)}/g, ':host {$1}');

    // Transform .dark and body.dark selectors to :host(.dark) for Shadow DOM
    css = css.replace(/\.dark\s*,\s*body\.dark\s*{([^}]*)}/g, ':host(.dark) {$1}');
    css = css.replace(/\.dark\s*{([^}]*)}/g, ':host(.dark) {$1}');
    css = css.replace(/body\.dark\s*{([^}]*)}/g, ':host(.dark) {$1}');

    // Also transform any remaining body selectors to :host
    css = css.replace(/body\s*{([^}]*)}/g, ':host {$1}');

    return css;
  }

  /**
   * Complete fallback CSS
   */
  private static getFallbackCSS(): string {
    return `
      ${this.getShadowDOMStyles()}
      ${this.getCSSVariables()}
      ${this.getTailwindFallback()}
    `;
  }

  /**
   * Comprehensive Tailwind CSS fallback for Shadow DOM
   */
  private static getTailwindFallback(): string {
    return `
      /* Tailwind-like reset */
      *,
      ::before,
      ::after {
        box-sizing: border-box;
        border-width: 0;
        border-style: solid;
        border-color: hsl(var(--border));
      }

      /* Display utilities */
      .flex { display: flex !important; }
      .grid { display: grid !important; }
      .block { display: block !important; }
      .inline-block { display: inline-block !important; }
      .inline { display: inline !important; }
      .hidden { display: none !important; }
      .inline-flex { display: inline-flex !important; }

      /* Sizing utilities */
      .w-full { width: 100% !important; }
      .w-8 { width: 2rem !important; }
      .w-auto { width: auto !important; }
      .h-full { height: 100% !important; }
      .h-8 { height: 2rem !important; }
      .h-auto { height: auto !important; }
      .min-h-full { min-height: 100% !important; }

      /* Spacing utilities - Padding */
      .p-0 { padding: 0 !important; }
      .p-1 { padding: 0.25rem !important; }
      .p-2 { padding: 0.5rem !important; }
      .p-4 { padding: 1rem !important; }
      .p-6 { padding: 1.5rem !important; }
      .p-8 { padding: 2rem !important; }
      .px-3 { padding-left: 0.75rem !important; padding-right: 0.75rem !important; }
      .px-4 { padding-left: 1rem !important; padding-right: 1rem !important; }
      .px-6 { padding-left: 1.5rem !important; padding-right: 1.5rem !important; }
      .py-2 { padding-top: 0.5rem !important; padding-bottom: 0.5rem !important; }
      .py-4 { padding-top: 1rem !important; padding-bottom: 1rem !important; }
      .py-12 { padding-top: 3rem !important; padding-bottom: 3rem !important; }

      /* Spacing utilities - Margin */
      .m-0 { margin: 0 !important; }
      .m-1 { margin: 0.25rem !important; }
      .m-2 { margin: 0.5rem !important; }
      .m-4 { margin: 1rem !important; }
      .m-6 { margin: 1.5rem !important; }
      .m-8 { margin: 2rem !important; }
      .mb-2 { margin-bottom: 0.5rem !important; }
      .mb-4 { margin-bottom: 1rem !important; }
      .mb-5 { margin-bottom: 1.25rem !important; }
      .mb-6 { margin-bottom: 1.5rem !important; }
      .mb-8 { margin-bottom: 2rem !important; }
      .mt-4 { margin-top: 1rem !important; }
      .mt-5 { margin-top: 1.25rem !important; }
      .mr-3 { margin-right: 0.75rem !important; }

      /* Typography */
      .text-sm { font-size: 0.875rem !important; line-height: 1.25rem !important; }
      .text-base { font-size: 1rem !important; line-height: 1.5rem !important; }
      .text-lg { font-size: 1.125rem !important; line-height: 1.75rem !important; }
      .text-xl { font-size: 1.25rem !important; line-height: 1.75rem !important; }
      .text-2xl { font-size: 1.5rem !important; line-height: 2rem !important; }
      .text-3xl { font-size: 1.875rem !important; line-height: 2.25rem !important; }

      .font-medium { font-weight: 500 !important; }
      .font-semibold { font-weight: 600 !important; }
      .font-bold { font-weight: 700 !important; }

      /* Text colors */
      .text-white { color: #ffffff !important; }
      .text-gray-600 { color: hsl(var(--muted-foreground)) !important; }
      .text-gray-700 { color: hsl(var(--foreground)) !important; }
      .text-gray-800 { color: hsl(var(--foreground)) !important; }
      .text-gray-900 { color: hsl(var(--foreground)) !important; }

      /* Background colors */
      .bg-white { background-color: hsl(var(--background)) !important; }
      .bg-gray-50 { background-color: hsl(var(--muted)) !important; }
      .bg-gray-100 { background-color: hsl(var(--muted)) !important; }
      .bg-gray-700 { background-color: hsl(var(--muted)) !important; }
      .bg-gray-800 { background-color: hsl(var(--card)) !important; }
      .bg-gradient-to-r { background-image: linear-gradient(to right, var(--tw-gradient-stops)) !important; }

      /* Gradient colors */
      .from-purple-600 { --tw-gradient-from: #9333ea !important; --tw-gradient-to: rgb(147 51 234 / 0) !important; --tw-gradient-stops: var(--tw-gradient-from), var(--tw-gradient-to) !important; }
      .to-indigo-600 { --tw-gradient-to: #4f46e5 !important; }

      /* Border utilities */
      .border { border-width: 1px !important; }
      .border-b { border-bottom-width: 1px !important; }
      .border-b-2 { border-bottom-width: 2px !important; }
      .border-t-lg { border-top-width: 4px !important; }
      .border-gray-200 { border-color: hsl(var(--border)) !important; }
      .border-gray-300 { border-color: hsl(var(--border)) !important; }
      .border-gray-700 { border-color: hsl(var(--border)) !important; }
      .border-blue-600 { border-color: #2563eb !important; }
      .border-purple-600 { border-color: #9333ea !important; }
      .border-transparent { border-color: transparent !important; }

      /* Border radius */
      .rounded { border-radius: 0.25rem !important; }
      .rounded-md { border-radius: 0.375rem !important; }
      .rounded-lg { border-radius: 0.5rem !important; }
      .rounded-xl { border-radius: 0.75rem !important; }
      .rounded-2xl { border-radius: 1rem !important; }
      .rounded-t-lg { border-top-left-radius: 0.5rem !important; border-top-right-radius: 0.5rem !important; }
      .rounded-full { border-radius: 9999px !important; }
      .rounded-none { border-radius: 0px !important; }

      /* Box shadow */
      .shadow { box-shadow: 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1) !important; }
      .shadow-lg { box-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1) !important; }

      /* Grid utilities */
      .grid-cols-1 { grid-template-columns: repeat(1, minmax(0, 1fr)) !important; }
      .grid-cols-2 { grid-template-columns: repeat(2, minmax(0, 1fr)) !important; }
      .grid-cols-3 { grid-template-columns: repeat(3, minmax(0, 1fr)) !important; }

      .gap-1 { gap: 0.25rem !important; }
      .gap-2 { gap: 0.5rem !important; }
      .gap-4 { gap: 1rem !important; }

      /* Flexbox utilities */
      .items-center { align-items: center !important; }
      .items-start { align-items: flex-start !important; }
      .justify-center { justify-content: center !important; }
      .justify-between { justify-content: space-between !important; }
      .justify-start { justify-content: flex-start !important; }

      /* Position utilities */
      .relative { position: relative !important; }
      .absolute { position: absolute !important; }
      .z-50 { z-index: 50 !important; }

      /* Overflow utilities */
      .overflow-hidden { overflow: hidden !important; }
      .overflow-auto { overflow: auto !important; }

      /* Cursor utilities */
      .cursor-pointer { cursor: pointer !important; }

      /* Transition utilities */
      .transition-all { transition-property: all !important; transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1) !important; transition-duration: 150ms !important; }
      .duration-200 { transition-duration: 200ms !important; }

      /* Hover utilities */
      .hover\\:bg-gray-50:hover { background-color: hsl(var(--muted)) !important; }
      .hover\\:bg-gray-700:hover { background-color: hsl(var(--muted)) !important; }

      /* Animation utilities */
      @keyframes spin {
        to {
          transform: rotate(360deg);
        }
      }
      .animate-spin {
        animation: spin 1s linear infinite !important;
      }

      /* Dark mode utilities */
      .dark\\:bg-gray-700 { background-color: hsl(var(--muted)) !important; }
      .dark\\:bg-gray-800 { background-color: hsl(var(--card)) !important; }
      .dark\\:border-gray-700 { border-color: hsl(var(--border)) !important; }
      .dark\\:text-gray-300 { color: hsl(var(--muted-foreground)) !important; }
      .dark\\:hover\\:bg-gray-700:hover { background-color: hsl(var(--muted)) !important; }

      /* Data attribute utilities for Radix components */
      [data-state="active"] {
        background: linear-gradient(to right, #9333ea, #4f46e5) !important;
        color: white !important;
        box-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1) !important;
        border-bottom: 2px solid #9333ea !important;
      }

      /* White space utilities */
      .whitespace-nowrap { white-space: nowrap !important; }

      /* Text alignment */
      .text-center { text-align: center !important; }
      .text-left { text-align: left !important; }

      /* Responsive utilities (basic support) */
      @media (min-width: 640px) {
        .sm\\:text-base { font-size: 1rem !important; line-height: 1.5rem !important; }
      }
    `;
  }
}
