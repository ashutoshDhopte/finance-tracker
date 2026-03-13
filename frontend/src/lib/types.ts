export interface Transaction {
  id: string;
  user_id: string;
  account_id?: string;
  amount: number;
  currency: string;
  merchant_name?: string;
  merchant_raw?: string;
  category_id?: string;
  category_name?: string;
  transaction_date: string;
  posted_date?: string;
  txn_type: "credit" | "debit";
  source: string;
  source_hash?: string;
  ai_confidence?: number;
  raw_text?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateTransactionRequest {
  account_id?: string;
  amount: number;
  currency?: string;
  merchant_name?: string;
  category_id?: string;
  transaction_date: string;
  posted_date?: string;
  txn_type: "credit" | "debit";
  source?: string;
  notes?: string;
}

export interface UpdateTransactionRequest {
  account_id?: string;
  amount?: number;
  merchant_name?: string;
  category_id?: string;
  transaction_date?: string;
  txn_type?: string;
  notes?: string;
}

export interface Category {
  id: string;
  name: string;
  icon: string;
  color: string;
  parent_id?: string;
  created_at: string;
}

export interface Account {
  id: string;
  user_id: string;
  name: string;
  institution: string;
  account_type: "checking" | "savings" | "credit_card" | "investment";
  last_four?: string;
  last_synced_at?: string;
  created_at: string;
  updated_at: string;
}

export interface ReportSummary {
  total_income: number;
  total_expenses: number;
  total_transfers: number;
  net: number;
  by_category: CategorySummary[];
  by_account: AccountSummary[];
}

export interface AccountSummary {
  account_id: string;
  account_name: string;
  institution: string;
  account_type: string;
  last_four?: string;
  income: number;
  expenses: number;
  net: number;
}

export interface CategorySummary {
  category_id: string;
  category_name: string;
  color: string;
  icon: string;
  total: number;
  count: number;
}

export interface TrendPoint {
  month: string;
  total_income: number;
  total_expenses: number;
  net: number;
}

export interface Alert {
  id: string;
  user_id: string;
  name: string;
  category_id?: string;
  category_name?: string;
  threshold: number;
  period: "daily" | "weekly" | "biweekly" | "monthly";
  enabled: boolean;
  last_triggered_at?: string;
  created_at: string;
  updated_at: string;
}

export interface TriggeredAlert {
  alert: Alert;
  current_spend: number;
  exceeded: boolean;
}
