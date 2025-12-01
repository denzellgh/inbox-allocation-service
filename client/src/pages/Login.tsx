import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { z } from "zod";
import { LogIn, Loader2, AlertCircle } from "lucide-react";
import api from "../lib/api";
import { useAuthStore } from "../store/authStore";

const loginSchema = z.string().guid({
  message: "Invalid Operator ID format (must be UUID)",
});

export default function Login() {
  const [operatorId, setOperatorId] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const login = useAuthStore((state) => state.login);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    // Validate UUID
    const validation = loginSchema.safeParse(operatorId);
    if (!validation.success) {
      setError(validation.error.issues[0].message);
      return;
    }

    setLoading(true);

    try {
      // Try to fetch operator details
      // We temporarily set the header for this request or rely on the backend allowing it
      // Since we don't have the ID in store yet, the interceptor won't send it.
      // If the backend requires X-Operator-ID to fetch an operator, we might need to send it explicitly here.

      const response = await api.get(`/api/v1/operators/${operatorId}`, {
        headers: {
          "X-Operator-ID": operatorId, // Self-identification for login check
        },
      });

      const operator = response.data;

      if (operator) {
        login(operator.id, operator.role);
        navigate("/");
      } else {
        setError("Operator not found");
      }
    } catch (err: any) {
      console.error("Login error:", err);
      if (err.response?.status === 404) {
        setError("Operator not found");
      } else if (err.response?.status === 403) {
        setError("Access denied");
      } else {
        setError("Failed to login. Please check your connection.");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 px-4">
      <div className="max-w-md w-full space-y-8 bg-white dark:bg-gray-800 p-8 rounded-xl shadow-lg border border-gray-100 dark:border-gray-700">
        <div className="text-center">
          <div className="mx-auto h-12 w-12 bg-indigo-100 dark:bg-indigo-900/50 rounded-full flex items-center justify-center">
            <LogIn className="h-6 w-6 text-indigo-600 dark:text-indigo-400" />
          </div>
          <h2 className="mt-6 text-3xl font-extrabold text-gray-900 dark:text-white">
            Operator Login
          </h2>
          <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
            Enter your Operator ID to access the dashboard
          </p>
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          <div>
            <label htmlFor="operator-id" className="sr-only">
              Operator ID
            </label>
            <input
              id="operator-id"
              name="operatorId"
              type="text"
              required
              className="appearance-none rounded-lg relative block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 placeholder-gray-500 dark:placeholder-gray-400 text-gray-900 dark:text-white dark:bg-gray-700 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 focus:z-10 sm:text-sm"
              placeholder="Operator UUID (e.g., 30000000-...)"
              value={operatorId}
              onChange={(e) => setOperatorId(e.target.value)}
            />
          </div>

          {error && (
            <div className="rounded-md bg-red-50 dark:bg-red-900/20 p-4">
              <div className="flex">
                <div className="shrink-0">
                  <AlertCircle
                    className="h-5 w-5 text-red-400"
                    aria-hidden="true"
                  />
                </div>
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-red-800 dark:text-red-200">
                    {error}
                  </h3>
                </div>
              </div>
            </div>
          )}

          <div>
            <button
              type="submit"
              disabled={loading}
              className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {loading ? (
                <Loader2 className="animate-spin h-5 w-5" />
              ) : (
                "Sign in"
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
