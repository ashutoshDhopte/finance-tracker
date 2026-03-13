"use client";

import { useEffect, useState, useCallback } from "react";
import { api } from "@/lib/api";
import { formatCurrency, formatDate } from "@/lib/utils";
import type { Transaction, Category, Account, CreateTransactionRequest, UpdateTransactionRequest } from "@/lib/types";
import Modal from "@/components/Modal";
import {
  Plus,
  Search,
  Filter,
  ArrowUpRight,
  ArrowDownRight,
  Upload,
  Trash2,
  Pencil,
  ChevronLeft,
  ChevronRight,
  X,
} from "lucide-react";

const PAGE_SIZE = 20;

export default function TransactionsPage() {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(true);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);

  // Filters
  const [search, setSearch] = useState("");
  const [filterCategory, setFilterCategory] = useState("");
  const [filterAccount, setFilterAccount] = useState("");
  const [filterType, setFilterType] = useState("");
  const [filterStart, setFilterStart] = useState("");
  const [filterEnd, setFilterEnd] = useState("");
  const [showFilters, setShowFilters] = useState(false);

  // Modals
  const [showCreate, setShowCreate] = useState(false);
  const [editingTxn, setEditingTxn] = useState<Transaction | null>(null);
  const [showImport, setShowImport] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const loadTransactions = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.getTransactions({
        search: search || undefined,
        category_id: filterCategory || undefined,
        account_id: filterAccount || undefined,
        txn_type: filterType || undefined,
        start_date: filterStart || undefined,
        end_date: filterEnd || undefined,
        limit: PAGE_SIZE,
        offset: page * PAGE_SIZE,
      });
      setTransactions(data.transactions);
      setTotal(data.count);
    } catch {
      // handled by api client
    } finally {
      setLoading(false);
    }
  }, [search, filterCategory, filterAccount, filterType, filterStart, filterEnd, page]);

  useEffect(() => {
    api.getCategories().then((d) => setCategories(d.categories)).catch(() => {});
    api.getAccounts().then((d) => setAccounts(d.accounts)).catch(() => {});
  }, []);

  useEffect(() => {
    loadTransactions();
  }, [loadTransactions]);

  async function handleDelete() {
    if (!deletingId) return;
    try {
      await api.deleteTransaction(deletingId);
      setDeletingId(null);
      loadTransactions();
    } catch {
      // ignore
    }
  }

  function clearFilters() {
    setSearch("");
    setFilterCategory("");
    setFilterAccount("");
    setFilterType("");
    setFilterStart("");
    setFilterEnd("");
    setPage(0);
  }

  const hasFilters = search || filterCategory || filterAccount || filterType || filterStart || filterEnd;

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white">Transactions</h1>
          <p className="text-zinc-400 text-sm mt-1">{total} transaction{total !== 1 ? "s" : ""}</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => setShowImport(true)}
            className="px-3 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-sm rounded-lg flex items-center gap-2 transition-colors cursor-pointer"
          >
            <Upload className="w-4 h-4" />
            Import CSV
          </button>
          <button
            onClick={() => setShowCreate(true)}
            className="px-3 py-2 bg-emerald-600 hover:bg-emerald-500 text-white text-sm rounded-lg flex items-center gap-2 transition-colors cursor-pointer"
          >
            <Plus className="w-4 h-4" />
            Add
          </button>
        </div>
      </div>

      {/* Search & filter bar */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500" />
          <input
            type="text"
            placeholder="Search by merchant or notes..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setPage(0); }}
            className="w-full pl-10 pr-4 py-2.5 bg-zinc-900 border border-zinc-800 rounded-lg text-sm text-white placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent"
          />
        </div>
        <button
          onClick={() => setShowFilters(!showFilters)}
          className={`px-3 py-2.5 border rounded-lg text-sm flex items-center gap-2 transition-colors cursor-pointer ${
            hasFilters ? "bg-emerald-600/15 border-emerald-500/30 text-emerald-400" : "bg-zinc-900 border-zinc-800 text-zinc-400 hover:text-white"
          }`}
        >
          <Filter className="w-4 h-4" />
          Filters
          {hasFilters && (
            <button onClick={(e) => { e.stopPropagation(); clearFilters(); }} className="ml-1 cursor-pointer">
              <X className="w-3 h-3" />
            </button>
          )}
        </button>
      </div>

      {/* Expanded filters */}
      {showFilters && (
        <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-4 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3">
          <div>
            <label className="block text-xs text-zinc-400 mb-1">Category</label>
            <select
              value={filterCategory}
              onChange={(e) => { setFilterCategory(e.target.value); setPage(0); }}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            >
              <option value="">All categories</option>
              {categories.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
            </select>
          </div>
          <div>
            <label className="block text-xs text-zinc-400 mb-1">Account</label>
            <select
              value={filterAccount}
              onChange={(e) => { setFilterAccount(e.target.value); setPage(0); }}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            >
              <option value="">All accounts</option>
              {accounts.map((a) => (
                <option key={a.id} value={a.id}>
                  {a.name} ({a.institution}){a.last_four ? ` ••${a.last_four}` : ""}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs text-zinc-400 mb-1">Type</label>
            <select
              value={filterType}
              onChange={(e) => { setFilterType(e.target.value); setPage(0); }}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            >
              <option value="">All types</option>
              <option value="credit">Income</option>
              <option value="debit">Expense</option>
            </select>
          </div>
          <div>
            <label className="block text-xs text-zinc-400 mb-1">Start date</label>
            <input
              type="date"
              value={filterStart}
              onChange={(e) => { setFilterStart(e.target.value); setPage(0); }}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            />
          </div>
          <div>
            <label className="block text-xs text-zinc-400 mb-1">End date</label>
            <input
              type="date"
              value={filterEnd}
              onChange={(e) => { setFilterEnd(e.target.value); setPage(0); }}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            />
          </div>
        </div>
      )}

      {/* Transaction list */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden">
        {/* Table header */}
        <div className="hidden sm:grid grid-cols-[40%_18%_15%_15%_12%] px-5 py-3 border-b border-zinc-800 text-xs font-medium text-zinc-500 uppercase tracking-wider">
          <span>Transaction</span>
          <span>Category</span>
          <span className="text-right">Amount</span>
          <span className="text-right">Date</span>
          <span />
        </div>

        {loading ? (
          <div className="p-8 flex justify-center">
            <div className="w-6 h-6 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin" />
          </div>
        ) : transactions.length === 0 ? (
          <p className="text-zinc-500 text-sm py-12 text-center">No transactions found</p>
        ) : (
          <div className="divide-y divide-zinc-800">
            {transactions.map((txn) => (
              <div
                key={txn.id}
                className="px-5 py-3 flex flex-col sm:grid sm:grid-cols-[40%_18%_15%_15%_12%] gap-2 items-start sm:items-center hover:bg-zinc-800/40 transition-colors"
              >
                <div className="flex items-center gap-3 min-w-0">
                  <div className={`w-8 h-8 rounded-lg flex items-center justify-center shrink-0 ${txn.txn_type === "credit" ? "bg-emerald-500/15 text-emerald-400" : "bg-red-500/15 text-red-400"}`}>
                    {txn.txn_type === "credit" ? <ArrowDownRight className="w-4 h-4" /> : <ArrowUpRight className="w-4 h-4" />}
                  </div>
                  <div className="min-w-0">
                    <p className="text-sm font-medium text-white break-words">{txn.merchant_name || "Unknown"}</p>
                    {txn.notes && <p className="text-xs text-zinc-500 break-words">{txn.notes}</p>}
                  </div>
                </div>
                <span className="text-xs text-zinc-400 bg-zinc-800 px-2 py-1 rounded-md break-words">
                  {txn.category_name || "—"}
                </span>
                <span className={`text-right text-sm font-mono font-medium ${txn.txn_type === "credit" ? "text-emerald-400" : "text-red-400"}`}>
                  {txn.txn_type === "credit" ? "+" : "-"}{formatCurrency(txn.amount)}
                </span>
                <span className="text-right text-xs text-zinc-500">{formatDate(txn.transaction_date)}</span>
                <div className="flex justify-end gap-1">
                  <button
                    onClick={() => setEditingTxn(txn)}
                    className="p-1.5 text-zinc-500 hover:text-white transition-colors cursor-pointer"
                  >
                    <Pencil className="w-3.5 h-3.5" />
                  </button>
                  <button
                    onClick={() => setDeletingId(txn.id)}
                    className="p-1.5 text-zinc-500 hover:text-red-400 transition-colors cursor-pointer"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Pagination */}
        {total > PAGE_SIZE && (
          <div className="px-5 py-3 border-t border-zinc-800 flex items-center justify-between">
            <span className="text-xs text-zinc-500">
              Showing {page * PAGE_SIZE + 1}–{Math.min((page + 1) * PAGE_SIZE, total)} of {total}
            </span>
            <div className="flex gap-1">
              <button
                onClick={() => setPage((p) => Math.max(0, p - 1))}
                disabled={page === 0}
                className="p-1.5 text-zinc-400 hover:text-white disabled:text-zinc-700 cursor-pointer disabled:cursor-not-allowed"
              >
                <ChevronLeft className="w-4 h-4" />
              </button>
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={(page + 1) * PAGE_SIZE >= total}
                className="p-1.5 text-zinc-400 hover:text-white disabled:text-zinc-700 cursor-pointer disabled:cursor-not-allowed"
              >
                <ChevronRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Create modal */}
      <TransactionFormModal
        open={showCreate}
        onClose={() => setShowCreate(false)}
        categories={categories}
        accounts={accounts}
        onSaved={loadTransactions}
      />

      {/* Edit modal */}
      {editingTxn && (
        <TransactionFormModal
          open
          onClose={() => setEditingTxn(null)}
          categories={categories}
          accounts={accounts}
          transaction={editingTxn}
          onSaved={loadTransactions}
        />
      )}

      {/* Delete confirmation */}
      <Modal open={!!deletingId} onClose={() => setDeletingId(null)} title="Delete Transaction">
        <p className="text-zinc-300 text-sm mb-4">Are you sure you want to delete this transaction? This cannot be undone.</p>
        <div className="flex justify-end gap-3">
          <button onClick={() => setDeletingId(null)} className="px-4 py-2 bg-zinc-800 text-zinc-300 rounded-lg text-sm hover:bg-zinc-700 cursor-pointer">
            Cancel
          </button>
          <button onClick={handleDelete} className="px-4 py-2 bg-red-600 text-white rounded-lg text-sm hover:bg-red-500 cursor-pointer">
            Delete
          </button>
        </div>
      </Modal>

      {/* CSV import modal */}
      <CSVImportModal open={showImport} onClose={() => setShowImport(false)} onDone={loadTransactions} accounts={accounts} />
    </div>
  );
}

function TransactionFormModal({
  open,
  onClose,
  categories,
  accounts,
  transaction,
  onSaved,
}: {
  open: boolean;
  onClose: () => void;
  categories: Category[];
  accounts: Account[];
  transaction?: Transaction;
  onSaved: () => void;
}) {
  const isEdit = !!transaction;
  const [form, setForm] = useState<CreateTransactionRequest>({
    amount: transaction?.amount ?? 0,
    merchant_name: transaction?.merchant_name ?? "",
    category_id: transaction?.category_id ?? "",
    account_id: transaction?.account_id ?? "",
    transaction_date: transaction?.transaction_date?.split("T")[0] ?? new Date().toISOString().split("T")[0],
    txn_type: transaction?.txn_type ?? "debit",
    notes: transaction?.notes ?? "",
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError("");
    try {
      if (isEdit && transaction) {
        const update: UpdateTransactionRequest = {};
        if (form.amount !== transaction.amount) update.amount = form.amount;
        if (form.merchant_name !== transaction.merchant_name) update.merchant_name = form.merchant_name || undefined;
        if (form.category_id !== transaction.category_id) update.category_id = form.category_id || undefined;
        if (form.account_id !== transaction.account_id) update.account_id = form.account_id || undefined;
        if (form.transaction_date !== transaction.transaction_date?.split("T")[0]) update.transaction_date = form.transaction_date;
        if (form.txn_type !== transaction.txn_type) update.txn_type = form.txn_type;
        if (form.notes !== transaction.notes) update.notes = form.notes || undefined;
        await api.updateTransaction(transaction.id, update);
      } else {
        await api.createTransaction(form);
      }
      onClose();
      onSaved();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save");
    } finally {
      setSaving(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={isEdit ? "Edit Transaction" : "Add Transaction"}>
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && <p className="text-red-400 text-sm bg-red-500/10 rounded-lg px-3 py-2">{error}</p>}

        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="block text-xs text-zinc-400 mb-1">Amount</label>
            <input
              type="number"
              step="0.01"
              value={form.amount || ""}
              onChange={(e) => setForm({ ...form, amount: parseFloat(e.target.value) || 0 })}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
              required
            />
          </div>
          <div>
            <label className="block text-xs text-zinc-400 mb-1">Type</label>
            <select
              value={form.txn_type}
              onChange={(e) => setForm({ ...form, txn_type: e.target.value as "credit" | "debit" })}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            >
              <option value="debit">Expense</option>
              <option value="credit">Income</option>
            </select>
          </div>
        </div>

        <div>
          <label className="block text-xs text-zinc-400 mb-1">Merchant</label>
          <input
            type="text"
            value={form.merchant_name || ""}
            onChange={(e) => setForm({ ...form, merchant_name: e.target.value })}
            className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            placeholder="e.g. Whole Foods"
          />
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="block text-xs text-zinc-400 mb-1">Category</label>
            <select
              value={form.category_id || ""}
              onChange={(e) => setForm({ ...form, category_id: e.target.value || undefined })}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            >
              <option value="">None</option>
              {categories.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
            </select>
          </div>
          <div>
            <label className="block text-xs text-zinc-400 mb-1">Account</label>
            <select
              value={form.account_id || ""}
              onChange={(e) => setForm({ ...form, account_id: e.target.value || undefined })}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            >
              <option value="">None</option>
              {accounts.map((a) => <option key={a.id} value={a.id}>{a.name} ({a.institution})</option>)}
            </select>
          </div>
        </div>

        <div>
          <label className="block text-xs text-zinc-400 mb-1">Date</label>
          <input
            type="date"
            value={form.transaction_date}
            onChange={(e) => setForm({ ...form, transaction_date: e.target.value })}
            className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            required
          />
        </div>

        <div>
          <label className="block text-xs text-zinc-400 mb-1">Notes</label>
          <textarea
            value={form.notes || ""}
            onChange={(e) => setForm({ ...form, notes: e.target.value || undefined })}
            rows={2}
            className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500 resize-none"
            placeholder="Optional notes..."
          />
        </div>

        <div className="flex justify-end gap-3 pt-2">
          <button type="button" onClick={onClose} className="px-4 py-2 bg-zinc-800 text-zinc-300 rounded-lg text-sm hover:bg-zinc-700 cursor-pointer">
            Cancel
          </button>
          <button type="submit" disabled={saving} className="px-4 py-2 bg-emerald-600 text-white rounded-lg text-sm hover:bg-emerald-500 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed">
            {saving ? "Saving..." : isEdit ? "Update" : "Create"}
          </button>
        </div>
      </form>
    </Modal>
  );
}

function CSVImportModal({ open, onClose, onDone, accounts }: { open: boolean; onClose: () => void; onDone: () => void; accounts: Account[] }) {
  const [file, setFile] = useState<File | null>(null);
  const [accountId, setAccountId] = useState("");
  const [importing, setImporting] = useState(false);
  const [result, setResult] = useState<{ imported: number; skipped: number; failed: number } | null>(null);
  const [error, setError] = useState("");

  async function handleImport() {
    if (!file) return;
    setImporting(true);
    setError("");
    try {
      const res = await api.importCSV(file, accountId || undefined);
      setResult(res);
      onDone();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Import failed");
    } finally {
      setImporting(false);
    }
  }

  function handleClose() {
    setFile(null);
    setAccountId("");
    setResult(null);
    setError("");
    onClose();
  }

  return (
    <Modal open={open} onClose={handleClose} title="Import CSV">
      {result ? (
        <div className="space-y-3">
          <div className="bg-emerald-500/10 border border-emerald-500/20 rounded-lg p-4 text-sm">
            <p className="text-emerald-300 font-medium">Import complete</p>
            <p className="text-emerald-200/80 mt-1">
              {result.imported} imported, {result.skipped} skipped, {result.failed} failed
            </p>
          </div>
          <div className="flex justify-end">
            <button onClick={handleClose} className="px-4 py-2 bg-zinc-800 text-zinc-300 rounded-lg text-sm hover:bg-zinc-700 cursor-pointer">
              Close
            </button>
          </div>
        </div>
      ) : (
        <div className="space-y-4">
          <p className="text-zinc-400 text-sm">Upload a CSV file from Chase or PNC. The format will be auto-detected.</p>
          {error && <p className="text-red-400 text-sm bg-red-500/10 rounded-lg px-3 py-2">{error}</p>}

          <div>
            <label className="block text-xs text-zinc-400 mb-1">Bank account</label>
            <select
              value={accountId}
              onChange={(e) => setAccountId(e.target.value)}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            >
              <option value="">None</option>
              {accounts.map((a) => (
                <option key={a.id} value={a.id}>
                  {a.name} ({a.institution}){a.last_four ? ` ••${a.last_four}` : ""}
                </option>
              ))}
            </select>
          </div>

          <div
            className="border-2 border-dashed border-zinc-700 rounded-xl p-8 text-center hover:border-zinc-500 transition-colors cursor-pointer"
            onClick={() => document.getElementById("csv-input")?.click()}
          >
            <Upload className="w-8 h-8 text-zinc-500 mx-auto mb-2" />
            <p className="text-sm text-zinc-400">
              {file ? file.name : "Click to select a CSV file"}
            </p>
            <input
              id="csv-input"
              type="file"
              accept=".csv"
              className="hidden"
              onChange={(e) => setFile(e.target.files?.[0] || null)}
            />
          </div>
          <div className="flex justify-end gap-3">
            <button onClick={handleClose} className="px-4 py-2 bg-zinc-800 text-zinc-300 rounded-lg text-sm hover:bg-zinc-700 cursor-pointer">
              Cancel
            </button>
            <button
              onClick={handleImport}
              disabled={!file || importing}
              className="px-4 py-2 bg-emerald-600 text-white rounded-lg text-sm hover:bg-emerald-500 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
            >
              {importing ? "Importing..." : "Import"}
            </button>
          </div>
        </div>
      )}
    </Modal>
  );
}
