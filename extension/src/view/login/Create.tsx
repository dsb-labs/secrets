import { useEffect, useState } from "preact/hooks";
import { useLocation } from "preact-iso";
import browser from "webextension-polyfill";
import { Client, UnauthorizedError, UnreachableError } from "@/lib/client";

type Props = {
  client: Client;
  onExpired: () => Promise<void>;
};

// Create renders a form to create a new login entry. The domains field is pre-populated with
// the current tab's hostname.
export function Create({ client, onExpired }: Props) {
  const { route } = useLocation();
  const [name, setName] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [passwordVisible, setPasswordVisible] = useState(false);
  const [domains, setDomains] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    async function prefill() {
      const tabs = await browser.tabs.query({ active: true, currentWindow: true });
      const url = tabs[0]?.url;
      if (url) {
        const hostname = new URL(url).hostname;
        setDomains(hostname);
      }
    }
    prefill();
  }, []);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    setSubmitting(true);
    setError("");
    try {
      const domainList = domains
        .split(",")
        .map((d) => d.trim())
        .filter(Boolean);
      await client.createLogin({ name, username, password, domains: domainList });
      route("/logins");
    } catch (err) {
      if (err instanceof UnauthorizedError) {
        await onExpired();
      } else if (err instanceof UnreachableError) {
        setError(err.message);
      } else {
        setError("An unexpected error occurred.");
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} class="flex flex-col gap-4 p-4">
      <div class="flex items-center gap-2">
        <button
          type="button"
          onClick={() => route("/logins")}
          class="flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
        >
          <svg class="h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
          </svg>
          Back
        </button>
        <h1 class="text-sm font-semibold text-gray-900 dark:text-white">New login</h1>
      </div>
      <div class="flex flex-col gap-3">
        <input
          type="text"
          value={name}
          onInput={(e) => setName((e.target as HTMLInputElement).value)}
          placeholder="Name"
          class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-white"
          required
        />
        <input
          type="text"
          value={username}
          onInput={(e) => setUsername((e.target as HTMLInputElement).value)}
          placeholder="Username"
          autocomplete="off"
          class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-white"
          required
        />
        <div class="relative">
          <input
            type={passwordVisible ? "text" : "password"}
            value={password}
            onInput={(e) => setPassword((e.target as HTMLInputElement).value)}
            placeholder="Password"
            autocomplete="new-password"
            class="w-full rounded-md border border-gray-300 px-3 py-2 pr-10 text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-white"
            required
          />
          <button
            type="button"
            onClick={() => setPasswordVisible((v) => !v)}
            aria-label="Toggle password visibility"
            class="absolute inset-y-0 right-2 flex items-center text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
          >
            {passwordVisible ? (
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
                  d="M3.98 8.223A10.477 10.477 0 0 0 1.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.451 10.451 0 0 1 12 4.5c4.756 0 8.773 3.162 10.065 7.498a10.522 10.522 0 0 1-4.293 5.774M6.228 6.228 3 3m3.228 3.228 3.65 3.65m7.894 7.894L21 21m-3.228-3.228-3.65-3.65m0 0a3 3 0 1 0-4.243-4.243m4.242 4.242L9.88 9.88"
                />
              </svg>
            ) : (
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
                  d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.641 0-8.574-3.007-9.964-7.178Z"
                />
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
              </svg>
            )}
          </button>
        </div>
        <input
          type="text"
          value={domains}
          onInput={(e) => setDomains((e.target as HTMLInputElement).value)}
          placeholder="Domains (comma-separated)"
          autocomplete="off"
          class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-white"
        />
        {error && <p class="text-xs text-red-500">{error}</p>}
      </div>
      <button
        type="submit"
        disabled={submitting}
        class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
      >
        {submitting ? "Saving…" : "Save login"}
      </button>
    </form>
  );
}
