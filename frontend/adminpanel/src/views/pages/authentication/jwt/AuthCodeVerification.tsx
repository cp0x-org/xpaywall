import { useState } from 'react';

// material-ui
import { useColorScheme, useTheme } from '@mui/material/styles';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';

// third party
import OTPInput from 'react-otp-input';

// project imports
import { ThemeMode } from 'config';
import useConfig from 'hooks/useConfig';
import { withAlpha } from 'utils/colorUtils';

// ============================|| JWT - CODE VERIFICATION ||============================ //

export default function AuthCodeVerification() {
  const theme = useTheme();
  const { colorScheme } = useColorScheme();
  const {
    state: { borderRadius, outlinedFilled }
  } = useConfig();

  const [otp, setOtp] = useState<string>();
  const borderColor = colorScheme === ThemeMode.DARK ? withAlpha(theme.vars.palette.text.primary, 0.28) : theme.vars.palette.grey[400];

  const bgcolor = outlinedFilled
    ? colorScheme === ThemeMode.DARK
      ? theme.vars.palette.dark[800]
      : theme.vars.palette.grey[50]
    : 'transparent';

  return (
    <Stack sx={{ gap: 3 }}>
      <Box
        sx={{
          '& input:focus-visible': {
            borderColor: `${theme.vars.palette.primary.main} !important`,
            boxShadow: `0px 0px 0px 1px ${theme.vars.palette.primary.main}`
          }
        }}
      >
        <OTPInput
          value={otp}
          onChange={(otpNumber: string) => setOtp(otpNumber)}
          numInputs={4}
          inputType="tel"
          shouldAutoFocus
          containerStyle={{ gap: 6 }}
          inputStyle={{
            width: '100%',
            margin: '8px',
            padding: '10px',
            border: `1px solid ${borderColor}`,
            borderRadius: `${borderRadius}px`,
            outline: 'none',
            background: bgcolor
          }}
          renderInput={(props) => <input {...props} />}
        />
      </Box>
      <Button disableElevation fullWidth size="large" type="submit" variant="contained">
        Continue
      </Button>
      <Stack direction="row" sx={{ justifyContent: 'space-between', alignItems: 'baseline' }}>
        <Typography>Did not receive the email? Check your spam filter, or</Typography>
        <Typography variant="body1" sx={{ minWidth: 85, ml: 2, textDecoration: 'none', cursor: 'pointer', color: 'primary.main' }}>
          Resend code
        </Typography>
      </Stack>
    </Stack>
  );
}
