// Application-wide constants
export const APP_CONSTANTS = {
  CHART: {
    DEFAULT_HEIGHT: 400,
    DEFAULT_TICK_MARGIN: 8,
    ROUTE_TICKS: [10, 50, 100, 300, 500, 1000],
    ROUTE_DOMAIN: [0, 1100],
  },
  PERFORMANCE: {
    MAX_ROUTES: 1000,
    MIN_ROUTES: 10,
    DEFAULT_RPS: 10000,
    DEFAULT_CONNECTIONS: 100,
    DEFAULT_DURATION: 30,
  },
  FORMATS: {
    MEMORY: 'MB',
    CPU: '%',
    LATENCY: 'ms',
    THROUGHPUT: 'RPS',
  }
};

// Route definitions
export const ROUTES = {
  HOME: '/',
  TESTING_DETAILS: '/testing-details',
  RESILIENCE_DETAILS: '/resilience-details',
  CONFORMANCE_DETAILS: '/conformance-details',
} as const;

// Test phases
export const TEST_PHASES = {
  SCALING_UP: 'scaling-up',
  SCALING_DOWN: 'scaling-down',
} as const;

// Chart stroke patterns
export const STROKE_PATTERNS = {
  SOLID: 'none',
  DASHED: '5 5',
  DOTTED: '3 3',
} as const;
