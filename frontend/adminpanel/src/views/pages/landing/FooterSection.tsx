import { Link as RouterLink } from 'react-router-dom';

// material-ui
import { useTheme, styled } from '@mui/material/styles';
import Container from '@mui/material/Container';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';
import Link from '@mui/material/Link';
import Box from '@mui/material/Box';

// project imports
import Logo from 'ui-component/Logo';

// assets
import GitHubIcon from '@mui/icons-material/GitHub';

const FooterLink = styled(Link)(({ theme }) => ({
  color: theme.vars.palette.text.hint,
  ...theme.applyStyles('dark', { color: theme.vars.palette.text.secondary }),
  '&:hover, &:active': { color: theme.vars.palette.secondary[200] }
}));

// =============================|| LANDING - FOOTER SECTION ||============================= //

export default function FooterSection() {
  const theme = useTheme();
  const textColor = { color: 'text.hint', ...theme.applyStyles('dark', { color: 'text.secondary' }) };

  return (
    <>
      <Container sx={{ mb: 10 }}>
        <Grid container spacing={6}>
          <Grid size={{ xs: 12, md: 4 }}>
            <Stack sx={{ gap: { xs: 2, md: 5 } }}>
              <Typography component={RouterLink} to="/" aria-label="theme-logo">
                <Logo dark />
              </Typography>
              <Typography variant="body2" sx={{ ...textColor }}>
                xpaywall is an open-source x402/MPP payment gateway that enforces micropayments in front of your APIs. Self-host and
                monetize any HTTP endpoint.
              </Typography>
            </Stack>
          </Grid>
          <Grid size={{ xs: 12, md: 8 }}>
            <Grid container spacing={{ xs: 5, md: 2 }}>
              <Grid size={{ xs: 6, sm: 4 }}>
                <Stack sx={{ gap: { xs: 3, md: 5 } }}>
                  <Typography variant="h4" sx={{ fontWeight: 500, ...textColor }}>
                    Product
                  </Typography>
                  <Stack sx={{ gap: { xs: 1.5, md: 2.5 } }}>
                    <RouterLink to="/dashboard" style={{ textDecoration: 'none' }}>
                      <FooterLink underline="none">Dashboard</FooterLink>
                    </RouterLink>
                    <RouterLink to="/login" style={{ textDecoration: 'none' }}>
                      <FooterLink underline="none">Sign In</FooterLink>
                    </RouterLink>
                    <RouterLink to="/projects" style={{ textDecoration: 'none' }}>
                      <FooterLink underline="none">Projects</FooterLink>
                    </RouterLink>
                    <RouterLink to="/routes" style={{ textDecoration: 'none' }}>
                      <FooterLink underline="none">Routes</FooterLink>
                    </RouterLink>
                  </Stack>
                </Stack>
              </Grid>
              <Grid size={{ xs: 6, sm: 4 }}>
                <Stack sx={{ gap: { xs: 3, md: 5 } }}>
                  <Typography variant="h4" sx={{ fontWeight: 500, ...textColor }}>
                    Services
                  </Typography>
                  <Stack sx={{ gap: { xs: 1.5, md: 2.5 } }}>
                    <Typography variant="body2" sx={{ ...textColor }}>
                      xgateway — port 8081
                    </Typography>
                    <Typography variant="body2" sx={{ ...textColor }}>
                      control-api — port 9091
                    </Typography>
                    <Typography variant="body2" sx={{ ...textColor }}>
                      adminpanel — port 3000
                    </Typography>
                  </Stack>
                </Stack>
              </Grid>
              <Grid size={{ xs: 6, sm: 4 }}>
                <Stack sx={{ gap: { xs: 3, md: 5 } }}>
                  <Typography variant="h4" sx={{ fontWeight: 500, ...textColor }}>
                    Protocols
                  </Typography>
                  <Stack sx={{ gap: { xs: 1.5, md: 2.5 } }}>
                    <Typography variant="body2" sx={{ ...textColor }}>
                      x402
                    </Typography>
                    <Typography variant="body2" sx={{ ...textColor }}>
                      MPP / Tempo
                    </Typography>
                    <Typography variant="body2" sx={{ ...textColor }}>
                      Stripe
                    </Typography>
                  </Stack>
                </Stack>
              </Grid>
            </Grid>
          </Grid>
        </Grid>
      </Container>
      <Box sx={{ bgcolor: 'dark.dark', py: { xs: 3, sm: 1.5 } }}>
        <Container>
          <Stack
            direction={{ xs: 'column', sm: 'row' }}
            sx={{ gap: { xs: 1.5, sm: 1, md: 3 }, alignItems: 'center', justifyContent: 'space-between' }}
          >
            <Typography sx={{ color: 'text.secondary' }}>© {new Date().getFullYear()} xpaywall</Typography>
            <Stack direction="row" sx={{ gap: { xs: 2, sm: 1.5, md: 2 }, alignItems: 'center' }}>
              <IconButton
                size="small"
                aria-label="GitHub"
                component={Link}
                href="https://github.com"
                target="_blank"
              >
                <GitHubIcon sx={{ color: 'text.secondary', '&:hover': { color: 'success.main' } }} />
              </IconButton>
            </Stack>
          </Stack>
        </Container>
      </Box>
    </>
  );
}
