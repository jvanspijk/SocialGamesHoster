import styled from 'styled-components';

export const AbilitiesSectionTitle = styled.h3`
    font-family: ${props => props.theme.typography.fontFamilyHeadings};
    font-size: ${props => props.theme.typography.fontSizeLg};
    color: ${props => props.theme.colors.primary};
    text-align: center;
    margin: ${props => props.theme.spacing.sm} 0;
    padding: ${props => props.theme.spacing.sm};
    background-color: ${props => props.theme.colors.secondary};
    border-radius: ${props => props.theme.borderRadius.lg};
    box-shadow: 0px 4px 8px ${props => props.theme.colors.shadowMedium};
    display: block;
    max-width: 600px;
    width: 90%;
    box-sizing: border-box;
    align-self: center;
    border: 1px solid ${props => props.theme.colors.mutedText};
`;

export const NumberCircle = styled.div`
    position: absolute;
    left: -20px;
    top: 50%;
    transform: translateY(-50%);
    background-color: ${props => props.theme.colors.primary};
    color: ${props => props.theme.colors.white};
    border-radius: ${props => props.theme.borderRadius.circle};
    width: 40px;
    height: 40px;
    display: flex;
    justify-content: center;
    align-items: center;
    font-size: ${props => props.theme.typography.fontSizeMd};
    font-weight: bold;
    border: 3px solid ${props => props.theme.colors.secondary};
    box-shadow: 0px 2px 5px ${props => props.theme.colors.shadowMedium};
`;