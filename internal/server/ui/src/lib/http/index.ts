/**
 * The Method enum contains values for HTTP verbs.
 */
enum Method {
  Get = "GET",
  Put = "PUT",
  Post = "POST",
  Delete = "DELETE",
}

/**
 * The HTTPError type represents an error as returned by the API.
 */
export type HTTPError = {
  /**
   * The error message.
   */
  message: string;
  /**
   * The HTTP status code.
   */
  code: number;
};

/**
 * The HTTPClient class is to be inherited for building HTTP-based clients that make calls to the
 * API.
 */
export abstract class HTTPClient {
  /**
   * Performs an HTTP GET request to the specified path, returning the response.
   * @param path The HTTP endpoint.
   * @protected
   */
  protected async get<Response>(path: string): Promise<Response> {
    return await call<void, Response>(Method.Get, path);
  }

  /**
   * Performs an HTTP POST request to the specified path with the JSON-encoded body, returning the response.
   * @param path The HTTP endpoint.
   * @param body The HTTP request body.
   * @protected
   */
  protected async post<Request, Response>(path: string, body: Request): Promise<Response> {
    return await call<Request, Response>(Method.Post, path, body);
  }
}

/**
 * Performs an HTTP request with an optional body, returning the response/
 * @param method The HTTP method to use.
 * @param path The HTTP endpoint.
 * @param body The HTTP request body.
 */
async function call<Request, Response>(method: Method, path: string, body?: Request): Promise<Response> {
  const response = await fetch(path, {
    method: method,
    body: body ? JSON.stringify(body) : null,
  });

  if (!response.ok) {
    throw (await response.json()) as HTTPError;
  }

  return (await response.json()) as Response;
}
