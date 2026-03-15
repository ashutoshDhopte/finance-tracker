"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import { formatCurrency, formatDate, formatMonth, buildTransactionUrl } from "@/lib/utils";
import type { Transaction, ReportSummary, TrendPoint, TriggeredAlert, AccountSummary } from "@/lib/types";
import {
  TrendingUp,
  TrendingDown,
  Wallet,
  ArrowLeftRight,
  AlertTriangle,
  ArrowUpRight,
  ArrowDownRight,
  Landmark,
  PiggyBank,
  CreditCard,
  Building2,
} from "lucide-react";
import {
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Tooltip,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
} from "recharts";

export default function DashboardPage() {
  const [summary, setSummary] = useState<ReportSummary | null>(null);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [trends, setTrends] = useState<TrendPoint[]>([]);
  const [alerts, setAlerts] = useState<TriggeredAlert[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      try {
        const [summaryData, txnData, trendData, alertData] = await Promise.all([
          api.getSummary(),
          api.getTransactions({ limit: 8 }),
          api.getTrends(6),
          api.checkAlerts(),
        ]);
        setSummary(summaryData);
        setTransactions(txnData.transactions);
        setTrends(trendData.trends);
        setAlerts(alertData.triggered);
      } catch {
        // Will redirect to login if 401
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  if (loading) {
    return <DashboardSkeleton />;
  }

  const router = useRouter();

  const pieData = summary?.by_category
    .filter((c) => c.total > 0)
    .map((c) => ({ name: c.category_name, value: c.total, color: c.color, category_id: c.category_id })) || [];

  const trendData = trends.map((t) => ({
    month: formatMonth(t.month),
    Income: t.total_income,
    Expenses: t.total_expenses,
  }));

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-white">Dashboard</h1>
        <p className="text-zinc-400 text-sm mt-1">All time overview</p>
      </div>

      {/* Triggered alerts */}
      {alerts.length > 0 && (
        <div className="bg-amber-500/10 border border-amber-500/20 rounded-xl p-4">
          <div className="flex items-center gap-2 mb-2">
            <AlertTriangle className="w-5 h-5 text-amber-400" />
            <span className="font-medium text-amber-300">
              {alerts.length} alert{alerts.length > 1 ? "s" : ""} triggered
            </span>
          </div>
          <div className="space-y-1">
            {alerts.map((a) => (
              <p key={a.alert.id} className="text-sm text-amber-200/80">
                {a.alert.name}: {formatCurrency(a.current_spend)} spent
                (limit: {formatCurrency(a.alert.threshold)})
              </p>
            ))}
          </div>
        </div>
      )}

      {/* Summary cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <SummaryCard
          label="Income"
          amount={summary?.total_income ?? 0}
          icon={<TrendingUp className="w-5 h-5" />}
          color="emerald"
          href={buildTransactionUrl({ type: "credit" })}
        />
        <SummaryCard
          label="Expenses"
          amount={summary?.total_expenses ?? 0}
          icon={<TrendingDown className="w-5 h-5" />}
          color="red"
          href={buildTransactionUrl({ type: "debit" })}
        />
        <SummaryCard
          label="Transfers"
          amount={summary?.total_transfers ?? 0}
          icon={<ArrowLeftRight className="w-5 h-5" />}
          color="blue"
          href={buildTransactionUrl({ category_name: "Transfer" })}
        />
        <SummaryCard
          label="Net"
          amount={summary?.net ?? 0}
          icon={<Wallet className="w-5 h-5" />}
          color={(summary?.net ?? 0) >= 0 ? "emerald" : "red"}
          href="/transactions"
        />
      </div>

      {/* Charts row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Spending by category */}
        <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
          <h2 className="text-sm font-semibold text-zinc-300 mb-4">Spending by Category</h2>
          {pieData.length > 0 ? (
            <div className="flex items-center gap-4">
              <div className="w-48 h-48">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={pieData}
                      cx="50%"
                      cy="50%"
                      innerRadius={50}
                      outerRadius={80}
                      paddingAngle={3}
                      dataKey="value"
                      className="cursor-pointer"
                      onClick={(_, index) => {
                        const entry = pieData[index];
                        if (entry) router.push(buildTransactionUrl({ category: entry.category_id }));
                      }}
                    >
                      {pieData.map((entry, i) => (
                        <Cell key={i} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip
                      formatter={(value) => formatCurrency(Number(value))}
                      contentStyle={{
                        backgroundColor: "#18181b",
                        border: "1px solid #3f3f46",
                        borderRadius: "8px",
                        color: "#fff",
                      }}
                    />
                  </PieChart>
                </ResponsiveContainer>
              </div>
              <div className="flex-1 space-y-2">
                {pieData.slice(0, 6).map((c) => (
                  <Link
                    key={c.name}
                    href={buildTransactionUrl({ category: c.category_id })}
                    className="flex items-center justify-between text-sm hover:bg-zinc-800/50 rounded-md px-2 py-1 -mx-2 transition-colors"
                  >
                    <div className="flex items-center gap-2">
                      <div className="w-3 h-3 rounded-full" style={{ backgroundColor: c.color }} />
                      <span className="text-zinc-300">{c.name}</span>
                    </div>
                    <span className="text-zinc-400 font-mono">{formatCurrency(c.value)}</span>
                  </Link>
                ))}
              </div>
            </div>
          ) : (
            <p className="text-zinc-500 text-sm py-8 text-center">No spending data yet</p>
          )}
        </div>

        {/* Trends chart */}
        <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
          <h2 className="text-sm font-semibold text-zinc-300 mb-4">Income vs Expenses</h2>
          {trendData.length > 0 ? (
            <div className="h-48">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={trendData}>
                  <defs>
                    <linearGradient id="incomeGrad" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stopColor="#10b981" stopOpacity={0.3} />
                      <stop offset="100%" stopColor="#10b981" stopOpacity={0} />
                    </linearGradient>
                    <linearGradient id="expenseGrad" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stopColor="#ef4444" stopOpacity={0.3} />
                      <stop offset="100%" stopColor="#ef4444" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
                  <XAxis dataKey="month" tick={{ fill: "#71717a", fontSize: 12 }} tickLine={false} axisLine={false} />
                  <YAxis tick={{ fill: "#71717a", fontSize: 12 }} tickLine={false} axisLine={false} tickFormatter={(v) => `$${(v / 1000).toFixed(0)}k`} />
                  <Tooltip
                    formatter={(value) => formatCurrency(Number(value))}
                    contentStyle={{
                      backgroundColor: "#18181b",
                      border: "1px solid #3f3f46",
                      borderRadius: "8px",
                      color: "#fff",
                    }}
                  />
                  <Area type="monotone" dataKey="Income" stroke="#10b981" fill="url(#incomeGrad)" strokeWidth={2} />
                  <Area type="monotone" dataKey="Expenses" stroke="#ef4444" fill="url(#expenseGrad)" strokeWidth={2} />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          ) : (
            <p className="text-zinc-500 text-sm py-8 text-center">No trend data yet</p>
          )}
        </div>
      </div>

      {/* Account balances */}
      {(summary?.by_account?.length ?? 0) > 0 && (
        <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
          <h2 className="text-sm font-semibold text-zinc-300 mb-4">Account Balances</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            {summary!.by_account.map((acc) => (
              <Link key={acc.account_id} href={buildTransactionUrl({ account: acc.account_id })}>
                <AccountCard account={acc} />
              </Link>
            ))}
          </div>
        </div>
      )}

      {/* Recent transactions */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl">
        <div className="px-5 py-4 border-b border-zinc-800 flex items-center justify-between">
          <h2 className="text-sm font-semibold text-zinc-300">Recent Transactions</h2>
          <a href="/transactions" className="text-xs text-emerald-400 hover:text-emerald-300 transition-colors">
            View all
          </a>
        </div>
        <div className="divide-y divide-zinc-800">
          {transactions.length > 0 ? (
            transactions.map((txn) => (
              <div key={txn.id} className="px-5 py-3 flex items-center justify-between hover:bg-zinc-800/40 transition-colors">
                <div className="flex items-center gap-3">
                  <div className={`w-8 h-8 rounded-lg flex items-center justify-center ${txn.txn_type === "credit" ? "bg-emerald-500/15 text-emerald-400" : "bg-red-500/15 text-red-400"}`}>
                    {txn.txn_type === "credit" ? <ArrowDownRight className="w-4 h-4" /> : <ArrowUpRight className="w-4 h-4" />}
                  </div>
                  <div>
                    <p className="text-sm font-medium text-white">{txn.merchant_name || "Unknown"}</p>
                    <p className="text-xs text-zinc-500">
                      {txn.category_name || "Uncategorized"} &middot; {formatDate(txn.transaction_date)}
                      {txn.account_name && <> &middot; {txn.account_name}</>}
                    </p>
                  </div>
                </div>
                <span className={`text-sm font-mono font-medium ${txn.txn_type === "credit" ? "text-emerald-400" : "text-red-400"}`}>
                  {txn.txn_type === "credit" ? "+" : "-"}{formatCurrency(txn.amount)}
                </span>
              </div>
            ))
          ) : (
            <p className="text-zinc-500 text-sm py-8 text-center">No transactions yet</p>
          )}
        </div>
      </div>
    </div>
  );
}

function SummaryCard({ label, amount, icon, color, href }: { label: string; amount: number; icon: React.ReactNode; color: string; href?: string }) {
  const colorMap: Record<string, { bg: string; text: string; iconBg: string }> = {
    emerald: { bg: "bg-emerald-500/10", text: "text-emerald-400", iconBg: "bg-emerald-500/15" },
    red: { bg: "bg-red-500/10", text: "text-red-400", iconBg: "bg-red-500/15" },
    blue: { bg: "bg-blue-500/10", text: "text-blue-400", iconBg: "bg-blue-500/15" },
  };
  const c = colorMap[color] || colorMap.emerald;

  const card = (
    <div className={`bg-zinc-900 border border-zinc-800 rounded-xl p-5 ${href ? "hover:border-zinc-600 transition-colors cursor-pointer" : ""}`}>
      <div className="flex items-center justify-between mb-3">
        <span className="text-sm text-zinc-400">{label}</span>
        <div className={`w-9 h-9 rounded-lg ${c.iconBg} ${c.text} flex items-center justify-center`}>
          {icon}
        </div>
      </div>
      <p className={`text-2xl font-bold ${c.text}`}>{formatCurrency(amount)}</p>
    </div>
  );

  return href ? <Link href={href}>{card}</Link> : card;
}

const ACCOUNT_TYPE_ICONS: Record<string, React.ReactNode> = {
  checking: <Landmark className="w-5 h-5" />,
  savings: <PiggyBank className="w-5 h-5" />,
  credit_card: <CreditCard className="w-5 h-5" />,
  investment: <Building2 className="w-5 h-5" />,
};

function AccountCard({ account }: { account: AccountSummary }) {
  const netColor = account.net >= 0 ? "text-emerald-400" : "text-red-400";

  return (
    <div className="bg-zinc-800/50 border border-zinc-700/50 rounded-lg p-4 hover:border-zinc-600 transition-colors cursor-pointer">
      <div className="flex items-center gap-3 mb-3">
        <div className="w-9 h-9 rounded-lg bg-zinc-700/50 flex items-center justify-center text-zinc-400">
          {ACCOUNT_TYPE_ICONS[account.account_type] || <Building2 className="w-5 h-5" />}
        </div>
        <div className="min-w-0">
          <p className="text-sm font-medium text-white truncate">
            {account.account_name}
            {account.last_four && <span className="text-zinc-500 font-mono ml-1.5">••{account.last_four}</span>}
          </p>
          <p className="text-xs text-zinc-500">{account.institution}</p>
        </div>
      </div>
      <div className="flex items-baseline justify-between">
        <span className="text-xs text-zinc-500">Net</span>
        <span className={`text-lg font-bold font-mono ${netColor}`}>
          {account.net >= 0 ? "+" : ""}{formatCurrency(account.net)}
        </span>
      </div>
      <div className="flex justify-between mt-1.5 text-xs text-zinc-500">
        <span>In: <span className="text-emerald-400/80 font-mono">{formatCurrency(account.income)}</span></span>
        <span>Out: <span className="text-red-400/80 font-mono">{formatCurrency(account.expenses)}</span></span>
      </div>
    </div>
  );
}

function DashboardSkeleton() {
  return (
    <div className="space-y-6 animate-pulse">
      <div>
        <div className="h-7 w-36 bg-zinc-800 rounded" />
        <div className="h-4 w-28 bg-zinc-800 rounded mt-2" />
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="bg-zinc-900 border border-zinc-800 rounded-xl p-5 h-28" />
        ))}
      </div>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-zinc-900 border border-zinc-800 rounded-xl h-64" />
        <div className="bg-zinc-900 border border-zinc-800 rounded-xl h-64" />
      </div>
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl h-80" />
    </div>
  );
}
