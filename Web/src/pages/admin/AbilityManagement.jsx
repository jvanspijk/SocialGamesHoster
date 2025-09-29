import React, { useState, useEffect } from 'react';

const AbilityManagement = () => {
    const [abilities, setAbilities] = useState([]);
    const [newAbility, setNewAbility] = useState('');
    const [loading, setLoading] = useState(false);

    const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

    useEffect(() => {
        const fetchAbilities = async () => {
            setLoading(true);
            try {
                const token = localStorage.getItem('jwt-token');
                const response = await fetch(`${apiUrl}/abilities`, {
                    method: 'GET',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${token}`,
                    }
                });
                if (!response.ok) {
                    throw new Error('Failed to fetch abilities');
                }
                const data = await response.json();
                setAbilities(data);
            } catch (error) {
                console.error('Error fetching abilities:', error);
            } finally {
                setLoading(false);
            }
        }
        fetchAbilities();
    }, []);

    const handleAddAbility = () => {
        if (!newAbility.trim()) return;
        setLoading(true);
        const token = localStorage.getItem('jwt-token');
        fetch(`${apiUrl}/abilities`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`,
            },
            body: JSON.stringify({ name: newAbility }),
        })
        .then((res) => res.json())
        .then((added) => setAbilities((prev) => [...prev, added]))
        .finally(() => {
            setNewAbility('');
            setLoading(false);
        });
    };

    const handleDeleteAbility = (id) => {
        setLoading(true);
        fetch(`${apiUrl}/abilities/${id}`, { method: 'DELETE' })
            .then(() => setAbilities((prev) => prev.filter((a) => a.id !== id)))
            .finally(() => setLoading(false));
    };

    // Each child in a list should have a unique "key" prop.
    return (
        <div>
            <h1>Ability Management</h1>
            <div>
                <input
                    type="text"
                    value={newAbility}
                    onChange={(e) => setNewAbility(e.target.value)}
                    placeholder="New ability name"
                    disabled={loading}
                />
                <button onClick={handleAddAbility} disabled={loading || !newAbility.trim()}>
                    Add Ability
                </button>
            </div>
            {loading && <p>Loading...</p>}
            <ul>
                {abilities.map((ability) => (
                    <li key={ability.id}>
                        {ability.name}
                        <button onClick={() => handleDeleteAbility(ability.id)} disabled={loading}>
                            Delete
                        </button>
                    </li>
                ))}
            </ul>
        </div>
    );
};

export default AbilityManagement;