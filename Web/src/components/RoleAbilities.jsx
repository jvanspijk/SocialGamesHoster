import React from 'react';
import { AbilitiesSectionTitle, NumberCircle } from './component-styles/RoleAbilities.jsx';
import { ScrollBox } from './component-styles/Common.jsx';

export function RoleAbilities({ abilities }) {
    if (!abilities || abilities.length === 0) {
        console.warn("RoleAbilities component used but abilities array was null or empty.")
        return null;
    }

    return (
        <>
            <AbilitiesSectionTitle>Abilities:</AbilitiesSectionTitle>
            {abilities.map((ability, index) => (
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
    );
}