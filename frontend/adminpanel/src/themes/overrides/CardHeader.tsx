// material-ui
import { Theme } from '@mui/material/styles';

// ==============================|| OVERRIDES - CARD HEADER ||============================== //

export default function CardHeader(theme: Theme) {
  return {
    MuiCardHeader: {
      styleOverrides: {
        root: {
          color: theme.vars.palette.text.dark,
          padding: '24px'
        },
        title: {
          fontSize: '1.125rem'
        }
      }
    }
  };
}
