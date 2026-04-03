import { useEffect, useState } from "preact/hooks";
import browser from "webextension-polyfill";
import { KeeperClient, UnauthorizedError, type Login } from "@/lib/client";
import { EmptyState } from "@/component/EmptyState";
import { LoginList } from "@/component/LoginList";

type Props = {
  client: KeeperClient;
  onExpired: () => void;
};

// Logins fetches and renders the list of logins stored for the current tab's domain. Shows a
// loading state while fetching, an empty state when no logins are found, and a list of rows
// otherwise.
export function Logins({ client, onExpired }: Props) {
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
          onExpired();
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
      <div>
        <h1 class="text-sm font-semibold text-gray-900 dark:text-white">Logins</h1>
        {domain && <p class="mt-0.5 truncate text-xs text-gray-500 dark:text-gray-400">{domain}</p>}
      </div>
      {logins.length === 0 ? <EmptyState /> : <LoginList logins={logins} />}
    </div>
  );
}
