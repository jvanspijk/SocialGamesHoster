import React from 'react';
import { useEffect, useState } from 'react';
import UserNamesDropdown from '../components/userNamesDropdown';
import LoginButton from '../components/loginButton';
import { useNavigate } from "react-router";
import { ThemeProvider } from 'styled-components'
import Theme from '../styles/Theme.jsx'
import '../styles/App.css';

const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

function LoginPage() {
    const [players, setPlayers] = useState([]);
    const [selectedName, setSelectedName] = useState('');
    const navigate = useNavigate();

    useEffect(() => {
        populateUserData();
    }, []);

    useEffect(() => {
        if (players.length > 0 && !selectedName) {
            setSelectedName(players[0].name);
        }
    }, [players, selectedName]);

    async function loginSelectedPlayer() {
        if (!apiUrl) {
            console.error("API URL is not defined.");
            return;
        }
        if (!selectedName) {
            console.error("No username selected for login.");
            return;
        }
        try {
            const response = await fetch(`${apiUrl}/Player/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(selectedName),
            });
            if (!response.ok) {
                const errorBody = await response.text();
                console.error("Login failed:", response.status, errorBody);
                return;
            }
            const data = await response.json();
            localStorage.setItem('jwt-token', data.token);
            navigate(`/game/Player/${selectedName}`);
            return;
        } catch (error) {
            console.error("Error during login: ", error);
        }
    }

    async function populateUserData() {
        try {
            const response = await fetch(`${apiUrl}/Player`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();
            setPlayers(data);
        } catch (error) {
            console.error("Error fetching users: ", error);
        }
    }

    return (
        <ThemeProvider theme ={Theme}>
        <div className="min-h-screen text-gray-300 flex flex-col items-center justify-center py-8">
            <h1 id="tableLabel" className="text-4xl font-bold mb-4 text-gray-200">Login</h1>
            {players ? (
                <UserNamesDropdown users={players} selectedUsername={selectedName} onUsernameChange={setSelectedName} />
            ) : (
                <p>Loading users...</p>
            )}
            {<LoginButton onClick={loginSelectedPlayer} disabled={!selectedName} />}
            </div>
        </ThemeProvider>
    );
}

export default LoginPage;