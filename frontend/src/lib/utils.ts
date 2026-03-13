export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(amount);
}

export function formatDate(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export function formatDateShort(dateStr: string): string {
  const d = new Date(dateStr);
  return d.toLocaleDateString("en-US", { month: "short", day: "numeric" });
}

export function formatMonth(monthStr: string): string {
  const [year, month] = monthStr.split("-");
  const d = new Date(parseInt(year), parseInt(month) - 1);
  return d.toLocaleDateString("en-US", { month: "short", year: "numeric" });
}

export function getCurrentMonth(): string {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}`;
}

export function getMonthRange(monthStr: string): { start: string; end: string } {
  const [year, month] = monthStr.split("-").map(Number);
  const start = new Date(year, month - 1, 1);
  const end = new Date(year, month, 0);
  return {
    start: start.toISOString().split("T")[0],
    end: end.toISOString().split("T")[0],
  };
}

export function cn(...classes: (string | false | null | undefined)[]): string {
  return classes.filter(Boolean).join(" ");
}

const CATEGORY_COLORS: Record<string, string> = {
  "#4CAF50": "bg-green-500",
  "#FF9800": "bg-orange-500",
  "#F44336": "bg-red-500",
  "#2196F3": "bg-blue-500",
  "#9C27B0": "bg-purple-500",
  "#009688": "bg-teal-500",
  "#795548": "bg-amber-800",
  "#E91E63": "bg-pink-500",
  "#3F51B5": "bg-indigo-500",
  "#00BCD4": "bg-cyan-500",
  "#FF5722": "bg-orange-600",
  "#607688": "bg-slate-500",
  "#8BC34A": "bg-lime-500",
  "#CDDC39": "bg-yellow-500",
  "#9E9E9E": "bg-gray-500",
};

export function getCategoryColorClass(hex: string): string {
  return CATEGORY_COLORS[hex] || "bg-gray-500";
}
