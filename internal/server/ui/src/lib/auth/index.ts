import { HTTPClient } from "@/lib/http";

/**
 * The LoginRequest type represents the HTTP request body when calling the /api/v1/auth
 * endpoint.
 */
type LoginRequest = {
  /**
   * The user's email address.
   */
  email: string;
  /**
   * The user's password.
   */
  password: string;
};

/**
 * The LoginResponse type represents the HTTP response body when calling the /api/v1/auth
 * endpoint.
 */
type LoginResponse = {
  /**
   * The authentication token.
   */
  token: string;
};

/**
 * The AuthClient class is an extension of the HTTPClient class that performs authentication
 * specific API calls.
 */
export class AuthClient extends HTTPClient {
  /**
   * Login attempts to authenticate using the provided email and password combination. On success, a cookie
   * should be set.
   * @param email The user's email address.
   * @param password The user's password.
   */
  async login(email: string, password: string): Promise<void> {
    await this.post<LoginRequest, LoginResponse>("/api/v1/auth", {
      email,
      password,
    });
  }
}
