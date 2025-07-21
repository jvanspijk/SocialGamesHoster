import React, { useEffect, useState } from 'react';
//TODO: add role description and abilities management
const RoleManagement = () => {
    const [roles, setRoles] = useState([]);
    const [newRole, setNewRole] = useState('');
    const [loading, setLoading] = useState(false);

    const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

    useEffect(() => {
        const fetchRoles = async () => {
            setLoading(true);
            try {
                const token = localStorage.getItem('jwt-token');
                const response = await fetch(`${apiUrl}/roles`, {
                    method: 'GET',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${token}`,
                    },
                });
                if (!response.ok) {
                    throw new Error('Failed to fetch roles');
                }
                const data = await response.json();
                setRoles(data);
            } catch (error) {
                console.error('Error fetching roles:', error);
            } finally {
                setLoading(false);
            }
        }
        fetchRoles();
    }, []);

    const handleAddRole = async (e) => {
        e.preventDefault();
        if (!newRole.trim()) return;
        setLoading(true);
        const token = localStorage.getItem('jwt-token');
        const response = await fetch(`${apiUrl}/roles`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`, 
            },
            body: JSON.stringify({ name: newRole })
        });
        if (response.ok) {
            const createdRole = await response.json();
            setRoles([...roles, createdRole]);
            setNewRole('');
        }
        setLoading(false);
    };

    const handleDeleteRole = async (roleId) => {
        setLoading(true);
        const token = localStorage.getItem('jwt-token');
        const response = await fetch(`${apiUrl}/roles/${roleId}`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`,
            }
        });
        if (response.ok) {
            setRoles(roles.filter(role => role.id !== roleId));
        }
        setLoading(false);
    };

    return (
        <div>
            <h1>Role Management</h1>
            <form onSubmit={handleAddRole}>
                <input
                    type="text"
                    value={newRole}
                    onChange={e => setNewRole(e.target.value)}
                    placeholder="New role name"
                    disabled={loading}
                />
                <button type="submit" disabled={loading || !newRole.trim()}>Add Role</button>
            </form>
            {loading && <p>Loading...</p>}
            <ul>
                {roles.map(role => (
                    <li key={role.id}>
                        {role.name}
                        <button onClick={() => handleDeleteRole(role.id)} disabled={loading}>Delete</button>
                    </li>
                ))}
            </ul>
        </div>
    );
};

export default RoleManagement;