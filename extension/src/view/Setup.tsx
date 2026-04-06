import { useState } from "preact/hooks";
import { setServerURL } from "@/lib/storage";
import { KeeperClient, UnreachableError } from "@/lib/client";

type Props = {
  onConfigured: (url: string) => void;
};

// Setup renders a form that prompts the user to enter their Keeper server URL. On submission, it
// validates the URL format and checks reachability via the health endpoint before persisting the
// URL to storage and calling onConfigured.
export function Setup({ onConfigured }: Props) {
  const [url, setUrl] = useState("");
  const [error, setError] = useState("");
  const [checking, setChecking] = useState(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    try {
      new URL(url);
    } catch {
      setError("Please enter a valid URL.");
      return;
    }

    setChecking(true);
    setError("");
    try {
      await new KeeperClient(url).ping();
    } catch (err) {
      setError(err instanceof UnreachableError ? err.message : "An unexpected error occurred.");
      return;
    } finally {
      setChecking(false);
    }

    await setServerURL(url);
    onConfigured(url);
  }

  return (
    <form onSubmit={handleSubmit} class="flex flex-col gap-4 p-4">
      <div>
        <h1 class="text-sm font-semibold text-gray-900 dark:text-white">Connect to Keeper</h1>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">Enter the URL of your Keeper server.</p>
      </div>
      <div class="flex flex-col gap-1">
        <input
          type="url"
          value={url}
          onInput={(e) => setUrl((e.target as HTMLInputElement).value)}
          placeholder="https://your-secrets-server.com"
          class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-white"
          required
        />
        {error && <p class="text-xs text-red-500">{error}</p>}
      </div>
      <button
        type="submit"
        disabled={checking}
        class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
      >
        {checking ? "Connecting…" : "Connect"}
      </button>
    </form>
  );
}
