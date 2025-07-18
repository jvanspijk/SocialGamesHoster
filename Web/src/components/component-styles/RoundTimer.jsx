import styled from 'styled-components';
import Theme from '../../styles/Theme'; // Import the theme

export const TimerContainer = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 0;
  gap: 0;
  font-size: ${Theme.typography.fontSizeSm};
`;

export const TimerLabel = styled.p`
  font-family: ${Theme.typography.fontFamilyHeadings};
  font-size: ${Theme.typography.fontSizeLg};
  color: ${Theme.colors.textPrimary};
  margin-bottom: ${Theme.spacing.sm};
  text-align: center;
`;

export const TimerCircle = styled.div`
  width: 100px;
  height: 100px;
  border-radius: ${Theme.borderRadius.circle};
  background-color: ${(props) => props.$background || Theme.colors.background};
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: ${Theme.colors.shadowMedium};
  border: 4px solid ${Theme.colors.primary};
`;

export const TimeDisplay = styled.span`
  font-family: ${Theme.typography.fontFamilyBase};
  font-size: ${Theme.typography.fontSizeMd};
  color: ${Theme.colors.textPrimary};
  font-weight: bold;
`;