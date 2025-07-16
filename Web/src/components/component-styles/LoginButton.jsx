import styled from 'styled-components';

const StyledLoginButton = styled.button`
    background-color: ${props => props.theme.colors.primary};
    color: ${props => props.theme.colors.white};
    padding: ${props => props.theme.spacing.md} ${props => props.theme.spacing.lg};
    border: none;
    border-radius: ${props => props.theme.borderRadius.md};
    font-family: ${props => props.theme.typography.fontFamilyBase};
    font-size: ${props => props.theme.typography.fontSizeBase};
    font-weight: bold;
    cursor: pointer;
    transition: background-color 0.3s ease, box-shadow 0.3s ease;
    box-shadow: 0px 4px 8px ${props => props.theme.colors.shadowMedium};

    &:hover {
        background-color: #A0522D;
        box-shadow: 0px 6px 12px ${props => props.theme.colors.shadowMedium};
    }

    &:active {
        background-color: #6B2E0C; 
        box-shadow: 0px 2px 4px ${props => props.theme.colors.shadowLight};
    }

    &:disabled {
        background-color: ${props => props.theme.colors.mutedText};
        cursor: not-allowed;
        box-shadow: none;
    }
`;

export { StyledLoginButton };