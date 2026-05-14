// material-ui
import { Theme } from '@mui/material/styles';
import { AlertProps } from '@mui/material/Alert';

// project imports
import { withAlpha } from 'utils/colorUtils';

// assets
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';

// ==============================|| OVERRIDES - ALERT ||============================== //

export default function Alert(theme: Theme) {
  const { vars } = theme;

  const getPaletteColor = (severity: AlertProps['severity']) =>
    severity ? vars.palette[severity as keyof typeof vars.palette] : vars.palette.info;

  const getCommonStyles = (ownerState: AlertProps) => {
    const isWarningOrSuccess = ownerState.severity === 'warning' || ownerState.severity === 'success';
    return { isWarningOrSuccess };
  };

  const standardVariant = ({ ownerState }: { ownerState: AlertProps }) => {
    const paletteColor = getPaletteColor(ownerState.severity);
    const { isWarningOrSuccess } = getCommonStyles(ownerState);

    return {
      color: isWarningOrSuccess ? paletteColor.dark : paletteColor.main,
      backgroundColor: withAlpha(paletteColor.main, 0.075),
      '& .MuiAlert-icon': { color: isWarningOrSuccess ? paletteColor.dark : paletteColor.main }
    };
  };

  const outlinedVariant = ({ ownerState }: { ownerState: AlertProps }) => {
    const paletteColor = getPaletteColor(ownerState.severity);
    const { isWarningOrSuccess } = getCommonStyles(ownerState);

    return {
      color: isWarningOrSuccess ? paletteColor.dark : paletteColor.main,
      borderColor: paletteColor.dark,
      '& .MuiAlert-icon': { color: isWarningOrSuccess ? paletteColor.dark : paletteColor.main }
    };
  };

  const filledVariant = ({ ownerState }: { ownerState: AlertProps }) => {
    const paletteColor = getPaletteColor(ownerState.severity);
    const { isWarningOrSuccess } = getCommonStyles(ownerState);

    return {
      color: isWarningOrSuccess ? vars.palette.common.black : vars.palette.common.white,
      backgroundColor: isWarningOrSuccess ? paletteColor.dark : paletteColor.main,
      '& .MuiAlert-icon': {
        color: isWarningOrSuccess ? vars.palette.common.black : vars.palette.common.white
      }
    };
  };

  return {
    MuiAlert: {
      defaultProps: {
        iconMapping: {
          primary: <InfoOutlinedIcon sx={{ fontSize: 'inherit' }} />
        }
      },
      styleOverrides: {
        root: {
          alignItems: 'center',
          variants: [
            { props: { variant: 'standard' }, style: standardVariant },
            { props: { variant: 'outlined' }, style: outlinedVariant },
            { props: { variant: 'filled' }, style: filledVariant }
          ]
        },
        outlined: { border: '1px dashed' }
      }
    }
  };
}
