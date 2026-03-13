import {
  ShoppingCart,
  Utensils,
  Fuel,
  Car,
  ShoppingBag,
  FileText,
  Home,
  HeartPulse,
  Film,
  Repeat,
  DollarSign,
  ArrowRightLeft,
  Landmark,
  AlertTriangle,
  Tag,
  Plane,
  GraduationCap,
  Baby,
  Dumbbell,
  Wifi,
  Gift,
  Briefcase,
  PawPrint,
  Wrench,
  Music,
  Gamepad2,
  BookOpen,
  Coffee,
  type LucideIcon,
} from "lucide-react";

const ICON_MAP: Record<string, LucideIcon> = {
  "shopping-cart": ShoppingCart,
  "utensils": Utensils,
  "fuel": Fuel,
  "car": Car,
  "bag": ShoppingBag,
  "shopping-bag": ShoppingBag,
  "file-text": FileText,
  "home": Home,
  "heart-pulse": HeartPulse,
  "film": Film,
  "repeat": Repeat,
  "dollar-sign": DollarSign,
  "arrow-right-left": ArrowRightLeft,
  "landmark": Landmark,
  "alert-triangle": AlertTriangle,
  "tag": Tag,
  "plane": Plane,
  "graduation-cap": GraduationCap,
  "baby": Baby,
  "dumbbell": Dumbbell,
  "wifi": Wifi,
  "gift": Gift,
  "briefcase": Briefcase,
  "paw-print": PawPrint,
  "wrench": Wrench,
  "music": Music,
  "gamepad-2": Gamepad2,
  "book-open": BookOpen,
  "coffee": Coffee,
};

export default function CategoryIcon({ icon, color, size = "w-5 h-5" }: { icon: string; color?: string; size?: string }) {
  const LucideComponent = ICON_MAP[icon];
  if (LucideComponent) {
    return <LucideComponent className={size} style={color ? { color } : undefined} />;
  }

  // Treat short non-mapped strings as emoji
  if (icon && icon.length <= 4) {
    return <span>{icon}</span>;
  }

  return <Tag className={size} style={color ? { color } : undefined} />;
}
