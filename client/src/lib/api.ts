import axios from "axios";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || "http://localhost:8080",
});

api.interceptors.request.use((config) => {
  // Default Tenant ID
  config.headers["X-Tenant-ID"] = "10000000-0000-0000-0000-000000000001";

  // Operator ID from localStorage
  const operatorId = localStorage.getItem("operator_id");
  if (operatorId) {
    config.headers["X-Operator-ID"] = operatorId;
  }

  return config;
});

export default api;
