import React from 'react';
import { StyledLoginButton } from './component-styles/LoginButton';

const LoginButton = ({ onClick, children, disabled }) => {
    return (
        <StyledLoginButton onClick={onClick} disabled={disabled}>
            {children || 'Log In'}
        </StyledLoginButton>
    );
};

export default LoginButton;