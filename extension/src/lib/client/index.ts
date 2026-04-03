const HEALTH_PATH = "/api/v1/health";
const AUTH_PATH = "/api/v1/auth";

export class UnreachableError extends Error {
  constructor(url: string) {
    super(`Could not reach Keeper server at ${url}`);
  }
}

export class InvalidCredentialsError extends Error {
  constructor() {
    super("Invalid email or password.");
  }
}

// ping sends a request to the health endpoint of the server at the given base URL. Throws
// UnreachableError if the server cannot be reached or returns a non-OK response.
export async function ping(baseURL: string): Promise<void> {
  let response: Response;
  try {
    response = await fetch(`${baseURL}${HEALTH_PATH}`);
  } catch {
    throw new UnreachableError(baseURL);
  }

  if (!response.ok) {
    throw new UnreachableError(baseURL);
  }
}

// login authenticates against the server at the given base URL with the provided email and password,
// returning the session token on success. Throws InvalidCredentialsError for bad credentials or
// UnreachableError if the server cannot be reached.
export async function login(baseURL: string, email: string, password: string): Promise<string> {
  let response: Response;
  try {
    response = await fetch(`${baseURL}${AUTH_PATH}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });
  } catch {
    throw new UnreachableError(baseURL);
  }

  if (response.status === 400 || response.status === 404) {
    throw new InvalidCredentialsError();
  }

  if (!response.ok) {
    throw new UnreachableError(baseURL);
  }

  const { token } = await response.json();
  return token as string;
}
