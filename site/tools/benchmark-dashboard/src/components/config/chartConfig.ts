import { ChartConfig } from '@/components/ui/chart';

// Standardized color palette
export const CHART_COLORS = {
  PRIMARY: '#8b5cf6',     // Purple-500
  SECONDARY: '#6366f1',   // Indigo-500
  ACCENT: '#a855f7',      // Purple-400
  DARK: '#4f46e5',        // Indigo-600
  LIGHT: '#c084fc',       // Purple-300
  VARIANT: '#7c3aed',     // Violet-600
  MUTED: '#818cf8',       // Indigo-400
} as const;

// Standard chart configurations
export const DEFAULT_CHART_CONFIG: ChartConfig = {
  throughput: {
    label: "Throughput",
    color: CHART_COLORS.PRIMARY,
  },
  latency: {
    label: "Latency",
    color: CHART_COLORS.SECONDARY,
  },
  gateway: {
    label: "Gateway",
    color: CHART_COLORS.ACCENT,
  },
  proxy: {
    label: "Proxy",
    color: CHART_COLORS.DARK,
  },
  memory: {
    label: "Memory",
    color: CHART_COLORS.PRIMARY,
  },
  cpu: {
    label: "CPU",
    color: CHART_COLORS.SECONDARY,
  },
  efficiency: {
    label: "Efficiency",
    color: CHART_COLORS.VARIANT,
  },
};

// Specific chart configurations
export const PERFORMANCE_CHART_CONFIG: ChartConfig = {
  throughput: {
    label: "Throughput",
    color: CHART_COLORS.PRIMARY,
  },
  latency: {
    label: "Latency",
    color: CHART_COLORS.SECONDARY,
  },
  gateway: {
    label: "Gateway",
    color: CHART_COLORS.ACCENT,
  },
  proxy: {
    label: "Proxy",
    color: CHART_COLORS.DARK,
  },
  rpsPerMB: {
    label: "RPS per MB",
    color: CHART_COLORS.VARIANT,
  },
};

export const RESOURCE_CHART_CONFIG: ChartConfig = {
  gateway: {
    label: "Gateway",
    color: CHART_COLORS.ACCENT,
  },
  proxy: {
    label: "Proxy",
    color: CHART_COLORS.DARK,
  },
  gatewayMean: {
    label: "Gateway Mean",
    color: CHART_COLORS.PRIMARY,
  },
  proxyMean: {
    label: "Proxy Mean",
    color: CHART_COLORS.SECONDARY,
  },
  gatewayMax: {
    label: "Gateway Max",
    color: CHART_COLORS.LIGHT,
  },
  proxyMax: {
    label: "Proxy Max",
    color: CHART_COLORS.MUTED,
  },
};

export const LATENCY_CHART_CONFIG: ChartConfig = {
  latency: {
    label: "Latency",
    color: CHART_COLORS.PRIMARY,
  },
  p50: {
    label: "P50",
    color: CHART_COLORS.PRIMARY,
  },
  p75: {
    label: "P75",
    color: CHART_COLORS.SECONDARY,
  },
  p90: {
    label: "P90",
    color: CHART_COLORS.DARK,
  },
  p95: {
    label: "P95",
    color: CHART_COLORS.ACCENT,
  },
  p99: {
    label: "P99",
    color: CHART_COLORS.VARIANT,
  },
};

// Chart styling configurations
export const CHART_STYLES = {
  AREA: {
    FILL_OPACITY: 0.4,
    STROKE_WIDTH: 2,
  },
  LINE: {
    STROKE_WIDTH: 3,
    DOT_RADIUS: 3,
    DOT_STROKE_WIDTH: 2,
  },
  BAR: {
    RADIUS: 4,
  },
  GRID: {
    STROKE_DASH_ARRAY: '3 3',
  },
} as const;
