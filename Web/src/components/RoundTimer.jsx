import React, { useState, useEffect } from 'react';
import { ScrollBox } from './component-styles/Common.jsx';
import Theme from '../styles/Theme.jsx';
import { TimerContainer, TimerLabel, TimerCircle, TimeDisplay } from './component-styles/RoundTimer.jsx';

// API URL for fetching the round end time
const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080';

/**
 * Formats a given time in milliseconds into a human-readable HH:MM:SS string.
 * @param {number | null} ms - The time in milliseconds to format.
 * @returns {string} The formatted time string or 'Loading...' if ms is null.
 */
function formatTime(ms) {
    if (ms == null) return 'Loading...'; // Display loading state if time is not yet available
    const totalSeconds = Math.floor(ms / 1000);
    const hours = Math.floor(totalSeconds / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;

    // Pad with leading zeros to ensure consistent HH:MM:SS format
    return `${hours.toString().padStart(2, '0')}:${minutes
        .toString()
        .padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
}

/**
 * RoundTimer component displays a countdown timer for the current round.
 * It fetches the round end time from an API and updates the countdown every second.
 */
export function RoundTimer() {
    const [endTime, setEndTime] = useState(null);
    const [remainingTime, setRemainingTime] = useState(null);

    useEffect(() => {
        async function fetchEndTime() {
            try {
                const response = await fetch(`${apiUrl}/Rounds/end-time`, {
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
    }, []); // Empty dependency array ensures this runs only once on mount

    useEffect(() => {
        if (!endTime) return;

        /**
         * Calculates and updates the remaining time.
         * Sets remainingTime to 0 if the countdown has finished.
         */
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
            <TimerContainer>
                <TimerLabel>Time left in current round:</TimerLabel>
                <TimerCircle $background={Theme.colors.background}>
                    <TimeDisplay>{formatTime(remainingTime)}</TimeDisplay>
                </TimerCircle>
            </TimerContainer>
        </ScrollBox>
    );
}