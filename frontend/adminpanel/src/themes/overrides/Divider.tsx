// material-ui
import { Theme } from '@mui/material/styles';

// ==============================|| OVERRIDES - DIVIDER ||============================== //

export default function Divider(theme: Theme) {
  return {
    MuiDivider: {
      styleOverrides: {
        root: {
          borderColor: theme.vars.palette.divider
        }
      }
    }
  };
}
