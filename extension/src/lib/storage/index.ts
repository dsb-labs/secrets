import browser from "webextension-polyfill";

const KEY_SERVER_URL = "serverUrl";
const KEY_TOKEN = "token";

export async function getServerURL(): Promise<string> {
  const result = await browser.storage.local.get(KEY_SERVER_URL);
  return (result[KEY_SERVER_URL] as string) ?? "";
}

export async function setServerURL(url: string): Promise<void> {
  await browser.storage.local.set({ [KEY_SERVER_URL]: url });
}

export async function getToken(): Promise<string> {
  const result = await browser.storage.session.get(KEY_TOKEN);
  return (result[KEY_TOKEN] as string) ?? "";
}

export async function setToken(token: string): Promise<void> {
  await browser.storage.session.set({ [KEY_TOKEN]: token });
}

export async function clearToken(): Promise<void> {
  await browser.storage.session.remove(KEY_TOKEN);
}
