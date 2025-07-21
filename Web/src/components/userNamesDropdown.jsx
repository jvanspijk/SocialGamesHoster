import React from 'react';
import { StyledDropdownContainer, StyledDropdownLabel, StyledDropdownSelect } from './component-styles/UserNamesDropdown.jsx';

const UserNamesDropdown = ({ users, selectedUsername, onUsernameChange }) => {
    if (!users || !Array.isArray(users)) {
        return null;
    }

    const handleChange = (event) => {
        console.log("Selected username:", event.target.value);
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