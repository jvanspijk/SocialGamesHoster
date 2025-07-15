import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from "react-router";

const apiUrl = "http://localhost:8080";

function MyDetails() {
    const [role, setRole] = useState(null);
    const [endTime, setEndTime] = useState(null);
    const [remaining, setRemaining] = useState(null);
    const { username } = useParams();

    const navigate = useNavigate();

    // Fetch role
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
                    navigate("not-found");
                    return;
                }
                const data = await response.json();
                console.log("Role data:", data);
                setRole(data);
            } catch (error) {
                console.error("Error fetching role:", error);
            }
        }
        fetchRole();
    }, [username]);

    useEffect(() => {
        async function fetchEndTime() {
            try {
                const response = await fetch(`${apiUrl}/Round/end-time`, {
                    method: 'GET',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                });
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                const data = await response.json();
                setEndTime(new Date(data));
            } catch (error) {
                console.error("Error fetching end time:", error);
            }
        }
        fetchEndTime();
    }, []);

    useEffect(() => {
        if (!endTime) return;

        function updateRemaining() {
            const now = new Date();
            const diff = endTime - now;
            setRemaining(diff > 0 ? diff : 0);
        }

        updateRemaining();
        const interval = setInterval(updateRemaining, 1000);
        return () => clearInterval(interval);
    }, [endTime]);

    function formatTime(ms) {
        if (ms == null) return '';
        const totalSeconds = Math.floor(ms / 1000);
        const hours = Math.floor(totalSeconds / 3600);
        const minutes = Math.floor((totalSeconds % 3600) / 60);
        const seconds = totalSeconds % 60;
        return `${hours.toString().padStart(2, '0')}:${minutes
            .toString()
            .padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    }

    return (
        <section>
            <header>
                <h1 style={{ fontSize: '5em', marginTop: '0.67em', marginBottom: '0.67em' }}>{username}</h1>
            </header>

            {role ? (
                <article>
                    <h2>{role.name}</h2>
                    <p>{role.description}</p>
                    {role.abilities && role.abilities.length > 0 && (
                        <div>
                            <h3>Abilities:</h3>
                            <ul>
                                {role.abilities.map((ability, index) => (
                                    <li key={index}>
                                        <b>{ability.name}:</b> {ability.description}
                                    </li>
                                ))}
                            </ul>
                        </div>
                    )}
                </article>
            ) : (
                <p>Loading role information...</p>
            )}

            <div style={{ marginTop: '1em', padding: '0.5em' }}>
                <strong>Time left:</strong>{' '}
                {remaining != null ? formatTime(remaining) : 'Loading...'}
            </div>
        </section>
    );
}

export default MyDetails;