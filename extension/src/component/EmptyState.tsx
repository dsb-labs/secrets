// EmptyState renders a centred empty-state card with a key icon, title, and description.
export function EmptyState() {
  return (
    <div class="flex flex-col items-center justify-center rounded-xl border border-dashed border-gray-300 bg-white py-10 text-center dark:border-gray-600 dark:bg-gray-800">
      <div class="mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-gray-100 dark:bg-gray-700">
        <svg
          class="h-6 w-6 text-gray-400 dark:text-gray-500"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke-width="1.5"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M15.75 5.25a3 3 0 0 1 3 3m3 0a6 6 0 0 1-7.029 5.912c-.563-.097-1.159.026-1.563.43L10.5 17.25H8.25v2.25H6v2.25H2.25v-2.818c0-.597.237-1.17.659-1.591l6.499-6.499c.404-.404.527-1 .43-1.563A6 6 0 0 1 21.75 8.25Z"
          />
        </svg>
      </div>
      <p class="text-sm font-medium text-gray-900 dark:text-white">No logins for this page</p>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">No saved logins match this domain.</p>
    </div>
  );
}
