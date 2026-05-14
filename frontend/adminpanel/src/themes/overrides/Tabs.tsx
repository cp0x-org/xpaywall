// material-ui
import { Theme } from '@mui/material/styles';

// project imports
import { withAlpha } from 'utils/colorUtils';

// ==============================|| OVERRIDES - TABS ||============================== //

export default function Tabs(theme: Theme) {
  return {
    MuiTabs: {
      styleOverrides: {
        flexContainer: {
          borderBottom: '1px solid',
          borderColor: theme.vars.palette.grey[200],

          ...theme.applyStyles('dark', {
            borderColor: withAlpha(theme.vars.palette.text.primary, 0.2)
          })
        }
      }
    }
  };
}
