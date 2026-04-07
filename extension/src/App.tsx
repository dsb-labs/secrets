import { type ComponentChildren } from "preact";
import { useEffect, useRef, useState } from "preact/hooks";
import { LocationProvider, Router, Route, useLocation } from "preact-iso";
import { getServerURL, getToken, setToken as persistToken, clearToken } from "@/lib/storage";
import { Client } from "@/lib/client";
import { Setup } from "@/view/auth/Setup";
import { Login } from "@/view/auth/Login";
import { List } from "@/view/login/List";
import { Detail } from "@/view/login/Detail";
import { Create } from "@/view/login/Create";

type Status = "loading" | "setup" | "login" | "authenticated";

export function App() {
  const [status, setStatus] = useState<Status>("loading");
  const clientRef = useRef(new Client(""));

  useEffect(() => {
    async function init() {
      const [url, token] = await Promise.all([getServerURL(), getToken()]);
      if (!url) {
        setStatus("setup");
        return;
      }
      clientRef.current = new Client(url, token ?? "");
      setStatus(token ? "authenticated" : "login");
    }

    init();
  }, []);

  async function handleConfigured(url: string) {
    clientRef.current = new Client(url);
    setStatus("login");
  }

  async function handleAuthenticated() {
    await persistToken(clientRef.current.token());
    setStatus("authenticated");
  }

  async function handleExpired() {
    await clearToken();
    setStatus("login");
  }

  const client = clientRef.current;

  return (
    <LocationProvider>
      <AuthGuard status={status}>
        <Router>
          <Route path="/setup" component={() => <Setup onConfigured={handleConfigured} />} />
          <Route path="/login" component={() => <Login client={client} onAuthenticated={handleAuthenticated} />} />
          <Route path="/logins" component={() => <List client={client} onExpired={handleExpired} />} />
          <Route path="/logins/new" component={() => <Create client={client} onExpired={handleExpired} />} />
          <Route
            path="/logins/:id"
            component={({ id }: { id: string }) => <Detail id={id} client={client} onExpired={handleExpired} />}
          />
        </Router>
      </AuthGuard>
    </LocationProvider>
  );
}

function AuthGuard({ status, children }: { status: Status; children: ComponentChildren }) {
  const { path, route } = useLocation();

  useEffect(() => {
    if (status === "loading") return;
    if (status === "setup" && path !== "/setup") route("/setup");
    else if (status === "login" && path !== "/login") route("/login");
    else if (status === "authenticated" && path !== "/logins" && !path.startsWith("/logins/")) route("/logins");
  }, [status]);

  if (status === "loading") {
    return (
      <div class="flex min-h-48 items-center justify-center">
        <p class="text-sm text-gray-500 dark:text-gray-400">Loading…</p>
      </div>
    );
  }

  return <>{children}</>;
}
