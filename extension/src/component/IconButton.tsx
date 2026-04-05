import { type ComponentChildren } from "preact";

type Props = {
  onClick: () => void;
  label: string;
  children: ComponentChildren;
};

// IconButton renders a small square icon button with a consistent border and hover style.
export function IconButton({ onClick, label, children }: Props) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={label}
      class="flex h-9 w-9 shrink-0 cursor-pointer items-center justify-center rounded-lg border border-gray-200 bg-white text-gray-500 transition-colors hover:bg-gray-50 hover:text-gray-700 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-gray-200"
    >
      {children}
    </button>
  );
}
