import { useEffect, useMemo, useState } from "preact/hooks";
import { getServerURL, getToken, setToken as persistToken, clearToken } from "@/lib/storage";
import { KeeperClient } from "@/lib/client";
import { Setup } from "@/view/Setup";
import { Login } from "@/view/Login";
import { Logins } from "@/view/Logins";

export function App() {
  const [serverURL, setServerURL] = useState<string | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([getServerURL(), getToken()]).then(([url, tkn]) => {
      setServerURL(url || null);
      setToken(tkn || null);
      setLoading(false);
    });
  }, []);

  const client = useMemo(() => new KeeperClient(serverURL ?? "", token ?? ""), [serverURL, token]);

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

  if (!token) {
    async function handleAuthenticated() {
      await persistToken(client.token());
      setToken(client.token());
    }

    return <Login client={client} onAuthenticated={handleAuthenticated} />;
  }

  async function handleExpired() {
    await clearToken();
    setToken(null);
  }

  return <Logins client={client} onExpired={handleExpired} />;
}
