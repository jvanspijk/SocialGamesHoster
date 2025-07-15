import React from 'react';
import styled from 'styled-components';

const StyledDropdownContainer = styled.div`
    font-family: 'Inter', sans-serif;
    padding: 1.5rem;
    background-color: #f5f5dc; /* Cream/Beige background */
    border-radius: 12px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    max-width: 300px;
    margin: 2rem auto;
`;

const StyledDropdownLabel = styled.label`
    font-size: 1rem;
    color: #5a4a3a; /* Darker brown for text */
    margin-bottom: 0.5rem;
    font-weight: 600;
    display: block;
`;

const StyledDropdownSelect = styled.select`
    width: 100%;
    padding: 0.75rem 1rem;
    border: 1px solid #a08c7a; /* Muted brown border */
    border-radius: 8px;
    background-color: #fffaf0; /* Off-white for select background */
    color: #4a3b2e; /* Deep brown for select text */
    font-size: 1rem;
    appearance: none; /* Remove default browser arrow */
    -webkit-appearance: none;
    -moz-appearance: none;
    background-image: url('data:image/svg+xml;utf8,<svg fill="%234a3b2e" height="24" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg"><path d="M7 10l5 5 5-5z"/><path d="M0 0h24v24H0z" fill="none"/></svg>'); /* Custom arrow */
    background-repeat: no-repeat;
    background-position: right 1rem center;
    background-size: 1.5em;
    cursor: pointer;
    transition: all 0.3s ease;

    &:focus {
        outline: none;
        border-color: #7b6a5a; /* Slightly darker brown on focus */
        box-shadow: 0 0 0 3px rgba(123, 106, 90, 0.3);
    }

    option {
        background-color: #fffaf0;
        color: #4a3b2e;
    }

    option:disabled {
        color: #b0a090; /* Lighter brown for disabled options */
        font-style: italic;
    }
`;

const UserNamesDropdown = ({ users, selectedUsername, onUsernameChange }) => {
    if (!users || !Array.isArray(users)) {
        return null;
    }

    const handleChange = (event) => {
        onUsernameChange(event.target.value);
    };

    return (
        <StyledDropdownContainer>
            <div className="flex flex-col">
                <StyledDropdownLabel htmlFor="userNameSelect">
                    Choose your name:
                </StyledDropdownLabel>
                <StyledDropdownSelect
                    id="userNameSelect"
                    name="userNameSelect"
                    value={selectedUsername}
                    onChange={handleChange}
                >
                    <option value="">Select your name...</option>
                    {users.map(user => (
                        <option
                            key={user.name}
                            value={user.name}
                            disabled={user.IsOnline}
                        >
                            {user.name}
                        </option>
                    ))}
                </StyledDropdownSelect>
            </div>
        </StyledDropdownContainer>
    );
};

export default UserNamesDropdown;

//TODO: styled components?
// Or Chakra / Material UI
