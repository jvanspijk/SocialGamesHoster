import React, { useState } from 'react';
import { useNavigate } from 'react-router';

const AdminLogin = () => {
    const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');

    const navigate = useNavigate();

    async function attemptLogin(e) {
        e.preventDefault();
        if (!apiUrl) {
            setError("API URL is not defined.");
            console.error("API URL is not defined.");
            return;
        }
        if (!username) {
            setError("No username provided for login.");
            console.error("No username provided for login.");
            return;
        }
        if (!password) {
            setError("No password provided for login.");
            console.error("No password provided for login.");
            return;
        }
        try {
            const response = await fetch(`${apiUrl}/Admin/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json'
                },
                body: JSON.stringify({ username: username, password: password }),
            });
            let responseText = await response.text();
            let data;
            try {
                data = JSON.parse(responseText);
            } catch {
                setError(responseText || "Login failed.");
                console.error("Login failed:", response.status, responseText);
                return;
            }
            if (!response.ok) {
                if (data && data.errors) {
                    const errorMessages = Object.values(data.errors)
                        .flat()
                        .join(' ');
                    setError(errorMessages || "Login failed.");
                    console.error("Login failed:", response.status, errorMessages);
                } else {
                    setError(data.title || "Login failed.");
                    console.error("Login failed:", response.status, data.title);
                }
                return;
            }
            if (data.token) {
                localStorage.setItem('jwt-token', data.token);
                navigate('/admin/dashboard', { replace: true });
                setError('');
            } else {
                setError("Login failed: No token returned.");
                console.error("Login failed: No token returned.");
            }
            return;
        } catch (error) {
            console.error("Error during login: ", error);
            setError(error.message || "An error occurred during login.");
        }
    }

    return (
        <div className="admin-login-container">
            <h2>Admin Login</h2>
            <form onSubmit={attemptLogin}>
                <div>
                    <label htmlFor="admin-username">Username </label>
                    <input
                        id="admin-username"
                        type="text"
                        value={username}
                        onChange={e => setUsername(e.target.value)}
                        autoComplete="username"
                        required
                    />
                </div>
                <div>
                    <label htmlFor="admin-password">Password </label>
                    <input
                        id="admin-password"
                        type="password"
                        value={password}
                        onChange={e => setPassword(e.target.value)}
                        autoComplete="current-password"
                        required
                    />
                </div>
                {error && <div className="error-message">{error}</div>}
                <button type="submit">Login</button>
            </form>
        </div>
    );
};

export default AdminLogin;