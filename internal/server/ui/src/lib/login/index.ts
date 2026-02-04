import { HTTPClient } from "@/lib/http";

/**
 * The Login type represents login details as returned by the API.
 */
export type Login = {
  /**
   * The login's unique identifier.
   */
  id: string;
  /**
   * The username.
   */
  username: string;
  /**
   * The password.
   */
  password: string;
  /**
   * The domains where this login can be used.
   */
  domains: string[];
};

/**
 * The ListResponse type represents the HTTP response when calling the GET /api/v1/login
 * endpoint.
 */
type ListResponse = {
  /**
   * The account's logins.
   */
  logins: Login[];
};

/**
 * The LoginClient class is an extension of the HTTPClient class that performs API requests for user
 * login records.
 */
export class LoginClient extends HTTPClient {
  /**
   * List the caller's login records.
   */
  async list(): Promise<Login[]> {
    const { logins } = await this.get<ListResponse>("/api/v1/login");

    return logins;
  }
}
