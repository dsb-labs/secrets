import { useEffect, useState } from "preact/hooks";
import browser from "webextension-polyfill";
import { listLogins, type Login } from "@/lib/client";

type Props = {
  serverURL: string;
  token: string;
};

// Logins fetches and renders the list of logins stored for the current tab's domain. Shows a
// loading state while fetching, an empty state when no logins are found, and a list of rows
// otherwise.
export function Logins({ serverURL, token }: Props) {
  const [logins, setLogins] = useState<Login[]>([]);
  const [domain, setDomain] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    browser.tabs
      .query({ active: true, currentWindow: true })
      .then((tabs) => {
        const url = tabs[0]?.url;
        if (!url) return "";
        return new URL(url).hostname;
      })
      .then((hostname) => {
        setDomain(hostname);
        return listLogins(serverURL, token, hostname);
      })
      .then(setLogins)
      .catch(() => setError("Failed to load logins."))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div class="flex min-h-48 items-center justify-center">
        <p class="text-sm text-gray-500 dark:text-gray-400">Loading…</p>
      </div>
    );
  }

  if (error) {
    return (
      <div class="flex min-h-48 items-center justify-center p-4">
        <p class="text-sm text-red-500">{error}</p>
      </div>
    );
  }

  return (
    <div class="flex flex-col gap-4 p-4">
      <div>
        <h1 class="text-sm font-semibold text-gray-900 dark:text-white">Logins</h1>
        {domain && <p class="mt-0.5 truncate text-xs text-gray-500 dark:text-gray-400">{domain}</p>}
      </div>
      {logins.length === 0 ? <EmptyState /> : <LoginList logins={logins} />}
    </div>
  );
}

function EmptyState() {
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

function LoginList({ logins }: { logins: Login[] }) {
  return (
    <div class="divide-y divide-gray-200 rounded-xl border border-gray-200 bg-white shadow-sm dark:divide-gray-700 dark:border-gray-700 dark:bg-gray-800">
      {logins.map((login) => (
        <LoginRow key={login.id} login={login} />
      ))}
    </div>
  );
}

function LoginRow({ login }: { login: Login }) {
  const title = login.name || login.domains[0] || login.username;
  const favicon = login.domains[0] ? `https://www.google.com/s2/favicons?domain=${login.domains[0]}&sz=32` : null;

  return (
    <div class="flex items-center justify-between px-4 py-3.5 transition-colors hover:bg-gray-50 dark:hover:bg-gray-700/50">
      <div class="flex min-w-0 items-center gap-3">
        {favicon ? (
          <img
            src={favicon}
            alt=""
            class="h-5 w-5 shrink-0 rounded"
            onError={(e) => ((e.target as HTMLImageElement).style.display = "none")}
          />
        ) : (
          <svg
            class="h-5 w-5 shrink-0 text-gray-400 dark:text-gray-500"
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
        )}
        <div class="min-w-0">
          <p class="truncate text-sm font-medium text-gray-900 dark:text-white">{title}</p>
          <p class="mt-0.5 truncate text-xs text-gray-500 dark:text-gray-400">{login.username}</p>
        </div>
      </div>
      <svg
        class="h-4 w-4 shrink-0 text-gray-400 dark:text-gray-500"
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        stroke-width="1.5"
        stroke="currentColor"
      >
        <path stroke-linecap="round" stroke-linejoin="round" d="m8.25 4.5 7.5 7.5-7.5 7.5" />
      </svg>
    </div>
  );
}
