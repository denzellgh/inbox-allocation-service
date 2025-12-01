import { BrowserRouter, Routes, Route } from "react-router-dom";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route
          path="/login"
          element={<div className="p-4">Login Placeholder</div>}
        />
        <Route
          path="/"
          element={<div className="p-4">Dashboard Placeholder</div>}
        />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
