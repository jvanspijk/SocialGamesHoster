import React from 'react';
import { useEffect, useState } from 'react';
import UserNamesDropdown from '../components/userNamesDropdown';
import LoginButton from '../components/loginButton';
import { useNavigate } from "react-router";
import '../styles/App.css';
// TODO: Use environment variable for target URL
const apiUrl = "http://localhost:8080";

function LoginPage() {
    const [users, setUsers] = useState([]);
    const [selectedUsername, setSelectedUsername] = useState('');
    const navigate = useNavigate();

    useEffect(() => {
        populateUserData();
    }, []);

    useEffect(() => {
        if (users.length > 0 && !selectedUsername) {
            setSelectedUsername(users[0].name);
        }
    }, [users, selectedUsername]);

    async function loginSelectedUser() {
        if (!apiUrl) {
            console.error("API URL is not defined.");
            return;
        }
        if (!selectedUsername) {
            console.error("No username selected for login.");
            return;
        }
        try {
            const response = await fetch(`${apiUrl}/Auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(selectedUsername),
            });
            if (!response.ok) {
                const errorBody = await response.text();
                console.error("Login failed:", response.status, errorBody);
                return;
            }
            const data = await response.json();
            localStorage.setItem('jwt-token', data.token);
            navigate(`/user/${selectedUsername}`);
            return;
        } catch (error) {
            console.error("Error during login: ", error);
        }
    }

    async function populateUserData() {
        try {
            const response = await fetch(`${apiUrl}/User`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();
            setUsers(data);
        } catch (error) {
            console.error("Error fetching users: ", error);
        }
    }

    return (
        <div className="min-h-screen text-gray-300 flex flex-col items-center justify-center py-8">
            <h1 id="tableLabel" className="text-4xl font-bold mb-4 text-gray-200">Login</h1>
            {users ? (
                <UserNamesDropdown users={users} selectedUsername={selectedUsername} onUsernameChange={setSelectedUsername} />
            ) : (
                <p>Loading users...</p>
            )}
            {<LoginButton onClick={loginSelectedUser} disabled={!selectedUsername} />}
        </div>
    );
}

export default LoginPage;