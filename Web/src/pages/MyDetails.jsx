import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from "react-router";
import styled from 'styled-components';

const apiUrl = "http://localhost:8080";

const PageContainer = styled.div`
    background-color: #F5DEB3;
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 2em;
    font-family: 'Old Standard TT', serif;
    color: #4A3A2B;
`;

const StyledHeader = styled.header`
    text-align: center;
    margin-bottom: 0em;
    padding: 0.5em 1em;    
    border-radius: 10px;
`;

const MainTitle = styled.h1`
    font-size: 4em;
    color: #8B4513;
    text-shadow: 3px 3px 5px rgba(0, 0, 0, 0.3);
    margin: 0;
    font-family: 'IM Fell DW Pica', serif;
    letter-spacing: 1px;
`;

const AbilitiesSectionTitle = styled.h3`
    font-family: 'IM Fell DW Pica', serif;
    font-size: 2em;
    color: #8B4513;
    text-align: center;
    margin: 1em 0 1em 0; /* Increased top and bottom margins for better separation */
    padding: 1em 1em; /* More padding for a bolder banner */
    background-color: #D2B48C;
    border-radius: 10px; /* Slightly more rounded corners */
    box-shadow: 0px 4px 8px rgba(0, 0, 0, 0.3); /* Stronger shadow for more depth */
    display: block;
    max-width: 600px;
    width: 90%;
    box-sizing: border-box;
    align-self: center;
    border: 1px solid #A08060; /* Added a subtle border */
`;

const ScrollBox = styled.article`
    background-color: #FDF5E6;
    border: 1px solid #D2B48C;
    box-shadow: 
        0px 5px 15px rgba(0, 0, 0, 0.1), 
        inset 0 0 15px rgba(210, 180, 140, 0.2);
    padding: 2.5em 3em;
    margin: 1.5em 0;
    width: 600px;
    max-width: 90%;
    box-sizing: border-box;
    text-align: center;
    position: relative; /* For number circle positioning */
    border-radius: 8px;

    h2, h3 {
        color: #6B4226;
        margin-top: 0;
        text-shadow: 1px 1px 2px rgba(0,0,0,0.1);
    }

    p {
        line-height: 1.6;
        margin-bottom: 1.5em;
    }

    ul {
        list-style: none;
        padding: 0;
        text-align: left;
        margin-left: 2em;
    }

    li {
        margin-bottom: 0.5em;
        font-size: 0.95em;
    }
`;

const NumberCircle = styled.div`
    position: absolute;
    left: -20px;
    top: 50%;
    transform: translateY(-50%);
    background-color: #8B4513;
    color: white;
    border-radius: 50%;
    width: 40px;
    height: 40px;
    display: flex;
    justify-content: center;
    align-items: center;
    font-size: 1.5em;
    font-weight: bold;
    border: 3px solid #D2B48C;
    box-shadow: 0px 2px 5px rgba(0,0,0,0.3);
`;



function MyDetails() {
    const [role, setRole] = useState(null);
    const [endTime, setEndTime] = useState(null);
    const [remaining, setRemaining] = useState(null);
    const { username } = useParams();

    const navigate = useNavigate();

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
    }, [username, navigate]);

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
        <PageContainer>
            <ScrollBox>
                <strong>Time left:</strong>{' '}
                {remaining != null ? formatTime(remaining) : 'Loading...'}
            </ScrollBox>

            <StyledHeader>
                <MainTitle>{username}</MainTitle>
            </StyledHeader>

            {role ? (
                <>
                    <ScrollBox>
                        <h2>{role.name}</h2>
                        <p>{role.description}</p>
                    </ScrollBox>

                    {role.abilities && role.abilities.length > 0 && (
                        <>
                            <AbilitiesSectionTitle>Abilities:</AbilitiesSectionTitle>
                            {role.abilities.map((ability, index) => (
                                <ScrollBox key={index}>
                                    <ul>
                                        <li>
                                            <b>{ability.name}:</b> {ability.description}
                                        </li>
                                    </ul>
                                    <NumberCircle>{index + 1}</NumberCircle>
                                </ScrollBox>
                            ))}
                        </>
                    )}
                </>
            ) : (
                <ScrollBox>
                    <p>Loading role information...</p>
                    <NumberCircle>1</NumberCircle>
                </ScrollBox>
            )}


        </PageContainer>
    );
}

export default MyDetails;