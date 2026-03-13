"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { formatCurrency, getCurrentMonth, getMonthRange } from "@/lib/utils";
import type { Category, CategorySummary } from "@/lib/types";
import Modal from "@/components/Modal";
import { Plus, Tag } from "lucide-react";

export default function CategoriesPage() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [spending, setSpending] = useState<Record<string, CategorySummary>>({});
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);

  async function load() {
    try {
      const { start, end } = getMonthRange(getCurrentMonth());
      const [catData, spendData] = await Promise.all([
        api.getCategories(),
        api.getCategoryReport(start, end),
      ]);
      setCategories(catData.categories);
      const map: Record<string, CategorySummary> = {};
      spendData.categories.forEach((c) => { map[c.category_id] = c; });
      setSpending(map);
    } catch {
      // handled
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { load(); }, []);

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="h-7 w-36 bg-zinc-800 rounded animate-pulse" />
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {Array.from({ length: 9 }).map((_, i) => (
            <div key={i} className="h-24 bg-zinc-900 border border-zinc-800 rounded-xl animate-pulse" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Categories</h1>
          <p className="text-zinc-400 text-sm mt-1">{categories.length} categories</p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="px-3 py-2 bg-emerald-600 hover:bg-emerald-500 text-white text-sm rounded-lg flex items-center gap-2 transition-colors cursor-pointer"
        >
          <Plus className="w-4 h-4" />
          Add Category
        </button>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {categories.map((cat) => {
          const spend = spending[cat.id];
          return (
            <div key={cat.id} className="bg-zinc-900 border border-zinc-800 rounded-xl p-4 hover:border-zinc-700 transition-colors">
              <div className="flex items-center gap-3 mb-3">
                <div
                  className="w-10 h-10 rounded-lg flex items-center justify-center text-lg"
                  style={{ backgroundColor: cat.color + "22" }}
                >
                  {cat.icon || <Tag className="w-5 h-5" style={{ color: cat.color }} />}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-white truncate">{cat.name}</p>
                  {spend ? (
                    <p className="text-xs text-zinc-500">{spend.count} transaction{spend.count !== 1 ? "s" : ""} this month</p>
                  ) : (
                    <p className="text-xs text-zinc-600">No spending this month</p>
                  )}
                </div>
              </div>
              {spend ? (
                <div>
                  <div className="flex justify-between items-baseline mb-1">
                    <span className="text-lg font-bold text-white">{formatCurrency(spend.total)}</span>
                  </div>
                  <div className="h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                    <div className="h-full rounded-full" style={{ backgroundColor: cat.color, width: "100%" }} />
                  </div>
                </div>
              ) : (
                <p className="text-sm text-zinc-600 font-mono">{formatCurrency(0)}</p>
              )}
            </div>
          );
        })}
      </div>

      <CreateCategoryModal
        open={showCreate}
        onClose={() => setShowCreate(false)}
        onCreated={load}
      />
    </div>
  );
}

function CreateCategoryModal({ open, onClose, onCreated }: { open: boolean; onClose: () => void; onCreated: () => void }) {
  const [name, setName] = useState("");
  const [icon, setIcon] = useState("");
  const [color, setColor] = useState("#10b981");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const presetColors = ["#10b981", "#ef4444", "#f59e0b", "#3b82f6", "#8b5cf6", "#ec4899", "#14b8a6", "#f97316", "#6366f1"];

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError("");
    try {
      await api.createCategory({ name, icon: icon || undefined, color });
      setName("");
      setIcon("");
      setColor("#10b981");
      onClose();
      onCreated();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create category");
    } finally {
      setSaving(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="New Category">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && <p className="text-red-400 text-sm bg-red-500/10 rounded-lg px-3 py-2">{error}</p>}

        <div>
          <label className="block text-xs text-zinc-400 mb-1">Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            placeholder="e.g. Pet Expenses"
            required
          />
        </div>

        <div>
          <label className="block text-xs text-zinc-400 mb-1">Icon (emoji)</label>
          <input
            type="text"
            value={icon}
            onChange={(e) => setIcon(e.target.value)}
            className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-lg text-sm text-white focus:outline-none focus:ring-2 focus:ring-emerald-500"
            placeholder="e.g. 🐕"
          />
        </div>

        <div>
          <label className="block text-xs text-zinc-400 mb-1">Color</label>
          <div className="flex gap-2 flex-wrap">
            {presetColors.map((c) => (
              <button
                key={c}
                type="button"
                onClick={() => setColor(c)}
                className={`w-8 h-8 rounded-lg cursor-pointer transition-transform ${color === c ? "scale-110 ring-2 ring-white ring-offset-2 ring-offset-zinc-900" : "hover:scale-105"}`}
                style={{ backgroundColor: c }}
              />
            ))}
          </div>
        </div>

        <div className="flex justify-end gap-3 pt-2">
          <button type="button" onClick={onClose} className="px-4 py-2 bg-zinc-800 text-zinc-300 rounded-lg text-sm hover:bg-zinc-700 cursor-pointer">
            Cancel
          </button>
          <button type="submit" disabled={saving} className="px-4 py-2 bg-emerald-600 text-white rounded-lg text-sm hover:bg-emerald-500 disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed">
            {saving ? "Creating..." : "Create"}
          </button>
        </div>
      </form>
    </Modal>
  );
}
