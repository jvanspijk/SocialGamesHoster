// src/theme/Theme.jsx

const colors = {
    primary: '#8B4513', // SaddleBrown
    secondary: '#D2B48C', // Tan
    background: '#F5DEB3', // Wheat
    surface: '#FDF5E6', // OldLace
    textPrimary: '#4A3A2B', // Darker Brown
    textSecondary: '#6B4226', // Lighter Brown
    mutedText: '#A08060', // Muted Brown
    white: '#FFFFFF',
    black: '#000000',
    shadowLight: 'rgba(0, 0, 0, 0.1)',
    shadowMedium: 'rgba(0, 0, 0, 0.3)',
};

const spacing = {
    xs: '0.25em',
    sm: '0.5em',
    md: '1em',
    lg: '1.5em',
    xl: '2em',
    xxl: '4em',
};

const typography = {
    fontFamilyBase: `'Old Standard TT', serif`,
    fontFamilyHeadings: `'IM Fell DW Pica', serif`,
    fontSizeSm: '1em',
    fontSizeBase: '1.2em',
    fontSizeMd: '2em',
    fontSizeLg: '2em',
    fontSizeXl: '3.5em',
    lineHeightBase: 1.6,
};

const borderRadius = {
    sm: '4px',
    md: '8px',
    lg: '10px',
    circle: '50%',
};

const Theme = {
    colors,
    spacing,
    typography,
    borderRadius,
};

export default Theme;