import styled from 'styled-components';

const StyledDropdownContainer = styled.div`
    font-family: ${props => props.theme.typography.fontFamilyBase};
    padding: ${props => props.theme.spacing.lg};
    background-color: ${props => props.theme.colors.surface};
    border-radius: ${props => props.theme.borderRadius.lg};
    box-shadow: 0 ${props => props.theme.spacing.sm} 12px ${props => props.theme.colors.shadowLight};
    max-width: 300px;
    margin: ${props => props.theme.spacing.xl} auto;
`;

const StyledDropdownLabel = styled.label`
    font-size: ${props => props.theme.typography.fontSizeBase};
    color: ${props => props.theme.colors.textSecondary}; /* Darker brown for text */
    margin-bottom: ${props => props.theme.spacing.sm};
    font-weight: 600;
    display: block;
`;

const StyledDropdownSelect = styled.select`
    width: 100%;
    padding: 0.75rem ${props => props.theme.spacing.md};
    border: 1px solid ${props => props.theme.colors.mutedText}
    border-radius: ${props => props.theme.borderRadius.md};
    background-color: ${props => props.theme.colors.white}; 
    color: ${props => props.theme.colors.textPrimary};
    font-size: ${props => props.theme.typography.fontSizeBase};
    appearance: none; 
    -webkit-appearance: none;
    -moz-appearance: none;
    background-image: url('data:image/svg+xml;utf8,<svg fill="%234a3b2e" height="24" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg"><path d="M7 10l5 5 5-5z"/><path d="M0 0h24v24H0z" fill="none"/></svg>');
    background-repeat: no-repeat;
    background-position: right ${props => props.theme.spacing.md} center;
    background-size: 1.5em;
    cursor: pointer;
    transition: all 0.3s ease;

    &:focus {
        outline: none;
        border-color: ${props => props.theme.colors.textSecondary};
        box-shadow: 0 0 0 3px ${props => props.theme.colors.shadowLight};
    }

    option {
        background-color: ${props => props.theme.colors.white};
        color: ${props => props.theme.colors.textPrimary};
    }

    option:disabled {
        color: ${props => props.theme.colors.mutedText};
        font-style: italic;
    }
`;

export { StyledDropdownContainer, StyledDropdownLabel, StyledDropdownSelect };