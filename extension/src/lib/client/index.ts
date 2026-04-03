const HEALTH_PATH = "/api/v1/health";

export class UnreachableError extends Error {
  constructor(url: string) {
    super(`Could not reach Keeper server at ${url}`);
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
