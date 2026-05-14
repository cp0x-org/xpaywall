// material-ui
import { Theme } from '@mui/material';

// ==============================|| OVERRIDES - TOOLTIP ||============================== //

export default function Tooltip(theme: Theme) {
  return {
    MuiTooltip: {
      styleOverrides: {
        tooltip: {
          margin: 0,
          lineHeight: 1.4,
          color: theme.vars.palette.background.paper,
          background: theme.vars.palette.text.primary
        }
      }
    }
  };
}
