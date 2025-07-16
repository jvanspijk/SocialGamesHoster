import styled from 'styled-components';

const MainTitle = styled.h1`
    font-size: ${props => props.theme.typography.fontSizeLg};
    color: ${props => props.theme.colors.primary};
    text-shadow: 3px 3px 5px ${props => props.theme.colors.shadowMedium};
    margin: 0;
    font-family: ${props => props.theme.typography.fontFamilyHeadings};
    letter-spacing: 1px;
`;

const PageContainer = styled.div`
    background-color: ${props => props.theme.colors.background};
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: ${props => props.theme.spacing.sm};
    font-family: ${props => props.theme.typography.fontFamilyBase};
    font-size: ${props => props.theme.typography.fontSizeBase};
    color: ${props => props.theme.colors.textPrimary};
`;

const StyledHeader = styled.header`
    text-align: center;
    margin-bottom: 0;
    padding: ${props => props.theme.spacing.sm} ${props => props.theme.spacing.sm};
    border-radius: ${props => props.theme.borderRadius.lg};
    font-size: ${props => props.theme.typography.fontSizeMd};
`;

const ScrollBox = styled.article`
    background-color: ${props => props.theme.colors.surface};
    border: 1px solid ${props => props.theme.colors.secondary};
    box-shadow:
        0px 5px 15px ${props => props.theme.colors.shadowLight},
        inset 0 0 15px rgba(210, 180, 140, 0.2); 
    padding: 1em 1em;
    margin: ${props => props.theme.spacing.md} 0;
    width: 600px;
    max-width: 90%;
    box-sizing: border-box;
    text-align: center;
    position: relative;
    border-radius: ${props => props.theme.borderRadius.md};

    h2, h3 {
        color: ${props => props.theme.colors.textSecondary};
        margin-top: 0;
        margin-bottom: ${props => props.theme.spacing.sm}
        text-shadow: 1px 1px 2px ${props => props.theme.colors.shadowLight};
    }

    p {
        line-height: ${props => props.theme.typography.lineHeightBase};
        margin-bottom: ${props => props.theme.spacing.lg};
    }

    ul {
        list-style: none;
        padding: 0;
        text-align: left;
        margin-left: ${props => props.theme.spacing.xl};
    }

    li {
        margin-bottom: 0; 
        font-size: ${props => props.theme.typography.fontSizeBase};
    }
`;

export { MainTitle, PageContainer, StyledHeader, ScrollBox };