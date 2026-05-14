// material-ui
import { Theme } from '@mui/material/styles';

// ==============================|| OVERRIDES - INPUT BASE ||============================== //

export default function InputBase(theme: Theme) {
  return {
    MuiInputBase: {
      styleOverrides: {
        input: {
          color: theme.vars.palette.text.dark,
          '&::placeholder': {
            color: theme.vars.palette.text.secondary,
            fontSize: '0.875rem'
          }
        }
      }
    }
  };
}
