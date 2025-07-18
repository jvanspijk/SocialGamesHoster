import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter, Routes, Route } from "react-router"
import LoginPage from './pages/Login.jsx'
import MyDetails from './pages/MyDetails.jsx'
import NotFoundPage from "./pages/NotFound.jsx";
import App from "./App.jsx"
import AdminLogin from "./pages/admin/AdminLogin.jsx";

const root = document.getElementById("root");

ReactDOM.createRoot(root).render(
    <BrowserRouter>
        <Routes>
            <Route path="/" element={<App />} />

            <Route path="game">
                <Route index element={<LoginPage />} />
                <Route path="login" element={<LoginPage />} />
                <Route path="player/:name" element={<MyDetails />} />
                <Route path="not-found" element={<NotFoundPage />} />
            </Route>

            <Route path="admin">
                <Route index element={<AdminLogin />} />
                {/*<Route path="players" element={<Players />} />*/}
                {/*<Route path="roles" element={<Roles />} />*/}
            </Route>

            <Route path="*" element={<NotFoundPage />} />
        </Routes>
    </BrowserRouter>
);

