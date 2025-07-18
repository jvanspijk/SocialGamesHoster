import React from 'react';
import { useParams, useNavigate } from 'react-router';
import { PageContainer, StyledHeader, MainTitle } from '../components/component-styles/Common.jsx';
import { RoundTimer } from '../components/RoundTimer';
import { RoleDescription } from '../components/RoleDescription'; 
import { RoleAbilities } from '../components/RoleAbilities';
import { ThemeProvider } from 'styled-components'
import Theme from '../styles/Theme.jsx'

function MyDetails() {
    const { name } = useParams();
    const navigate = useNavigate();

    return (
        <ThemeProvider theme={Theme}>
            <PageContainer>
                <RoundTimer />

                <StyledHeader>
                    <MainTitle>{name}</MainTitle>
                </StyledHeader>

                <RoleDescription username={name}>
                    {(role, loadingRole) => (
                        !loadingRole && role && role.abilities && role.abilities.length > 0 && (
                            <RoleAbilities abilities={role.abilities} />
                        )
                    )}
                </RoleDescription>
            </PageContainer>
        </ThemeProvider>
    );
}

export default MyDetails;