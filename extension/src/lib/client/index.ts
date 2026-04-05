const HEALTH_PATH = "/api/v1/health";
const AUTH_PATH = "/api/v1/auth";
const LOGIN_PATH = "/api/v1/login";

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

export class UnauthorizedError extends Error {
  constructor() {
    super("Session expired. Please sign in again.");
  }
}

export type Login = {
  id: string;
  username: string;
  password: string;
  domains: string[];
  createdAt: string;
  name: string;
};

// KeeperClient is a client for the Keeper API. It encapsulates the server URL and session token,
// centralising error handling for all requests. The token may be set at construction time (when
// restoring from storage) or populated later via login().
export class KeeperClient {
  private _token: string;

  constructor(
    private readonly baseURL: string,
    token: string = "",
  ) {
    this._token = token;
  }

  // token returns the current session token.
  token(): string {
    return this._token;
  }

  // ping checks the reachability of the server. Throws UnreachableError if the server cannot be
  // reached or returns a non-OK response.
  async ping(): Promise<void> {
    let response: Response;
    try {
      response = await fetch(`${this.baseURL}${HEALTH_PATH}`);
    } catch {
      throw new UnreachableError(this.baseURL);
    }

    if (!response.ok) {
      throw new UnreachableError(this.baseURL);
    }
  }

  // login authenticates with the provided email and password, storing the returned session token
  // internally. Throws InvalidCredentialsError for bad credentials or UnreachableError if the
  // server cannot be reached.
  async login(email: string, password: string): Promise<void> {
    let response: Response;
    try {
      response = await fetch(`${this.baseURL}${AUTH_PATH}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
    } catch {
      throw new UnreachableError(this.baseURL);
    }

    if (response.status === 400 || response.status === 404) {
      throw new InvalidCredentialsError();
    }

    if (!response.ok) {
      throw new UnreachableError(this.baseURL);
    }

    const { token } = await response.json();
    this._token = token as string;
  }

  // get sends an authenticated GET request to the given path and returns the parsed response body.
  // Throws UnauthorizedError on 401, or UnreachableError on network failure or any other non-OK response.
  private async get<T>(path: string): Promise<T> {
    let response: Response;
    try {
      response = await fetch(`${this.baseURL}${path}`, {
        headers: { Authorization: `Bearer ${this._token}` },
      });
    } catch {
      throw new UnreachableError(this.baseURL);
    }

    if (response.status === 401) {
      throw new UnauthorizedError();
    }

    if (!response.ok) {
      throw new UnreachableError(this.baseURL);
    }

    return response.json() as Promise<T>;
  }

  // listLogins returns all logins stored for the given domain.
  async listLogins(domain: string): Promise<Login[]> {
    const { logins } = await this.get<{ logins: Login[] }>(`${LOGIN_PATH}?domain=${encodeURIComponent(domain)}`);
    return logins ?? [];
  }

  // getLogin returns the login with the given ID.
  async getLogin(id: string): Promise<Login> {
    const { login } = await this.get<{ login: Login }>(`${LOGIN_PATH}/${encodeURIComponent(id)}`);
    return login;
  }
}
