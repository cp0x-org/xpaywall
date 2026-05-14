import { Link as RouterLink } from 'react-router-dom';

// material-ui
import Button from '@mui/material/Button';
import Container from '@mui/material/Container';
import Grid from '@mui/material/Grid';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';

// third party
import { motion } from 'framer-motion';

// project imports
import AnimateButton from 'ui-component/extended/AnimateButton';
import { DASHBOARD_PATH } from 'config';

// assets
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import ShieldIcon from '@mui/icons-material/Shield';
import DashboardImg from 'assets/images/landing/dashboard.png';

// ==============================|| LANDING - HEADER PAGE ||============================== //

export default function HeaderSection() {
  const headerSX = { fontSize: { xs: '2rem', sm: '3rem', md: '3.5rem', lg: '3.5rem' } };

  return (
    <Container sx={{ minHeight: '100vh', display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
      <Grid
        container
        sx={{ justifyContent: 'space-between', alignItems: 'center', mt: { xs: 10, sm: 6, md: 18.75 }, mb: { xs: 2.5, md: 10 } }}
      >
        <Grid size={{ xs: 12, md: 6 }}>
          <Grid container spacing={6}>
            <Grid size={12}>
              <motion.div
                initial={{ opacity: 0, translateY: 550 }}
                animate={{ opacity: 1, translateY: 0 }}
                transition={{ type: 'spring', stiffness: 150, damping: 30 }}
              >
                <Stack sx={{ gap: 2 }}>
                  <Chip
                    label="x402 · MPP · Stripe"
                    color="secondary"
                    size="small"
                    icon={<ShieldIcon />}
                    sx={{ width: 'fit-content', fontWeight: 600 }}
                  />
                  <Stack sx={{ gap: 1 }}>
                    <Typography variant="h1" sx={{ textAlign: { xs: 'center', md: 'left' }, ...headerSX }}>
                      API Payment
                    </Typography>
                    <Typography variant="h1" sx={{ color: 'primary.main', textAlign: { xs: 'center', md: 'left' }, ...headerSX }}>
                      Gateway
                    </Typography>
                  </Stack>
                </Stack>
              </motion.div>
            </Grid>
            <Grid sx={{ mt: -2.5, textAlign: { xs: 'center', md: 'left' } }} size={12}>
              <motion.div
                initial={{ opacity: 0, translateY: 550 }}
                animate={{ opacity: 1, translateY: 0 }}
                transition={{ type: 'spring', stiffness: 150, damping: 30, delay: 0.2 }}
              >
                <Typography
                  variant="body1"
                  sx={{ textAlign: { xs: 'center', md: 'left' }, color: 'text.primary', fontSize: { xs: '1rem', md: '1.125rem' } }}
                >
                  xpaywall sits in front of your APIs and enforces micropayments before proxying requests. Define payment rules per route,
                  verify proofs automatically, and manage everything from the admin dashboard.
                </Typography>
              </motion.div>
            </Grid>
            <Grid size={12}>
              <motion.div
                initial={{ opacity: 0, translateY: 550 }}
                animate={{ opacity: 1, translateY: 0 }}
                transition={{ type: 'spring', stiffness: 150, damping: 30, delay: 0.4 }}
              >
                <Grid container spacing={2} sx={{ justifyContent: { xs: 'center', md: 'flex-start' } }}>
                  <Grid>
                    <AnimateButton>
                      <Button
                        component={RouterLink}
                        to={DASHBOARD_PATH}
                        size="large"
                        variant="contained"
                        color="secondary"
                        startIcon={<PlayArrowIcon />}
                      >
                        Open Dashboard
                      </Button>
                    </AnimateButton>
                  </Grid>
                  <Grid>
                    <Button component={RouterLink} to="/login" size="large" variant="outlined">
                      Sign In
                    </Button>
                  </Grid>
                </Grid>
              </motion.div>
            </Grid>
          </Grid>
        </Grid>

        {/* Dashboard screenshot */}
        <Grid sx={{ display: { xs: 'none', md: 'flex' } }} size={{ xs: 12, md: 6 }}>
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ type: 'spring', stiffness: 150, damping: 30, delay: 0.3 }}
            style={{ width: '100%' }}
          >
            <Box
              sx={{
                ml: 4,
                mt: 8.75,
                borderRadius: 3,
                overflow: 'hidden',
                boxShadow: '0 8px 40px rgba(0,0,0,0.18)'
              }}
            >
              <Box component="img" src={DashboardImg} alt="xpaywall admin dashboard" sx={{ width: '100%', display: 'block' }} />
            </Box>
          </motion.div>
        </Grid>
      </Grid>
    </Container>
  );
}
