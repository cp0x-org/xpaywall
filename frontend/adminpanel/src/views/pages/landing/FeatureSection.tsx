// material-ui
import { useTheme } from '@mui/material/styles';
import Container from '@mui/material/Container';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';

// project imports
import FadeInWhenVisible from './Animation';
import SubCard from 'ui-component/cards/SubCard';
import Avatar from 'ui-component/extended/Avatar';

// assets
import {
  IconShieldCheck,
  IconRoute,
  IconDatabase,
  IconBrandGithub,
  IconCoin,
  IconLayoutDashboard
} from '@tabler/icons-react';

const features = [
  {
    title: 'x402 Protocol',
    caption: 'Native HTTP 402 payment enforcement. Clients pay per request using on-chain micropayments with zero API keys.',
    icon: <IconCoin size={28} />
  },
  {
    title: 'MPP / Tempo',
    caption: 'Multi-Payment Protocol support for cross-chain payments via Tempo. One gateway, multiple networks.',
    icon: <IconShieldCheck size={28} />
  },
  {
    title: 'Route-based Rules',
    caption: 'Configure payment requirements per route: price, payment method, merchant address, and upstream target.',
    icon: <IconRoute size={28} />
  },
  {
    title: 'Multi-project',
    caption: 'Manage multiple API backends from one dashboard. Isolate routes, stats, and logs per project.',
    icon: <IconLayoutDashboard size={28} />
  },
  {
    title: 'Request Logging',
    caption: 'Every request is logged with payment status, route, and timestamp. Export or query via the control API.',
    icon: <IconDatabase size={28} />
  },
  {
    title: 'Open Source',
    caption: 'Self-host xpaywall on your own infrastructure. Full control over keys, data, and configuration.',
    icon: <IconBrandGithub size={28} />
  }
];

// =============================|| LANDING - FEATURE SECTION ||============================= //

export default function FeatureSection() {
  const theme = useTheme();
  const avatarSx = { bgcolor: 'transparent', color: 'secondary.main', width: 56, height: 56 };

  return (
    <Container>
      <Grid container spacing={7.5} sx={{ justifyContent: 'center' }}>
        <Grid sx={{ textAlign: 'center' }} size={{ xs: 12, md: 6 }}>
          <Grid container spacing={1.5}>
            <Grid size={12}>
              <Typography variant="h2" sx={{ fontSize: { xs: '1.5rem', sm: '2.125rem' } }}>
                Everything you need to monetize your API
              </Typography>
            </Grid>
            <Grid size={12}>
              <Typography variant="body2" sx={{ fontSize: '1rem' }}>
                xpaywall handles the full payment lifecycle so you can focus on building your API.
              </Typography>
            </Grid>
          </Grid>
        </Grid>
        <Grid size={12}>
          <Grid container spacing={5} sx={{ justifyContent: 'center', '&> .MuiGrid-root > div': { height: '100%' } }}>
            {features.map((feature, index) => (
              <Grid key={index} size={{ md: 4, sm: 6 }}>
                <FadeInWhenVisible>
                  <SubCard
                    sx={{
                      bgcolor: 'grey.100',
                      ...theme.applyStyles('dark', { bgcolor: 'dark.800' }),
                      borderColor: 'divider',
                      '&:hover': { boxShadow: 'none' },
                      height: 1
                    }}
                  >
                    <Stack sx={{ gap: 3 }}>
                      <Avatar variant="rounded" sx={avatarSx}>
                        {feature.icon}
                      </Avatar>
                      <Stack sx={{ gap: 1.5 }}>
                        <Typography variant="h3" sx={{ fontWeight: 500 }}>
                          {feature.title}
                        </Typography>
                        <Typography variant="body2" sx={{ fontSize: '1rem' }}>
                          {feature.caption}
                        </Typography>
                      </Stack>
                    </Stack>
                  </SubCard>
                </FadeInWhenVisible>
              </Grid>
            ))}
          </Grid>
        </Grid>
      </Grid>
    </Container>
  );
}
