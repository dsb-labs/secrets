import { useState } from "preact/hooks";
import { Client, InvalidCredentialsError, UnreachableError } from "@/lib/client";

type Props = {
  client: Client;
  onAuthenticated: () => Promise<void>;
};

// Login renders a form that prompts the user to sign in with their email and password. On
// submission, it authenticates against the configured server via the provided client and
// calls onAuthenticated.
export function Login({ client, onAuthenticated }: Props) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    setSubmitting(true);
    setError("");
    try {
      await client.login(email, password);
      await onAuthenticated();
    } catch (err) {
      if (err instanceof InvalidCredentialsError || err instanceof UnreachableError) {
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
      <div>
        <h1 class="text-sm font-semibold text-gray-900 dark:text-white">Sign in to Secrets</h1>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">Enter your account credentials.</p>
      </div>
      <div class="flex flex-col gap-3">
        <input
          type="email"
          value={email}
          onInput={(e) => setEmail((e.target as HTMLInputElement).value)}
          placeholder="Email"
          class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-white"
          required
        />
        <input
          type="password"
          value={password}
          onInput={(e) => setPassword((e.target as HTMLInputElement).value)}
          placeholder="Password"
          class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none dark:border-gray-600 dark:bg-gray-800 dark:text-white"
          required
        />
        {error && <p class="text-xs text-red-500">{error}</p>}
      </div>
      <button
        type="submit"
        disabled={submitting}
        class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
      >
        {submitting ? "Signing in…" : "Sign in"}
      </button>
    </form>
  );
}
