import { useEffect, useState } from "preact/hooks";
import { useLocation } from "preact-iso";
import { type ComponentChildren } from "preact";
import { KeeperClient, UnauthorizedError, type Login } from "@/lib/client";
import { CopyButton } from "@/component/CopyButton";
import { PasswordField } from "@/component/PasswordField";

type Props = {
  id: string;
  client: KeeperClient;
  onExpired: () => Promise<void>;
};

// LoginDetail fetches and renders the full details of a single login, including username, password
// (with show/hide and copy controls), domains, and creation date.
export function LoginDetail({ id, client, onExpired }: Props) {
  const { route } = useLocation();
  const [login, setLogin] = useState<Login | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    async function load() {
      try {
        setLogin(await client.getLogin(id));
      } catch (err) {
        if (err instanceof UnauthorizedError) {
          await onExpired();
        } else {
          setError("Failed to load login.");
        }
      } finally {
        setLoading(false);
      }
    }

    load();
  }, [id]);

  if (loading) {
    return (
      <div class="flex min-h-48 items-center justify-center">
        <p class="text-sm text-gray-500 dark:text-gray-400">Loading…</p>
      </div>
    );
  }

  if (error || !login) {
    return (
      <div class="flex min-h-48 items-center justify-center p-4">
        <p class="text-sm text-red-500">{error || "Login not found."}</p>
      </div>
    );
  }

  return (
    <div class="flex flex-col gap-4 p-4">
      <button
        onClick={() => route("/logins")}
        class="flex items-center gap-1.5 self-start text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
      >
        <svg class="h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
        </svg>
        Back to logins
      </button>
      <div class="divide-y divide-gray-200 rounded-xl border border-gray-200 bg-white shadow-sm dark:divide-gray-700 dark:border-gray-700 dark:bg-gray-800">
        <div class="px-4 py-4">
          <h1 class="text-base font-semibold text-gray-900 dark:text-white">{login.name || "Login details"}</h1>
        </div>
        <div class="space-y-5 px-4 py-4">
          <Field label="Username">
            <div class="flex items-center gap-2">
              <p class="flex-1 truncate text-sm text-gray-900 dark:text-white">{login.username}</p>
              <CopyButton value={login.username} />
            </div>
          </Field>
          <Field label="Password">
            <PasswordField password={login.password} />
          </Field>
          {login.domains.length > 0 && (
            <Field label="Domains">
              <ul class="space-y-2">
                {login.domains.map((domain) => (
                  <li key={domain} class="flex min-w-0 items-center gap-2">
                    <img
                      src={`https://www.google.com/s2/favicons?domain=${domain}&sz=32`}
                      alt=""
                      class="h-5 w-5 shrink-0 rounded"
                      onError={(e) => ((e.target as HTMLImageElement).style.display = "none")}
                    />
                    <span class="truncate text-sm text-indigo-600 dark:text-indigo-400">{domain}</span>
                  </li>
                ))}
              </ul>
            </Field>
          )}
        </div>
        <div class="px-4 py-3">
          <p class="text-xs font-medium tracking-wide text-gray-500 uppercase dark:text-gray-400">Created</p>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{new Date(login.createdAt).toLocaleDateString()}</p>
        </div>
      </div>
    </div>
  );
}

function Field({ label, children }: { label: string; children: ComponentChildren }) {
  return (
    <div>
      <p class="mb-1.5 text-xs font-medium tracking-wide text-gray-500 uppercase dark:text-gray-400">{label}</p>
      {children}
    </div>
  );
}
