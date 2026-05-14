// ==============================|| OVERRIDES - PAPER ||============================== //

export default function Paper(borderRadius: number) {
  return {
    MuiPaper: {
      defaultProps: {
        elevation: 0
      },
      styleOverrides: {
        root: {
          backgroundImage: 'none'
        },
        rounded: {
          borderRadius: `${borderRadius}px`
        }
      }
    }
  };
}
