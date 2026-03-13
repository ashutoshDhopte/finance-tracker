"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { formatCurrency, getCurrentMonth, formatMonth } from "@/lib/utils";
import type { ReportSummary, TrendPoint, CategorySummary } from "@/lib/types";
import { ChevronLeft, ChevronRight, TrendingUp, TrendingDown, Wallet } from "lucide-react";
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

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-white">Reports</h1>
        <p className="text-zinc-400 text-sm mt-1">Financial insights and breakdowns</p>
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

      {tab === "monthly" && <MonthlyReport />}
      {tab === "biweekly" && <BiweeklyReport />}
      {tab === "trends" && <TrendsReport />}
    </div>
  );
}

function MonthlyReport() {
  const [month, setMonth] = useState(getCurrentMonth());
  const [summary, setSummary] = useState<ReportSummary | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    api.getMonthlyReport(month).then(setSummary).catch(() => {}).finally(() => setLoading(false));
  }, [month]);

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

      {summary && <ReportContent summary={summary} />}
    </div>
  );
}

function BiweeklyReport() {
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
    api.getBiweeklyReport(start, end).then(setSummary).catch(() => {}).finally(() => setLoading(false));
  }, [start, end]);

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

      {summary && <ReportContent summary={summary} />}
    </div>
  );
}

function TrendsReport() {
  const [months, setMonths] = useState(6);
  const [trends, setTrends] = useState<TrendPoint[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    api.getTrends(months).then((d) => setTrends(d.trends)).catch(() => {}).finally(() => setLoading(false));
  }, [months]);

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

function ReportContent({ summary }: { summary: ReportSummary }) {
  const pieData = summary.by_category
    .filter((c) => c.total > 0)
    .map((c) => ({ name: c.category_name, value: c.total, color: c.color, icon: c.icon, count: c.count }));

  return (
    <div className="space-y-6">
      {/* Summary cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <MiniCard label="Income" amount={summary.total_income} icon={<TrendingUp className="w-4 h-4" />} color="emerald" />
        <MiniCard label="Expenses" amount={summary.total_expenses} icon={<TrendingDown className="w-4 h-4" />} color="red" />
        <MiniCard label="Net" amount={summary.net} icon={<Wallet className="w-4 h-4" />} color={summary.net >= 0 ? "emerald" : "red"} />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Pie chart */}
        <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
          <h3 className="text-sm font-semibold text-zinc-300 mb-4">Spending Breakdown</h3>
          {pieData.length > 0 ? (
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie data={pieData} cx="50%" cy="50%" innerRadius={60} outerRadius={100} paddingAngle={3} dataKey="value">
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
              pieData.map((c) => <CategoryBar key={c.name} category={c} maxValue={pieData[0].value} />)
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
    <div>
      <div className="flex items-center justify-between mb-1">
        <div className="flex items-center gap-2">
          <span className="text-base">{category.icon}</span>
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

function MiniCard({ label, amount, icon, color }: { label: string; amount: number; icon: React.ReactNode; color: string }) {
  const cls = color === "emerald" ? "text-emerald-400" : "text-red-400";
  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-4 flex items-center justify-between">
      <div>
        <p className="text-xs text-zinc-500">{label}</p>
        <p className={`text-xl font-bold ${cls} mt-1`}>{formatCurrency(amount)}</p>
      </div>
      <div className={cls}>{icon}</div>
    </div>
  );
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
