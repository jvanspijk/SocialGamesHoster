// RoundTimer.jsx
import React, { useState, useEffect } from 'react';
import { ScrollBox } from './component-styles/Common.jsx';

const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

function formatTime(ms) {
    if (ms == null) return 'Loading...';
    const totalSeconds = Math.floor(ms / 1000);
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;
    return `${hours.toString().padStart(2, '0')}:${minutes
        .toString()
        .padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
}

export function RoundTimer() {
    const [endTime, setEndTime] = useState(null);
    const [remainingTime, setRemainingTime] = useState(null);

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
            setRemainingTime(diff > 0 ? diff : 0);
        }

        updateRemaining();
        const interval = setInterval(updateRemaining, 1000);
        return () => clearInterval(interval);
    }, [endTime]);

    return (
        <ScrollBox>
            <strong>Time left:</strong>{' '}
            {formatTime(remainingTime)}
        </ScrollBox>
    );
}