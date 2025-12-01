import { useAuthStore } from "../store/authStore";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import api from "../lib/api";
import { LogOut, Loader2 } from "lucide-react";
import clsx from "clsx";

export default function Header() {
  const { operatorId, logout } = useAuthStore();
  const queryClient = useQueryClient();

  const { data: operator, isLoading } = useQuery({
    queryKey: ["operator", operatorId],
    queryFn: async () => {
      const res = await api.get(`/api/v1/operators/${operatorId}`);
      return res.data;
    },
    enabled: !!operatorId,
    retry: false,
  });

  const statusMutation = useMutation({
    mutationFn: async (newStatus: string) => {
      await api.put("/api/v1/operator/status", { status: newStatus });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["operator", operatorId] });
    },
  });

  const isAvailable = operator?.status === "AVAILABLE";

  const toggleStatus = () => {
    statusMutation.mutate(isAvailable ? "OFFLINE" : "AVAILABLE");
  };

  return (
    <header className="h-16 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between px-6 shrink-0">
      <h1 className="text-lg font-semibold text-gray-900 dark:text-white">
        {/* Breadcrumbs or Page Title could go here */}
      </h1>
      <div className="flex items-center space-x-4">
        <div className="flex items-center space-x-2">
          <button
            onClick={toggleStatus}
            disabled={statusMutation.isPending || isLoading}
            className={clsx(
              "flex items-center px-3 py-1 rounded-full text-sm font-medium border transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500",
              isAvailable
                ? "bg-green-50 text-green-700 border-green-200 dark:bg-green-900/20 dark:text-green-300 dark:border-green-800"
                : "bg-gray-50 text-gray-700 border-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:border-gray-600"
            )}
          >
            {statusMutation.isPending ? (
              <Loader2 className="h-3 w-3 animate-spin mr-2" />
            ) : (
              <div
                className={clsx(
                  "h-2 w-2 rounded-full mr-2",
                  isAvailable ? "bg-green-500" : "bg-gray-400"
                )}
              />
            )}
            {isLoading ? "Loading..." : isAvailable ? "Available" : "Offline"}
          </button>
        </div>
        <div className="h-6 w-px bg-gray-200 dark:bg-gray-700" />
        <div className="flex items-center space-x-3">
          <span className="text-sm font-medium text-gray-700 dark:text-gray-200">
            {operator?.name || operatorId?.slice(0, 8)}
          </span>
          <button
            onClick={logout}
            className="p-2 text-gray-500 hover:text-red-600 dark:text-gray-400 dark:hover:text-red-400 transition-colors rounded-full hover:bg-gray-100 dark:hover:bg-gray-700"
            title="Logout"
          >
            <LogOut className="h-5 w-5" />
          </button>
        </div>
      </div>
    </header>
  );
}
