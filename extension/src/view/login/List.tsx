import { useEffect, useState } from "preact/hooks";
import { useLocation } from "preact-iso";
import browser from "webextension-polyfill";
import { Client, UnauthorizedError, type Login } from "@/lib/client";
import { EmptyState } from "@/component/EmptyState";
import { LoginList } from "@/component/LoginList";

type Props = {
  client: Client;
  onExpired: () => Promise<void>;
};

// List fetches and renders the list of logins stored for the current tab's domain. Shows a
// loading state while fetching, an empty state when no logins are found, and a list of rows
// otherwise.
export function List({ client, onExpired }: Props) {
  const { route } = useLocation();
  const [logins, setLogins] = useState<Login[]>([]);
  const [domain, setDomain] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    async function load() {
      try {
        const tabs = await browser.tabs.query({ active: true, currentWindow: true });
        const url = tabs[0]?.url;
        const hostname = url ? new URL(url).hostname : "";
        setDomain(hostname);
        setLogins(await client.listLogins(hostname));
      } catch (err) {
        if (err instanceof UnauthorizedError) {
          await onExpired();
        } else {
          setError("Failed to load logins.");
        }
      } finally {
        setLoading(false);
      }
    }

    load();
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
      <div class="flex items-start justify-between">
        <div>
          <h1 class="text-sm font-semibold text-gray-900 dark:text-white">Logins</h1>
          {domain && <p class="mt-0.5 truncate text-xs text-gray-500 dark:text-gray-400">{domain}</p>}
        </div>
        <button
          onClick={() => route("/logins/new")}
          aria-label="Add login"
          class="flex items-center justify-center rounded-md p-1 text-gray-500 hover:bg-gray-100 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-gray-200"
        >
          <svg class="h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
          </svg>
        </button>
      </div>
      {logins.length === 0 ? <EmptyState /> : <LoginList logins={logins} onSelect={(id) => route(`/logins/${id}`)} />}
    </div>
  );
}
