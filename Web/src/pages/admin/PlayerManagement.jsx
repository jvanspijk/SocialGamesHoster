import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router';

const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

const PlayerManagement = () => {
    const [roles, setRoles] = useState([]);
    const [players, setPlayers] = useState([]);
    const [loading, setLoading] = useState(true);
    const navigate = useNavigate();

    useEffect(() => {
        const fetchPlayers = async () => {
            try {
                const response = await fetch(`${apiUrl}/Players`);
                const data = await response.json();
                setPlayers(data);
            } catch (error) {
                console.error('Failed to fetch players:', error);
            } finally {
                setPlayers(prevPlayers => prevPlayers.sort((a, b) => a.id - b.id));
                setLoading(false);
            }
        };
        fetchPlayers();
        fetchRoles();
        // eslint-disable-next-line
    }, []);

    const fetchRoles = async () => {
        try {
            const token = localStorage.getItem('jwt-token');
            const response = await fetch(`${apiUrl}/Roles`, {
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });
            if (!response.ok) {
                throw new Error('Failed to fetch roles');
            }
            const data = await response.json();
            data.sort((a, b) => a.id - b.id);
            setRoles(data);
        }
        catch (error) {
            console.error('Failed to fetch roles:', error);
        }
    }

    const handleDelete = async (playerId) => {
        if (!window.confirm('Are you sure you want to delete this player?')) return;
        try {
            const token = localStorage.getItem('jwt-token');
            await fetch(`${apiUrl}/Players/${playerId}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });
            setPlayers(players.filter(player => player.id !== playerId));
        } catch (error) {
            console.error('Failed to delete player:', error);
        }
    };

    const handleNameClick = (playerName) => {
        const uriEncodedName = encodeURIComponent(playerName);
        navigate(`/game/Player/${uriEncodedName}`);
    };

    const handleRoleChange = async (playerName, newRoleId) => {
        if (!newRoleId) {
            console.warn('No role selected for player:', playerName);
            return;
        }
        newRoleId = parseInt(newRoleId, 10);
        try {
            const token = localStorage.getItem('jwt-token');
            const response = await fetch(`${apiUrl}/Players/${playerName}/Role`, {
                method: 'PATCH',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify(newRoleId)
            });
            if (!response.ok) throw new Error('Failed to update role');
            const updatedPlayer = await response.json();
            setPlayers(prevPlayers => {
                const newPlayersArray = prevPlayers.map(player =>
                    player.id === updatedPlayer.id
                        ? { ...player, role: roles.find(r => r.id === newRoleId) }
                        : player
                );
                return newPlayersArray.sort((a, b) => a.id - b.id);
            });
        } catch (error) {
            console.error('Failed to update player role:', error);
        }
    };

    if (loading) {
        return <div>Loading players...</div>;
    }

    return (
        <div>
            <h1>Player Management</h1>
            <table>
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>Name</th>
                        <th>Role</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {players.length === 0 ? (
                        <tr>
                            <td colSpan="4">No players found.</td>
                        </tr>
                    ) : (
                        players.map(player => (
                            <tr key={player.id}>
                                <td>{player.id}</td>
                                <td>
                                    <span
                                        style={{ color: 'blue', textDecoration: 'underline', cursor: 'pointer' }}
                                        onClick={() => handleNameClick(player.name)}
                                    >
                                        {player.name}
                                    </span>
                                </td>
                                <td>
                                    <select
                                        value={player.role?.id || ''}
                                        onChange={e => handleRoleChange(player.name, e.target.value)}
                                    >
                                        <option value="" disabled>Select role</option>
                                        {roles.map(role => (
                                            <option key={role.id} value={role.id}>
                                                {role.name}
                                            </option>
                                        ))}
                                    </select>
                                </td>
                                <td>
                                    <button onClick={() => handleDelete(player.id)}>Delete</button>
                                </td>
                            </tr>
                        ))
                    )}
                </tbody>
            </table>
        </div>
    );
};

export default PlayerManagement;