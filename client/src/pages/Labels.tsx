import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import api from "../lib/api";
import type { Inbox, Label } from "../lib/types";
import { Tag, Loader2, AlertCircle, Plus } from "lucide-react";

export default function Labels() {
  const [selectedInboxId, setSelectedInboxId] = useState<string>("");
  const [newLabelName, setNewLabelName] = useState("");
  const [newLabelColor, setNewLabelColor] = useState("#3B82F6");
  const [error, setError] = useState<string | null>(null);
  const queryClient = useQueryClient();

  // Fetch inboxes for dropdown
  const { data: inboxes, isLoading: loadingInboxes } = useQuery({
    queryKey: ["inboxes"],
    queryFn: async () => {
      const res = await api.get<Inbox[]>("/api/v1/inboxes");
      return res.data;
    },
  });

  // Fetch labels for selected inbox
  const { data: labels, isLoading: loadingLabels } = useQuery({
    queryKey: ["labels", selectedInboxId],
    queryFn: async () => {
      const res = await api.get<Label[]>(
        `/api/v1/labels?inbox_id=${selectedInboxId}`
      );
      return res.data;
    },
    enabled: !!selectedInboxId,
  });

  // Create label mutation
  const createLabelMutation = useMutation({
    mutationFn: async (data: {
      inbox_id: string;
      name: string;
      color: string;
    }) => {
      await api.post("/api/v1/labels", data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["labels", selectedInboxId] });
      setNewLabelName("");
      setNewLabelColor("#3B82F6");
      setError(null);
    },
    onError: (err: any) => {
      setError(err.response?.data?.message || "Failed to create label");
    },
  });

  const handleCreateLabel = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedInboxId) {
      setError("Please select an inbox first");
      return;
    }
    if (!newLabelName.trim()) {
      setError("Label name is required");
      return;
    }
    createLabelMutation.mutate({
      inbox_id: selectedInboxId,
      name: newLabelName.trim(),
      color: newLabelColor,
    });
  };

  if (loadingInboxes) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-indigo-600" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          Labels
        </h1>
        <p className="text-gray-600 dark:text-gray-400 mt-1">
          Manage labels for your inboxes
        </p>
      </div>

      {/* Inbox Selector */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6">
        <label
          htmlFor="inbox-select"
          className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
        >
          Select Inbox
        </label>
        <select
          id="inbox-select"
          value={selectedInboxId}
          onChange={(e) => setSelectedInboxId(e.target.value)}
          className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500"
        >
          <option value="">-- Select an Inbox --</option>
          {inboxes?.map((inbox) => (
            <option key={inbox.id} value={inbox.id}>
              {inbox.display_name}
            </option>
          ))}
        </select>
      </div>

      {/* Create Label Form */}
      {selectedInboxId && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Create New Label
          </h2>
          <form onSubmit={handleCreateLabel} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label
                  htmlFor="label-name"
                  className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  Label Name
                </label>
                <input
                  id="label-name"
                  type="text"
                  value={newLabelName}
                  onChange={(e) => setNewLabelName(e.target.value)}
                  placeholder="e.g., Hot Lead"
                  className="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500"
                />
              </div>
              <div>
                <label
                  htmlFor="label-color"
                  className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  Color
                </label>
                <div className="flex items-center space-x-2">
                  <input
                    id="label-color"
                    type="color"
                    value={newLabelColor}
                    onChange={(e) => setNewLabelColor(e.target.value)}
                    className="h-10 w-20 border border-gray-300 dark:border-gray-600 rounded cursor-pointer"
                  />
                  <input
                    type="text"
                    value={newLabelColor}
                    onChange={(e) => setNewLabelColor(e.target.value)}
                    placeholder="#3B82F6"
                    className="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-indigo-500"
                  />
                </div>
              </div>
            </div>

            {error && (
              <div className="rounded-md bg-red-50 dark:bg-red-900/20 p-4">
                <div className="flex">
                  <AlertCircle className="h-5 w-5 text-red-400" />
                  <p className="ml-3 text-sm text-red-800 dark:text-red-200">
                    {error}
                  </p>
                </div>
              </div>
            )}

            <button
              type="submit"
              disabled={createLabelMutation.isPending}
              className="flex items-center space-x-2 px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {createLabelMutation.isPending ? (
                <Loader2 className="h-5 w-5 animate-spin" />
              ) : (
                <Plus className="h-5 w-5" />
              )}
              <span>
                {createLabelMutation.isPending ? "Creating..." : "Create Label"}
              </span>
            </button>
          </form>
        </div>
      )}

      {/* Labels List */}
      {selectedInboxId && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Existing Labels
          </h2>

          {loadingLabels ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-indigo-600" />
            </div>
          ) : !labels || labels.length === 0 ? (
            <div className="text-center py-8">
              <Tag className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-600 dark:text-gray-400">
                No labels found for this inbox
              </p>
            </div>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
              {labels.map((label) => (
                <div
                  key={label.id}
                  className="flex items-center space-x-3 p-4 border border-gray-200 dark:border-gray-700 rounded-lg hover:shadow-md transition-shadow"
                >
                  <div
                    className="w-8 h-8 rounded-full shrink-0"
                    style={{ backgroundColor: label.color }}
                  />
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                      {label.name}
                    </p>
                    <p className="text-xs text-gray-500 dark:text-gray-400 font-mono">
                      {label.color}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
