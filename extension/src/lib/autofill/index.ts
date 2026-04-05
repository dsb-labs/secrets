import browser from "webextension-polyfill";

// fillCredentials is injected into the active tab via executeScript. It must be self-contained:
// no imports, no references to module-scope variables. Only built-in DOM APIs are used.
function fillCredentials(username: string, password: string): boolean {
  function fill(el: HTMLInputElement, value: string) {
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value")?.set;
    if (setter) setter.call(el, value);
    else el.value = value;
    el.dispatchEvent(new Event("input", { bubbles: true }));
    el.dispatchEvent(new Event("change", { bubbles: true }));
  }

  const pw = document.querySelector<HTMLInputElement>("input[type='password']");
  if (!pw) return false;

  const scope = pw.closest("form") ?? document;
  const un = scope.querySelector<HTMLInputElement>(
    "input[type='email'], input[autocomplete='username'], input[type='text'], input:not([type])",
  );
  if (un) fill(un, username);
  fill(pw, password);
  return true;
}

// autofill injects credentials into the active tab's login form. It returns true if both fields
// were found and filled, false if no password field was found on the page.
export async function autofill(username: string, password: string): Promise<boolean> {
  const tabs = await browser.tabs.query({ active: true, currentWindow: true });
  const tabId = tabs[0]?.id;
  if (tabId === undefined) return false;

  // Serialize the typed function and invoke it immediately with the credentials.
  // JSON.stringify safely escapes quotes, backslashes, and other special characters in passwords.
  const code = `(${fillCredentials.toString()})(${JSON.stringify(username)}, ${JSON.stringify(password)})`;
  const results = await browser.tabs.executeScript(tabId, { code });
  return results?.[0] === true;
}
