import { useEffect, useState } from "preact/hooks";
import { getServerURL, getToken } from "@/lib/storage";
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
    return <Login serverURL={serverURL} onAuthenticated={setToken} />;
  }

  return <Logins serverURL={serverURL} token={token} />;
}
