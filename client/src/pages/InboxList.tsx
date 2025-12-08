import { useQuery } from "@tanstack/react-query";
import api from "../lib/api";
import type { Inbox } from "../lib/types";
import { useAuthStore } from "../store/authStore";
import { Inbox as InboxIcon, Loader2, AlertCircle } from "lucide-react";

export default function InboxList() {
  const role = useAuthStore((state) => state.role);

  const {
    data: inboxes,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["inboxes"],
    queryFn: async () => {
      const res = await api.get<Inbox[]>("/api/v1/inboxes");
      return res.data;
    },
  });

  if (role !== "ADMIN" && role !== "MANAGER") {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <AlertCircle className="h-12 w-12 text-red-500 mx-auto mb-4" />
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
            Access Denied
          </h2>
          <p className="text-gray-600 dark:text-gray-400 mt-2">
            This page is only accessible to Managers and Admins.
          </p>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-indigo-600" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-center">
          <AlertCircle className="h-12 w-12 text-red-500 mx-auto mb-4" />
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
            Error Loading Inboxes
          </h2>
          <p className="text-gray-600 dark:text-gray-400 mt-2">
            {error.message}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          Inboxes
        </h1>
        <p className="text-gray-600 dark:text-gray-400 mt-1">
          Manage and view all inboxes
        </p>
      </div>

      {!inboxes || inboxes.length === 0 ? (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-12 text-center">
          <InboxIcon className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 dark:text-white">
            No inboxes found
          </h3>
          <p className="text-gray-600 dark:text-gray-400 mt-2">
            There are no inboxes available at the moment.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {inboxes.map((inbox) => (
            <div
              key={inbox.id}
              className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6 hover:shadow-md transition-shadow"
            >
              <div className="flex items-start justify-between">
                <div className="flex items-center space-x-3">
                  <div className="shrink-0">
                    <InboxIcon className="h-8 w-8 text-indigo-600 dark:text-indigo-400" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-white truncate">
                      {inbox.display_name}
                    </h3>
                    <p className="text-sm text-gray-500 dark:text-gray-400 font-mono truncate">
                      {inbox.id}
                    </p>
                  </div>
                </div>
              </div>
              <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-gray-600 dark:text-gray-400">
                    Created
                  </span>
                  <span className="text-gray-900 dark:text-white font-medium">
                    {new Date(inbox.created_at).toLocaleDateString()}
                  </span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
