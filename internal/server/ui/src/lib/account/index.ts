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
 * The GetAccountResponse type represents the HTTP response when calling the /api/v1/account
 * endpoint.
 */
type GetAccountResponse = {
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
   * Find returns the caller's account details.
   */
  async find(): Promise<Account> {
    const { account } = await this.get<GetAccountResponse>("/api/v1/account");

    return account;
  }
}
