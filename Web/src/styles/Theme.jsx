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
    xs: '0rem',
    sm: '0.1rem',
    md: '0.5rem',
    lg: '1rem',
    xl: '1.5rem',
    xxl: '2rem',
};

const typography = {
    fontFamilyBase: `'Old Standard TT', serif`,
    fontFamilyHeadings: `'IM Fell DW Pica', serif`,
    fontSizeSm: '0.75em',
    fontSizeBase: '1em',
    fontSizeMd: '1.5em',
    fontSizeLg: '1.75em',
    fontSizeXl: '2.5em',
    lineHeightBase: 1.5,
};

const borderRadius = {
    sm: '2px',
    md: '4px',
    lg: '5px',
    circle: '50%',
};

const Theme = {
    colors,
    spacing,
    typography,
    borderRadius,
};

export default Theme;