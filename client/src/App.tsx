import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import Login from "./pages/Login";
import ProtectedRoute from "./components/ProtectedRoute";
import Layout from "./components/Layout";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          <Route
            index
            element={
              <div className="space-y-4">
                <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
                  Dashboard
                </h1>
                <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700">
                  <p className="text-gray-600 dark:text-gray-300">
                    Welcome to Inbox Allocation Service.
                  </p>
                </div>
              </div>
            }
          />
          <Route
            path="inbox"
            element={<div className="p-4">My Inbox Placeholder</div>}
          />
          <Route
            path="admin/inboxes"
            element={<div className="p-4">Manage Inboxes Placeholder</div>}
          />
          <Route
            path="admin/operators"
            element={<div className="p-4">Manage Operators Placeholder</div>}
          />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
