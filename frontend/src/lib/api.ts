import type {
  Transaction,
  CreateTransactionRequest,
  UpdateTransactionRequest,
  Category,
  Account,
  ReportSummary,
  TrendPoint,
  Alert,
  TriggeredAlert,
} from "./types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

class ApiClient {
  private token: string | null = null;

  setToken(token: string | null) {
    this.token = token;
    if (token) {
      localStorage.setItem("auth_token", token);
    } else {
      localStorage.removeItem("auth_token");
    }
  }

  getToken(): string | null {
    if (this.token) return this.token;
    if (typeof window !== "undefined") {
      this.token = localStorage.getItem("auth_token");
    }
    return this.token;
  }

  private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const token = this.getToken();
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...((options.headers as Record<string, string>) || {}),
    };
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    const res = await fetch(`${API_BASE}${path}`, { ...options, headers });

    if (res.status === 401) {
      this.setToken(null);
      if (typeof window !== "undefined") {
        window.location.href = "/login";
      }
      throw new Error("Unauthorized");
    }

    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      throw new Error(body.error || `Request failed: ${res.status}`);
    }

    return res.json();
  }

  // Auth
  async login(username: string, password: string): Promise<{ token: string }> {
    const data = await this.request<{ token: string }>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });
    this.setToken(data.token);
    return data;
  }

  logout() {
    this.setToken(null);
  }

  // Transactions
  async getTransactions(params?: {
    start_date?: string;
    end_date?: string;
    category_id?: string;
    account_id?: string;
    txn_type?: string;
    source?: string;
    search?: string;
    limit?: number;
    offset?: number;
  }): Promise<{ transactions: Transaction[]; count: number }> {
    const query = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([k, v]) => {
        if (v !== undefined && v !== "") query.set(k, String(v));
      });
    }
    const qs = query.toString();
    return this.request(`/transactions${qs ? `?${qs}` : ""}`);
  }

  async getTransaction(id: string): Promise<Transaction> {
    return this.request(`/transactions/${id}`);
  }

  async createTransaction(data: CreateTransactionRequest): Promise<{ id: string }> {
    return this.request("/transactions", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateTransaction(id: string, data: UpdateTransactionRequest): Promise<void> {
    await this.request(`/transactions/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async deleteTransaction(id: string): Promise<void> {
    await this.request(`/transactions/${id}`, { method: "DELETE" });
  }

  // Categories
  async getCategories(): Promise<{ categories: Category[] }> {
    return this.request("/categories");
  }

  async createCategory(data: { name: string; icon?: string; color?: string }): Promise<{ id: string }> {
    return this.request("/categories", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateCategory(id: string, data: { name?: string; icon?: string; color?: string }): Promise<void> {
    await this.request(`/categories/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async deleteCategory(id: string): Promise<void> {
    await this.request(`/categories/${id}`, { method: "DELETE" });
  }

  // Accounts
  async getAccounts(): Promise<{ accounts: Account[] }> {
    return this.request("/accounts");
  }

  async createAccount(data: {
    name: string;
    institution: string;
    account_type: string;
    last_four?: string;
    debit_card_last_four?: string;
  }): Promise<{ id: string }> {
    return this.request("/accounts", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateAccount(id: string, data: {
    name?: string;
    institution?: string;
    account_type?: string;
    last_four?: string | null;
    debit_card_last_four?: string | null;
    inactive_date?: string | null;
  }): Promise<void> {
    await this.request(`/accounts/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async deleteAccount(id: string): Promise<void> {
    await this.request(`/accounts/${id}`, { method: "DELETE" });
  }

  // Reports
  async getSummary(accountId?: string): Promise<ReportSummary> {
    const params = new URLSearchParams();
    if (accountId) params.set("account_id", accountId);
    const qs = params.toString();
    return this.request(`/reports/monthly${qs ? `?${qs}` : ""}`);
  }

  async getMonthlyReport(month: string, accountId?: string): Promise<ReportSummary> {
    const params = new URLSearchParams({ month });
    if (accountId) params.set("account_id", accountId);
    return this.request(`/reports/monthly?${params}`);
  }

  async getBiweeklyReport(start: string, end: string, accountId?: string): Promise<ReportSummary> {
    const params = new URLSearchParams({ start, end });
    if (accountId) params.set("account_id", accountId);
    return this.request(`/reports/biweekly?${params}`);
  }

  async getCategoryReport(from: string, to: string): Promise<{ categories: import("./types").CategorySummary[]; from: string; to: string }> {
    return this.request(`/reports/categories?from=${from}&to=${to}`);
  }

  async getTrends(months?: number, accountId?: string): Promise<{ trends: TrendPoint[]; months: number }> {
    const params = new URLSearchParams();
    if (months) params.set("months", String(months));
    if (accountId) params.set("account_id", accountId);
    const qs = params.toString();
    return this.request(`/reports/trends${qs ? `?${qs}` : ""}`);
  }

  // Alerts
  async getAlerts(): Promise<{ alerts: Alert[] }> {
    return this.request("/alerts");
  }

  async createAlert(data: {
    name: string;
    category_id?: string;
    threshold: number;
    period: string;
  }): Promise<{ id: string }> {
    return this.request("/alerts", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async updateAlert(id: string, data: Partial<Alert>): Promise<void> {
    await this.request(`/alerts/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
  }

  async deleteAlert(id: string): Promise<void> {
    await this.request(`/alerts/${id}`, { method: "DELETE" });
  }

  async checkAlerts(): Promise<{ triggered: TriggeredAlert[]; count: number }> {
    return this.request("/alerts/check");
  }

  // Sync
  async syncGmail(days: number = 30): Promise<{ message: string; days: number; imported: number; skipped: number; failed: number }> {
    return this.request(`/sync/gmail?days=${days}`, { method: "POST" });
  }

  // CSV Import
  async importCSV(file: File, accountId?: string): Promise<{ imported: number; skipped: number; failed: number }> {
    const token = this.getToken();
    const formData = new FormData();
    formData.append("file", file);
    if (accountId) {
      formData.append("account_id", accountId);
    }
    const res = await fetch(`${API_BASE}/transactions/import/csv`, {
      method: "POST",
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: formData,
    });
    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      throw new Error(body.error || "Import failed");
    }
    return res.json();
  }
}

export const api = new ApiClient();
