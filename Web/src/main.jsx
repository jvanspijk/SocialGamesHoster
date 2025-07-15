import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter, Routes, Route } from "react-router"
import App from "./app";
import LoginPage from './pages/Login.jsx'
import MyDetails from './pages/MyDetails.jsx'
import NotFoundPage from "./pages/NotFound.jsx";



const root = document.getElementById("root");

ReactDOM.createRoot(root).render(
    <BrowserRouter>
        <Routes>
            <Route path="/" element={<App />} />
            <Route path="login" element={<LoginPage />} />
            <Route path="user/:username" element={<MyDetails />} />
            <Route path="not-found" element={<NotFoundPage />} />"
            <Route path="*" element={<NotFoundPage />} />
        </Routes>
    </BrowserRouter>
);

