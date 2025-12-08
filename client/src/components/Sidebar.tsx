import { NavLink } from "react-router-dom";
import { useAuthStore } from "../store/authStore";
import { LayoutDashboard, Inbox, Users, Settings, Tag } from "lucide-react";
import clsx from "clsx";

export default function Sidebar() {
  const role = useAuthStore((state) => state.role);
  const isAdminOrManager = role === "ADMIN" || role === "MANAGER";

  const navItems = [
    { to: "/", label: "Dashboard", icon: LayoutDashboard },
    { to: "/inbox", label: "My Inbox", icon: Inbox },
    { to: "/labels", label: "Labels", icon: Tag },
    // Admin/Manager only
    ...(isAdminOrManager
      ? [
          { to: "/admin/inboxes", label: "Inboxes", icon: Settings },
          { to: "/admin/operators", label: "Operators", icon: Users },
        ]
      : []),
  ];

  return (
    <aside className="w-64 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 flex flex-col shrink-0">
      <div className="h-16 flex items-center px-6 border-b border-gray-200 dark:border-gray-700">
        <span className="text-xl font-bold text-indigo-600 dark:text-indigo-400">
          InboxAlloc
        </span>
      </div>
      <nav className="flex-1 p-4 space-y-1">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              clsx(
                "flex items-center px-4 py-2 text-sm font-medium rounded-md transition-colors",
                isActive
                  ? "bg-indigo-50 text-indigo-700 dark:bg-indigo-900/20 dark:text-indigo-300"
                  : "text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
              )
            }
          >
            <item.icon className="mr-3 h-5 w-5" />
            {item.label}
          </NavLink>
        ))}
      </nav>
    </aside>
  );
}
