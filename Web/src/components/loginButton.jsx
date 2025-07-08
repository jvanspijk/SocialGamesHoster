import React from 'react';

const LoginButton = ({ onClick, children }) => {
  return (
    <button className="login-button" onClick={onClick}>
      {children || 'Log In'}
    </button>
  );
};

export default LoginButton;