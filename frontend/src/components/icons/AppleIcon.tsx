interface AppleIconProps {
  className?: string;
}

export default function AppleIcon({ className }: AppleIconProps) {
  return (
    <svg
      className={className}
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="currentColor"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <path d="M17.05 20.28c-.98.95-2.05.88-3.08.4-1.09-.5-2.08-.52-3.23 0-1.44.62-2.2.44-3.06-.4C3.79 16.17 4.36 9.98 8.87 9.71c1.28.07 2.17.74 2.92.78.99-.2 1.95-.78 3.01-.7 1.28.1 2.24.6 2.87 1.5-2.63 1.58-2.01 5.07.38 6.04-.45 1.18-1.04 2.35-2 2.95zM12.03 9.64c-.14-2.24 1.72-4.11 3.87-4.3.28 2.47-2.24 4.41-3.87 4.3z" />
    </svg>
  );
}
