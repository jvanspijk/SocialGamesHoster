import { useEffect, useState } from 'react';
import './App.css';
import UserNamesDropdown from './components/userNamesDropdown';
import LoginButton from './components/loginButton';

//const target = import.meta.env.VITE_API_URL;
// TODO: Use environment variable for target URL
const target = "http://localhost:32678";

function App() {

    const [users, setUsers] = useState([]);
    const [selectedUsername, setSelectedUsername] = useState('');

    useEffect(() => {
        populateUserData();
    }, []);  

    useEffect(() => {
        if (users.length > 0 && !selectedUsername) {
            setSelectedUsername(users[0].name);
        }
    }, [users, selectedUsername]);

    async function loginSelectedUser() {
        console.log("Logging in user: ", selectedUsername);
        if (!target) {
            console.error("API URL is not defined.");
            return;
        }
        if (!selectedUsername) {
            console.error("No username selected for login.");
            return;
        }
        try {
            const response = await fetch(`${target}/user/login/${selectedUsername}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'text/plain',
                }
            });
            if (!response.ok) {
                const errorBody = await response.text();
                console.error("Login failed:", response.status, errorBody);
                return;
            }
            const data = await response.json();
            console.log("Login succeeded:", data);
        } catch (error) {
            console.error("Error during login: ", error);
        }
    }
    
    async function populateUserData() {

        try {
            const response = await fetch(`${target}/user`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            console.log("Response: ", response);
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

export default App;