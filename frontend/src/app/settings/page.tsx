"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { formatCurrency, formatDate } from "@/lib/utils";
import type { Account, Alert, Category } from "@/lib/types";
import Modal from "@/components/Modal";
import {
  Plus,
  Building2,
  CreditCard,
  Landmark,
  PiggyBank,
  Bell,
  BellOff,
  Trash2,
  AlertTriangle,
  Mail,
  RefreshCw,
  Check,
} from "lucide-react";

const ACCOUNT_ICONS: Record<string, React.ReactNode> = {
  checking: <Landmark className="w-5 h-5" />,
  savings: <PiggyBank className="w-5 h-5" />,
  credit_card: <CreditCard className="w-5 h-5" />,
  investment: <Building2 className="w-5 h-5" />,
};

export default function SettingsPage() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAddAccount, setShowAddAccount] = useState(false);
  const [showAddAlert, setShowAddAlert] = useState(false);

  async function load() {
    try {
      const [accData, alertData, catData] = await Promise.all([
        api.getAccounts(),
        api.getAlerts(),
        api.getCategories(),
      ]);
      setAccounts(accData.accounts);
      setAlerts(alertData.alerts);
      setCategories(catData.categories);
    } catch {
      // handled
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { load(); }, []);

  async function toggleAlert(alert: Alert) {
    try {
      await api.updateAlert(alert.id, { enabled: !alert.enabled });
      load();
    } catch {
      // ignore
    }
  }

  async function deleteAlert(id: string) {
    try {
      await api.deleteAlert(id);
      load();
    } catch {
      // ignore
    }
  }

  if (loading) {
    return (
      <div className="space-y-6 animate-pulse">
        <div className="h-7 w-28 bg-zinc-800 rounded" />
        <div className="h-48 bg-zinc-900 border border-zinc-800 rounded-xl" />
        <div className="h-48 bg-zinc-900 border border-zinc-800 rounded-xl" />
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-white">Settings</h1>
        <p className="text-zinc-400 text-sm mt-1">Manage accounts and alerts</p>
      </div>

      {/* Accounts section */}
      <section>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-white">Accounts</h2>
          <button
            onClick={() => setShowAddAccount(true)}
            className="px-3 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-white text-sm rounded-lg flex items-center gap-2 transition-colors cursor-pointer"
          >
            <Plus className="w-4 h-4" />
            Add
          </button>
        </div>

        {accounts.length === 0 ? (
          <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-8 text-center">
            <Building2 className="w-10 h-10 text-zinc-700 mx-auto mb-3" />
            <p className="text-zinc-400 text-sm">No accounts added yet</p>
            <p className="text-zinc-600 text-xs mt-1">Add your bank accounts to organize transactions</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {accounts.map((acc) => (
              <div key={acc.id} className="bg-zinc-900 border border-zinc-800 rounded-xl p-4 hover:border-zinc-700 transition-colors">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-lg bg-zinc-800 flex items-center justify-center text-zinc-400">
                    {ACCOUNT_ICONS[acc.account_type] || <Building2 className="w-5 h-5" />}
                  </div>
                  <div className="flex-1">
                    <p className="text-sm font-medium text-white">{acc.name}</p>
                    <p className="text-xs text-zinc-500">{acc.institution} &middot; {acc.account_type.replace("_", " ")}</p>
                  </div>
                </div>
                {acc.last_synced_at && (
                  <p className="text-xs text-zinc-600 mt-3">Last synced: {formatDate(acc.last_synced_at)}</p>
                )}
              </div>
            ))}
          </div>
        )}
      </section>

      {/* Alerts section */}
      <section>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-white">Spending Alerts</h2>
          <button
            onClick={() => setShowAddAlert(true)}
            className="px-3 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-white text-sm rounded-lg flex items-center gap-2 transition-colors cursor-pointer"
          >
            <Plus className="w-4 h-4" />
            Add
          </button>
        </div>

        {alerts.length === 0 ? (
          <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-8 text-center">
            <Bell className="w-10 h-10 text-zinc-700 mx-auto mb-3" />
            <p className="text-zinc-400 text-sm">No alerts configured</p>
            <p className="text-zinc-600 text-xs mt-1">Set spending limits to get notified when you overspend</p>
          </div>
        ) : (
          <div className="space-y-3">
            {alerts.map((alert) => (
              <div key={alert.id} className={`bg-zinc-900 border rounded-xl p-4 flex items-center gap-4 ${alert.enabled ? "border-zinc-800" : "border-zinc-800/50 opacity-60"}`}>
                <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${alert.enabled ? "bg-amber-500/15 text-amber-400" : "bg-zinc-800 text-zinc-500"}`}>
                  <AlertTriangle className="w-5 h-5" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-white">{alert.name}</p>
                  <p className="text-xs text-zinc-500">
                    {formatCurrency(alert.threshold)} / {alert.period}
                    {alert.category_name && ` in ${alert.category_name}`}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => toggleAlert(alert)}
                    className={`p-2 rounded-lg transition-colors cursor-pointer ${alert.enabled ? "text-amber-400 hover:bg-amber-500/15" : "text-zinc-500 hover:bg-zinc-800"}`}
                    title={alert.enabled ? "Disable alert" : "Enable alert"}
                  >
                    {alert.enabled ? <Bell className="w-4 h-4" /> : <BellOff className="w-4 h-4" />}
                  </button>
                  <button
                    onClick={() => deleteAlert(alert.id)}
                    className="p-2 text-zinc-500 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors cursor-pointer"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      {/* Email Sync section */}
      <SyncSection />

      {/* API info */}
      <section>
        <h2 className="text-lg font-semibold text-white mb-4">API Connection</h2>
        <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <div className="flex items-center gap-2 mb-2">
            <div className="w-2 h-2 rounded-full bg-emerald-500" />
            <span className="text-sm text-zinc-300">Connected</span>
          </div>
          <p className="text-xs text-zinc-500 font-mono">
            {process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1"}
          </p>
        </div>
      </section>

      {/* Add account modal */}
      <AddAccountModal open={showAddAccount} onClose={() => setShowAddAccount(false)} onCreated={load} />

      {/* Add alert modal */}
      <AddAlertModal open={showAddAlert} onClose={() => setShowAddAlert(false)} categories={categories} onCreated={load} />
    </div>
  );
}

function AddAccountModal({ open, onClose, onCreated }: { open: boolean; onClose: () => void; onCreated: () => void }) {
  const [name, setName] = useState("");
  const [institution, setInstitution] = useState("");
  const [accountType, setAccountType] = useState("checking");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError("");
    try {
      await api.createAccount({ name, institution, account_type: accountType });
      setName("");
      setInstitution("");
      setAccountType("checking");
      onClose();
      onCreated();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create");
    } finally {
      setSaving(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Add Account">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && <p className="text-red-400 text-sm bg-red-500/10 rounded-lg px-3 py-2">{error}</p>}

        <div>
          <label className="block text-xs text-zinc-400 mb-1">Account name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            placeholder="e.g. PNC Checking"
            required
          />
        </div>

        <div>
          <label className="block text-xs text-zinc-400 mb-1">Institution</label>
          <input
            type="text"
            value={institution}
            onChange={(e) => setInstitution(e.target.value)}
            className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            placeholder="e.g. PNC Bank"
            required
          />
        </div>

        <div>
          <label className="block text-xs text-zinc-400 mb-1">Type</label>
          <select
            value={accountType}
            onChange={(e) => setAccountType(e.target.value)}
            className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
          >
            <option value="checking">Checking</option>
            <option value="savings">Savings</option>
            <option value="credit_card">Credit Card</option>
            <option value="investment">Investment</option>
          </select>
        </div>

        <div className="flex justify-end gap-3 pt-2">
          <button type="button" onClick={onClose} className="px-4 py-2 bg-zinc-800 text-zinc-300 rounded-lg text-sm hover:bg-zinc-700 cursor-pointer">
            Cancel
          </button>
          <button type="submit" disabled={saving} className="px-4 py-2 bg-emerald-600 text-white rounded-lg text-sm hover:bg-emerald-500 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed">
            {saving ? "Adding..." : "Add Account"}
          </button>
        </div>
      </form>
    </Modal>
  );
}

function SyncSection() {
  const [syncing, setSyncing] = useState(false);
  const [days, setDays] = useState(30);
  const [result, setResult] = useState<{ imported: number; skipped: number; failed: number } | null>(null);
  const [error, setError] = useState("");

  async function handleSync() {
    setSyncing(true);
    setError("");
    setResult(null);
    try {
      const res = await api.syncGmail(days);
      setResult(res);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Sync failed");
    } finally {
      setSyncing(false);
    }
  }

  return (
    <section>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-white">Email Sync</h2>
      </div>
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5 space-y-4">
        <div className="flex items-start gap-3">
          <div className="w-10 h-10 rounded-lg bg-blue-500/15 flex items-center justify-center text-blue-400 shrink-0">
            <Mail className="w-5 h-5" />
          </div>
          <div>
            <p className="text-sm font-medium text-white">Sync bank emails from Gmail</p>
            <p className="text-xs text-zinc-500 mt-0.5">
              Fetches bank alert emails, parses them with AI, and imports as transactions.
              Duplicates are automatically skipped.
            </p>
          </div>
        </div>

        <div className="flex flex-col sm:flex-row gap-3 items-start sm:items-end">
          <div>
            <label className="block text-xs text-zinc-400 mb-1">Time range</label>
            <select
              value={days}
              onChange={(e) => setDays(Number(e.target.value))}
              disabled={syncing}
              className="px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
            >
              <option value={7}>Last 7 days</option>
              <option value={14}>Last 14 days</option>
              <option value={30}>Last 30 days</option>
              <option value={60}>Last 60 days</option>
              <option value={90}>Last 90 days</option>
            </select>
          </div>
          <button
            onClick={handleSync}
            disabled={syncing}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm rounded-lg flex items-center gap-2 transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {syncing ? (
              <RefreshCw className="w-4 h-4 animate-spin" />
            ) : (
              <RefreshCw className="w-4 h-4" />
            )}
            {syncing ? "Syncing..." : "Sync Emails"}
          </button>
        </div>

        {result && (
          <div className="bg-emerald-500/10 border border-emerald-500/20 rounded-lg p-3 flex items-start gap-2">
            <Check className="w-4 h-4 text-emerald-400 mt-0.5 shrink-0" />
            <div className="text-sm">
              <p className="text-emerald-300 font-medium">Sync complete</p>
              <p className="text-emerald-200/80 text-xs mt-0.5">
                {result.imported} imported, {result.skipped} duplicates skipped, {result.failed} failed
              </p>
            </div>
          </div>
        )}

        {error && (
          <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-3 text-sm text-red-300">
            {error}
          </div>
        )}
      </div>
    </section>
  );
}

function AddAlertModal({
  open,
  onClose,
  categories,
  onCreated,
}: { open: boolean; onClose: () => void; categories: Category[]; onCreated: () => void }) {
  const [name, setName] = useState("");
  const [categoryId, setCategoryId] = useState("");
  const [threshold, setThreshold] = useState("");
  const [period, setPeriod] = useState("monthly");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError("");
    try {
      await api.createAlert({
        name,
        category_id: categoryId || undefined,
        threshold: parseFloat(threshold),
        period,
      });
      setName("");
      setCategoryId("");
      setThreshold("");
      setPeriod("monthly");
      onClose();
      onCreated();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create");
    } finally {
      setSaving(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="New Spending Alert">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && <p className="text-red-400 text-sm bg-red-500/10 rounded-lg px-3 py-2">{error}</p>}

        <div>
          <label className="block text-xs text-zinc-400 mb-1">Alert name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            placeholder="e.g. Dining budget exceeded"
            required
          />
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="block text-xs text-zinc-400 mb-1">Threshold ($)</label>
            <input
              type="number"
              step="0.01"
              value={threshold}
              onChange={(e) => setThreshold(e.target.value)}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
              placeholder="500"
              required
            />
          </div>
          <div>
            <label className="block text-xs text-zinc-400 mb-1">Period</label>
            <select
              value={period}
              onChange={(e) => setPeriod(e.target.value)}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            >
              <option value="daily">Daily</option>
              <option value="weekly">Weekly</option>
              <option value="biweekly">Bi-weekly</option>
              <option value="monthly">Monthly</option>
            </select>
          </div>
        </div>

        <div>
          <label className="block text-xs text-zinc-400 mb-1">Category (optional)</label>
          <select
            value={categoryId}
            onChange={(e) => setCategoryId(e.target.value)}
            className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
          >
            <option value="">All categories</option>
            {categories.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
          </select>
        </div>

        <div className="flex justify-end gap-3 pt-2">
          <button type="button" onClick={onClose} className="px-4 py-2 bg-zinc-800 text-zinc-300 rounded-lg text-sm hover:bg-zinc-700 cursor-pointer">
            Cancel
          </button>
          <button type="submit" disabled={saving} className="px-4 py-2 bg-emerald-600 text-white rounded-lg text-sm hover:bg-emerald-500 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed">
            {saving ? "Creating..." : "Create Alert"}
          </button>
        </div>
      </form>
    </Modal>
  );
}
