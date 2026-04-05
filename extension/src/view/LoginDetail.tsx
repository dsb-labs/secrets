import { useEffect, useState } from "preact/hooks";
import { useLocation } from "preact-iso";
import { type ComponentChildren } from "preact";
import { KeeperClient, UnauthorizedError, type Login } from "@/lib/client";
import { CopyButton } from "@/component/CopyButton";
import { PasswordField } from "@/component/PasswordField";
import { autofill } from "@/lib/autofill";

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
  const [fillStatus, setFillStatus] = useState<"idle" | "filled" | "not-found">("idle");

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

  async function handleAutofill() {
    if (!login) return;
    const filled = await autofill(login.username, login.password);
    setFillStatus(filled ? "filled" : "not-found");
    setTimeout(() => setFillStatus("idle"), 2000);
  }

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
      <button
        onClick={handleAutofill}
        class="flex w-full items-center justify-center gap-2 rounded-lg bg-indigo-600 px-4 py-2.5 text-sm font-medium text-white hover:bg-indigo-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500 focus-visible:ring-offset-2 dark:focus-visible:ring-offset-gray-900"
      >
        {fillStatus === "filled" ? (
          <>
            <svg
              class="h-4 w-4"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              stroke-width="1.5"
              stroke="currentColor"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
            </svg>
            Filled
          </>
        ) : fillStatus === "not-found" ? (
          <>
            <svg
              class="h-4 w-4"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              stroke-width="1.5"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z"
              />
            </svg>
            No fields found
          </>
        ) : (
          <>
            <svg
              class="h-4 w-4"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              stroke-width="1.5"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M16.862 4.487l1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10"
              />
            </svg>
            Auto-fill
          </>
        )}
      </button>
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
