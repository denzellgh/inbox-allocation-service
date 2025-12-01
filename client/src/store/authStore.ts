import { create } from "zustand";
import { persist } from "zustand/middleware";

interface AuthState {
  operatorId: string | null;
  role: string | null;
  tenantId: string;
  login: (id: string, role: string) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      operatorId: null,
      role: null,
      tenantId: "10000000-0000-0000-0000-000000000001", // Default Tenant
      login: (id, role) => set({ operatorId: id, role }),
      logout: () => {
        localStorage.removeItem("operator_id"); // Clear legacy/direct storage if any
        set({ operatorId: null, role: null });
      },
    }),
    {
      name: "auth-storage",
      onRehydrateStorage: () => (state) => {
        // Sync with localStorage for API interceptor
        if (state?.operatorId) {
          localStorage.setItem("operator_id", state.operatorId);
        } else {
          localStorage.removeItem("operator_id");
        }
      },
    }
  )
);

// Subscribe to changes to keep localStorage in sync for the API interceptor
useAuthStore.subscribe((state) => {
  if (state.operatorId) {
    localStorage.setItem("operator_id", state.operatorId);
  } else {
    localStorage.removeItem("operator_id");
  }
});
