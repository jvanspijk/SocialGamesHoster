import React, { useState, useEffect } from 'react';
import { ScrollBox } from './component-styles/Common.jsx';
import { useNavigate } from 'react-router';

const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export function RoleDescription({ username, children }) {
    const navigate = useNavigate();

    const [role, setRole] = useState(null);
    const [loadingRole, setLoadingRole] = useState(true);

    useEffect(() => {
        async function fetchRole() {
            if (!username) {
                console.error("Username is not provided.");
                setLoadingRole(false);
                return;
            }
            const token = localStorage.getItem('jwt-token');
            try {
                const response = await fetch(`${apiUrl}/Roles/${username}`, {
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
                setRole(data);
            } catch (error) {
                console.error("Error fetching role:", error);
            } finally {
                setLoadingRole(false);
            }
        }
        fetchRole();
    }, [username, navigate]);

    if (loadingRole) {
        return (
            <ScrollBox>
                <p>Loading role information...</p>
            </ScrollBox>
        );
    }

    if (!role) {
        return null;
    }

    return (
        <>
            <ScrollBox>
                <h2>{role.name}</h2>
                <p>{role.description}</p>
            </ScrollBox>
            {children(role, loadingRole)}
        </>
    );
}