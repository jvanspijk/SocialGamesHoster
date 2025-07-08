import { useEffect, useState } from 'react';
import './App.css';
import UserNamesDropdown from './components/userNamesDropdown';
import LoginButton from './components/loginButton';



function App() {
    const [users, setUsers] = useState();

    useEffect(() => {
        populateUserData();
    }, []);    
    
    async function populateUserData() {
        try {
            const response = await fetch('user');
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
            {/*<p className="mb-8 text-lg text-white-400">Demo.</p>*/}

            {users ? (
                <UserNamesDropdown users={users} />
            ) : (
                <p>Loading users...</p>
            )}
            {<LoginButton/>}
        </div>
    );
}

export default App;