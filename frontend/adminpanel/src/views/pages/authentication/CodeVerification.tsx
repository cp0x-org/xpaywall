import { Link } from 'react-router-dom';

// material-ui
import { Theme } from '@mui/material/styles';
import useMediaQuery from '@mui/material/useMediaQuery';
import Button from '@mui/material/Button';
import Divider from '@mui/material/Divider';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';

// project imports
import AuthWrapper1 from './AuthWrapper1';
import AuthCardWrapper from './AuthCardWrapper';
import AuthCodeVerification from './jwt/AuthCodeVerification';

import Logo from 'ui-component/Logo';
import AnimateButton from 'ui-component/extended/AnimateButton';
import AuthFooter from 'ui-component/cards/AuthFooter';

// ===========================|| AUTH3 - CODE VERIFICATION ||=========================== //

export default function CodeVerification() {
  const downMD = useMediaQuery((theme: Theme) => theme.breakpoints.down('md'));

  return (
    <AuthWrapper1>
      <Stack sx={{ justifyContent: 'flex-end', minHeight: '100vh' }}>
        <Stack sx={{ justifyContent: 'center', alignItems: 'center', minHeight: 'calc(100vh - 68px)' }}>
          <Box sx={{ m: { xs: 1, sm: 3 }, mb: 0 }}>
            <AuthCardWrapper>
              <Stack sx={{ gap: 2, alignItems: 'center', justifyContent: 'center' }}>
                <Box sx={{ mb: 3 }}>
                  <Link to="#" aria-label="theme logo">
                    <Logo />
                  </Link>
                </Box>
                <Stack direction={downMD ? 'column-reverse' : 'row'} sx={{ alignItems: 'center', justifyContent: 'center' }}>
                  <Stack sx={{ alignItems: 'center', justifyContent: 'center' }}>
                    <Typography gutterBottom variant={downMD ? 'h3' : 'h2'} sx={{ color: 'secondary.main', mb: 1 }}>
                      Enter Verification Code
                    </Typography>
                    <Typography variant="subtitle1" sx={{ fontSize: '1rem' }}>
                      We sent you an email.
                    </Typography>
                    <Typography variant="caption" sx={{ fontSize: '0.875rem', textAlign: downMD ? 'center' : 'inherit', mt: 1 }}>
                      We’ve sent you a code on jone.****@company.com
                    </Typography>
                  </Stack>
                </Stack>
                <AuthCodeVerification />
                <Divider sx={{ width: 1 }} />
                <Typography
                  component={Link}
                  to="#"
                  variant="subtitle1"
                  sx={{ textAlign: downMD ? 'center' : 'inherit', textDecoration: 'none' }}
                >
                  Did not receive the email? Check your spam filter, or
                </Typography>
                <Stack sx={{ width: 1 }}>
                  <AnimateButton>
                    <Button disableElevation fullWidth size="large" type="submit" variant="outlined" color="secondary">
                      Resend Code
                    </Button>
                  </AnimateButton>
                </Stack>
              </Stack>
            </AuthCardWrapper>
          </Box>
        </Stack>
        <Box sx={{ px: 3, my: 3 }}>
          <AuthFooter />
        </Box>
      </Stack>
    </AuthWrapper1>
  );
}
