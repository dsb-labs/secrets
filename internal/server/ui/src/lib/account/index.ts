import { HTTPClient } from "@/lib/http";

/**
 * The Account type represents account details as returned by the API.
 */
export type Account = {
  /**
   * The user's email address.
   */
  email: string;
  /**
   * The user's display name.
   */
  displayName: string;
};

/**
 * The GetResponse type represents the HTTP response when calling the GET /api/v1/account
 * endpoint.
 */
type GetResponse = {
  /**
   * The caller's account details.
   */
  account: Account;
};

/**
 * The AccountClient class is an extension of the HTTPClient class that performs account
 * specific API calls.
 */
export class AccountClient extends HTTPClient {
  /**
   * Find the caller's account details.
   */
  async find(): Promise<Account> {
    const { account } = await this.get<GetResponse>("/api/v1/account");

    return account;
  }
}
