import * as React from 'react';

interface ProjectBarChartIconProps extends React.SVGProps<SVGSVGElement> {
  className?: string;
}

/**
 * Simple bar chart icon (Lucide BarChart3).
 */
export function ProjectBarChartIcon({ className, ...props }: ProjectBarChartIconProps) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className || 'w-5 h-5'}
      {...props}
    >
      <path d="M3 3v16a2 2 0 0 0 2 2h16" />
      <path d="M18 17V9" />
      <path d="M13 17V5" />
      <path d="M8 17v-3" />
    </svg>
  );
}
