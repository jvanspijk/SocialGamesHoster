import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router';

const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

const PlayerManagement = () => {
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
                setLoading(false);
            }
        };
        fetchPlayers();
    }, []);

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
                                <td>{player.role?.name}</td>
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