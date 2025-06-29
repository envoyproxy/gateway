import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Helper function to apply Tailwind classes with the bd- prefix
 * This helps during the migration from unprefixed to prefixed classes
 */
export function bdcn(...inputs: ClassValue[]) {
  // This function can be used to gradually migrate classes
  // For now, it just calls the regular cn function
  return cn(...inputs)
}

/**
 * Utility to convert regular Tailwind classes to bd- prefixed classes
 * Useful for automated migration
 */
export function addBdPrefix(className: string): string {
  if (!className) return className;

  // List of common Tailwind prefixes that should get the bd- prefix
  const tailwindPrefixes = [
    'flex', 'grid', 'block', 'inline', 'hidden', 'visible',
    'absolute', 'relative', 'fixed', 'sticky', 'static',
    'top-', 'right-', 'bottom-', 'left-', 'inset-',
    'z-', 'order-',
    'w-', 'h-', 'min-w-', 'min-h-', 'max-w-', 'max-h-',
    'm-', 'mx-', 'my-', 'mt-', 'mr-', 'mb-', 'ml-',
    'p-', 'px-', 'py-', 'pt-', 'pr-', 'pb-', 'pl-',
    'text-', 'font-', 'leading-', 'tracking-', 'align-',
    'bg-', 'border-', 'outline-', 'ring-',
    'rounded-', 'shadow-',
    'opacity-', 'cursor-', 'select-', 'pointer-events-',
    'transition-', 'duration-', 'ease-', 'delay-',
    'transform', 'rotate-', 'scale-', 'translate-', 'skew-',
    'overflow-', 'truncate', 'whitespace-',
    'items-', 'justify-', 'content-', 'self-',
    'space-x-', 'space-y-', 'gap-',
    'sr-only', 'not-sr-only'
  ];

  return className
    .split(' ')
    .map(cls => {
      if (!cls.trim()) return cls;

      // Check if class starts with any Tailwind prefix
      const needsPrefix = tailwindPrefixes.some(prefix =>
        cls.startsWith(prefix) || cls === prefix.slice(0, -1)
      );

      // Don't add prefix if it already has bd- or if it's not a Tailwind class
      if (needsPrefix && !cls.startsWith('bd-')) {
        return `bd-${cls}`;
      }

      return cls;
    })
    .join(' ');
}
