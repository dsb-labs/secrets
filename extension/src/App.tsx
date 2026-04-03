import { useEffect, useState } from "preact/hooks";
import { getServerURL } from "@/lib/storage";
import { Setup } from "@/view/Setup";

export function App() {
  const [serverURL, setServerURL] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getServerURL().then((url) => {
      setServerURL(url || null);
      setLoading(false);
    });
  }, []);

  if (loading) {
    return (
      <div class="flex min-h-48 items-center justify-center">
        <p class="text-sm text-gray-500 dark:text-gray-400">Loading…</p>
      </div>
    );
  }

  if (!serverURL) {
    return <Setup onConfigured={setServerURL} />;
  }

  return (
    <div class="flex flex-col gap-2 p-4">
      <h1 class="text-base font-semibold text-gray-900 dark:text-white">Keeper</h1>
      <p class="text-sm text-gray-500 dark:text-gray-400">Connected to {serverURL}</p>
    </div>
  );
}
