// material-ui
import Container from '@mui/material/Container';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';
import Box from '@mui/material/Box';

// project imports
import FadeInWhenVisible from './Animation';

// assets
import { IconCircleNumber1, IconCircleNumber2, IconCircleNumber3, IconCircleNumber4 } from '@tabler/icons-react';

const steps = [
  {
    icon: <IconCircleNumber1 size={40} />,
    title: 'Client hits xgateway',
    description: 'Any HTTP request to a protected API route is intercepted by xgateway first.'
  },
  {
    icon: <IconCircleNumber2 size={40} />,
    title: 'No proof → 402 response',
    description: 'xgateway resolves the payment rule and returns HTTP 402 with payment instructions if no valid proof is present.'
  },
  {
    icon: <IconCircleNumber3 size={40} />,
    title: 'Client pays & retries',
    description: 'Client pays via x402, MPP/Tempo, or Stripe and retries the request with the payment proof header.'
  },
  {
    icon: <IconCircleNumber4 size={40} />,
    title: 'Verified & proxied',
    description: 'xgateway verifies the proof, proxies the request to your upstream API, and logs the transaction.'
  }
];

// =============================|| LANDING - HOW IT WORKS ||============================= //

export default function HowItWorksSection() {
  return (
    <Container>
      <Grid container spacing={7.5} sx={{ justifyContent: 'center' }}>
        <Grid sx={{ textAlign: 'center' }} size={{ xs: 12, md: 8 }}>
          <Grid container spacing={1.5}>
            <Grid size={12}>
              <Typography variant="h2" sx={{ fontSize: { xs: '1.5rem', sm: '2.125rem' } }}>
                How xpaywall works
              </Typography>
            </Grid>
            <Grid size={12}>
              <Typography variant="body2" sx={{ fontSize: '1rem' }}>
                A simple 4-step flow from API request to verified payment — fully transparent to your upstream service.
              </Typography>
            </Grid>
          </Grid>
        </Grid>
        <Grid size={12}>
          <Grid container spacing={4} sx={{ justifyContent: 'center' }}>
            {steps.map((step, index) => (
              <Grid key={index} size={{ xs: 12, sm: 6, md: 3 }}>
                <FadeInWhenVisible>
                  <Stack sx={{ alignItems: 'center', textAlign: 'center', gap: 2 }}>
                    <Box sx={{ color: 'secondary.main' }}>{step.icon}</Box>
                    <Typography variant="h4" sx={{ fontWeight: 600 }}>
                      {step.title}
                    </Typography>
                    <Typography variant="body2" sx={{ color: 'text.secondary', fontSize: '0.95rem' }}>
                      {step.description}
                    </Typography>
                  </Stack>
                </FadeInWhenVisible>
              </Grid>
            ))}
          </Grid>
        </Grid>
      </Grid>
    </Container>
  );
}
