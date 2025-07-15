import React from 'react';
import { useEffect, useState } from 'react';
import { useParams } from "react-router";
//import '../styles/App.css';

const apiUrl = "http://localhost:8080";

function MyDetails() {
    const [role, setRole] = useState(null);
    //const [userDetails, setUserDetails] = useState(null);
    const { username } = useParams();

    useEffect(() => {
        async function fetchRole() {
            if (!username) {
                console.error("Username is not provided.");
                return;
            }
            const token = localStorage.getItem('jwt-token');
            try {
                const response = await fetch(`${apiUrl}/Role/${username}`, {
                    method: 'GET',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${token || ''}`,
                    },
                });
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                const data = await response.json();
                setRole(data);
            } catch (error) {
                console.error("Error fetching role:", error);
            }
        }
        fetchRole();
    }, [username]);

    // TODO: make role details component
    return (
        <div className="my-details">
            <h1>{username}</h1>
            {role ? (
                <>
                    <h2>{role.name}</h2>      
                    <p>{role.description}</p>
                </>
            ) : (
                "Loading..."
            )}
        </div>
    );
}

export default MyDetails;