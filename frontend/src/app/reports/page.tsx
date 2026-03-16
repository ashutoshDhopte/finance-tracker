"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import { formatCurrency, getCurrentMonth, formatMonth, getMonthRange, buildTransactionUrl } from "@/lib/utils";
import type { ReportSummary, TrendPoint, CategorySummary, Account } from "@/lib/types";
import { ChevronLeft, ChevronRight, TrendingUp, TrendingDown, Wallet, ArrowLeftRight, Filter } from "lucide-react";
import CategoryIcon from "@/components/CategoryIcon";
import {
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Tooltip,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Legend,
  AreaChart,
  Area,
} from "recharts";

type ReportTab = "monthly" | "biweekly" | "trends";

export default function ReportsPage() {
  const [tab, setTab] = useState<ReportTab>("monthly");
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [selectedAccount, setSelectedAccount] = useState("");

  useEffect(() => {
    api.getAccounts().then((d) => setAccounts(d.accounts)).catch(() => {});
  }, []);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between flex-wrap gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white">Reports</h1>
          <p className="text-zinc-400 text-sm mt-1">Financial insights and breakdowns</p>
        </div>
        <div className="flex items-center gap-2">
          <Filter className="w-4 h-4 text-zinc-400" />
          <select
            value={selectedAccount}
            onChange={(e) => setSelectedAccount(e.target.value)}
            className="px-3 py-1.5 bg-zinc-900 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500 min-w-[180px] cursor-pointer"
          >
            <option value="">All accounts</option>
            {accounts.map((acc) => (
              <option key={acc.id} value={acc.id}>
                {acc.name}{acc.last_four ? ` ••${acc.last_four}` : ""}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 bg-zinc-900 border border-zinc-800 rounded-lg p-1 w-fit">
        {(["monthly", "biweekly", "trends"] as ReportTab[]).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors capitalize cursor-pointer ${
              tab === t ? "bg-zinc-800 text-white" : "text-zinc-400 hover:text-white"
            }`}
          >
            {t === "biweekly" ? "Bi-weekly" : t}
          </button>
        ))}
      </div>

      {tab === "monthly" && <MonthlyReport accountId={selectedAccount} />}
      {tab === "biweekly" && <BiweeklyReport accountId={selectedAccount} />}
      {tab === "trends" && <TrendsReport accountId={selectedAccount} />}
    </div>
  );
}

function MonthlyReport({ accountId }: { accountId: string }) {
  const [month, setMonth] = useState(getCurrentMonth());
  const [summary, setSummary] = useState<ReportSummary | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    api.getMonthlyReport(month, accountId || undefined).then(setSummary).catch(() => {}).finally(() => setLoading(false));
  }, [month, accountId]);

  function prevMonth() {
    const [y, m] = month.split("-").map(Number);
    const d = new Date(y, m - 2);
    setMonth(`${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`);
  }

  function nextMonth() {
    const [y, m] = month.split("-").map(Number);
    const d = new Date(y, m);
    setMonth(`${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`);
  }

  if (loading) return <ReportSkeleton />;

  return (
    <div className="space-y-6">
      {/* Month selector */}
      <div className="flex items-center gap-4">
        <button onClick={prevMonth} className="p-2 text-zinc-400 hover:text-white bg-zinc-900 border border-zinc-800 rounded-lg cursor-pointer">
          <ChevronLeft className="w-4 h-4" />
        </button>
        <span className="text-lg font-semibold text-white min-w-[140px] text-center">{formatMonth(month)}</span>
        <button onClick={nextMonth} className="p-2 text-zinc-400 hover:text-white bg-zinc-900 border border-zinc-800 rounded-lg cursor-pointer">
          <ChevronRight className="w-4 h-4" />
        </button>
      </div>

      {summary && (() => {
        const range = getMonthRange(month);
        return <ReportContent summary={summary} startDate={range.start} endDate={range.end} />;
      })()}
    </div>
  );
}

function BiweeklyReport({ accountId }: { accountId: string }) {
  const today = new Date();
  const [start, setStart] = useState(() => {
    const d = new Date(today.getFullYear(), today.getMonth(), today.getDate() >= 15 ? 15 : 1);
    return d.toISOString().split("T")[0];
  });
  const [end, setEnd] = useState(() => {
    const d = today.getDate() >= 15
      ? new Date(today.getFullYear(), today.getMonth() + 1, 0)
      : new Date(today.getFullYear(), today.getMonth(), 14);
    return d.toISOString().split("T")[0];
  });
  const [summary, setSummary] = useState<ReportSummary | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    api.getBiweeklyReport(start, end, accountId || undefined).then(setSummary).catch(() => {}).finally(() => setLoading(false));
  }, [start, end, accountId]);

  if (loading) return <ReportSkeleton />;

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-end gap-4">
        <div>
          <label className="block text-xs text-zinc-400 mb-1">Start date</label>
          <input
            type="date"
            value={start}
            onChange={(e) => setStart(e.target.value)}
            className="px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
          />
        </div>
        <div>
          <label className="block text-xs text-zinc-400 mb-1">End date</label>
          <input
            type="date"
            value={end}
            onChange={(e) => setEnd(e.target.value)}
            className="px-3 py-2 bg-zinc-900 border border-zinc-800 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
          />
        </div>
      </div>

      {summary && <ReportContent summary={summary} startDate={start} endDate={end} />}
    </div>
  );
}

function TrendsReport({ accountId }: { accountId: string }) {
  const [months, setMonths] = useState(6);
  const [trends, setTrends] = useState<TrendPoint[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    api.getTrends(months, accountId || undefined).then((d) => setTrends(d.trends)).catch(() => {}).finally(() => setLoading(false));
  }, [months, accountId]);

  const chartData = trends.map((t) => ({
    month: formatMonth(t.month),
    Income: t.total_income,
    Expenses: t.total_expenses,
    Net: t.net,
  }));

  if (loading) return <ReportSkeleton />;

  return (
    <div className="space-y-6">
      <div className="flex gap-2">
        {[3, 6, 12].map((m) => (
          <button
            key={m}
            onClick={() => setMonths(m)}
            className={`px-3 py-1.5 rounded-lg text-sm cursor-pointer ${
              months === m ? "bg-emerald-600/20 text-emerald-400 border border-emerald-500/30" : "bg-zinc-900 border border-zinc-800 text-zinc-400 hover:text-white"
            }`}
          >
            {m}mo
          </button>
        ))}
      </div>

      {/* Income vs Expenses bar chart */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
        <h3 className="text-sm font-semibold text-zinc-300 mb-4">Income vs Expenses</h3>
        {chartData.length > 0 ? (
          <div className="h-72">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
                <XAxis dataKey="month" tick={{ fill: "#71717a", fontSize: 12 }} tickLine={false} axisLine={false} />
                <YAxis tick={{ fill: "#71717a", fontSize: 12 }} tickLine={false} axisLine={false} tickFormatter={(v) => `$${(v / 1000).toFixed(0)}k`} />
                <Tooltip
                  formatter={(value) => formatCurrency(Number(value))}
                  contentStyle={{ backgroundColor: "#18181b", border: "1px solid #3f3f46", borderRadius: "8px", color: "#fff" }}
                />
                <Legend wrapperStyle={{ color: "#a1a1aa" }} />
                <Bar dataKey="Income" fill="#10b981" radius={[4, 4, 0, 0]} />
                <Bar dataKey="Expenses" fill="#ef4444" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        ) : (
          <p className="text-zinc-500 text-sm py-12 text-center">No trend data</p>
        )}
      </div>

      {/* Net savings area chart */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
        <h3 className="text-sm font-semibold text-zinc-300 mb-4">Net Savings Trend</h3>
        {chartData.length > 0 ? (
          <div className="h-56">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={chartData}>
                <defs>
                  <linearGradient id="netGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#10b981" stopOpacity={0.3} />
                    <stop offset="100%" stopColor="#10b981" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
                <XAxis dataKey="month" tick={{ fill: "#71717a", fontSize: 12 }} tickLine={false} axisLine={false} />
                <YAxis tick={{ fill: "#71717a", fontSize: 12 }} tickLine={false} axisLine={false} tickFormatter={(v) => formatCurrency(v)} />
                <Tooltip
                  formatter={(value) => formatCurrency(Number(value))}
                  contentStyle={{ backgroundColor: "#18181b", border: "1px solid #3f3f46", borderRadius: "8px", color: "#fff" }}
                />
                <Area type="monotone" dataKey="Net" stroke="#10b981" fill="url(#netGrad)" strokeWidth={2} />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        ) : (
          <p className="text-zinc-500 text-sm py-12 text-center">No trend data</p>
        )}
      </div>
    </div>
  );
}

function ReportContent({ summary, startDate, endDate }: { summary: ReportSummary; startDate?: string; endDate?: string }) {
  const router = useRouter();
  const dateFilters = { start: startDate, end: endDate };

  const pieData = summary.by_category
    .filter((c) => c.total > 0)
    .map((c) => ({ name: c.category_name, value: c.total, color: c.color, icon: c.icon, count: c.count, category_id: c.category_id }));

  return (
    <div className="space-y-6">
      {/* Summary cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <MiniCard label="Income" amount={summary.total_income} icon={<TrendingUp className="w-4 h-4" />} color="emerald" href={buildTransactionUrl({ type: "credit", ...dateFilters })} />
        <MiniCard label="Expenses" amount={summary.total_expenses} icon={<TrendingDown className="w-4 h-4" />} color="red" href={buildTransactionUrl({ type: "debit", ...dateFilters })} />
        <MiniCard label="Transfers" amount={summary.total_transfers} icon={<ArrowLeftRight className="w-4 h-4" />} color="blue" href={buildTransactionUrl({ category_name: "Transfer", ...dateFilters })} />
        <MiniCard label="Net" amount={summary.net} icon={<Wallet className="w-4 h-4" />} color={summary.net >= 0 ? "emerald" : "red"} href={buildTransactionUrl(dateFilters)} />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Pie chart */}
        <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
          <h3 className="text-sm font-semibold text-zinc-300 mb-4">Spending Breakdown</h3>
          {pieData.length > 0 ? (
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={pieData}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={100}
                    paddingAngle={3}
                    dataKey="value"
                    className="cursor-pointer"
                    onClick={(_, index) => {
                      const entry = pieData[index];
                      if (entry) router.push(buildTransactionUrl({ category: entry.category_id, ...dateFilters }));
                    }}
                  >
                    {pieData.map((entry, i) => <Cell key={i} fill={entry.color} />)}
                  </Pie>
                  <Tooltip
                    formatter={(value) => formatCurrency(Number(value))}
                    contentStyle={{ backgroundColor: "#18181b", border: "1px solid #3f3f46", borderRadius: "8px", color: "#fff" }}
                  />
                </PieChart>
              </ResponsiveContainer>
            </div>
          ) : (
            <p className="text-zinc-500 text-sm py-12 text-center">No spending data</p>
          )}
        </div>

        {/* Category breakdown table */}
        <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
          <h3 className="text-sm font-semibold text-zinc-300 mb-4">By Category</h3>
          <div className="space-y-3">
            {pieData.length > 0 ? (
              pieData.map((c) => (
                <Link key={c.name} href={buildTransactionUrl({ category: c.category_id, ...dateFilters })}>
                  <CategoryBar category={c} maxValue={pieData[0].value} />
                </Link>
              ))
            ) : (
              <p className="text-zinc-500 text-sm py-12 text-center">No data</p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

function CategoryBar({ category, maxValue }: { category: { name: string; value: number; color: string; icon: string; count: number }; maxValue: number }) {
  const pct = maxValue > 0 ? (category.value / maxValue) * 100 : 0;
  return (
    <div className="hover:bg-zinc-800/50 rounded-lg p-2 -m-2 transition-colors cursor-pointer">
      <div className="flex items-center justify-between mb-1">
        <div className="flex items-center gap-2">
          <span className="text-base text-zinc-400"><CategoryIcon icon={category.icon} size="w-4 h-4" /></span>
          <span className="text-sm text-zinc-300">{category.name}</span>
          <span className="text-xs text-zinc-600">{category.count} txn{category.count !== 1 ? "s" : ""}</span>
        </div>
        <span className="text-sm font-mono text-zinc-400">{formatCurrency(category.value)}</span>
      </div>
      <div className="h-2 bg-zinc-800 rounded-full overflow-hidden">
        <div className="h-full rounded-full transition-all duration-500" style={{ width: `${pct}%`, backgroundColor: category.color }} />
      </div>
    </div>
  );
}

function MiniCard({ label, amount, icon, color, href }: { label: string; amount: number; icon: React.ReactNode; color: string; href?: string }) {
  const cls = color === "emerald" ? "text-emerald-400" : color === "blue" ? "text-blue-400" : "text-red-400";
  const card = (
    <div className={`bg-zinc-900 border border-zinc-800 rounded-xl p-4 flex items-center justify-between ${href ? "hover:border-zinc-600 transition-colors cursor-pointer" : ""}`}>
      <div>
        <p className="text-xs text-zinc-500">{label}</p>
        <p className={`text-xl font-bold ${cls} mt-1`}>{formatCurrency(amount)}</p>
      </div>
      <div className={cls}>{icon}</div>
    </div>
  );
  return href ? <Link href={href}>{card}</Link> : card;
}

function ReportSkeleton() {
  return (
    <div className="space-y-6 animate-pulse">
      <div className="grid grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => <div key={i} className="h-20 bg-zinc-900 border border-zinc-800 rounded-xl" />)}
      </div>
      <div className="grid grid-cols-2 gap-6">
        <div className="h-72 bg-zinc-900 border border-zinc-800 rounded-xl" />
        <div className="h-72 bg-zinc-900 border border-zinc-800 rounded-xl" />
      </div>
    </div>
  );
}
